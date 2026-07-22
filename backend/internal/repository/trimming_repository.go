package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/trimming"
)

// AppointmentTrimmingDetailRepository is a BE9-2E facade; remove with the repository aggregator in BE9-2F.
type AppointmentTrimmingDetailRepository = trimming.AppointmentTrimmingDetailRepository

func NewAppointmentTrimmingDetailRepository(db *gorm.DB) AppointmentTrimmingDetailRepository {
	return trimming.NewAppointmentTrimmingDetailRepository(db)
}
