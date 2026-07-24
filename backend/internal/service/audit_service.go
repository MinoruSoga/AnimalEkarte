package service

import (
	domainaudit "github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// Compatibility aliases keep the former service boundary source-compatible
// while the implementation is owned by internal/audit.
type AuditService = domainaudit.Service
type AuditTxLogger = domainaudit.TxLogger
type AuditKernel = domainaudit.Kernel
type AuditLogInput = domainaudit.Entry

// NewAuditService delegates construction to the audit domain.
func NewAuditService(repo repository.AuditRepository) AuditKernel {
	return domainaudit.NewService(repo)
}

// validateAuditLog preserves package-local legacy tests during the cutover.
func validateAuditLog(log *model.AuditLog) error {
	return domainaudit.ValidateLog(log)
}

func auditActorTypeFor(actorID *uint64) string {
	return sharedkernel.AuditActorTypeFor(actorID)
}
