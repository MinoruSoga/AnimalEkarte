package service

import (
	"context"
)

// billingAuditTxAdapter は AuditTxLogger.LogEntryTx(ctx, *AuditLogInput) を billing の
// consumer-side view（LogEntryTx(ctx, *billing.AuditEntry)）へ写像する（#211 fail-closed・
// medicalrecord の tx adapter 先例・BE9-2C B③）。
type billingAuditTxAdapter struct {
	inner AuditTxLogger
}

func (a billingAuditTxAdapter) LogEntryTx(ctx context.Context, e *billingAuditEntry) error {
	return a.inner.LogEntryTx(ctx, &AuditLogInput{
		ClinicID:   e.ClinicID,
		ActorID:    e.ActorID,
		ActorType:  e.ActorType,
		Action:     e.Action,
		Resource:   e.Resource,
		ResourceID: e.ResourceID,
		OldValue:   e.OldValue,
		NewValue:   e.NewValue,
		Metadata:   e.Metadata,
	})
}
