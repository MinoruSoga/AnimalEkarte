package service

import (
	"context"
)

// billingAuditAdapter は AuditService.LogEntry(ctx, *AuditLogInput) を
// billing の consumer-side view（LogEntry(ctx, *billing.AuditEntry)）へ写像する
// （medicalrecord の adapter 先例・BE9-2C B②）。
type billingAuditAdapter struct {
	inner AuditService
}

func (a billingAuditAdapter) LogEntry(ctx context.Context, e *billingAuditEntry) error {
	return a.inner.LogEntry(ctx, &AuditLogInput{
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
