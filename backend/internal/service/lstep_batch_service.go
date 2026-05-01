package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// LstepBatchService はバッチ処理でノーショウ検知・休眠検知を行うサービス（BE-005, BE-014）。
type LstepBatchService interface {
	// DetectNoShowReservations は指定クリニックのノーショウ予約を検知しタグ付与・ステータス更新を行う。
	// 処理件数と個別エラーのスライスを返す（全体は失敗しない）。
	DetectNoShowReservations(ctx context.Context, clinicID uint64) (int, []error)
	// RunNoShowCheckAllClinics は全クリニックに対してノーショウ検知を実行するcronエントリポイント。
	RunNoShowCheckAllClinics(ctx context.Context) error
	// DetectDormantOwners は指定クリニックの休眠飼い主を検知しタグを同期する。
	// 処理件数と個別エラーのスライスを返す（全体は失敗しない）。
	DetectDormantOwners(ctx context.Context, clinicID uint64) (int, []error)
	// RunDormantDetectionAllClinics は全クリニックに対して休眠検知を実行するcronエントリポイント（02:00 JST）。
	RunDormantDetectionAllClinics(ctx context.Context) error
	// RunLTVTopPercentSyncAllClinics は全クリニックに対して LTV 上位 20% タグを同期するcronエントリポイント（FEAT-377）。
	RunLTVTopPercentSyncAllClinics(ctx context.Context) error
	// RunVisitDormantSyncAllClinics は全クリニックに対して VISIT_* タグ（120/180/220/240日超）を同期するcronエントリポイント（FEAT-377）。
	RunVisitDormantSyncAllClinics(ctx context.Context) error
	// RunHealthPreventionTagSyncAllClinics は全クリニックに対して健診・予防・物販タグを同期するcronエントリポイント（FEAT-379）。
	RunHealthPreventionTagSyncAllClinics(ctx context.Context) error
}

type lstepBatchService struct {
	reservationRepo repository.ReservationRepository
	tagSyncSvc      LstepTagSyncService
	clinicRepo      repository.ClinicRepository
	medRecordRepo   repository.MedicalRecordRepository
	auditSvc        AuditService
	settingsSvc     LstepSettingsService
}

// NewLstepBatchService は LstepBatchService を初期化して返す。
func NewLstepBatchService(
	reservationRepo repository.ReservationRepository,
	tagSyncSvc LstepTagSyncService,
	clinicRepo repository.ClinicRepository,
	medRecordRepo repository.MedicalRecordRepository,
	auditSvc AuditService,
	settingsSvc LstepSettingsService,
) LstepBatchService {
	return &lstepBatchService{
		reservationRepo: reservationRepo,
		tagSyncSvc:      tagSyncSvc,
		clinicRepo:      clinicRepo,
		medRecordRepo:   medRecordRepo,
		auditSvc:        auditSvc,
		settingsSvc:     settingsSvc,
	}
}

// DetectNoShowReservations は指定クリニックの no-show 予約を検知してステータス更新・タグ付与を行う。
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

		if r.OwnerID != nil {
			if tagErr := s.tagSyncSvc.SyncNoShowTag(ctx, clinicID, *r.OwnerID, r.StartTime); tagErr != nil {
				slog.ErrorContext(ctx, "no-show batch: failed to sync tag", "reservation_id", r.ID, "owner_id", *r.OwnerID, "error", tagErr)
				errs = append(errs, apperrors.Wrap(tagErr, "failed to sync no-show tag"))
			}
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
			_ = s.auditSvc.LogLstepOperationWithMetadata(ctx, clinic.ID, nil,
				"batch_no_show_detect", "clinic", &clinic.ID,
				map[string]any{
					"operation":       "batch_no_show_detect",
					"processed_count": count,
					"error_count":     len(errs),
				},
			)
		}
	}
	return nil
}

// DetectDormantOwners は指定クリニックの休眠飼い主を検知してタグを同期する（閾値: 180日）。
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

// RunDormantDetectionAllClinics は全クリニックに対して休眠検知を実行する（02:00 JST バッチ）。
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
