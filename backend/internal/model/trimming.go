package model

import (
	"time"

	"github.com/google/uuid"
)

// Trimming トリミング記録モデル
type Trimming struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
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
