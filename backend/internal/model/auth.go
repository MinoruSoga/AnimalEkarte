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

// IsExpired はトークンが期限切れかどうかを返す
func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsRevoked はトークンが無効化されているかどうかを返す
func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// IsValid はトークンが有効（期限内かつ未無効化）かどうかを返す
func (t *RefreshToken) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked()
}
