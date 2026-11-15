package ridge

import (
	"reflect"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/springbear2020/self-hub/server/global"
	"github.com/springbear2020/self-hub/server/model/ridge/dto"
	"gorm.io/gorm"
)

func TestGetRidgeContactStat(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	defer sqlDB.Close()
	previous := global.GVA_DB
	global.GVA_DB = db
	defer func() { global.GVA_DB = previous }()
	for _, sql := range []string{
		`CREATE TABLE ridge_contacts (friended_at datetime, gender integer, category_id integer, status integer, user_id integer)`,
		`CREATE TABLE sys_dictionaries (id integer, type text, deleted_at datetime)`,
		`CREATE TABLE sys_dictionary_details (id integer, sys_dictionary_id integer, label text, value text, deleted_at datetime)`,
		`INSERT INTO sys_dictionaries (id, type) VALUES (1, 'ridge_contact_category'), (2, 'other')`,
		`INSERT INTO sys_dictionary_details (sys_dictionary_id, label, value) VALUES (1, '同学', '1'), (2, '错误分类', '1')`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	service := &RidgeContactService{}
	empty, err := service.GetRidgeContactStat()
	if err != nil {
		t.Fatal(err)
	}
	if empty.Total != 0 || empty.Years == nil || empty.Categories == nil || len(empty.Genders) != 3 {
		t.Fatalf("invalid empty result: %+v", empty)
	}
	if err := db.Exec(`INSERT INTO ridge_contacts VALUES
 ('2022-03-01', 1, 1, 0, 1),
 ('2024-06-01', 2, 1, 2, 2),
 (NULL, 0, 0, 1, 1),
 (NULL, NULL, 99, 2, 2)`).Error; err != nil {
		t.Fatal(err)
	}
	got, err := service.GetRidgeContactStat()
	if err != nil {
		t.Fatal(err)
	}
	want := dto.RidgeContactStat{
		Total: 4, UnknownYear: 2,
		Years:      []dto.ContactStatItem{{Name: "2022", Value: 1}, {Name: "2023", Value: 0}, {Name: "2024", Value: 1}},
		Genders:    []dto.ContactStatItem{{Name: "男", Value: 1}, {Name: "女", Value: 1}, {Name: "未知", Value: 2}},
		Categories: []dto.ContactStatItem{{Name: "同学", Value: 2}, {Name: "未分类", Value: 1}, {Name: "未知分类（99）", Value: 1}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("statistics mismatch\ngot: %+v\nwant: %+v", got, want)
	}
}
