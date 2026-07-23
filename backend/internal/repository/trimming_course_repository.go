package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/trimming"
)

// TrimmingCourseRepository is a BE9-2E facade; remove with the repository aggregator in BE9-2F.
type TrimmingCourseRepository = trimming.TrimmingCourseRepository

// NewTrimmingCourseRepository constructs a TrimmingCourseRepository.
func NewTrimmingCourseRepository(db *gorm.DB) TrimmingCourseRepository {
	return trimming.NewTrimmingCourseRepository(db)
}
