package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// lstepNoShowAuditTxAdapter maps the lstep-owned semantic event to the shared audit schema while
// preserving the ambient transaction context used by the appointment transition.
type lstepNoShowAuditTxAdapter struct{ inner AuditTxLogger }

func (a lstepNoShowAuditTxAdapter) LogNoShowTransitionTx(ctx context.Context, entry *lstep.NoShowAuditEntry) error {
	return a.inner.LogEntryTx(ctx, &AuditLogInput{
		ClinicID:   &entry.ClinicID,
		ActorType:  model.AuditActorTypeSystem,
		Action:     model.AuditActionReservationNoShow,
		Resource:   model.AuditResourceReservation,
		ResourceID: &entry.AppointmentID,
		OldValue: map[string]any{
			"status": entry.PreviousStatus,
		},
		NewValue: map[string]any{
			"status": model.ReservationStatusNoShow,
		},
		Metadata: map[string]any{
			"evaluated_at": entry.EvaluatedAt,
			"rule_version": entry.RuleVersion,
			"batch_run_id": entry.BatchRunID,
		},
	})
}
