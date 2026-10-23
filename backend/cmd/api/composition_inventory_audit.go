package main

import (
	"context"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/inventory"
)

// inventoryAuditTxAdapter は inventory の consumer-side 監査 port（inventory.AuditTxLogger）を
// 共有監査カーネル audit.TxLogger へ field-for-field で adapt する（trimmingAuditTxAdapter と同じ先例）。
// EMR-449: is_bill_discount_excluded の監査は業務書込と同じ ambient transaction 経路で fail-closed。
type inventoryAuditTxAdapter struct {
	inner audit.TxLogger
}

func (a inventoryAuditTxAdapter) LogEntryTx(
	ctx context.Context,
	entry *inventory.AuditEntry,
) error {
	return a.inner.LogEntryTx(ctx, &audit.Entry{
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
