package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *lstepBatchService) DetectNoShowReservations(ctx context.Context, clinicID uint64) (int, []error) {
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
func (s *lstepBatchService) RunNoShowCheckAllClinics(ctx context.Context) error {
	clinics, err := s.clinicRepo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "no-show batch: failed to fetch clinics", "error", err)
		return apperrors.Wrap(err, "failed to fetch clinics for no-show batch")
	}

	for i := range clinics {
		clinic := &clinics[i]
		if s.settingsSvc != nil {
			enabled, syncErr := s.settingsSvc.IsSyncEnabled(ctx, clinic.ID)
			if syncErr != nil {
				slog.ErrorContext(ctx, "no-show batch: failed to check sync enabled", "clinic_id", clinic.ID, "error", syncErr)
				continue
			}
			if !enabled {
				continue
			}
		}
		count, errs := s.DetectNoShowReservations(ctx, clinic.ID)
		if len(errs) > 0 {
			slog.ErrorContext(ctx, "no-show batch: partial errors", "clinic_id", clinic.ID, "error_count", len(errs))
		}
		if count > 0 {
			slog.InfoContext(ctx, "no-show batch: updated reservations", "clinic_id", clinic.ID, "count", count)
			// ISSUE-010: 処理件数とエラー件数をメタデータとして永続化する。
			if err := s.auditSvc.LogLstepOperationWithMetadata(ctx, clinic.ID, nil,
				"batch_no_show_detect", "clinic", &clinic.ID,
				map[string]any{
					"operation":       "batch_no_show_detect",
					"processed_count": count,
					"error_count":     len(errs),
				},
			); err != nil {
				slog.WarnContext(ctx, "audit log failed for no-show batch", "error", err, "clinic_id", clinic.ID)
			}
		}
	}
	return nil
}

// DetectDormantOwners は指定クリニックの休眠飼い主を検知してタグを同期する（閾値: 180日）。
