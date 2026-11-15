package ridge

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/springbear2020/self-hub/server/global"
	"github.com/springbear2020/self-hub/server/model/ridge"
	"gorm.io/gorm"
)

func TestNextContactBirthday(t *testing.T) {
	cases := []struct {
		name, birthday, today, want string
		lunar, leap                 bool
	}{
		{name: "今天", birthday: "1990-11-15", today: "2026-11-15", want: "2026-11-15"},
		{name: "跨年", birthday: "1990-01-02", today: "2026-12-31", want: "2027-01-02"},
		{name: "公历闰日平年", birthday: "2000-02-29", today: "2026-01-01", want: "2026-02-28"},
		{name: "公历闰日闰年", birthday: "2000-02-29", today: "2028-01-01", want: "2028-02-29"},
		{name: "农历正月初一", birthday: "1990-01-01", today: "2026-01-01", want: "2026-02-17", lunar: true},
		{name: "农历腊月跨公历年", birthday: "1990-12-08", today: "2026-01-01", want: "2026-01-26", lunar: true},
		{name: "农历生日已过", birthday: "1990-01-01", today: "2026-02-18", want: "2027-02-06", lunar: true},
		{name: "闰六月有对应闰月", birthday: "2017-06-01", today: "2025-01-01", want: "2025-07-25", lunar: true, leap: true},
		{name: "闰六月回退普通月", birthday: "2017-06-01", today: "2026-01-01", want: "2026-07-14", lunar: true, leap: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			today, _ := time.ParseInLocation("2006-01-02", tc.today, birthdayLocation)
			got, err := nextContactBirthday(ridge.RidgeContact{Birthday: &tc.birthday, BirthdayIsLunar: tc.lunar, BirthdayIsLeapMonth: tc.leap}, today)
			if err != nil {
				t.Fatal(err)
			}
			if got.Format("2006-01-02") != tc.want {
				t.Fatalf("got %s want %s", got.Format("2006-01-02"), tc.want)
			}
		})
	}
}

func TestValidateContactBirthday(t *testing.T) {
	for _, tc := range []struct {
		date               string
		lunar, leap, valid bool
	}{
		{"", false, false, true},
		{"1990-08-15", false, false, true},
		{"2000-02-29", false, false, true},
		{"1990-02-30", false, false, false},
		{"1990-13-01", true, false, false},
		{"1990-01-31", true, false, false},
		{"1990-01-01", true, true, false},
		{"2017-06-01", true, true, true},
		{"9999-01-01", true, false, false},
		{"2099-01-01", false, false, false},
		{"1990-1-1", false, false, false},
	} {
		contact := ridge.RidgeContact{Birthday: &tc.date, BirthdayIsLunar: tc.lunar, BirthdayIsLeapMonth: tc.leap}
		err := validateContactBirthday(&contact)
		if (err == nil) != tc.valid {
			t.Errorf("date %s lunar=%t leap=%t: %v", tc.date, tc.lunar, tc.leap, err)
		}
	}
	// 确实存在的农历二月三十应可保存，不能按公历校验。
	for year := 1980; year <= 2000; year++ {
		value := time.Date(year, 2, 1, 0, 0, 0, 0, time.UTC).Format("2006-01") + "-30"
		contact := ridge.RidgeContact{Birthday: &value, BirthdayIsLunar: true}
		if validateContactBirthday(&contact) == nil {
			now := time.Date(2026, 1, 1, 0, 0, 0, 0, birthdayLocation)
			if _, err := nextContactBirthday(contact, now); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatal("no valid lunar February 30 found")
}

func birthdayTestDB(t *testing.T) *gorm.DB {
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
	if err := db.AutoMigrate(&ridge.RidgeContact{}); err != nil {
		t.Fatal(err)
	}
	previous := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previous })
	return db
}

func TestContactBirthdayPersistenceAndTimeline(t *testing.T) {
	db := birthdayTestDB(t)
	service := &RidgeContactService{}
	name, date := "好友", "2017-06-01"
	contact := ridge.RidgeContact{RemarkName: &name, Birthday: &date, BirthdayIsLunar: true, BirthdayIsLeapMonth: true}
	if err := service.CreateRidgeContact(1, &contact); err != nil {
		t.Fatal(err)
	}
	solar := "1990-11-15"
	contact.Birthday = &solar
	contact.BirthdayIsLunar = false
	contact.BirthdayIsLeapMonth = false
	if err := service.UpdateRidgeContact(1, contact); err != nil {
		t.Fatal(err)
	}
	saved, err := service.GetRidgeContact(1, "1")
	if err != nil {
		t.Fatal(err)
	}
	if saved.BirthdayIsLunar || saved.BirthdayIsLeapMonth || *saved.Birthday != solar {
		t.Fatalf("switch to solar failed: %+v", saved)
	}
	// 北京时间11月15日零点，今天生日必须算0天；包含已删除状态。
	now := time.Date(2026, 11, 14, 16, 0, 0, 0, time.UTC)
	deleted := 2
	otherDate := "1990-11-22"
	other := ridge.RidgeContact{UserId: 2, RemarkName: &name, Birthday: &otherDate, Status: &deleted}
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	outside := "1990-11-23"
	if err := db.Create(&ridge.RidgeContact{Birthday: &outside}).Error; err != nil {
		t.Fatal(err)
	}
	items, err := service.GetUpcomingContactBirthdays(7, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].DaysUntil != 0 || items[1].DaysUntil != 7 {
		t.Fatalf("wrong timeline: %+v", items)
	}
	contact.Birthday = nil
	if err := service.UpdateRidgeContact(1, contact); err != nil {
		t.Fatal(err)
	}
	saved, err = service.GetRidgeContact(1, "1")
	if err != nil {
		t.Fatal(err)
	}
	if saved.Birthday != nil {
		t.Fatal("birthday was not cleared to NULL")
	}
	// 越权修改不得影响另一个用户的生日。
	other.Birthday = nil
	if err := service.UpdateRidgeContact(99, other); err != nil {
		t.Fatal(err)
	}
	otherSaved, err := service.GetRidgeContact(2, "2")
	if err != nil {
		t.Fatal(err)
	}
	if otherSaved.Birthday == nil || *otherSaved.Birthday != otherDate {
		t.Fatal("unauthorized update changed another user birthday")
	}
	if _, err := service.GetUpcomingContactBirthdays(365, now); err == nil {
		t.Fatal("invalid range accepted")
	}
}

