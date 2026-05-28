package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func (s *lstepBatchService) DetectDormantOwners(ctx context.Context, clinicID uint64) (int, []error) {
	const minDaysSince = 180
	entries, err := s.medRecordRepo.FindDormantOwnerEntries(ctx, clinicID, minDaysSince)
	if err != nil {
		slog.ErrorContext(ctx, "dormant batch: failed to find dormant owners", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find dormant owners")}
	}

	var errs []error
	count := 0
	for _, entry := range entries {
		if tagErr := s.tagSyncSvc.SyncDormantTags(ctx, clinicID, entry.OwnerID, entry.DaysSince); tagErr != nil {
			slog.ErrorContext(ctx, "dormant batch: failed to sync dormant tag", "clinic_id", clinicID, "owner_id", entry.OwnerID, "error", tagErr)
			errs = append(errs, apperrors.Wrap(tagErr, "failed to sync dormant tag"))
			continue
		}
		count++
	}
	return count, errs
}

// RunLTVTopPercentSyncAllClinics は全クリニックに対して LTV 上位 20% タグを同期する（FEAT-377）。

func (s *lstepBatchService) RunDormantDetectionAllClinics(ctx context.Context) error {
	clinics, err := s.clinicRepo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "dormant batch: failed to fetch clinics", "error", err)
		return apperrors.Wrap(err, "failed to fetch clinics for dormant batch")
	}

	for i := range clinics {
		clinic := &clinics[i]
		if s.settingsSvc != nil {
			enabled, syncErr := s.settingsSvc.IsSyncEnabled(ctx, clinic.ID)
			if syncErr != nil {
				slog.ErrorContext(ctx, "dormant batch: failed to check sync enabled", "clinic_id", clinic.ID, "error", syncErr)
				continue
			}
			if !enabled {
				continue
			}
		}
		count, errs := s.DetectDormantOwners(ctx, clinic.ID)
		if len(errs) > 0 {
			slog.ErrorContext(ctx, "dormant batch: partial errors", "clinic_id", clinic.ID, "error_count", len(errs))
		}
		if count > 0 {
			slog.InfoContext(ctx, "dormant batch: synced dormant tags", "clinic_id", clinic.ID, "count", count)
			// ISSUE-010: 処理件数・エラー件数・閾値を永続化する。閾値は後から判定基準を再現できるよう含める。
			_ = s.auditSvc.LogLstepOperationWithMetadata(ctx, clinic.ID, nil,
				"batch_dormant_detect", "clinic", &clinic.ID,
				map[string]any{
					"operation":       "batch_dormant_detect",
					"processed_count": count,
					"error_count":     len(errs),
					"min_days_since":  180,
				},
			)
		}
	}
	return nil
}
