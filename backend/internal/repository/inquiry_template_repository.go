package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// InquiryTemplateRepository is a stable facade alias for medicalrecord.InquiryTemplateRepository
// (BE9-2D roll-up of the former inquirytemplate subpackage). Kept only because the repositories.go aggregator still constructs it and
// cmd/api/main.go passes that instance into the medicalrecord constructors; no other production
// consumer remains. Delete when main.go switches to calling medicalrecord.New* directly (BE9-2F).
type InquiryTemplateRepository = medicalrecord.InquiryTemplateRepository

// NewInquiryTemplateRepository constructs the inquiry template repository.
func NewInquiryTemplateRepository(db *gorm.DB) InquiryTemplateRepository {
	return medicalrecord.NewInquiryTemplateRepository(db)
}
