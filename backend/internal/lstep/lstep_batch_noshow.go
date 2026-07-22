package lstep

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/reservation"
)

const noShowRuleVersion = "end-time-plus-4h-v1"

func (s *lstepBatchService) detectNoShowReservations(ctx context.Context, clinicID uint64) (int, []error) {
	candidates, err := s.reservationRepo.FindNoShowCandidates(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "no-show batch: failed to find candidates", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find no-show candidates")}
	}
	if len(candidates) > 0 && (s.transactor == nil || s.noShowAuditTx == nil) {
		err := apperrors.WrapInternalServerError("no-show transaction or audit dependency is not configured")
		slog.ErrorContext(ctx, "no-show batch: transaction-local audit is not configured", "clinic_id", clinicID)
		return 0, []error{err}
	}

	var errs []error
	count := 0
	evaluatedAt := time.Now().UTC()
	if s.nowFn != nil {
		evaluatedAt = s.nowFn().UTC()
	}
	batchRunID := fmt.Sprintf("no-show:%d:%d", clinicID, evaluatedAt.UnixNano())
	for i := range candidates {
		r := candidates[i]
		var transition reservation.NoShowTransition
		txErr := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
			var markErr error
			transition, markErr = s.reservationRepo.MarkNoShow(txCtx, clinicID, r.ID)
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

// RunNoShowCheckAllClinics は全クリニックに対してノーショウ検知を実行する。
// ISSUE-010: 処理件数とエラー件数をメタデータとして永続化する。
func (s *lstepBatchService) RunNoShowCheckAllClinics(ctx context.Context) error {
	return s.runBatchAllClinics(ctx,
		"no-show batch", "no-show batch", "updated reservations", "batch_no_show_detect",
		nil,
		s.detectNoShowReservations,
	)
}
