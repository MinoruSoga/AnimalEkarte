package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/inquirytemplate"
)

// InquiryTemplateRepository is a stable facade alias for the inquirytemplate domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type InquiryTemplateRepository = inquirytemplate.Repository

// NewInquiryTemplateRepository constructs the inquiry template repository.
func NewInquiryTemplateRepository(db *gorm.DB) InquiryTemplateRepository {
	return inquirytemplate.New(db)
}
