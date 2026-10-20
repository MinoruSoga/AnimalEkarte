package model

import (
	"time"
)

// StaffReservationCapability records which reservation types a staff member can handle.
type StaffReservationCapability struct {
	ID                uint64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID          uint64           `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	StaffID           uint64           `gorm:"column:staff_id;type:bigint;not null" json:"staff_id"`
	ReservationTypeID uint64           `gorm:"column:reservation_type_id;type:bigint;not null" json:"reservation_type_id"`
	CreatedAt         time.Time        `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	Clinic            *Clinic          `gorm:"foreignKey:ClinicID;references:ID"          json:"clinic,omitempty"`
	Staff             *Staff           `gorm:"foreignKey:StaffID;references:ID"           json:"staff,omitempty"`
	ReservationType   *ReservationType `gorm:"foreignKey:ReservationTypeID;references:ID" json:"reservation_type,omitempty"`
}

func (StaffReservationCapability) TableName() string { return "staff_reservation_capabilities" }
