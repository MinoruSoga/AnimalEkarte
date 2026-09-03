package model

import "time"

// LineLinkToken は LINE User ID 紐付け用の一時トークン（BE-021）。
type LineLinkToken struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID  uint64     `gorm:"not null"                 json:"clinic_id"`
	OwnerID   uint64     `gorm:"not null"                 json:"owner_id"`
	Token     string     `gorm:"not null;uniqueIndex"     json:"-"`
	ExpiresAt time.Time  `gorm:"not null"                 json:"expires_at"`
	UsedAt    *time.Time `                                json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime"           json:"created_at"`
}

func (LineLinkToken) TableName() string { return "line_link_tokens" }
