package model

import (
	"time"
)

type LstepSyncErrorCounter struct {
	ID           uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID     uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	OwnerID      uint64    `gorm:"column:owner_id;type:bigint;not null" json:"owner_id"`
	FailureCount int       `gorm:"column:failure_count;type:int;not null;default:0" json:"failure_count"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now()" json:"updated_at"`
}

func (LstepSyncErrorCounter) TableName() string { return "lstep_sync_error_counters" }
