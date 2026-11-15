package task

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/springbear2020/self-hub/server/model/ridge"
	"gorm.io/gorm"
)

func TestAnalyzeRidgeContactRules(t *testing.T) {
	now := time.Date(2026, 11, 15, 2, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	daysAgo := func(days int) *time.Time { v := now.Add(-time.Duration(days) * 24 * time.Hour); return &v }
	normal, chatOnly := 0, 1
	zero := time.Time{}
	exact90 := daysAgo(90)
	over90 := exact90.Add(-time.Nanosecond)
	before180 := daysAgo(180).Add(time.Nanosecond)
	cases := []struct {
		name    string
		contact ridge.RidgeContact
		action  uint
		restore bool
	}{
		{"正常超过90天", ridge.RidgeContact{Status: &normal, LastChatAt: &over90}, 1, false},
		{"正常恰好90天", ridge.RidgeContact{Status: &normal, LastChatAt: exact90}, 0, false},
		{"新互动优先于旧聊天", ridge.RidgeContact{Status: &normal, LastChatAt: daysAgo(200), LastInteractionAt: daysAgo(10)}, 0, false},
		{"新聊天优先于旧互动", ridge.RidgeContact{Status: &normal, LastChatAt: daysAgo(10), LastInteractionAt: daysAgo(200)}, 0, false},
		{"仅互动记录", ridge.RidgeContact{Status: &normal, LastInteractionAt: daysAgo(91)}, 1, false},
		{"无联系记录回退添加时间", ridge.RidgeContact{Status: &normal, FriendedAt: daysAgo(91)}, 1, false},
		{"新添加但无联系记录", ridge.RidgeContact{Status: &normal, FriendedAt: daysAgo(10)}, 0, false},
		{"缺失全部时间", ridge.RidgeContact{Status: &normal}, 3, false},
		{"零时间视为缺失", ridge.RidgeContact{Status: &normal, FriendedAt: &zero, LastChatAt: &zero}, 3, false},
		{"仅聊天缺起点", ridge.RidgeContact{Status: &chatOnly, LastChatAt: daysAgo(1)}, 3, false},
		{"仅聊天不足180天", ridge.RidgeContact{Status: &chatOnly, ChatOnlySince: &before180, LastChatAt: daysAgo(400)}, 0, false},
		{"仅聊天恰好180天", ridge.RidgeContact{Status: &chatOnly, ChatOnlySince: daysAgo(180)}, 2, false},
		{"仅聊天超过180天", ridge.RidgeContact{Status: &chatOnly, ChatOnlySince: daysAgo(181)}, 2, false},
		{"旧联系不恢复", ridge.RidgeContact{Status: &chatOnly, ChatOnlySince: daysAgo(180), LastChatAt: daysAgo(200)}, 2, false},
		{"起点同一时刻不算新联系", ridge.RidgeContact{Status: &chatOnly, ChatOnlySince: daysAgo(180), LastChatAt: daysAgo(180)}, 2, false},
		{"观察期新聊天恢复", ridge.RidgeContact{Status: &chatOnly, ChatOnlySince: daysAgo(100), LastChatAt: daysAgo(10)}, 0, true},
		{"观察期新互动恢复", ridge.RidgeContact{Status: &chatOnly, ChatOnlySince: daysAgo(100), LastInteractionAt: daysAgo(10)}, 0, true},
		{"超过180天仍优先恢复", ridge.RidgeContact{Status: &chatOnly, ChatOnlySince: daysAgo(200), LastInteractionAt: daysAgo(1)}, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			action, restore := analyzeRidgeContact(tc.contact, now)
			if action != tc.action || restore != tc.restore {
				t.Fatalf("got action=%d restore=%t; want action=%d restore=%t", action, restore, tc.action, tc.restore)
			}
		})
	}
}

