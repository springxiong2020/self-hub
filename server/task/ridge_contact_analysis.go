package task

import (
	"errors"
	"time"

	"github.com/springbear2020/self-hub/server/model/ridge"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ridgeAnalysisLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

const (
	contactNormal   = 0
	contactChatOnly = 1

	contactActionNone       uint = 0
	contactActionChatOnly   uint = 1
	contactActionDelete     uint = 2
	contactActionSupplement uint = 3
)

// AnalyzeRidgeContacts 分析指定用户的联系人，更新建议，并在重新联系后恢复后台好友状态。
//
// 触发方式：用户通过身份校验登录或恢复登录状态时同步调用，每个用户按北京时间每天
// 最多成功分析一次，未登录的用户不分析。完成日期持久化到 ridge_contact_analyses，
// 服务重启不会丢失；同一用户的并发请求通过数据库记录锁串行化。分析失败则整体回滚，
// 不记为当天已完成，下次登录/恢复登录状态时可以重试；调用方记录错误但不阻断登录。
//
// 分析范围：只处理 user_id = uid、is_exempt = 0 且 status 为 0（正常）或 1（仅聊天）的联系人。
// 豁免、已删除、状态未知的记录完全跳过，保留其现有状态及操作建议。
//
// 时间口径：取 last_chat_at 和 last_interaction_at 中较晚的有效时间作为最近联系时间；
// 空值及 Go 零时间均视为无记录，不等同于从未联系。天数按连续 24 小时计算，
// 不使用后台记录的 created_at、updated_at 作为联系时间或观察起点。
//
// 正常好友（status = 0）：
//   - 有最近联系时间时以其起算；两个联系时间均无记录时，以 friended_at 起算。
//   - 起算时间仍缺失时，suggested_action = 3（补充资料）。
//   - 距起算时间严格超过 90 天，suggested_action = 1（设为仅聊天）；
//     恰好 90 天及以内，suggested_action = 0（无建议），同时清除失效的旧建议。
//   - 这里只生成建议，不自动将正常好友设为仅聊天，也不自动开始观察期。
//
// 仅聊天好友（status = 1），严格按以下顺序判断：
//   - chat_only_since 缺失时，suggested_action = 3（补充资料），不推断观察起点。
//   - 最近联系时间严格晚于 chat_only_since，说明本轮观察开始后有新聊天或新互动：
//     自动将后台 status 恢复为 0、chat_only_since 清空为 NULL、suggested_action 清为 0。
//     恢复优先于删除建议，即使任务执行时已满 180 天也先恢复；本次不再按正常好友重复分析。
//   - 没有新聊天或互动，且自 chat_only_since 起已满 180 天，suggested_action = 2（删除好友）。
//   - 尚未满 180 天，suggested_action = 0，清除失效的旧建议。
//   - 180 天从设置仅聊天的时间起算，不从最后联系时间或建议生成时间起算。
//
// 任务不实际操作微信、不物理删除联系人；自动恢复仅修改后台状态。
// 对确需变化的记录更新 updated_at；零值状态/建议及 NULL 时间通过 map 明确写入。
// 整次分析在事务和行锁内完成，避免依据旧数据覆盖并发编辑；任一读写失败则整体回滚。
// now 由调用方传入，生产使用当前时间，测试使用固定时间以复核临界点。
func AnalyzeRidgeContacts(db *gorm.DB, uid uint, now time.Time) error {
	if db == nil {
		return errors.New("联系人分析失败：数据库未初始化")
	}

	if uid == 0 {
		return errors.New("联系人分析失败：用户ID不能为空")
	}
	today := now.In(ridgeAnalysisLocation).Format("2006-01-02")
	return db.Transaction(func(tx *gorm.DB) error {
		// 唯一用户主键配合事务内的插入/行锁，保证多个实例也不能重复成功分析。
		run := ridge.RidgeContactAnalysis{UserId: uid}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&run).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", uid).First(&run).Error; err != nil {
			return err
		}
		if run.AnalysisDate == today {
			return nil
		}

		var contacts []ridge.RidgeContact
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND is_exempt = ? AND status IN ?", uid, 0, []int{contactNormal, contactChatOnly}).
			Order("id").Find(&contacts).Error; err != nil {
			return err
		}

		for _, contact := range contacts {
			action, restore := analyzeRidgeContact(contact, now)
			if action == contact.SuggestedAction && !restore {
				continue
			}
			updates := map[string]interface{}{
				"suggested_action": action,
				"updated_at":       now,
			}
			if restore {
				updates["status"] = contactNormal
				updates["chat_only_since"] = nil
			}
			if err := tx.Model(&ridge.RidgeContact{}).Where("user_id = ? AND id = ?", uid, contact.Id).Updates(updates).Error; err != nil {
				return err
			}
		}
		return tx.Model(&ridge.RidgeContactAnalysis{}).Where("user_id = ?", uid).
			Updates(map[string]interface{}{"analysis_date": today, "completed_at": now}).Error
	})
}

func analyzeRidgeContact(contact ridge.RidgeContact, now time.Time) (action uint, restore bool) {
	latest := contact.LastChatAt
	if !validContactTime(latest) || (validContactTime(contact.LastInteractionAt) && contact.LastInteractionAt.After(*latest)) {
		latest = contact.LastInteractionAt
	}

	if contact.Status != nil && *contact.Status == contactChatOnly {
		if !validContactTime(contact.ChatOnlySince) {
			return contactActionSupplement, false
		}
		if validContactTime(latest) && latest.After(*contact.ChatOnlySince) {
			return contactActionNone, true
		}
		if now.Sub(*contact.ChatOnlySince) >= 180*24*time.Hour {
			return contactActionDelete, false
		}
		return contactActionNone, false
	}

	if !validContactTime(latest) {
		latest = contact.FriendedAt
	}
	if !validContactTime(latest) {
		return contactActionSupplement, false
	}
	if now.Sub(*latest) > 90*24*time.Hour {
		return contactActionChatOnly, false
	}
	return contactActionNone, false
}

func validContactTime(value *time.Time) bool {
	return value != nil && !value.IsZero()
}
