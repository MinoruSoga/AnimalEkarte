package model

import "time"

type LstepSyncErrorCounter struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID     uint64    `gorm:"not null"                json:"clinic_id"`
	OwnerID      uint64    `gorm:"not null"                json:"owner_id"`
	FailureCount int       `gorm:"not null;default:0"      json:"failure_count"`
	CreatedAt    time.Time `gorm:"not null;default:now()"  json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;default:now()"  json:"updated_at"`
}

func (LstepSyncErrorCounter) TableName() string { return "lstep_sync_error_counters" }
