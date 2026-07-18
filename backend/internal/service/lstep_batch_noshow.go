package service

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *lstepBatchService) detectNoShowReservations(ctx context.Context, clinicID uint64) (int, []error) {
	candidates, err := s.reservationRepo.FindNoShowCandidates(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "no-show batch: failed to find candidates", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find no-show candidates")}
	}

	var errs []error
	count := 0
	for i := range candidates {
		r := candidates[i]
		if _, updateErr := s.reservationRepo.Update(ctx, clinicID, r.ID, map[string]any{
			"status": model.ReservationStatusNoShow,
		}); updateErr != nil {
			slog.ErrorContext(ctx, "no-show batch: failed to update status", "reservation_id", r.ID, "error", updateErr)
			errs = append(errs, apperrors.Wrap(updateErr, "failed to update no-show status"))
			continue
		}

		count++
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
