package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func (s *lstepBatchService) RunLTVTopPercentSyncAllClinics(ctx context.Context) error {
	clinics, err := s.clinicRepo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "ltv-top-percent batch: failed to fetch clinics", "error", err)
		return apperrors.Wrap(err, "failed to fetch clinics for ltv-top-percent batch")
	}

	for i := range clinics {
		clinic := &clinics[i]
		if s.settingsSvc != nil {
			enabled, syncErr := s.settingsSvc.IsSyncEnabled(ctx, clinic.ID)
			if syncErr != nil {
				slog.ErrorContext(ctx, "ltv-top-percent batch: failed to check sync enabled", "clinic_id", clinic.ID, "error", syncErr)
				continue
			}
			if !enabled {
				continue
			}
		}
		count, errs := s.tagSyncSvc.SyncLTVTopPercent(ctx, clinic.ID)
		if len(errs) > 0 {
			slog.ErrorContext(ctx, "ltv-top-percent batch: partial errors", "clinic_id", clinic.ID, "error_count", len(errs))
		}
		if count > 0 {
			slog.InfoContext(ctx, "ltv-top-percent batch: synced ltv tags", "clinic_id", clinic.ID, "count", count)
			_ = s.auditSvc.LogLstepOperationWithMetadata(ctx, clinic.ID, nil,
				"batch_ltv_top_percent", "clinic", &clinic.ID,
				map[string]any{
					"operation":       "batch_ltv_top_percent",
					"processed_count": count,
					"error_count":     len(errs),
				},
			)
		}
	}
	return nil
}

// RunVisitDormantSyncAllClinics は全クリニックに対して VISIT_* タグを同期する（FEAT-377）。
func (s *lstepBatchService) RunVisitDormantSyncAllClinics(ctx context.Context) error {
	const minDaysSince = 120
	clinics, err := s.clinicRepo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "visit-dormant batch: failed to fetch clinics", "error", err)
		return apperrors.Wrap(err, "failed to fetch clinics for visit-dormant batch")
	}

	for i := range clinics {
		clinic := &clinics[i]
		if s.settingsSvc != nil {
			enabled, syncErr := s.settingsSvc.IsSyncEnabled(ctx, clinic.ID)
			if syncErr != nil {
				slog.ErrorContext(ctx, "visit-dormant batch: failed to check sync enabled", "clinic_id", clinic.ID, "error", syncErr)
				continue
			}
			if !enabled {
				continue
			}
		}
		entries, findErr := s.medRecordRepo.FindDormantOwnerEntries(ctx, clinic.ID, minDaysSince)
		if findErr != nil {
			slog.ErrorContext(ctx, "visit-dormant batch: failed to find entries", "clinic_id", clinic.ID, "error", findErr)
			continue
		}
		count := 0
		var errs []error
		for _, entry := range entries {
			if tagErr := s.tagSyncSvc.SyncVisitDormantTags(ctx, clinic.ID, entry.OwnerID, entry.DaysSince); tagErr != nil {
				slog.ErrorContext(ctx, "visit-dormant batch: failed to sync tag", "clinic_id", clinic.ID, "owner_id", entry.OwnerID, "error", tagErr)
				errs = append(errs, apperrors.Wrap(tagErr, "failed to sync visit dormant tag"))
				continue
			}
			count++
		}
		if len(errs) > 0 {
			slog.ErrorContext(ctx, "visit-dormant batch: partial errors", "clinic_id", clinic.ID, "error_count", len(errs))
		}
		if count > 0 {
			slog.InfoContext(ctx, "visit-dormant batch: synced visit tags", "clinic_id", clinic.ID, "count", count)
			_ = s.auditSvc.LogLstepOperationWithMetadata(ctx, clinic.ID, nil,
				"batch_visit_dormant", "clinic", &clinic.ID,
				map[string]any{
					"operation":       "batch_visit_dormant",
					"processed_count": count,
					"error_count":     len(errs),
				},
			)
		}
	}
	return nil
}

// RunHealthPreventionTagSyncAllClinics は全クリニックに対して健診・予防・物販タグを同期する（FEAT-379）。
func (s *lstepBatchService) RunHealthPreventionTagSyncAllClinics(ctx context.Context) error {
	clinics, err := s.clinicRepo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "health-prevention batch: failed to fetch clinics", "error", err)
		return apperrors.Wrap(err, "failed to fetch clinics for health-prevention batch")
	}

	for i := range clinics {
		clinic := &clinics[i]
		if s.settingsSvc != nil {
			enabled, syncErr := s.settingsSvc.IsSyncEnabled(ctx, clinic.ID)
			if syncErr != nil {
				slog.ErrorContext(ctx, "health-prevention batch: failed to check sync enabled", "clinic_id", clinic.ID, "error", syncErr)
				continue
			}
			if !enabled {
				continue
			}
		}
		count, errs := s.tagSyncSvc.SyncHealthPreventionTagsForClinic(ctx, clinic.ID)
		if len(errs) > 0 {
			slog.ErrorContext(ctx, "health-prevention batch: partial errors", "clinic_id", clinic.ID, "error_count", len(errs))
		}
		if count > 0 {
			slog.InfoContext(ctx, "health-prevention batch: synced tags", "clinic_id", clinic.ID, "count", count)
			_ = s.auditSvc.LogLstepOperationWithMetadata(ctx, clinic.ID, nil,
				"batch_health_prevention", "clinic", &clinic.ID,
				map[string]any{
					"operation":       "batch_health_prevention",
					"processed_count": count,
					"error_count":     len(errs),
				},
			)
		}
	}
	return nil
}

// runDeliveryTriggersForClinic は 1 クリニック分の全配信トリガーバッチを実行する（FEAT-383）。
