package dto

type ContactBirthday struct {
	Id                  int    `json:"id"`
	Name                string `json:"name"`
	Birthday            string `json:"birthday"`
	BirthdayIsLunar     bool   `json:"birthdayIsLunar"`
	BirthdayIsLeapMonth bool   `json:"birthdayIsLeapMonth"`
	NextBirthday        string `json:"nextBirthday"`
	DaysUntil           int    `json:"daysUntil"`
}
