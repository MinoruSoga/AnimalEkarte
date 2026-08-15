package lstep

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

const noShowRuleVersion = "end-time-plus-4h-v1"

func (s *lstepBatchService) detectNoShowReservations(ctx context.Context, clinicID uint64) (int, []error) {
	evaluatedAt := time.Now().UTC()
	if s.nowFn != nil {
		evaluatedAt = s.nowFn().UTC()
	}
	// G2F-08: FindNoShowCandidates is hard-capped (noShowCandidateMax); remaining
	// rows drain on later cycles. Per-row WithTx+audit stays intentional (fail-closed).
	candidates, err := s.reservationRepo.FindNoShowCandidates(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "no-show batch: failed to find candidates", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find no-show candidates")}
	}
	batchRunID := fmt.Sprintf("no-show:%d:%d", clinicID, evaluatedAt.UnixNano())
	return s.applyNoShowCandidates(
		ctx,
		clinicID,
		candidates,
		evaluatedAt,
		batchRunID,
		s.legacyMarkNoShowAt,
	)
}

func (s *lstepBatchService) detectNoShowReservationsAt(
	ctx context.Context,
	clinicID uint64,
	evaluatedAt time.Time,
	batchRunID string,
) (int, []error) {
	repositoryAt, ok := s.reservationRepo.(lstepBatchReservationRepositoryAt)
	if !ok {
		err := apperrors.WrapInternalServerError("no-show scheduled-time repository is not configured")
		slog.ErrorContext(ctx, "no-show batch: scheduled-time repository is not configured", "clinic_id", clinicID)
		return 0, []error{err}
	}
	evaluatedAt = evaluatedAt.UTC()
	candidates, err := repositoryAt.FindNoShowCandidatesAt(ctx, clinicID, evaluatedAt)
	if err != nil {
		slog.ErrorContext(ctx, "no-show batch: failed to find candidates", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find no-show candidates")}
	}
	return s.applyNoShowCandidates(
		ctx,
		clinicID,
		candidates,
		evaluatedAt,
		batchRunID,
		repositoryAt.MarkNoShowAt,
	)
}

type markNoShowAtFunc func(
	ctx context.Context,
	clinicID uint64,
	id uint64,
	evaluatedAt time.Time,
) (reservation.NoShowTransition, error)

func (s *lstepBatchService) applyNoShowCandidates(
	ctx context.Context,
	clinicID uint64,
	candidates []model.Reservation,
	evaluatedAt time.Time,
	batchRunID string,
	markAt markNoShowAtFunc,
) (int, []error) {
	if len(candidates) > 0 && (s.transactor == nil || s.noShowAuditTx == nil) {
		err := apperrors.WrapInternalServerError("no-show transaction or audit dependency is not configured")
		slog.ErrorContext(ctx, "no-show batch: transaction-local audit is not configured", "clinic_id", clinicID)
		return 0, []error{err}
	}

	var errs []error
	count := 0
	for i := range candidates {
		r := candidates[i]
		var transition reservation.NoShowTransition
		txErr := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
			var markErr error
			transition, markErr = markAt(txCtx, clinicID, r.ID, evaluatedAt)
			if markErr != nil {
				return apperrors.Wrap(markErr, "failed to update no-show status")
			}
			if !transition.Changed {
				return nil
			}
			if auditErr := s.noShowAuditTx.LogNoShowTransitionTx(txCtx, &NoShowAuditEntry{
				ClinicID:       clinicID,
				AppointmentID:  r.ID,
				PreviousStatus: transition.PreviousStatus,
				EvaluatedAt:    evaluatedAt,
				RuleVersion:    noShowRuleVersion,
				BatchRunID:     batchRunID,
			}); auditErr != nil {
				return apperrors.Wrap(auditErr, "failed to write no-show transition audit")
			}
			return nil
		})
		if txErr != nil {
			slog.ErrorContext(ctx, "no-show batch: transition transaction failed", "reservation_id", r.ID, "error", txErr)
			errs = append(errs, apperrors.Wrap(txErr, "failed to commit no-show transition"))
			continue
		}
		if transition.Changed {
			count++
		}
	}
	return count, errs
}

func (s *lstepBatchService) legacyMarkNoShowAt(
	ctx context.Context,
	clinicID uint64,
	id uint64,
	_ time.Time,
) (reservation.NoShowTransition, error) {
	return s.reservationRepo.MarkNoShow(ctx, clinicID, id)
}

// RunNoShowCheckAllClinics は全クリニックに対してノーショウ検知を実行する。
// ISSUE-010: 処理件数とエラー件数をメタデータとして永続化する。
func (s *lstepBatchService) RunNoShowCheckAllClinics(ctx context.Context) error {
	return s.runBatchAllClinics(ctx,
		"no-show batch", "no-show batch", "updated reservations", "batch_no_show_detect",
		nil,
		s.detectNoShowReservations,
	)
}

// RunNoShowCheckAllClinicsAt runs the durable no-show job against one immutable
// scheduled instant. Selection, compare-and-set, and audit identity all use it.
func (s *lstepBatchService) RunNoShowCheckAllClinicsAt(
	ctx context.Context,
	scheduledAt time.Time,
	runID string,
) BatchRunResult {
	scheduledAt = scheduledAt.UTC()
	result, _ := s.runBatchAllClinicsWithResult(
		ctx,
		"no-show batch",
		"no-show batch",
		"updated reservations",
		"batch_no_show_detect",
		scheduledBatchMetadata(scheduledAt, runID),
		func(ctx context.Context, clinicID uint64) (int, []error) {
			return s.detectNoShowReservationsAt(ctx, clinicID, scheduledAt, runID)
		},
	)
	return result
}
