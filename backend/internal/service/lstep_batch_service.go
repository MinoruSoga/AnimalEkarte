package service

import (
	"context"
	"time"

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

// DetectNoShowReservations は指定クリニックの no-show 予約を検知してステータス更新・タグ付与を行う。
