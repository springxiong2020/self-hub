package dto

type ContactStatItem struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type RidgeContactStat struct {
	Total       int64             `json:"total"`
	UnknownYear int64             `json:"unknownYear"`
	Years       []ContactStatItem `json:"years"`
	Genders     []ContactStatItem `json:"genders"`
	Categories  []ContactStatItem `json:"categories"`
}
