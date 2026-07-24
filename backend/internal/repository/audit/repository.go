// Package audit preserves the legacy repository import path.
package audit

import (
	"gorm.io/gorm"

	domainaudit "github.com/animal-ekarte/backend/internal/audit"
)

// Repository is a compatibility alias for the domain audit repository.
type Repository = domainaudit.Repository

// New delegates construction to the domain audit package.
func New(db *gorm.DB) Repository {
	return domainaudit.NewRepository(db)
}

// MarshalAuditJSON preserves the legacy helper name.
var MarshalAuditJSON = domainaudit.MarshalJSON
