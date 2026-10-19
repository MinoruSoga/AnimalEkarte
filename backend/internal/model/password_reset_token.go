package model

import (
	"time"
)

// PasswordResetToken はパスワードリセットトークンを表す。
// token_hash に SHA256(rawToken) を保存し、rawToken はメールURLにのみ含める。
type PasswordResetToken struct {
	ID        uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	AccountID uint64    `gorm:"column:account_id;type:bigint;not null" json:"account_id"`
	TokenHash string    `gorm:"column:token_hash;type:text;not null" json:"-"`
	ExpiresAt time.Time `gorm:"column:expires_at;type:timestamptz;not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }
