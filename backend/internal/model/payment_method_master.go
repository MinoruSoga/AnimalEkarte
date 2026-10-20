package model

import (
	"gorm.io/gorm"
	"time"
)

// PaymentMethodMaster は診療所ごとの支払方法マスタ
type PaymentMethodMaster struct {
	ID           uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID     uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name         string         `gorm:"column:name;type:varchar(50);not null" json:"name"`
	SystemKey    *string        `gorm:"column:system_key;type:varchar(50)" json:"-"`
	DisplayOrder int            `gorm:"column:display_order;type:integer;not null;default:0" json:"display_order"`
	IsActive     bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (PaymentMethodMaster) TableName() string { return "payment_methods" }
