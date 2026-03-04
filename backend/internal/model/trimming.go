package model

import (
	"time"

	"github.com/google/uuid"
)

// Trimming トリミング記録モデル
type Trimming struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	PetID           uuid.UUID  `json:"pet_id" gorm:"type:uuid;not null"`
	OwnerID         uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null"`
	StaffID         *uuid.UUID `json:"staff_id" gorm:"type:uuid"`
	AppointmentDate time.Time  `json:"appointment_date"`
	Course          string     `json:"course" gorm:"type:varchar(100)"`
	Options         string     `json:"options" gorm:"type:json"` // JSON array
	StyleRequest    string     `json:"style_request" gorm:"type:text"`
	Status          string     `json:"status" gorm:"type:varchar(20);default:'予約'"` // 予約, 進行中, 完了
	TotalPrice      *float64   `json:"total_price" gorm:"type:decimal(10,2)"`
	Notes           string     `json:"notes" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	Pet   *Pet   `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Owner *Owner `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
}

// TableName テーブル名を指定
func (Trimming) TableName() string {
	return "trimmings"
}

// CreateTrimmingRequest トリミング作成リクエスト
type CreateTrimmingRequest struct {
	PetID           uuid.UUID  `json:"pet_id" binding:"required"`
	OwnerID         uuid.UUID  `json:"owner_id" binding:"required"`
	StaffID         *uuid.UUID `json:"staff_id"`
	AppointmentDate time.Time  `json:"appointment_date" binding:"required"`
	Course          string     `json:"course" binding:"required"`
	Options         string     `json:"options"`
	StyleRequest    string     `json:"style_request"`
	Notes           string     `json:"notes"`
}

// UpdateTrimmingRequest トリミング更新リクエスト
type UpdateTrimmingRequest struct {
	Status          string     `json:"status"`
	AppointmentDate *time.Time `json:"appointment_date"`
	Course          string     `json:"course"`
	Options         string     `json:"options"`
	StyleRequest    string     `json:"style_request"`
	TotalPrice      *float64   `json:"total_price"`
	Notes           string     `json:"notes"`
}
