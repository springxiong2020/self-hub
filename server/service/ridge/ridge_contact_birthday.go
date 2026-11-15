package ridge

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/6tail/lunar-go/calendar"
	"github.com/springbear2020/self-hub/server/global"
	"github.com/springbear2020/self-hub/server/model/ridge"
	"github.com/springbear2020/self-hub/server/model/ridge/dto"
)

var birthdayLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

func birthdayParts(value string) (year, month, day int, err error) {
	if len(value) != 10 || value[4] != '-' || value[7] != '-' {
		return 0, 0, 0, errors.New("生日格式应为 YYYY-MM-DD")
	}
	for index, char := range value {
		if index != 4 && index != 7 && (char < '0' || char > '9') {
			return 0, 0, 0, errors.New("生日格式应为 YYYY-MM-DD")
		}
	}
	year, e1 := strconv.Atoi(value[:4])
	month, e2 := strconv.Atoi(value[5:7])
	day, e3 := strconv.Atoi(value[8:])
	if e1 != nil || e2 != nil || e3 != nil || year < 1900 || year > 2100 || month < 1 || month > 12 || day < 1 || day > 31 {
		return 0, 0, 0, errors.New("生日应为1900至2100年之间的有效年月日")
	}
	return year, month, day, nil
}

func validateContactBirthday(contact *ridge.RidgeContact) error {
	if contact.Birthday == nil || strings.TrimSpace(*contact.Birthday) == "" {
		contact.Birthday = nil
		contact.BirthdayIsLunar = false
		contact.BirthdayIsLeapMonth = false
		return nil
	}
	value := strings.TrimSpace(*contact.Birthday)
	year, month, day, err := birthdayParts(value)
	if err != nil {
		return err
	}
	var birth time.Time
	if contact.BirthdayIsLunar {
		lunarMonth := month
		if contact.BirthdayIsLeapMonth {
			lunarMonth = -month
		}
		m := calendar.NewLunarMonthFromYm(year, lunarMonth)
		if m == nil {
			return errors.New("出生年份没有所选农历闰月，请检查生日")
		}
		if day > m.GetDayCount() {
			return fmt.Errorf("所选农历月份只有%d天", m.GetDayCount())
		}
		solar := calendar.NewLunarFromYmd(year, lunarMonth, day).GetSolar()
		birth = time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), 0, 0, 0, 0, birthdayLocation)
	} else {
		contact.BirthdayIsLeapMonth = false
		birth, err = time.ParseInLocation("2006-01-02", value, birthdayLocation)
		if err != nil {
			return errors.New("公历生日日期无效")
		}
	}
	if birth.After(time.Now().In(birthdayLocation)) {
		return errors.New("生日不能晚于今天")
	}
	contact.Birthday = &value
	return nil
}

// nextContactBirthday 返回不早于今天的下次生日（北京时间）。
// 农历闰月当年不存在时使用同名普通月；目标月没有三十则取廿九；
// 公历2月29日在非闰年按2月28日计算。原始生日保持不变。
func nextContactBirthday(contact ridge.RidgeContact, today time.Time) (time.Time, error) {
	if contact.Birthday == nil {
		return time.Time{}, errors.New("未填写生日")
	}
	birthYear, month, day, err := birthdayParts(*contact.Birthday)
	if err != nil {
		return time.Time{}, err
	}
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, birthdayLocation)
	for year := today.Year() - 1; year <= today.Year()+2; year++ {
		if year < birthYear {
			continue
		}
		var candidate time.Time
		if contact.BirthdayIsLunar {
			targetMonth := month
			if contact.BirthdayIsLeapMonth && calendar.NewLunarYear(year).GetLeapMonth() == month {
				targetMonth = -month
			}
			lunarMonth := calendar.NewLunarMonthFromYm(year, targetMonth)
			if lunarMonth == nil {
				continue
			}
			targetDay := min(day, lunarMonth.GetDayCount())
			solar := calendar.NewLunarFromYmd(year, targetMonth, targetDay).GetSolar()
			candidate = time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), 0, 0, 0, 0, birthdayLocation)
		} else {
			lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, birthdayLocation).Day()
			candidate = time.Date(year, time.Month(month), min(day, lastDay), 0, 0, 0, 0, birthdayLocation)
		}
		if !candidate.Before(today) {
			return candidate, nil
		}
	}
	return time.Time{}, errors.New("无法计算下次生日")
}

