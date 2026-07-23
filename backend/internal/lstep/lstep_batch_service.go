package lstep

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// lstepBatchPageSize は PERF-FOLLOWUP-02 のカーソルページネーション 1 ページあたりの件数。
const lstepDormantBatchPageSize = 500

// deliveryTriggerHourJST は仕様 §6.4 で固定された Lステップ自動配信バッチの実行時刻 (10:00 JST)。
// 設定可能化は意図的に廃止 (configurable fire hour 削除)。
const deliveryTriggerHourJST = 10

// LstepBatchService はバッチ処理でノーショウ検知・休眠検知を行うサービス（BE-005, BE-014）。
type LstepBatchService interface {
	// RunNoShowCheckAllClinics は全クリニックに対してノーショウ検知を実行するcronエントリポイント。
	RunNoShowCheckAllClinics(ctx context.Context) error
	// RunNoShowCheckAllClinicsAt は durable scheduler の scheduledAt と安定 runID を使って実行する。
	RunNoShowCheckAllClinicsAt(ctx context.Context, scheduledAt time.Time, runID string) BatchRunResult
	// RunDormantDetectionAllClinics は全クリニックに対して休眠検知を実行するcronエントリポイント（02:00 JST）。
	RunDormantDetectionAllClinics(ctx context.Context) error
	// RunDormantDetectionAllClinicsAt は durable scheduler の scheduledAt を休眠判定基準に使う。
	RunDormantDetectionAllClinicsAt(ctx context.Context, scheduledAt time.Time, runID string) BatchRunResult
	// RunLTVTopPercentSyncAllClinics は全クリニックに対して LTV 上位 20% タグを同期するcronエントリポイント（FEAT-377）。
	RunLTVTopPercentSyncAllClinics(ctx context.Context) error
	// RunVisitDormantSyncAllClinics は全クリニックに対して VISIT_* タグ（180/210/240日超）を同期するcronエントリポイント（FEAT-377）。
	RunVisitDormantSyncAllClinics(ctx context.Context) error
	// RunHealthPreventionTagSyncAllClinics は全クリニックに対して健診・予防・物販タグを同期するcronエントリポイント（FEAT-379）。
	RunHealthPreventionTagSyncAllClinics(ctx context.Context) error
	// RunDeliveryTriggerBatchAllClinics は全クリニックに対して自動配信トリガーバッチを実行するcronエントリポイント（FEAT-383: 10:00 JST）。
	RunDeliveryTriggerBatchAllClinics(ctx context.Context) error
	// RunDeliveryTriggerBatchAllClinicsAt は durable scheduler の scheduledAt を全配信判定へ渡す。
	RunDeliveryTriggerBatchAllClinicsAt(ctx context.Context, scheduledAt time.Time, runID string) BatchRunResult
}

// BatchRunResult is the strict aggregate returned to the durable scheduler.
// Processed must always equal Succeeded + Failed.
type BatchRunResult struct {
	Processed int
	Succeeded int
	Failed    int
}

// Validate checks the counter invariant before the result crosses the HTTP boundary.
func (r BatchRunResult) Validate() error {
	if r.Processed < 0 || r.Succeeded < 0 || r.Failed < 0 {
		return errors.New("batch counters must be non-negative")
	}
	if r.Processed != r.Succeeded+r.Failed {
		return errors.New("batch processed must equal succeeded plus failed")
	}
	return nil
}

func (r BatchRunResult) add(succeeded, failed int) BatchRunResult {
	return BatchRunResult{
		Processed: r.Processed + succeeded + failed,
		Succeeded: r.Succeeded + succeeded,
		Failed:    r.Failed + failed,
	}
}

func failedBatchRunResult() BatchRunResult {
	return BatchRunResult{Processed: 1, Failed: 1}
}

func scheduledBatchMetadata(scheduledAt time.Time, runID string) map[string]any {
	return map[string]any{
		"batch_run_id":   runID,
		"scheduled_time": scheduledAt.UnixMilli(),
	}
}

