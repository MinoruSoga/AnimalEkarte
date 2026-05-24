package model

import "time"

// LstepSettings はクリニックごとのLステップ同期設定。
type LstepSettings struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID      uint64     `gorm:"not null;uniqueIndex"     json:"clinic_id"`
	IsSyncEnabled bool       `gorm:"not null;default:false"   json:"is_sync_enabled"`
	SyncEnabledAt *time.Time `gorm:"column:sync_enabled_at"   json:"sync_enabled_at"`
	CreatedAt     time.Time  `gorm:"not null;default:now()"   json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null;default:now()"   json:"updated_at"`
}

func (LstepSettings) TableName() string { return "lstep_settings" }