func TestAnalyzeRidgeContactsDatabase(t *testing.T) {
	db := contactAnalysisTestDB(t)
	now := time.Date(2026, 11, 15, 2, 0, 0, 0, time.UTC)
	old, recent := now.Add(-200*24*time.Hour), now.Add(-24*time.Hour)
	normal, chatOnly, deleted, unknown := 0, 1, 2, 99
	no, yes := 0, 1
	name := "保持姓名"
	contacts := []ridge.RidgeContact{
		{Status: &chatOnly, IsExempt: &no, ChatOnlySince: &old, LastInteractionAt: &recent, SuggestedAction: 2, RemarkName: &name},
		{Status: &normal, IsExempt: &no, LastChatAt: &recent, SuggestedAction: 1},
		{Status: &normal, IsExempt: &no, FriendedAt: &old},
		{Status: &chatOnly, IsExempt: &no, ChatOnlySince: &old},
		{Status: &normal, IsExempt: &no, SuggestedAction: 2},
		{Status: &normal, IsExempt: &yes, FriendedAt: &old, SuggestedAction: 2},
		{Status: &deleted, IsExempt: &no, FriendedAt: &old, SuggestedAction: 1},
		{Status: &unknown, IsExempt: &no, SuggestedAction: 2},
		{IsExempt: &no, SuggestedAction: 2},
		{Status: &normal, SuggestedAction: 2},
	}
	for i := range contacts {
		contacts[i].UserId = 1
	}
	if err := db.Create(&contacts).Error; err != nil {
		t.Fatal(err)
	}
	if err := AnalyzeRidgeContacts(db, 1, now); err != nil {
		t.Fatal(err)
	}
	var got []ridge.RidgeContact
	if err := db.Order("id").Find(&got).Error; err != nil {
		t.Fatal(err)
	}
	wantActions := []uint{0, 0, 1, 2, 3, 2, 1, 2, 2, 2}
	for i, contact := range got {
		if contact.SuggestedAction != wantActions[i] {
			t.Errorf("contact %d: got action %d want %d", i, contact.SuggestedAction, wantActions[i])
		}
		if i >= 5 && !contact.UpdatedAt.Equal(*contacts[i].UpdatedAt) {
			t.Errorf("skipped contact %d was updated", i)
		}
	}
	if *got[0].Status != normal || got[0].ChatOnlySince != nil || *got[0].RemarkName != name {
		t.Fatalf("restore must write zero status and NULL observation start without changing name: %+v", got[0])
	}
	// An unchanged rerun must not touch updated_at.
	if err := AnalyzeRidgeContacts(db, 1, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	var rerun ridge.RidgeContact
	if err := db.First(&rerun, contacts[0].Id).Error; err != nil {
		t.Fatal(err)
	}
	if !rerun.UpdatedAt.Equal(*got[0].UpdatedAt) {
		t.Fatal("unchanged analysis updated timestamp")
	}
}

func TestAnalyzeRidgeContactsRollback(t *testing.T) {
	db := contactAnalysisTestDB(t)
	now := time.Now()
	old := now.Add(-200 * 24 * time.Hour)
	normal, no := 0, 0
	contacts := []ridge.RidgeContact{{Status: &normal, IsExempt: &no, FriendedAt: &old}, {Status: &normal, IsExempt: &no, FriendedAt: &old}}
	for i := range contacts {
		contacts[i].UserId = 1
	}
	if err := db.Create(&contacts).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TRIGGER fail_second_update BEFORE UPDATE ON ridge_contacts WHEN OLD.id = 2 BEGIN SELECT RAISE(ABORT, 'simulated write failure'); END`).Error; err != nil {
		t.Fatal(err)
	}
	if err := AnalyzeRidgeContacts(db, 1, now); err == nil {
		t.Fatal("expected write failure")
	}
	var count int64
	if err := db.Model(&ridge.RidgeContact{}).Where("suggested_action <> 0").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("failed analysis left partial updates")
	}
	if err := db.Model(&ridge.RidgeContactAnalysis{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("failed analysis was marked complete")
	}
	if err := db.Exec("DROP TRIGGER fail_second_update").Error; err != nil {
		t.Fatal(err)
	}
	if err := AnalyzeRidgeContacts(db, 1, now); err != nil {
		t.Fatalf("same-day retry failed: %v", err)
	}
	if err := db.Model(&ridge.RidgeContact{}).Where("suggested_action = 1").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal("same-day retry did not analyze contacts")
	}

	if err := AnalyzeRidgeContacts(nil, 1, now); err == nil {
		t.Fatal("expected nil database error")
	}
}

func contactAnalysisTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&ridge.RidgeContact{}, &ridge.RidgeContactAnalysis{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestRidgeContactAnalysisDailyPerUser(t *testing.T) {
	db := contactAnalysisTestDB(t)
	// 北京时间23:59触发首次分析，跨零点即进入新的自然日，不必等待24小时。
	now := time.Date(2026, 11, 15, 15, 59, 0, 0, time.UTC)
	recent := now.Add(-time.Hour)
	old := now.Add(-200 * 24 * time.Hour)
	normal, no := 0, 0
	contacts := []ridge.RidgeContact{
		{UserId: 1, Status: &normal, IsExempt: &no, LastChatAt: &recent},
		{UserId: 2, Status: &normal, IsExempt: &no, LastChatAt: &old},
	}
	if err := db.Create(&contacts).Error; err != nil {
		t.Fatal(err)
	}
	if err := AnalyzeRidgeContacts(db, 1, now); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Model(&ridge.RidgeContactAnalysis{}).Where("user_id = ?", 2).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("analysis created a completion record for an inactive user")
	}
	var other ridge.RidgeContact
	if err := db.First(&other, contacts[1].Id).Error; err != nil {
		t.Fatal(err)
	}
	if other.SuggestedAction != 0 {
		t.Fatal("analysis changed another user's contacts")
	}
	if err := db.Model(&ridge.RidgeContact{}).Where("id = ?", contacts[0].Id).Update("last_chat_at", old).Error; err != nil {
		t.Fatal(err)
	}
	// 使用新的GORM会话模拟新请求/服务重启；当日持久化记录仍会阻止重复分析。
	if err := AnalyzeRidgeContacts(db.Session(&gorm.Session{NewDB: true}), 1, now.Add(30*time.Second)); err != nil {
		t.Fatal(err)
	}
	var own ridge.RidgeContact
	if err := db.First(&own, contacts[0].Id).Error; err != nil {
		t.Fatal(err)
	}
	if own.SuggestedAction != 0 {
		t.Fatal("analysis ran twice on the same Beijing date")
	}
	if err := AnalyzeRidgeContacts(db, 1, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&own, contacts[0].Id).Error; err != nil {
		t.Fatal(err)
	}
	if own.SuggestedAction != 1 {
		t.Fatal("analysis did not run on the next Beijing date")
	}
	var run ridge.RidgeContactAnalysis
	if err := db.First(&run, "user_id = ?", 1).Error; err != nil {
		t.Fatal(err)
	}
	if run.AnalysisDate != "2026-11-16" {
		t.Fatalf("wrong completion date: %+v", run)
	}
	if err := AnalyzeRidgeContacts(db, 2, now); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&other, contacts[1].Id).Error; err != nil {
		t.Fatal(err)
	}
	if other.SuggestedAction != 1 {
		t.Fatal("second user was not analyzed independently")
	}
	if err := AnalyzeRidgeContacts(db, 0, now); err == nil {
		t.Fatal("expected invalid user ID error")
	}
}

func TestRidgeContactAnalysisEmptyUser(t *testing.T) {
	db := contactAnalysisTestDB(t)
	now := time.Now()
	if err := AnalyzeRidgeContacts(db, 1, now); err != nil {
		t.Fatal(err)
	}
	var run ridge.RidgeContactAnalysis
	if err := db.First(&run, "user_id = ?", 1).Error; err != nil {
		t.Fatal(err)
	}
	if run.AnalysisDate != now.In(ridgeAnalysisLocation).Format("2006-01-02") || run.CompletedAt == nil {
		t.Fatal("empty user analysis did not complete")
	}
}
