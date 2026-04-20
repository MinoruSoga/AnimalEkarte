package model

import "time"

// ClinicHoliday は病院が設定した個別休診日を表す
type ClinicHoliday struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"   json:"id"`
	ClinicID  uint64    `gorm:"not null"                   json:"clinic_id"`
	Date      time.Time `gorm:"type:date;not null"         json:"date"`
	Reason    string    `gorm:"not null;default:''"        json:"reason"`
	CreatedAt time.Time `gorm:"autoCreateTime"             json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"             json:"updated_at"`
}

func (ClinicHoliday) TableName() string { return "clinic_holidays" }