type lstepBatchDeliveryTrigger interface {
	TriggerFirstVisitFollowUp3D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerFirstVisitFollowUp7D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerNextVisitReminder(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerVaccineDeadline60(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerVaccineDeadline30(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerBirthdayMessage(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerDormantPrevention180(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerDormantPrevention210(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerDormantPrevention240(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerDormantPrevention365(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerFilariaAlert(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerFleaTickAlert(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	TriggerFoodRefillReminder(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
}

type lstepBatchService struct {
	reservationRepo      lstepBatchReservationRepository
	tagSyncSvc           lstepBatchTagSyncer
	clinicRepo           lstepBatchClinicRepository
	medRecordRepo        lstepBatchMedicalRecordRepository
	auditSvc             lstepBatchAuditService
	settingsSvc          lstepBatchSettingsService
	lstepDeliveryTrigger lstepBatchDeliveryTrigger
	transactor           lstepBatchTransactor
	noShowAuditTx        NoShowAuditTxLogger
	nowFn                func() time.Time
}

// NewLstepBatchService は LstepBatchService を初期化して返す。
func NewLstepBatchService(
	reservationRepo lstepBatchReservationRepository,
	tagSyncSvc lstepBatchTagSyncer,
	clinicRepo lstepBatchClinicRepository,
	medRecordRepo lstepBatchMedicalRecordRepository,
	auditSvc lstepBatchAuditService,
	settingsSvc lstepBatchSettingsService,
	lstepDeliveryTrigger lstepBatchDeliveryTrigger,
	transactor lstepBatchTransactor,
	noShowAuditTx NoShowAuditTxLogger,
) LstepBatchService {
	return &lstepBatchService{
		reservationRepo:      reservationRepo,
		tagSyncSvc:           tagSyncSvc,
		clinicRepo:           clinicRepo,
		medRecordRepo:        medRecordRepo,
		auditSvc:             auditSvc,
		settingsSvc:          settingsSvc,
		lstepDeliveryTrigger: lstepDeliveryTrigger,
		transactor:           transactor,
		noShowAuditTx:        noShowAuditTx,
		nowFn:                time.Now,
	}
}

type lstepBatchReservationRepository interface {
	FindNoShowCandidates(ctx context.Context, clinicID uint64) ([]model.Reservation, error)
	MarkNoShow(ctx context.Context, clinicID, id uint64) (reservation.NoShowTransition, error)
}

type lstepBatchReservationRepositoryAt interface {
	FindNoShowCandidatesAt(
		ctx context.Context,
		clinicID uint64,
		evaluatedAt time.Time,
	) ([]model.Reservation, error)
	MarkNoShowAt(
		ctx context.Context,
		clinicID uint64,
		id uint64,
		evaluatedAt time.Time,
	) (reservation.NoShowTransition, error)
}

type lstepBatchTransactor interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

// NoShowAuditEntry is the semantic, transaction-local audit contract owned by lstep.
// cmd/api maps it to the shared audit log without leaking that service type here.
type NoShowAuditEntry struct {
	ClinicID       uint64
	AppointmentID  uint64
	PreviousStatus model.ReservationStatus
	EvaluatedAt    time.Time
	RuleVersion    string
	BatchRunID     string
}

type NoShowAuditTxLogger interface {
	LogNoShowTransitionTx(ctx context.Context, entry *NoShowAuditEntry) error
}

type lstepBatchClinicRepository interface {
	FindAll(ctx context.Context) ([]model.Clinic, error)
}

type lstepBatchMedicalRecordRepository interface {
	FindDormantOwnerEntries(ctx context.Context, clinicID uint64, minDaysSince int) ([]medicalrecord.DormantOwnerEntry, error)
	FindDormantOwnerEntriesCursor(ctx context.Context, clinicID uint64, minDaysSince int, afterOwnerID uint64, limit int) ([]medicalrecord.DormantOwnerEntry, error)
}

type lstepBatchMedicalRecordRepositoryAt interface {
	FindDormantOwnerEntriesCursorAt(
		ctx context.Context,
		clinicID uint64,
		minDaysSince int,
		afterOwnerID uint64,
		limit int,
		evaluatedAt time.Time,
	) ([]medicalrecord.DormantOwnerEntry, error)
}

type lstepBatchTagSyncer interface {
	SyncDormantTagsWithThresholds(ctx context.Context, clinicID, ownerID uint64, daysSince int, thresholds model.DormantThresholds) error
	SyncLTVTopPercent(ctx context.Context, clinicID uint64) (int, []error)
	SyncVisitDormantTags(ctx context.Context, clinicID, ownerID uint64, daysSince int) error
	SyncHealthPreventionTagsForClinic(ctx context.Context, clinicID uint64) (int, []error)
}

type lstepBatchSettingsService interface {
	IsSyncEnabled(ctx context.Context, clinicID uint64) (bool, error)
	GetDormantThresholds(ctx context.Context, clinicID uint64) (model.DormantThresholds, error)
}

type lstepBatchAuditService interface {
	LogLstepOperationWithMetadata(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64, metadata any) error
}

// runBatchAllClinics は「全クリニック走査 → IsSyncEnabled ゲート → 1 クリニック分の処理 →
// 部分エラーログ（件数のみ）→ 処理件数>0 またはエラーあり時に audit 記録」という
// lstep cron バッチ共通骨格を集約する（G3-2, dup-lstep-batch-allclinics）。
// ログ文言・audit operation 文字列・extraMeta は呼び出し側が指定した値をそのまま使う。
func (s *lstepBatchService) runBatchAllClinics(
	ctx context.Context,
	label string,
	auditWarnLabel string,
	syncedSuffix string,
	operation string,
	extraMeta map[string]any,
	perClinic func(ctx context.Context, clinicID uint64) (int, []error),
) error {
	_, err := s.runBatchAllClinicsWithResult(
		ctx,
		label,
		auditWarnLabel,
		syncedSuffix,
		operation,
		extraMeta,
		perClinic,
	)
	return err
}

// runBatchAllClinicsWithResult preserves the legacy fatal error while also
// counting every business, settings, and audit failure for durable scheduling.
func (s *lstepBatchService) runBatchAllClinicsWithResult(
	ctx context.Context,
	label string,
	auditWarnLabel string,
	syncedSuffix string,
	operation string,
	extraMeta map[string]any,
	perClinic func(ctx context.Context, clinicID uint64) (int, []error),
) (BatchRunResult, error) {
	if s.settingsSvc == nil {
		slog.ErrorContext(ctx, label+": settings service is not configured")
		return failedBatchRunResult(), apperrors.WrapInternalServerError("LSTEP settings service is not configured")
	}
	clinics, err := s.clinicRepo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, label+": failed to fetch clinics", "error", err)
		return failedBatchRunResult(), apperrors.Wrap(err, "failed to fetch clinics for "+label)
	}

	result := BatchRunResult{}
	for i := range clinics {
		clinic := &clinics[i]
		enabled, syncErr := s.settingsSvc.IsSyncEnabled(ctx, clinic.ID)
		if syncErr != nil {
			slog.ErrorContext(ctx, label+": failed to check sync enabled", "clinic_id", clinic.ID, "error", syncErr)
			result = result.add(0, 1)
			continue
		}
		if !enabled {
			continue
		}
		count, errs := perClinic(ctx, clinic.ID)
		if count < 0 {
			slog.ErrorContext(ctx, label+": invalid negative processed count", "clinic_id", clinic.ID)
			count = 0
			errs = append(errs, errors.New("negative batch processed count"))
		}
		result = result.add(count, len(errs))
		if len(errs) > 0 {
			slog.ErrorContext(ctx, label+": partial errors",
				"clinic_id", clinic.ID, "error_count", len(errs))
		}
		if count > 0 {
			slog.InfoContext(ctx, label+": "+syncedSuffix, "clinic_id", clinic.ID, "count", count)
		}
		if count > 0 || len(errs) > 0 {
			meta := map[string]any{
				"operation":       operation,
				"processed_count": count,
				"error_count":     len(errs),
			}
			maps.Copy(meta, extraMeta)
			if s.auditSvc == nil {
				slog.WarnContext(ctx, "audit logger is not configured for "+auditWarnLabel, "clinic_id", clinic.ID)
				result = result.add(0, 1)
				continue
			}
			if err := s.auditSvc.LogLstepOperationWithMetadata(ctx, clinic.ID, nil,
				operation, "clinic", &clinic.ID, meta,
			); err != nil {
				slog.WarnContext(ctx, "audit log failed for "+auditWarnLabel, "error", err, "clinic_id", clinic.ID)
				result = result.add(0, 1)
			}
		}
	}
	return result, nil
}
