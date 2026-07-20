package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// InquiryRepository is a stable facade alias for medicalrecord.InquiryRepository (BE9-2D
// roll-up: the implementation moved from internal/repository/inquiry_repository.go into
// internal/medicalrecord). Kept because medical_record_service.go (BE9-2D
// sub-batch(4) scope, staying in internal/service) and the repositories.go aggregator still depend on it.
type InquiryRepository = medicalrecord.InquiryRepository

// NewInquiryRepository constructs the inquiry repository.
func NewInquiryRepository(db *gorm.DB) InquiryRepository {
	return medicalrecord.NewInquiryRepository(db)
}
