package ridge

import "time"

// RidgeContactAnalysis 每个用户保留一条分析记录，用于跨重启、跨实例的每日去重。
// AnalysisDate 采用北京时间 YYYY-MM-DD，仅在联系人分析成功后提交。
type RidgeContactAnalysis struct {
	UserId       uint       `gorm:"primaryKey;autoIncrement:false;column:user_id;comment:用户ID"`
	AnalysisDate string     `gorm:"column:analysis_date;size:10;not null;default:'';comment:最近成功分析日期（北京时间）"`
	CompletedAt  *time.Time `gorm:"column:completed_at;comment:最近成功分析时间"`
}

func (RidgeContactAnalysis) TableName() string {
	return "ridge_contact_analyses"
}