func TestContactBirthdayOccasion(t *testing.T) {
	for _, tc := range []struct {
		name, birth, next string
		lunar             bool
		age               int
		label, level      string
	}{
		{name: "普通生日", birth: "2001-11-15", next: "2026-11-15", age: 25},
		{name: "迎来的年龄而非当前年龄", birth: "2009-01-03", next: "2027-01-03", age: 18, label: "成人礼", level: "adult"},
		{name: "三十", birth: "1996-11-15", next: "2026-11-15", age: 30, label: "三十而立", level: "anniversary"},
		{name: "四十", birth: "1986-11-15", next: "2026-11-15", age: 40, label: "四十不惑", level: "anniversary"},
		{name: "五十", birth: "1976-11-15", next: "2026-11-15", age: 50, label: "五十知命", level: "anniversary"},
		{name: "花甲", birth: "1966-11-15", next: "2026-11-15", age: 60, label: "六十大寿 · 花甲之年", level: "longevity"},
		{name: "古稀", birth: "1956-11-15", next: "2026-11-15", age: 70, label: "七十大寿 · 古稀之年", level: "longevity"},
		{name: "八十", birth: "1946-11-15", next: "2026-11-15", age: 80, label: "八十大寿", level: "grand"},
		{name: "九十", birth: "1936-11-15", next: "2026-11-15", age: 90, label: "九十大寿", level: "grand"},
		{name: "期颐", birth: "1926-11-15", next: "2026-11-15", age: 100, label: "百岁大寿 · 期颐之年", level: "grand"},
		{name: "百岁以上", birth: "1925-11-15", next: "2026-11-15", age: 101, label: "101 岁高寿", level: "grand"},
		{name: "六六寿", birth: "1960-11-15", next: "2026-11-15", age: 66, label: "六六寿", level: "commemorative"},
		{name: "喜寿", birth: "1949-11-15", next: "2026-11-15", age: 77, label: "喜寿", level: "commemorative"},
		{name: "米寿", birth: "1938-11-15", next: "2026-11-15", age: 88, label: "米寿", level: "commemorative"},
		{name: "白寿", birth: "1927-11-15", next: "2026-11-15", age: 99, label: "白寿", level: "commemorative"},
		{name: "不提前做整寿", birth: "1967-11-15", next: "2026-11-15", age: 59},
		{name: "腊月生日跨公历年不多算一岁", birth: "1972-12-14", next: "2027-01-21", lunar: true, age: 54},
		{name: "截图农历生日", birth: "2001-12-20", next: "2027-01-27", lunar: true, age: 25},
		{name: "农历跨年成人礼", birth: "2008-12-14", next: "2027-01-21", lunar: true, age: 18, label: "成人礼", level: "adult"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			next, err := time.ParseInLocation("2006-01-02", tc.next, birthdayLocation)
			if err != nil {
				t.Fatal(err)
			}
			age, label, level := contactBirthdayOccasion(ridge.RidgeContact{Birthday: &tc.birth, BirthdayIsLunar: tc.lunar}, next)
			if age != tc.age || label != tc.label || level != tc.level {
				t.Fatalf("got age=%d label=%s level=%s; want age=%d label=%s level=%s", age, label, level, tc.age, tc.label, tc.level)
			}
		})
	}
}

func TestUpcomingBirthdayIncludesMilestone(t *testing.T) {
	db := birthdayTestDB(t)
	birth, name := "1946-11-15", "奶奶"
	if err := db.Create(&ridge.RidgeContact{Birthday: &birth, RemarkName: &name}).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 15, 0, 0, 0, 0, birthdayLocation)
	items, err := (&RidgeContactService{}).GetUpcomingContactBirthdays(30, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Age != 80 || items[0].MilestoneLabel != "八十大寿" || items[0].MilestoneLevel != "grand" {
		t.Fatalf("milestone missing from response: %+v", items)
	}
}
