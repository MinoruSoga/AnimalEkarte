package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/trimming"
)

// TrimmingOptionRepository is a BE9-2E facade; remove with the repository aggregator in BE9-2F.
type TrimmingOptionRepository = trimming.TrimmingOptionRepository

// NewTrimmingOptionRepository constructs the trimming option repository.
func NewTrimmingOptionRepository(db *gorm.DB) TrimmingOptionRepository {
	return trimming.NewTrimmingOptionRepository(db)
}
