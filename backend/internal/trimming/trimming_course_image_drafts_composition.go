package trimming

import (
	"github.com/animal-ekarte/backend/internal/persistence"
	"gorm.io/gorm"
)

func NewCourseImageDrafts(db *gorm.DB) CourseImageDrafts {
	return &courseImageDrafts{tx: persistence.NewTransactor(db), repo: &courseImageUploadRepository{db: db}, lifecycle: NewAppointmentTrimmingDetailRepository(db).(ImageLifecycleRepository)}
}
