package model

import "time"

// TokenBlacklist は失効済み refresh_token の JTI または family marker を記録する。
// family marker は既存の JTI denylist 上で token family 全体を失効させる。
type TokenBlacklist struct {
	JTI       string    `gorm:"primaryKey"   json:"jti"`
	ExpiresAt time.Time `gorm:"not null"     json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (TokenBlacklist) TableName() string { return "token_blacklist" }
