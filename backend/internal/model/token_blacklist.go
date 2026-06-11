package model

import "time"

// TokenBlacklist は失効済み refresh_token の JTI を記録する。
// ログアウト時に JTI を登録し、RefreshToken エンドポイントで照合する。
type TokenBlacklist struct {
	JTI       string    `gorm:"primaryKey"   json:"jti"`
	ExpiresAt time.Time `gorm:"not null"     json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (TokenBlacklist) TableName() string { return "token_blacklist" }
