package ridge

import (
	"github.com/springbear2020/self-hub/server/global"
	"github.com/springbear2020/self-hub/server/model/ridge"
	"github.com/springbear2020/self-hub/server/model/ridge/request"
	"gorm.io/gorm"
)

type RidgeContactService struct{}

func (poService *RidgeContactService) CreateRidgeContact(uid uint, po *ridge.RidgeContact) (err error) {
	if err = validateContactBirthday(po); err != nil {
		return err
	}
	po.UserId = uid
	po.SuggestedAction = 0
	err = global.GVA_DB.Create(po).Error
	return err
}

func (poService *RidgeContactService) DeleteRidgeContact(uid uint, id string) (err error) {
	return poService.DeleteRidgeContactByIds(uid, []string{id})
}

func (poService *RidgeContactService) DeleteRidgeContactByIds(uid uint, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]ridge.RidgeContact{}, "user_id = ? AND id in ?", uid, ids).Error
	return err
}

func (poService *RidgeContactService) UpdateRidgeContact(uid uint, po ridge.RidgeContact) (err error) {
	if err = validateContactBirthday(&po); err != nil {
		return err
	}
	// Updates(struct)会跳过nil/false，生日必须支持清空及农历切回公历。
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&ridge.RidgeContact{}).Where("user_id = ? AND id = ?", uid, po.Id)
		if err := query.Omit("Birthday", "BirthdayIsLunar", "BirthdayIsLeapMonth").Updates(&po).Error; err != nil {
			return err
		}
		return tx.Model(&ridge.RidgeContact{}).Where("user_id = ? AND id = ?", uid, po.Id).
			Updates(map[string]interface{}{"birthday": po.Birthday, "birthday_is_lunar": po.BirthdayIsLunar, "birthday_is_leap_month": po.BirthdayIsLeapMonth}).Error
	})
}

func (poService *RidgeContactService) GetRidgeContact(uid uint, id string) (po ridge.RidgeContact, err error) {
	err = global.GVA_DB.Where("user_id = ? AND id = ?", uid, id).First(&po).Error
	return
}

func (poService *RidgeContactService) GetRidgeContactInfoList(uid uint, info request.RidgeContactSearch) (list []ridge.RidgeContact, total int64, err error) {
	// 查询条件
	db := global.GVA_DB.Model(&ridge.RidgeContact{})
	if info.RemarkName != nil && *info.RemarkName != "" {
		db = db.Where("remark_name LIKE ?", "%"+*info.RemarkName+"%")
	}
	if info.Gender != nil {
		db = db.Where("gender = ?", *info.Gender)
	}
	if info.CategoryId != nil {
		db = db.Where("category_id = ?", *info.CategoryId)
	}
	if info.TagId != nil {
		db = db.Where("tag_id = ?", *info.TagId)
	}
	if info.Status != nil {
		db = db.Where("status = ?", *info.Status)
	}
	if info.IsExempt != nil {
		db = db.Where("is_exempt = ?", *info.IsExempt)
	}
	if info.SuggestedAction != nil {
		db = db.Where("suggested_action = ?", *info.SuggestedAction)
	}

	// 记录总数
	err = db.Count(&total).Error
	if err != nil {
		return
	}

	// 分页条件
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	// 查询记录
	var pos []ridge.RidgeContact
	err = db.Order("sort_order desc, is_exempt asc, suggested_action desc, status desc, friended_at desc, id desc").Find(&pos).Error
	return pos, total, err
}
