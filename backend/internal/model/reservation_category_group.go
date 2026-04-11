package model

import "time"

// ReservationCategoryGroup は予約区分のグループ（カレンダー凡例用）
type ReservationCategoryGroup struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID  uint64    `gorm:"not null"                 json:"clinic_id"`
	Name      string    `gorm:"not null"                 json:"name"`
	Color     string    `gorm:"not null;default:'#3B82F6'" json:"color"`
	SortOrder int       `gorm:"default:0"                json:"sort_order"`
	IsActive  bool      `gorm:"not null;default:true"    json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime"           json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"           json:"updated_at"`
}

func (ReservationCategoryGroup) TableName() string { return "reservation_type_groups" }
