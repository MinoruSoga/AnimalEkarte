package model

import (
	"time"
)

// LstepSettings はクリニックごとのLステップ同期設定。
type LstepSettings struct {
	ID            uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID      uint64     `gorm:"column:clinic_id;type:bigint;not null;uniqueIndex:lstep_settings_clinic_id_key" json:"clinic_id"`
	IsSyncEnabled bool       `gorm:"column:is_sync_enabled;type:boolean;not null;default:false" json:"is_sync_enabled"`
	SyncEnabledAt *time.Time `gorm:"column:sync_enabled_at;type:timestamptz" json:"sync_enabled_at"`
	CreatedAt     time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;type:timestamptz;not null;default:now()" json:"updated_at"`
}

func (LstepSettings) TableName() string { return "lstep_settings" }
