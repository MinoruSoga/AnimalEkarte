package model

import "time"

// ClosingSpecialPeriod は特別締め時間期間（GW・年末年始等）
type ClosingSpecialPeriod struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID     uint64    `gorm:"not null"                 json:"clinic_id"`
	StartDate    time.Time `gorm:"type:date;not null"       json:"start_date"`
	EndDate      time.Time `gorm:"type:date;not null"       json:"end_date"`
	AmPmBoundary string    `gorm:"type:time;not null"       json:"am_pm_boundary"`
	PmEnd        string    `gorm:"type:time;not null"       json:"pm_end"`
	Note         string    `gorm:"not null;default:''"      json:"note"`
	CreatedAt    time.Time `gorm:"autoCreateTime"           json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"           json:"updated_at"`
}

func (ClosingSpecialPeriod) TableName() string { return "closing_special_periods" }
