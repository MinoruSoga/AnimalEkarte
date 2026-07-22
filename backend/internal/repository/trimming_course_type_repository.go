package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/trimming"
)

// TrimmingCourseTypeRepository is a BE9-2E facade; remove with the repository aggregator in BE9-2F.
type TrimmingCourseTypeRepository = trimming.TrimmingCourseTypeRepository

// NewTrimmingCourseTypeRepository constructs the trimming course type repository.
func NewTrimmingCourseTypeRepository(db *gorm.DB) TrimmingCourseTypeRepository {
	return trimming.NewTrimmingCourseTypeRepository(db)
}
