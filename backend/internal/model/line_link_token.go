package model

import (
	"time"
)

// LineLinkToken は LINE User ID 紐付け用の一時トークン（BE-021）。
type LineLinkToken struct {
	ID        uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	OwnerID   uint64     `gorm:"column:owner_id;type:bigint;not null" json:"owner_id"`
	Token     string     `gorm:"column:token;type:varchar(64);not null;uniqueIndex:line_link_tokens_token_key" json:"-"`
	ExpiresAt time.Time  `gorm:"column:expires_at;type:timestamptz;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at;type:timestamptz" json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
}

func (LineLinkToken) TableName() string { return "line_link_tokens" }
