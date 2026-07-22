package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/lstep"
)

// lstepLifecycleAuditTxAdapter translates the lstep-owned audit contract at the
// composition boundary while preserving the ambient transaction context.
type lstepLifecycleAuditTxAdapter struct{ inner AuditTxLogger }

func (a lstepLifecycleAuditTxAdapter) LogEntryTx(ctx context.Context, entry *lstep.LifecycleAuditEntry) error {
	return a.inner.LogEntryTx(ctx, &AuditLogInput{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
	})
}
