package model

import (
	"time"

	"gorm.io/gorm"
)

// PaymentMethodMaster は診療所ごとの支払方法マスタ
type PaymentMethodMaster struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID     uint64         `gorm:"not null"                 json:"clinic_id"`
	Name         string         `gorm:"not null"                 json:"name"`
	DisplayOrder int            `gorm:"not null;default:0"       json:"display_order"`
	IsActive     bool           `gorm:"not null;default:true"    json:"is_active"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"           json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"           json:"updated_at"`
	DeletedAt    gorm.DeletedAt `                                json:"-"`
}

func (PaymentMethodMaster) TableName() string { return "payment_methods" }
