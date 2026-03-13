package model

import (
	"time"

	"gorm.io/gorm"
)

type StaffRole string

const (
	StaffRoleVeterinarian StaffRole = "veterinarian"
	StaffRoleNurse        StaffRole = "nurse"
	StaffRoleTrimmer      StaffRole = "trimmer"
	StaffRoleReception    StaffRole = "reception"
	StaffRoleManager      StaffRole = "manager"
)

type Staff struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID      uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name          string         `gorm:"not null"                                       json:"name"`
	IsActive      bool           `gorm:"default:true"                                   json:"is_active"`
	StaffRole     StaffRole      `gorm:"type:staff_role;not null"                       json:"staff_role"`
	JobTitleID    *uint64        `                                                      json:"job_title_id,omitempty"`
	LicenseNumber string         `gorm:"default:''"                                     json:"license_number"`
	SortOrder     int            `gorm:"default:0"                                      json:"sort_order"`
	DeletedAt     gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	JobTitle *JobTitle `gorm:"foreignKey:JobTitleID" json:"job_title,omitempty"`
}

func (Staff) TableName() string { return "staffs" }

type ShiftType string

const (
	ShiftTypeFull      ShiftType = "full"
	ShiftTypeMorning   ShiftType = "morning"
	ShiftTypeAfternoon ShiftType = "afternoon"
	ShiftTypeOff       ShiftType = "off"
	ShiftTypePaidLeave ShiftType = "paid_leave"
)

type ShiftEntry struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID  uint64    `gorm:"not null"                                       json:"clinic_id"`
	StaffID   uint64    `gorm:"not null"                                       json:"staff_id"`
	Date      time.Time `gorm:"type:date;not null"                             json:"date"`
	ShiftType ShiftType `gorm:"type:shift_type;not null"                       json:"shift_type"`
	StartTime string    `gorm:"default:''"                                     json:"start_time"`
	EndTime   string    `gorm:"default:''"                                     json:"end_time"`
	Note      string    `gorm:"default:''"                                     json:"note"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Staff Staff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (ShiftEntry) TableName() string { return "shift_entries" }
