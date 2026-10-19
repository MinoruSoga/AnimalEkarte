package model

import (
	"gorm.io/gorm"
	"time"
)

type Account struct {
	ID            uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	Email         string         `gorm:"column:email;type:text;not null" json:"email"`
	PasswordHash  string         `gorm:"column:password_hash;type:text;not null" json:"-"`
	IsActive      bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	IsSystemAdmin bool           `gorm:"column:is_system_admin;type:boolean;not null;default:false" json:"is_system_admin"`
	CreatedAt     time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Staff         *Staff         `gorm:"foreignKey:AccountID;references:ID" json:"staff,omitempty"`
}

func (Account) TableName() string { return "accounts" }

type StaffClinicAssignment struct {
	ID                       uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	StaffID                  uint64         `gorm:"column:staff_id;type:bigint;not null" json:"staff_id"`
	ClinicID                 uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	IsMain                   bool           `gorm:"column:is_main;type:boolean;not null;default:false" json:"is_main"`
	CreatedAt                time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt                time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	CapabilitiesConfiguredAt *time.Time     `gorm:"column:capabilities_configured_at;type:timestamptz" json:"capabilities_configured_at,omitempty"`
	NominationPaused         *bool          `gorm:"column:nomination_paused;type:boolean" json:"-"`
	Staff                    *Staff         `gorm:"foreignKey:StaffID;references:ID" json:"staff,omitempty"`
	Clinic                   *Clinic        `gorm:"foreignKey:ClinicID;references:ID" json:"clinic,omitempty"`
}

func (StaffClinicAssignment) TableName() string { return "staff_clinic_assignments" }
