package model

import (
	"time"

	"github.com/lib/pq"
)

// ClinicSettings は診療所の締め時間設定
type ClinicSettings struct {
	ClinicID            uint64        `gorm:"primaryKey"                            json:"clinic_id"`
	ClosingAmPmBoundary string        `gorm:"type:time;not null;default:'14:00'"    json:"closing_am_pm_boundary"`
	ClosingWeekdayEnd   string        `gorm:"type:time;not null;default:'18:30'"    json:"closing_weekday_end"`
	ClosingSundayEnd    string        `gorm:"type:time;not null;default:'17:30'"    json:"closing_sunday_end"`
	ClosedWeekdays      pq.Int64Array `gorm:"type:smallint[];not null;default:'{}'" json:"closed_weekdays"`
	LstepFireHourJST    int           `gorm:"not null;default:10"                  json:"lstep_fire_hour_jst"`
	CPMVersion          string        `gorm:"type:varchar(8);not null;default:'v1'" json:"cpm_version"`
	CreatedAt           time.Time     `gorm:"autoCreateTime"                       json:"created_at"`
	UpdatedAt           time.Time     `gorm:"autoUpdateTime"                       json:"updated_at"`
}

func (ClinicSettings) TableName() string { return "clinic_settings" }
