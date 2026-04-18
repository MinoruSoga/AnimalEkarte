package model

import "time"

// PasswordResetToken はパスワードリセットトークンを表す。
// token_hash に SHA256(rawToken) を保存し、rawToken はメールURLにのみ含める。
type PasswordResetToken struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID uint64    `gorm:"not null"                 json:"account_id"`
	TokenHash string    `gorm:"not null;uniqueIndex"     json:"-"`
	ExpiresAt time.Time `gorm:"not null"                 json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime"           json:"created_at"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }
