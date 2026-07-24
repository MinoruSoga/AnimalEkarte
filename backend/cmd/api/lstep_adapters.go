package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	appcrypto "github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/service"
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

type lstepLifecycleAuditTxAdapter struct{ inner service.AuditTxLogger }

func (a lstepLifecycleAuditTxAdapter) LogEntryTx(ctx context.Context, entry *lstep.LifecycleAuditEntry) error {
	return a.inner.LogEntryTx(ctx, &service.AuditLogInput{
		ClinicID: entry.ClinicID, ActorID: entry.ActorID, ActorType: entry.ActorType,
		Action: entry.Action, Resource: entry.Resource, ResourceID: entry.ResourceID,
	})
}

type lstepNoShowAuditTxAdapter struct{ inner service.AuditTxLogger }

func (a lstepNoShowAuditTxAdapter) LogNoShowTransitionTx(ctx context.Context, entry *lstep.NoShowAuditEntry) error {
	return a.inner.LogEntryTx(ctx, &service.AuditLogInput{
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

type ownerLifecycleWriterAdapter struct{ inner owner.LstepRepository }

func (a ownerLifecycleWriterAdapter) RecordLstepOptOut(ctx context.Context, clinicID, ownerID uint64, at time.Time, reason string) error {
	return a.inner.RecordLstepOptOut(ctx, clinicID, ownerID, at, reason)
}

func (a ownerLifecycleWriterAdapter) ClearLstepOptOut(ctx context.Context, clinicID, ownerID uint64) error {
	return a.inner.ClearLstepOptOut(ctx, clinicID, ownerID)
}

type petLifecycleWriterAdapter struct{ inner repository.PetRepository }

func (a petLifecycleWriterAdapter) RecordDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error {
	return a.inner.Update(ctx, clinicID, petID, map[string]any{
		"deceased_at": deceasedAt, "deceased_reason": reason, "status": model.PetStatusDeceased,
	})
}

func (a petLifecycleWriterAdapter) ClearDeath(ctx context.Context, clinicID, petID uint64) error {
	return a.inner.Update(ctx, clinicID, petID, map[string]any{
		"deceased_at": nil, "deceased_reason": nil, "status": model.PetStatusAlive,
	})
}

func newLegacyLstepDependencies(app *lstep.Application, cipher *appcrypto.AESGCMCipher) *service.LegacyLstepDependencies {
	return &service.LegacyLstepDependencies{
		TagSync:       app.TagSync,
		LineCustomers: app.LineCustomers,
		EncryptCredential: func(value string) (string, error) {
			return lstep.EncryptLineCredential(cipher, value)
		},
		DecryptCredential: func(ctx context.Context, value string) string {
			return lstep.DecryptLineCredential(ctx, cipher, value)
		},
		NewLinePusher: func(channelToken string) reservation.LinePusher {
			return lstep.NewLineMessagingService(channelToken)
		},
	}
}

func adaptPermissionAny(require func(...struct{ Resource, Action string }) gin.HandlerFunc) lstep.PermissionAnyMiddleware {
	return func(requirements ...lstep.PermissionRequirement) gin.HandlerFunc {
		legacy := make([]struct{ Resource, Action string }, len(requirements))
		for i, requirement := range requirements {
			legacy[i] = struct{ Resource, Action string }{
				Resource: requirement.Resource,
				Action:   requirement.Action,
			}
		}
		return require(legacy...)
	}
}
