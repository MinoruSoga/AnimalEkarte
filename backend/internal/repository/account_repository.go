package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/account"
)

// AccountRepository is a stable facade alias for the account domain package.
// Service/handler imports keep using repository.* so the domain split does
// not churn all importers.
type AccountRepository = account.Repository

// NewAccountRepository constructs the account repository.
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return account.New(db)
}
