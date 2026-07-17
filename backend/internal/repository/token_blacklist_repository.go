package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/tokenblacklist"
)

// TokenBlacklistRepository is a stable facade alias for the tokenblacklist domain package.
type TokenBlacklistRepository = tokenblacklist.Repository

// NewTokenBlacklistRepository constructs the token blacklist repository.
func NewTokenBlacklistRepository(db *gorm.DB) TokenBlacklistRepository {
	return tokenblacklist.New(db)
}
