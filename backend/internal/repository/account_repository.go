package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/auth"
)

// AccountRepository is a compatibility alias for auth persistence.
type AccountRepository = auth.AccountRepository

// NewAccountRepository constructs the account repository.
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return auth.NewAccountRepository(db)
}
