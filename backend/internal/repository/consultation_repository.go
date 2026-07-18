package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/consultation"
)

// ConsultationRepository is a stable facade alias for the consultation domain package.
type ConsultationRepository = consultation.Repository

// NewConsultationRepository constructs the consultation repository.
func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
	return consultation.New(db)
}
