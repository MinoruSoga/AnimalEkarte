package model

import (
	"time"
)

// TokenBlacklist は失効済み refresh_token の JTI または family marker を記録する。
// family marker は既存の JTI denylist 上で token family 全体を失効させる。
type TokenBlacklist struct {
	JTI       string    `gorm:"column:jti;type:text;primaryKey;autoIncrement:false;not null" json:"jti"`
	ExpiresAt time.Time `gorm:"column:expires_at;type:timestamptz;not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
}

func (TokenBlacklist) TableName() string { return "token_blacklist" }
