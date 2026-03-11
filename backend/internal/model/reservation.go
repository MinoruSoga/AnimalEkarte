package model

import (
	"time"

	"github.com/google/uuid"
)

// ReservationAppointment は予約テーブルに対応するモデル
// DB テーブル名: reservation_appointments
type ReservationAppointment struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StartTime    time.Time  `gorm:"column:start_time;not null" json:"start_time"`
	EndTime      time.Time  `gorm:"column:end_time;not null" json:"end_time"`
	OwnerName    string     `gorm:"column:owner_name;not null" json:"owner_name"`
	PetName      string     `gorm:"column:pet_name;not null" json:"pet_name"`
	PetID        *uuid.UUID `gorm:"column:pet_id;type:uuid" json:"pet_id,omitempty"`
	VisitType    string     `gorm:"column:visit_type;type:visit_type;not null" json:"visit_type"`
	Type         string     `gorm:"column:type;not null" json:"type"`
	Doctor       string     `gorm:"column:doctor;not null" json:"doctor"`
	IsDesignated bool       `gorm:"column:is_designated;default:false" json:"is_designated"`
	Status       string     `gorm:"column:status;type:reservation_status;default:'pending'" json:"status"`
	Notes        string     `gorm:"column:notes;default:''" json:"notes"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName は GORM に対してテーブル名を明示する
func (ReservationAppointment) TableName() string {
	return "reservation_appointments"
}

// CreateReservationRequest は予約作成リクエスト
type CreateReservationRequest struct {
	StartTime    string  `json:"start_time" binding:"required"`
	EndTime      string  `json:"end_time" binding:"required"`
	OwnerName    string  `json:"owner_name" binding:"required"`
	PetName      string  `json:"pet_name" binding:"required"`
	PetID        *string `json:"pet_id"`
	VisitType    string  `json:"visit_type" binding:"required"`
	Type         string  `json:"type" binding:"required"`
	Doctor       string  `json:"doctor" binding:"required"`
	IsDesignated bool    `json:"is_designated"`
	Status       string  `json:"status"`
	Notes        string  `json:"notes"`
}

// UpdateReservationRequest は予約更新リクエスト
type UpdateReservationRequest struct {
	StartTime    *string `json:"start_time"`
	EndTime      *string `json:"end_time"`
	OwnerName    *string `json:"owner_name"`
	PetName      *string `json:"pet_name"`
	PetID        *string `json:"pet_id"`
	VisitType    *string `json:"visit_type"`
	Type         *string `json:"type"`
	Doctor       *string `json:"doctor"`
	IsDesignated *bool   `json:"is_designated"`
	Status       *string `json:"status"`
	Notes        *string `json:"notes"`
}
