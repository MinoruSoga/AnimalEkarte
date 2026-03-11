package model

import (
	"time"

	"github.com/google/uuid"
)

// ShiftEntry はシフトエントリテーブルに対応するモデル
// staff_id は master_items (category='staff') を参照
type ShiftEntry struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StaffID   uuid.UUID `gorm:"column:staff_id;type:uuid;not null" json:"staff_id"`
	Date      time.Time `gorm:"column:date;type:date;not null" json:"date"`
	ShiftType string    `gorm:"column:shift_type;type:shift_type;not null" json:"shift_type"`
	StartTime string    `gorm:"column:start_time;default:''" json:"start_time"`
	EndTime   string    `gorm:"column:end_time;default:''" json:"end_time"`
	Note      string    `gorm:"column:note;default:''" json:"note"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// CreateShiftEntryRequest はシフト作成リクエスト
type CreateShiftEntryRequest struct {
	StaffID   string `json:"staff_id" binding:"required"`
	Date      string `json:"date" binding:"required"`
	ShiftType string `json:"shift_type" binding:"required"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Note      string `json:"note"`
}

// UpdateShiftEntryRequest はシフト更新リクエスト
type UpdateShiftEntryRequest struct {
	StaffID   *string `json:"staff_id"`
	Date      *string `json:"date"`
	ShiftType *string `json:"shift_type"`
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`
	Note      *string `json:"note"`
}
