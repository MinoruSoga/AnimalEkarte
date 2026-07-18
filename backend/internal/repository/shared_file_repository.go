package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/sharedfile"
)

// SharedFileRepository is a stable facade alias for the sharedfile domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type SharedFileRepository = sharedfile.Repository

// NewSharedFileRepository constructs the shared file repository.
func NewSharedFileRepository(db *gorm.DB) SharedFileRepository {
	return sharedfile.New(db)
}
