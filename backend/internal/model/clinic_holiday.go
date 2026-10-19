package model

import (
	"time"
)

// ClinicHoliday は病院が設定した個別休診日を表す
type ClinicHoliday struct {
	ID        uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Date      time.Time `gorm:"column:date;type:date;not null" json:"date"`
	Reason    string    `gorm:"column:reason;type:text;not null;default:''" json:"reason"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (ClinicHoliday) TableName() string { return "clinic_holidays" }
