package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/audit"
)

// AuditRepository is a stable facade alias for the audit domain package.
// Service/handler imports keep using repository.* so the domain split does
// not churn all importers.
type AuditRepository = audit.Repository

// NewAuditRepository constructs the audit repository.
func NewAuditRepository(db *gorm.DB) AuditRepository {
	return audit.New(db)
}

// MarshalAuditJSON re-exports audit.MarshalAuditJSON so existing callers in
// package repository (and its tests) do not need an import path change.
var MarshalAuditJSON = audit.MarshalAuditJSON
