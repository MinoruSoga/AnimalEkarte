package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/auth"
)

// TokenBlacklistRepository is a compatibility alias for auth persistence.
type TokenBlacklistRepository = auth.TokenBlacklistRepository

// NewTokenBlacklistRepository constructs the token blacklist repository.
func NewTokenBlacklistRepository(db *gorm.DB) TokenBlacklistRepository {
	return auth.NewTokenBlacklistRepository(db)
}
