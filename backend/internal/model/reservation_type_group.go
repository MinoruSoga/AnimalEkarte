package model

import (
	"gorm.io/gorm"
	"time"
)

// ReservationTypeGroup は予約区分のグループ（カレンダー凡例用）
type ReservationTypeGroup struct {
	ID        uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name      string         `gorm:"column:name;type:text;not null" json:"name"`
	Color     string         `gorm:"column:color;type:text;not null;default:'#3B82F6'" json:"color"`
	SortOrder *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	IsActive  bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Version   int64          `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
}

func (ReservationTypeGroup) TableName() string { return "reservation_type_groups" }