// contactBirthdayOccasion 按下次生日所属历法计算生日年龄，不加虚岁。
// 农历腊月可能落在下一公历年，必须使用农历年份，不能直接减公历年份。
// 18岁突出成人纪念，30/40/50标注人生节点，60/70强调整寿，80/90/100及百岁以上重点强调。
// 66/77/88/99作为次级民俗纪念标签；不统一采用地区性的“做九不做十”。
func contactBirthdayOccasion(contact ridge.RidgeContact, next time.Time) (age int, label, level string) {
	birthYear, _, _, err := birthdayParts(*contact.Birthday)
	if err != nil {
		return 0, "", ""
	}
	next = next.In(birthdayLocation)
	year := next.Year()
	if contact.BirthdayIsLunar {
		year = calendar.NewSolarFromYmd(next.Year(), int(next.Month()), next.Day()).GetLunar().GetYear()
	}
	age = year - birthYear
	switch age {
	case 18:
		label, level = "成人礼", "adult"
	case 30:
		label, level = "三十而立", "anniversary"
	case 40:
		label, level = "四十不惑", "anniversary"
	case 50:
		label, level = "五十知命", "anniversary"
	case 60:
		label, level = "六十大寿 · 花甲之年", "longevity"
	case 70:
		label, level = "七十大寿 · 古稀之年", "longevity"
	case 80:
		label, level = "八十大寿 · 杖朝之年", "grand"
	case 90:
		label, level = "九十大寿 · 鲐背之年", "grand"
	case 100:
		label, level = "百岁大寿 · 期颐之年", "grand"
	case 66:
		label, level = "六六寿", "commemorative"
	case 77:
		label, level = "喜寿", "commemorative"
	case 88:
		label, level = "米寿", "commemorative"
	case 99:
		label, level = "白寿", "commemorative"
	default:
		if age > 100 {
			label, level = fmt.Sprintf("%d 岁高寿", age), "grand"
		}
	}
	return
}

// GetUpcomingContactBirthdays 与页面概览一样按全表计算，不受分页或搜索条件影响。
func (s *RidgeContactService) GetUpcomingContactBirthdays(days int, now time.Time) ([]dto.ContactBirthday, error) {
	if days != 7 && days != 30 && days != 90 {
		return nil, errors.New("生日提醒范围仅支持7、30或90天")
	}
	contacts := []ridge.RidgeContact{}
	if err := global.GVA_DB.Select("id", "remark_name", "birthday", "birthday_is_lunar", "birthday_is_leap_month").
		Where("birthday IS NOT NULL AND birthday <> ''").Find(&contacts).Error; err != nil {
		return nil, err
	}
	now = now.In(birthdayLocation)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, birthdayLocation)
	result := []dto.ContactBirthday{}
	for _, contact := range contacts {
		next, err := nextContactBirthday(contact, today)
		if err != nil {
			continue
		}
		remaining := int(next.Sub(today).Hours() / 24)
		if remaining > days {
			continue
		}
		name := "未命名"
		if contact.RemarkName != nil {
			name = *contact.RemarkName
		}
		if contact.Id == nil {
			continue
		}
		age, milestoneLabel, milestoneLevel := contactBirthdayOccasion(contact, next)
		result = append(result, dto.ContactBirthday{Id: *contact.Id, Name: name, Birthday: *contact.Birthday,
			BirthdayIsLunar: contact.BirthdayIsLunar, BirthdayIsLeapMonth: contact.BirthdayIsLeapMonth,
			NextBirthday: next.Format("2006-01-02"), DaysUntil: remaining,
			Age: age, MilestoneLabel: milestoneLabel, MilestoneLevel: milestoneLevel})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].DaysUntil == result[j].DaysUntil {
			return result[i].Id < result[j].Id
		}
		return result[i].DaysUntil < result[j].DaysUntil
	})
	return result, nil
}
