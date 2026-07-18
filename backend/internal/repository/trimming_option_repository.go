package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/trimmingoption"
)

// TrimmingOptionRepository is a stable facade alias for trimmingoption.
type TrimmingOptionRepository = trimmingoption.Repository

// NewTrimmingOptionRepository constructs the trimming option repository.
func NewTrimmingOptionRepository(db *gorm.DB) TrimmingOptionRepository {
	return trimmingoption.New(db)
}
