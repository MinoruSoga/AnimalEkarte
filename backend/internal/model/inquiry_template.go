package model

import (
	"gorm.io/gorm"
	"time"
)

// InquiryTemplate は問診定型文マスタ（v11.0追加）
type InquiryTemplate struct {
	ID        uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Category  string         `gorm:"column:category;type:text;not null;default:''" json:"category"`
	Title     string         `gorm:"column:title;type:text;not null" json:"title"`
	Content   string         `gorm:"column:content;type:text;not null;default:''" json:"content"`
	IsActive  bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	SortOrder *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

func (InquiryTemplate) TableName() string { return "inquiry_templates" }
