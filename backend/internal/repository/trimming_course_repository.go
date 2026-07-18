package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/trimmingcourse"
)

// TrimmingCourseRepository is a facade alias to trimmingcourse.Repository (BE8-4 batch26).
type TrimmingCourseRepository = trimmingcourse.Repository

// NewTrimmingCourseRepository constructs a TrimmingCourseRepository.
func NewTrimmingCourseRepository(db *gorm.DB) TrimmingCourseRepository {
	return trimmingcourse.New(db)
}
