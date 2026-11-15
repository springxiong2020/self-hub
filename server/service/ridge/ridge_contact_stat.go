package ridge

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/springbear2020/self-hub/server/global"
	"github.com/springbear2020/self-hub/server/model/ridge"
	"github.com/springbear2020/self-hub/server/model/ridge/dto"
	"github.com/springbear2020/self-hub/server/model/system"
)

// GetRidgeContactStat uses the entire contact table, including deleted-friend statuses.
// Only statistical fields are loaded; contact details are never returned.
func (s *RidgeContactService) GetRidgeContactStat() (dto.RidgeContactStat, error) {
	result := dto.RidgeContactStat{
		Years:      []dto.ContactStatItem{},
		Genders:    []dto.ContactStatItem{{Name: "男"}, {Name: "女"}, {Name: "未知"}},
		Categories: []dto.ContactStatItem{},
	}
	var contacts []ridge.RidgeContact
	if err := global.GVA_DB.Select("friended_at", "gender", "category_id").Find(&contacts).Error; err != nil {
		return result, err
	}
	var categories []system.SysDictionaryDetail
	if err := global.GVA_DB.Model(&system.SysDictionaryDetail{}).
		Select("sys_dictionary_details.value", "sys_dictionary_details.label").
		Joins("JOIN sys_dictionaries d ON d.id = sys_dictionary_details.sys_dictionary_id").
		Where("d.type = ? AND d.deleted_at IS NULL", "ridge_contact_category").
		Find(&categories).Error; err != nil {
		return result, err
	}
	labels := map[int]string{0: "未分类"}
	for _, category := range categories {
		if id, err := strconv.Atoi(category.Value); err == nil {
			labels[id] = category.Label
		}
	}
	yearCounts := map[int]int64{}
	categoryCounts := map[int]int64{}
	result.Total = int64(len(contacts))
	for _, contact := range contacts {
		if contact.FriendedAt == nil || contact.FriendedAt.IsZero() {
			result.UnknownYear++
		} else {
			yearCounts[contact.FriendedAt.Year()]++
		}
		genderIndex := 2
		if contact.Gender != nil && (*contact.Gender == 1 || *contact.Gender == 2) {
			genderIndex = *contact.Gender - 1
		}
		result.Genders[genderIndex].Value++
		categoryID := 0
		if contact.CategoryId != nil {
			categoryID = *contact.CategoryId
		}
		categoryCounts[categoryID]++
	}
	years := make([]int, 0, len(yearCounts))
	for year := range yearCounts {
		years = append(years, year)
	}
	sort.Ints(years)
	if len(years) > 0 {
		for year := years[0]; year <= years[len(years)-1]; year++ {
			result.Years = append(result.Years, dto.ContactStatItem{Name: strconv.Itoa(year), Value: yearCounts[year]})
		}
	}
	for id, count := range categoryCounts {
		label, ok := labels[id]
		if !ok {
			label = fmt.Sprintf("未知分类（%d）", id)
		}
		result.Categories = append(result.Categories, dto.ContactStatItem{Name: label, Value: count})
	}
	sort.Slice(result.Categories, func(i, j int) bool {
		if result.Categories[i].Value == result.Categories[j].Value {
			return result.Categories[i].Name < result.Categories[j].Name
		}
		return result.Categories[i].Value > result.Categories[j].Value
	})
	return result, nil
}
