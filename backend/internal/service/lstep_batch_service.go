package service

import (
	"context"
	"log/slog"
	"maps"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/repository"
)

// deliveryTriggerHourJST は仕様 §6.4 で固定された Lステップ自動配信バッチの実行時刻 (10:00 JST)。
// 設定可能化は意図的に廃止 (configurable fire hour 削除)。
const deliveryTriggerHourJST = 10

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
	// RunVisitDormantSyncAllClinics は全クリニックに対して VISIT_* タグ（180/210/240日超）を同期するcronエントリポイント（FEAT-377）。
	RunVisitDormantSyncAllClinics(ctx context.Context) error
	// RunHealthPreventionTagSyncAllClinics は全クリニックに対して健診・予防・物販タグを同期するcronエントリポイント（FEAT-379）。
	RunHealthPreventionTagSyncAllClinics(ctx context.Context) error
	// RunDeliveryTriggerBatchAllClinics は全クリニックに対して自動配信トリガーバッチを実行するcronエントリポイント（FEAT-383: 10:00 JST）。
	RunDeliveryTriggerBatchAllClinics(ctx context.Context) error
}

type lstepBatchService struct {
	reservationRepo      repository.ReservationRepository
	tagSyncSvc           LstepTagSyncService
	clinicRepo           repository.ClinicRepository
	medRecordRepo        repository.MedicalRecordRepository
	auditSvc             AuditService
	settingsSvc          LstepSettingsService
	lstepDeliveryTrigger LstepDeliveryTriggerService
	nowFn                func() time.Time
}

// NewLstepBatchService は LstepBatchService を初期化して返す。
func NewLstepBatchService(
	reservationRepo repository.ReservationRepository,
	tagSyncSvc LstepTagSyncService,
	clinicRepo repository.ClinicRepository,
	medRecordRepo repository.MedicalRecordRepository,
	auditSvc AuditService,
	settingsSvc LstepSettingsService,
	lstepDeliveryTrigger LstepDeliveryTriggerService,
) LstepBatchService {
	return &lstepBatchService{
		reservationRepo:      reservationRepo,
		tagSyncSvc:           tagSyncSvc,
		clinicRepo:           clinicRepo,
		medRecordRepo:        medRecordRepo,
		auditSvc:             auditSvc,
		settingsSvc:          settingsSvc,
		lstepDeliveryTrigger: lstepDeliveryTrigger,
		nowFn:                time.Now,
	}
}

// runBatchAllClinics は「全クリニック走査 → IsSyncEnabled ゲート → 1 クリニック分の処理 →
// 部分エラーログ → 処理件数>0 時のみ audit 記録」という lstep cron バッチ共通骨格を集約する
// （G3-2, dup-lstep-batch-allclinics）。ログ文言・audit operation 文字列・extraMeta は
// 呼び出し側が指定した値をそのまま使うため、移行前後で出力はバイト級で同一になる。
func (s *lstepBatchService) runBatchAllClinics(
	ctx context.Context,
	label string,
	auditWarnLabel string,
	syncedSuffix string,
	operation string,
	extraMeta map[string]any,
	perClinic func(ctx context.Context, clinicID uint64) (int, []error),
) error {
	clinics, err := s.clinicRepo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, label+": failed to fetch clinics", "error", err)
		return apperrors.Wrap(err, "failed to fetch clinics for "+label)
	}

	for i := range clinics {
		clinic := &clinics[i]
		if s.settingsSvc != nil {
			enabled, syncErr := s.settingsSvc.IsSyncEnabled(ctx, clinic.ID)
			if syncErr != nil {
				slog.ErrorContext(ctx, label+": failed to check sync enabled", "clinic_id", clinic.ID, "error", syncErr)
				continue
			}
			if !enabled {
				continue
			}
		}
		count, errs := perClinic(ctx, clinic.ID)
		if len(errs) > 0 {
			slog.ErrorContext(ctx, label+": partial errors", "clinic_id", clinic.ID, "error_count", len(errs))
		}
		if count > 0 {
			slog.InfoContext(ctx, label+": "+syncedSuffix, "clinic_id", clinic.ID, "count", count)
			meta := map[string]any{
				"operation":       operation,
				"processed_count": count,
				"error_count":     len(errs),
			}
			maps.Copy(meta, extraMeta)
			if err := s.auditSvc.LogLstepOperationWithMetadata(ctx, clinic.ID, nil,
				operation, "clinic", &clinic.ID, meta,
			); err != nil {
				slog.WarnContext(ctx, "audit log failed for "+auditWarnLabel, "error", err, "clinic_id", clinic.ID)
			}
		}
	}
	return nil
}

// DetectNoShowReservations は指定クリニックの no-show 予約を検知してステータス更新・タグ付与を行う。
