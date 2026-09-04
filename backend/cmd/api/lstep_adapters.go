package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/pet"
)

type lstepAggregationRepositoryAdapter struct {
	inner owner.LtvRepository
}

func (a lstepAggregationRepositoryAdapter) FindOwnerLTV(ctx context.Context, params *lstep.FindOwnerLTVParams) ([]lstep.OwnerLTVRow, error) {
	rows, err := a.inner.FindOwnerLTV(ctx, toOwnerLTVParams(params))
	if err != nil {
		return nil, err
	}
	result := make([]lstep.OwnerLTVRow, len(rows))
	for i := range rows {
		result[i] = toLstepLTVRow(&rows[i])
	}
	return result, nil
}

func toOwnerLTVParams(params *lstep.FindOwnerLTVParams) *owner.FindOwnerLTVParams {
	return &owner.FindOwnerLTVParams{
		ClinicID: params.ClinicID, Sort: params.Sort,
		MinTotalAmount: params.MinTotalAmount, MaxTotalAmount: params.MaxTotalAmount,
		MinVisitCount: params.MinVisitCount, LineLinked: params.LineLinked,
		Year: params.Year, From: params.From, To: params.To, AmountBasis: params.AmountBasis,
		IncludeZero: params.IncludeZero, Search: params.Search, PeriodPreset: params.PeriodPreset,
		MaxVisitCount: params.MaxVisitCount, LastVisitBucket: params.LastVisitBucket,
		IncludeNoVisit: params.IncludeNoVisit, Order: params.Order,
	}
}

func toLstepLTVRow(row *owner.OwnerLTVRow) lstep.OwnerLTVRow {
	return lstep.OwnerLTVRow{
		OwnerID: row.OwnerID, OwnerName: row.OwnerName, LineUserID: row.LineUserID,
		LstepOptOut: row.LstepOptOut, TotalAmount: row.TotalAmount,
		TotalVisitCount: row.TotalVisitCount, AnnualVisitCount: row.AnnualVisitCount,
		LastVisitDate: row.LastVisitDate, FirstVisitDate: row.FirstVisitDate,
		AnnualAmount: row.AnnualAmount, BillingCount: row.BillingCount,
		PeriodVisitCount: row.PeriodVisitCount, DaysSinceLastVisit: row.DaysSinceLastVisit,
		LastVisitBucket: row.LastVisitBucket, MaxSingleVisitAmount: row.MaxSingleVisitAmount,
	}
}

type lstepLifecycleAuditTxAdapter struct{ inner audit.TxLogger }

func (a lstepLifecycleAuditTxAdapter) LogEntryTx(ctx context.Context, entry *lstep.LifecycleAuditEntry) error {
	return a.inner.LogEntryTx(ctx, &audit.Entry{
		ClinicID: entry.ClinicID, ActorID: entry.ActorID, ActorType: entry.ActorType,
		Action: entry.Action, Resource: entry.Resource, ResourceID: entry.ResourceID,
	})
}

type lstepNoShowAuditTxAdapter struct{ inner audit.TxLogger }

func (a lstepNoShowAuditTxAdapter) LogNoShowTransitionTx(ctx context.Context, entry *lstep.NoShowAuditEntry) error {
	return a.inner.LogEntryTx(ctx, &audit.Entry{
		ClinicID: &entry.ClinicID, ActorType: model.AuditActorTypeSystem,
		Action: model.AuditActionReservationNoShow, Resource: model.AuditResourceReservation,
		ResourceID: &entry.AppointmentID,
		OldValue:   map[string]any{"status": entry.PreviousStatus},
		NewValue:   map[string]any{"status": model.ReservationStatusNoShow},
		Metadata: map[string]any{
			"evaluated_at": entry.EvaluatedAt,
			"rule_version": entry.RuleVersion,
			"batch_run_id": entry.BatchRunID,
		},
	})
}

type lstepLineLinkAuditTxAdapter struct{ inner audit.TxLogger }

func (a lstepLineLinkAuditTxAdapter) LogOwnerLineLinkTx(
	ctx context.Context,
	clinicID, ownerID uint64,
) error {
	return a.inner.LogEntryTx(ctx, &audit.Entry{
		ClinicID:   &clinicID,
		ActorType:  model.AuditActorTypeSystem,
		Action:     model.AuditActionOwnerLineUserIDUpdate,
		Resource:   "owner",
		ResourceID: &ownerID,
		OldValue:   map[string]any{"linked": false},
		NewValue:   map[string]any{"linked": true},
	})
}

type ownerLifecycleWriterAdapter struct {
	inner owner.LifecycleOwnerRepository
}

func (a ownerLifecycleWriterAdapter) RecordLstepOptOut(ctx context.Context, clinicID, ownerID uint64, at time.Time, reason string) error {
	return a.inner.RecordLstepOptOut(ctx, clinicID, ownerID, at, reason)
}

func (a ownerLifecycleWriterAdapter) ClearLstepOptOut(ctx context.Context, clinicID, ownerID uint64) error {
	return a.inner.ClearLstepOptOut(ctx, clinicID, ownerID)
}

// petLifecycleWriterAdapter bridges pet.CompleteRepository to lstep.PetLifecycleWriter.
// CMD-04: call the typed LifecycleWriter surface only — never map[string]any Update.
// pet.CompleteRepository already implements the same methods; this adapter remains a
// thin composition-side type assertion until composition can pass PetLifecycle directly.
type petLifecycleWriterAdapter struct{ inner pet.LifecycleWriter }

func (a petLifecycleWriterAdapter) RecordDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error {
	return a.inner.RecordDeath(ctx, clinicID, petID, deceasedAt, reason)
}

func (a petLifecycleWriterAdapter) ClearDeath(ctx context.Context, clinicID, petID uint64) error {
	return a.inner.ClearDeath(ctx, clinicID, petID)
}

func adaptPermissionAny(
	require func(...auth.PermissionRequirement) gin.HandlerFunc,
) lstep.PermissionAnyMiddleware {
	return func(requirements ...lstep.PermissionRequirement) gin.HandlerFunc {
		target := make([]auth.PermissionRequirement, len(requirements))
		for i, requirement := range requirements {
			target[i] = auth.PermissionRequirement{
				Resource: requirement.Resource,
				Action:   requirement.Action,
			}
		}
		return require(target...)
	}
}
