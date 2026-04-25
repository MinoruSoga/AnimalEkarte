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
}

type lstepBatchService struct {
	reservationRepo repository.ReservationRepository
	tagSyncSvc      LstepTagSyncService
	clinicRepo      repository.ClinicRepository
	medRecordRepo   repository.MedicalRecordRepository
}

// NewLstepBatchService は LstepBatchService を初期化して返す。
func NewLstepBatchService(
	reservationRepo repository.ReservationRepository,
	tagSyncSvc LstepTagSyncService,
	clinicRepo repository.ClinicRepository,
	medRecordRepo repository.MedicalRecordRepository,
) LstepBatchService {
	return &lstepBatchService{
		reservationRepo: reservationRepo,
		tagSyncSvc:      tagSyncSvc,
		clinicRepo:      clinicRepo,
		medRecordRepo:   medRecordRepo,
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

	for _, clinic := range clinics {
		count, errs := s.DetectNoShowReservations(ctx, clinic.ID)
		if len(errs) > 0 {
			slog.ErrorContext(ctx, "no-show batch: partial errors", "clinic_id", clinic.ID, "error_count", len(errs))
		}
		if count > 0 {
			slog.InfoContext(ctx, "no-show batch: updated reservations", "clinic_id", clinic.ID, "count", count)
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

// RunDormantDetectionAllClinics は全クリニックに対して休眠検知を実行する（02:00 JST バッチ）。
func (s *lstepBatchService) RunDormantDetectionAllClinics(ctx context.Context) error {
	clinics, err := s.clinicRepo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "dormant batch: failed to fetch clinics", "error", err)
		return apperrors.Wrap(err, "failed to fetch clinics for dormant batch")
	}

	for _, clinic := range clinics {
		count, errs := s.DetectDormantOwners(ctx, clinic.ID)
		if len(errs) > 0 {
			slog.ErrorContext(ctx, "dormant batch: partial errors", "clinic_id", clinic.ID, "error_count", len(errs))
		}
		if count > 0 {
			slog.InfoContext(ctx, "dormant batch: synced dormant tags", "clinic_id", clinic.ID, "count", count)
		}
	}
	return nil
}
