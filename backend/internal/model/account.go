package model

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement"   json:"id"`
	Email         string         `gorm:"not null;uniqueIndex"       json:"email"`
	PasswordHash  string         `gorm:"not null"                   json:"-"`
	IsActive      bool           `gorm:"default:true"               json:"is_active"`
	IsSystemAdmin bool           `gorm:"default:false"              json:"is_system_admin"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"             json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"             json:"updated_at"`
	DeletedAt     gorm.DeletedAt `                                  json:"-"`

	// Relations
	Staff *Staff `gorm:"foreignKey:AccountID" json:"staff,omitempty"`
}

func (Account) TableName() string { return "accounts" }

type StaffClinicAssignment struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement"                       json:"id"`
	StaffID  uint64 `gorm:"not null;uniqueIndex:uk_staff_clinic"           json:"staff_id"`
	ClinicID uint64 `gorm:"not null;uniqueIndex:uk_staff_clinic"           json:"clinic_id"`
	IsMain   bool   `gorm:"default:false"                                  json:"is_main"`
	// CapabilitiesConfiguredAt は予約対応区分が PUT /v1/staffs/:id/reservation-capabilities
	// （または exclusion facade 経由）で明示設定された時刻。空選択による保存（junction 行 0 件）
	// でも設定される durable marker で、NULL = 未設定（backfill migration 023 の対象）。
	// write path のみが書き込み、消す経路はない。
	CapabilitiesConfiguredAt *time.Time     `gorm:"column:capabilities_configured_at"        json:"capabilities_configured_at,omitempty"`
	CreatedAt                time.Time      `gorm:"autoCreateTime"                           json:"created_at"`
	UpdatedAt                time.Time      `gorm:"autoUpdateTime"                           json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index"                                    json:"-"`

	// Relations
	Staff  *Staff  `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
	Clinic *Clinic `gorm:"foreignKey:ClinicID" json:"clinic,omitempty"`
}

func (StaffClinicAssignment) TableName() string { return "staff_clinic_assignments" }
