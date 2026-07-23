package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/trimming"
)

type trimmingAuditEntry = trimming.AuditEntry

type trimmingAuditTxAdapter struct {
	inner AuditTxLogger
}

func (a trimmingAuditTxAdapter) LogEntryTx(ctx context.Context, entry *trimmingAuditEntry) error {
	return a.inner.LogEntryTx(ctx, &AuditLogInput{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		Metadata:   entry.Metadata,
	})
}
