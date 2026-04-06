package model

import (
	"time"

	"gorm.io/gorm"
)

type Staff struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	AccountID     *uint64        `                                                      json:"account_id,omitempty"`
	Name          string         `gorm:"not null"                                       json:"name"`
	IsActive      bool           `gorm:"default:true"                                   json:"is_active"`
	OccupationID  *uint64        `                                                      json:"occupation_id,omitempty"`
	LicenseNumber string         `gorm:"default:''"                                     json:"license_number"`
	SortOrder     int            `gorm:"type:integer;default:0"                         json:"sort_order"`
	DeletedAt     gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Account           *Account                `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	Occupation        *Occupation             `gorm:"foreignKey:OccupationID" json:"occupation,omitempty"`
	ClinicAssignments []StaffClinicAssignment `gorm:"foreignKey:StaffID" json:"clinic_assignments,omitempty"`
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
	StartTime *string   `gorm:"type:time"                                     json:"start_time,omitempty"`
	EndTime   *string   `gorm:"type:time"                                     json:"end_time,omitempty"`
	Note      string    `gorm:"default:''"                                     json:"note"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Staff *Staff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (ShiftEntry) TableName() string { return "shift_entries" }
