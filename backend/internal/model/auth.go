package model

import "time"

// RefreshToken はリフレッシュトークンのDBモデル
type RefreshToken struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64     `gorm:"not null"                 json:"user_id"`
	ClinicID  uint64     `gorm:"not null"                 json:"clinic_id"`
	TokenHash string     `gorm:"not null;uniqueIndex"     json:"-"`
	ExpiresAt time.Time  `gorm:"not null"                 json:"expires_at"`
	RevokedAt *time.Time `                                json:"revoked_at,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime"           json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
