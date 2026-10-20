package model

import (
	"time"
)

// ClosingSpecialPeriod は特別締め時間期間（GW・年末年始等）
type ClosingSpecialPeriod struct {
	ID           uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID     uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	StartDate    time.Time `gorm:"column:start_date;type:date;not null" json:"start_date"`
	EndDate      time.Time `gorm:"column:end_date;type:date;not null" json:"end_date"`
	AmPmBoundary string    `gorm:"column:am_pm_boundary;type:time;not null" json:"am_pm_boundary"`
	PmEnd        string    `gorm:"column:pm_end;type:time;not null" json:"pm_end"`
	Note         string    `gorm:"column:note;type:varchar(100);not null;default:''" json:"note"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (ClosingSpecialPeriod) TableName() string { return "closing_special_periods" }
