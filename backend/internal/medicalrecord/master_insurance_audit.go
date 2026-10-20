package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MasterInsuranceWriteDependencies keeps an item setting and its durable audit
// in the same owner transaction. Missing dependencies fail before a mutation.
type MasterInsuranceWriteDependencies struct {
	Transactor Transactor
	Audit      AuditTxLogger
}

type masterInsuranceChange struct {
	ID                 uint64
	BeforeExists       bool
	SkipInsuranceAudit bool
	Before             *bool
	After              *bool
}

func masterInsuranceAuditValue(setting *bool) map[string]any {
	var value any
	if setting != nil {
		value = *setting
	}
	return map[string]any{"is_non_insurance": value}
}

func (d MasterInsuranceWriteDependencies) write(
	ctx context.Context, clinicID uint64, actorID *uint64, resource, action string,
	mutate func(context.Context) (masterInsuranceChange, error),
) error {
	if d.Transactor == nil || d.Audit == nil {
		return apperrors.WrapInternalServerError("master insurance transaction and audit dependencies are required")
	}
	if actorID == nil || *actorID == 0 || clinicID == 0 {
		return apperrors.WrapInvalidInput("authenticated clinic and staff actor are required for master insurance setting")
	}
	return d.Transactor.WithTx(ctx, func(txCtx context.Context) error {
		change, err := mutate(txCtx)
		if err != nil {
			return err
		}
		if change.SkipInsuranceAudit {
			return nil
		}
		var before any
		if change.BeforeExists {
			before = masterInsuranceAuditValue(change.Before)
		}
		if err := d.Audit.LogEntryTx(txCtx, &AuditEntry{
			ClinicID: &clinicID, ActorID: actorID, ActorType: model.AuditActorTypeStaff,
			Action: action, Resource: resource, ResourceID: &change.ID,
			OldValue: before, NewValue: masterInsuranceAuditValue(change.After),
			Metadata: map[string]any{"field": "is_non_insurance"},
		}); err != nil {
			return apperrors.Wrap(err, "failed to write master insurance audit")
		}
		return nil
	})
}
