package model

import (
	"encoding/json"
	"time"
)

type LineCustomer struct {
	ID               uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID         uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	LineUserID       string          `gorm:"column:line_user_id;type:text;not null" json:"line_user_id"`
	DisplayName      string          `gorm:"column:display_name;type:text;not null;default:''" json:"display_name"`
	RealName         string          `gorm:"column:real_name;type:text;not null;default:''" json:"real_name"`
	AdditionalFields json.RawMessage `gorm:"column:additional_fields;type:jsonb;not null;default:'{}'" json:"additional_fields"`
	CreatedAt        time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	FollowedAt       *time.Time      `gorm:"column:followed_at;type:timestamptz" json:"-"`
	BlockedAt        *time.Time      `gorm:"column:blocked_at;type:timestamptz" json:"-"`
	Owner            *Owner          `gorm:"-" json:"owner,omitempty"`
}

func (LineCustomer) TableName() string { return "line_customers" }
