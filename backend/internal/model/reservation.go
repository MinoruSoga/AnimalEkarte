package model

import (
	"time"

	"github.com/google/uuid"
)

// Reservation 予約モデル
type Reservation struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	PetID        uuid.UUID  `json:"pet_id" gorm:"type:uuid;not null;index:idx_res_pet_id"`
	OwnerID      uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null"`
	DoctorID     *uuid.UUID `json:"doctor_id" gorm:"type:uuid;index:idx_res_doctor_id"`
	StartTime    time.Time  `json:"start_time" gorm:"index:idx_res_start_time"`
	EndTime      time.Time  `json:"end_time"`
	VisitType    string     `json:"visit_type" gorm:"type:varchar(20)"`   // first, revisit
	ServiceType  string     `json:"service_type" gorm:"type:varchar(30)"` // 診療, 検診, 手術, etc.
	IsDesignated bool       `json:"is_designated" gorm:"default:false"`
	Status       string     `json:"status" gorm:"type:varchar(30);default:'pending'"` // pending, confirmed, checked_in, in_consultation, accounting, completed, canceled
	Notes        string     `json:"notes" gorm:"type:text"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Relations
	Pet   *Pet   `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Owner *Owner `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
}

// TableName テーブル名を指定
func (Reservation) TableName() string {
	return "reservations"
}

// CreateReservationRequest リクエスト用構造体
type CreateReservationRequest struct {
	PetID        uuid.UUID  `json:"pet_id" binding:"required"`
	OwnerID      uuid.UUID  `json:"owner_id" binding:"required"`
	DoctorID     *uuid.UUID `json:"doctor_id"`
	StartTime    time.Time  `json:"start_time" binding:"required"`
	EndTime      time.Time  `json:"end_time" binding:"required"`
	VisitType    string     `json:"visit_type" binding:"required"` // first, revisit
	ServiceType  string     `json:"service_type" binding:"required"`
	IsDesignated bool       `json:"is_designated"`
	Notes        string     `json:"notes"`
}

// UpdateReservationRequest 更新用リクエスト構造体
type UpdateReservationRequest struct {
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	VisitType    string     `json:"visit_type"`
	ServiceType  string     `json:"service_type"`
	DoctorID     *uuid.UUID `json:"doctor_id"`
	IsDesignated bool       `json:"is_designated"`
	Status       string     `json:"status"`
	Notes        string     `json:"notes"`
}
