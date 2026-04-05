package model

import (
	"time"

	"gorm.io/gorm"
)

// UserType defines authorization levels in the system
type UserType string

const (
	UserTypeSystemAdmin UserType = "system_admin"
	UserTypeClinicAdmin UserType = "clinic_admin"
	UserTypeStaff       UserType = "staff"
)

type Account struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	Email        string         `gorm:"not null;uniqueIndex"                           json:"email"`
	PasswordHash string         `gorm:"not null"                                       json:"-" swaggerignore:"true"`
	IsActive     bool           `gorm:"default:true"                                   json:"is_active"`
	UserType     string         `gorm:"type:user_type;not null;default:'staff'"        json:"user_type"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt    gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`

	// Relations
	Staff *Staff `gorm:"foreignKey:AccountID" json:"staff,omitempty"`
}

func (Account) TableName() string { return "accounts" }

type StaffClinicAssignment struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	StaffID   uint64    `gorm:"not null"                                       json:"staff_id"`
	ClinicID  uint64    `gorm:"not null"                                       json:"clinic_id"`
	IsMain    bool      `gorm:"default:false"                                  json:"is_main"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Staff  *Staff  `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
	Clinic *Clinic `gorm:"foreignKey:ClinicID" json:"clinic,omitempty"`
}

func (StaffClinicAssignment) TableName() string { return "staff_clinic_assignments" }
