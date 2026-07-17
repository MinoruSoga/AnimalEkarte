package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/trimmingcoursetype"
)

// TrimmingCourseTypeRepository is a stable facade alias for trimmingcoursetype.
type TrimmingCourseTypeRepository = trimmingcoursetype.Repository

// NewTrimmingCourseTypeRepository constructs the trimming course type repository.
func NewTrimmingCourseTypeRepository(db *gorm.DB) TrimmingCourseTypeRepository {
	return trimmingcoursetype.New(db)
}
