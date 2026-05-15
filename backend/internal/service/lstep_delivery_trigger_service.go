package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

const (
	exclTagDeliveryStop    = "EXCL_配信停止" // FE source: frontend/src/constants/lstep-tag-names.ts LSTEP_EXCL_DELIVERY_STOP
	exclTagDeliveryCaution = "EXCL_配信注意" // FE source: frontend/src/constants/lstep-tag-names.ts LSTEP_EXCL_DELIVERY_CAUTION
	tagFilariaAlert        = PrevFilariaTag     // FEAT-379 新命名 (旧: HEALTH_フィラリア対策中)
	tagFleaTickAlert       = PrevFleaTickTag    // FEAT-379 新命名 (旧: HEALTH_ノミダニ予防中)
	tagFoodRefill          = LtvFoodPurchaseTag // FEAT-379 新命名 (旧: PROD_フード購入)
)

// LstepDeliveryTriggerService は日次バッチで自動配信トリガーを判定し L ステップへタグ付与するサービス（FEAT-383）。
type LstepDeliveryTriggerService interface {
	// TriggerFirstVisitFollowUp3D は初診から3日後フォローアップ配信をトリガーする。
	TriggerFirstVisitFollowUp3D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerFirstVisitFollowUp7D は初診から7日後フォローアップ配信をトリガーする。
	TriggerFirstVisitFollowUp7D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerNextVisitReminder は次回来院推奨日当日のリマインダー配信をトリガーする。
	TriggerNextVisitReminder(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerVaccineDeadline60 はワクチン期限60日前リマインダー配信をトリガーする。
	TriggerVaccineDeadline60(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerVaccineDeadline30 はワクチン期限30日前リマインダー配信をトリガーする。
	TriggerVaccineDeadline30(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerBirthdayMessage はペット誕生日当日の誕生日メッセージ配信をトリガーする。
	TriggerBirthdayMessage(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerDormantPrevention180 は180日間未来院の飼い主に休眠予防配信をトリガーする。
	TriggerDormantPrevention180(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerDormantPrevention210 は210日間未来院の飼い主に休眠予防配信をトリガーする。
	TriggerDormantPrevention210(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerDormantPrevention240 は240日間未来院の飼い主に休眠予防配信をトリガーする。
	TriggerDormantPrevention240(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerDormantPrevention365 は365日間未来院の飼い主に休眠予防配信をトリガーする。
	TriggerDormantPrevention365(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerFilariaAlert はフィラリア対策タグを持つ飼い主に定期アラート配信をトリガーする。
	TriggerFilariaAlert(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerFleaTickAlert はノミダニ予防タグを持つ飼い主に定期アラート配信をトリガーする。
	TriggerFleaTickAlert(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerFoodRefillReminder はフード購入タグを持つ飼い主にリフィルリマインダー配信をトリガーする。
	TriggerFoodRefillReminder(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerFirstVisitWelcome はイベント駆動（初診完了直後）のウェルカム配信トリガー（スタブ）。
	TriggerFirstVisitWelcome(ctx context.Context, clinicID, ownerID uint64) error
	// TriggerCheckupFollowUp はイベント駆動（健診後）のフォローアップ配信トリガー（スタブ）。
	TriggerCheckupFollowUp(ctx context.Context, clinicID, ownerID uint64) error
}

type lstepDeliveryTriggerService struct {
	ownerRepo       repository.OwnerRepository
	medRecordRepo   repository.MedicalRecordRepository
	vaccinationRepo repository.VaccinationRepository
	billingItemRepo repository.BillingItemRepository
	petRepo         repository.PetRepository
	tagCacheRepo    repository.LstepTagCacheRepository
	triggerLogRepo  repository.LstepDeliveryTriggerLogRepository
	settingsSvc     LstepSettingsService
	prioritySvc     LstepTriggerPriorityService                                      // nil → suppression skip (Q23)
	clientBuilderFn func(ctx context.Context, clinicID uint64) (lstep.Client, error) // overridden in tests
}

// NewLstepDeliveryTriggerService は LstepDeliveryTriggerService を初期化して返す。
func NewLstepDeliveryTriggerService(
	ownerRepo repository.OwnerRepository,
	medRecordRepo repository.MedicalRecordRepository,
	vaccinationRepo repository.VaccinationRepository,
	billingItemRepo repository.BillingItemRepository,
	petRepo repository.PetRepository,
	tagCacheRepo repository.LstepTagCacheRepository,
	triggerLogRepo repository.LstepDeliveryTriggerLogRepository,
	settingsSvc LstepSettingsService,
	prioritySvc LstepTriggerPriorityService,
) LstepDeliveryTriggerService {
	return &lstepDeliveryTriggerService{
		ownerRepo:       ownerRepo,
		medRecordRepo:   medRecordRepo,
		vaccinationRepo: vaccinationRepo,
		billingItemRepo: billingItemRepo,
		petRepo:         petRepo,
		tagCacheRepo:    tagCacheRepo,
		triggerLogRepo:  triggerLogRepo,
		settingsSvc:     settingsSvc,
		prioritySvc:     prioritySvc,
	}
}

// ---- private helpers ----

// checkExclusion は飼い主が配信対象外かどうかを判定する。
// DeliveryExcluded フラグ または EXCL_配信停止 タグのいずれかが true の場合は除外理由を返す。
func (s *lstepDeliveryTriggerService) checkExclusion(ctx context.Context, clinicID, ownerID uint64) (excluded bool, reason string, err error) {
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		return false, "", apperrors.Wrap(err, "failed to find owner")
	}
	if owner.DeliveryExcluded {
		return true, "delivery_excluded_flag", nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return true, "no_line_user_id", nil
	}
	tags, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		return false, "", apperrors.Wrap(err, "failed to find owner tags")
	}
	for _, t := range tags {
		if t.TagName == exclTagDeliveryStop {
			return true, "excl_tag_delivery_stop", nil
		}
	}
	return false, "", nil
}

// alreadyFiredToday は当日同一トリガーが既に発火済みかを確認する（二重発火防止）。
func (s *lstepDeliveryTriggerService) alreadyFiredToday(ctx context.Context, clinicID, ownerID uint64, triggerType string, asOf time.Time) (bool, error) {
	exists, err := s.triggerLogRepo.ExistsTodayByOwnerAndType(ctx, clinicID, ownerID, triggerType, asOf)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to check trigger log")
	}
	return exists, nil
}

// recordTrigger はトリガーログを作成し生成された ID を返す。
func (s *lstepDeliveryTriggerService) recordTrigger(ctx context.Context, clinicID, ownerID uint64, triggerType string, asOf time.Time) (uint64, error) {
	log := &model.LstepDeliveryTriggerLog{
		OwnerID:     ownerID,
		ClinicID:    clinicID,
		TriggerType: triggerType,
		ScheduledAt: asOf,
		Status:      model.TriggerStatusScheduled,
	}
	if err := s.triggerLogRepo.Create(ctx, log); err != nil {
		return 0, apperrors.Wrap(err, "failed to create trigger log")
	}
	return log.ID, nil
}

// applyTagAndLog は L ステップへタグを付与しログを fired 状態に更新する。
func (s *lstepDeliveryTriggerService) applyTagAndLog(ctx context.Context, client lstep.Client, lineUserID, tagName string, logID uint64) error {
	if err := client.AddTag(ctx, lineUserID, tagName); err != nil {
		slog.ErrorContext(ctx, "delivery trigger: failed to add tag", "tag", tagName, "error", err)
		reason := fmt.Sprintf("lstep_add_tag_failed: %s", tagName)
		_ = s.triggerLogRepo.UpdateStatus(ctx, logID, model.TriggerStatusFailed, nil, &reason)
		return apperrors.Wrap(err, "failed to add lstep tag")
	}
	now := time.Now()
	return s.triggerLogRepo.UpdateStatus(ctx, logID, model.TriggerStatusFired, &now, nil)
}

// buildClient はクリニック設定から lstep.Client を構築する。
// 同期無効または API キー未設定の場合は nil, nil を返す（スキップ）。
func (s *lstepDeliveryTriggerService) buildClient(ctx context.Context, clinicID uint64) (lstep.Client, error) {
	if s.clientBuilderFn != nil {
		return s.clientBuilderFn(ctx, clinicID)
	}
	enabled, err := s.settingsSvc.IsSyncEnabled(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check lstep sync enabled")
	}
	if !enabled {
		return nil, nil
	}
	apiKey, baseURL, _, err := s.settingsSvc.GetRawCredentials(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get lstep credentials")
	}
	if apiKey == "" {
		return nil, nil
	}
	return lstep.NewClient(apiKey, baseURL), nil
}

// applySuppression は Q23 優先順位による配信抑制を判定・適用する。
// 戻り値 suppressed=true なら呼び出し元が抑制ログを作成して return すべき。
// nil prioritySvc の場合は抑制スキップ（後方互換）。
func (s *lstepDeliveryTriggerService) applySuppression(
	ctx context.Context,
	clinicID, ownerID uint64,
	triggerType string,
	asOf time.Time,
) (suppressed bool, err error) {
	if s.prioritySvc == nil {
		return false, nil
	}
	existing, err := s.triggerLogRepo.FindByOwnerAndDate(ctx, clinicID, ownerID, asOf)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to find existing trigger logs")
	}
	// SuppressedByPriority=true は既に抑制済みなので除外
	active := make([]model.LstepDeliveryTriggerLog, 0, len(existing))
	for _, l := range existing {
		if !l.SuppressedByPriority {
			active = append(active, l)
		}
	}
	if len(active) == 0 {
		return false, nil
	}

	currentPri, err := s.prioritySvc.GetPriorityFor(ctx, clinicID, triggerType)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to get priority for current trigger")
	}

	// 既存の中で最高優先度（最小値）を持つログを探す
	bestPri := int(^uint(0) >> 1) // MaxInt
	var bestLog model.LstepDeliveryTriggerLog
	for _, l := range active {
		p, pErr := s.prioritySvc.GetPriorityFor(ctx, clinicID, l.TriggerType)
		if pErr != nil {
			return false, apperrors.Wrap(pErr, "failed to get priority for existing trigger")
		}
		if p < bestPri {
			bestPri = p
			bestLog = l
		}
	}

	if currentPri > bestPri {
		// 現在のトリガーは既存より低優先度 → 抑制
		return true, nil
	}
	if currentPri < bestPri {
		// 現在のトリガーは既存より高優先度 → 全既存を降格
		for _, l := range active {
			reason := fmt.Sprintf("superseded by %s (priority %d < %d)", triggerType, currentPri, bestPri)
			if suppErr := s.triggerLogRepo.UpdateSuppressed(ctx, l.ID, reason); suppErr != nil {
				slog.ErrorContext(ctx, "delivery trigger: failed to suppress existing log", "log_id", l.ID, "error", suppErr)
				return false, apperrors.Wrap(suppErr, "failed to suppress existing trigger log")
			}
		}
		_ = bestLog // suppress "declared but not used" — bestLog is used only for logging context above
	}
	// 同優先度または降格完了 → 現在のトリガーは通常通り発火
	return false, nil
}

// runBatch は ownerID リストに対して除外チェック・重複チェック・タグ付与・ログ記録を行う汎用バッチ実行。
func (s *lstepDeliveryTriggerService) runBatch(
	ctx context.Context,
	clinicID uint64,
	ownerIDs []uint64,
	triggerType string,
	tagName string,
	asOf time.Time,
) (int, []error) {
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: failed to build lstep client", "clinic_id", clinicID, "trigger", triggerType, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to build lstep client")}
	}
	if client == nil {
		return 0, nil
	}

	var errs []error
	count := 0

	for _, ownerID := range ownerIDs {
		fired, loopErr := s.processSingleOwner(ctx, client, clinicID, ownerID, triggerType, tagName, asOf)
		if loopErr != nil {
			errs = append(errs, loopErr)
			continue
		}
		if fired {
			count++
		}
	}
	return count, errs
}

// processSingleOwner は1飼い主分のトリガー処理を行い、配信実行されたか否かを返す。
func (s *lstepDeliveryTriggerService) processSingleOwner(
	ctx context.Context,
	client lstep.Client,
	clinicID, ownerID uint64,
	triggerType, tagName string,
	asOf time.Time,
) (bool, error) {
	already, err := s.alreadyFiredToday(ctx, clinicID, ownerID, triggerType, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: alreadyFiredToday error", "owner_id", ownerID, "trigger", triggerType, "error", err)
		return false, err
	}
	if already {
		return false, nil
	}

	// Q23: 優先順位による配信抑制チェック
	suppressed, err := s.applySuppression(ctx, clinicID, ownerID, triggerType, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: applySuppression error", "owner_id", ownerID, "trigger", triggerType, "error", err)
		return false, err
	}
	if suppressed {
		suppressionReason := fmt.Sprintf("owner_id=%d already has higher-priority trigger on %s", ownerID, asOf.Format("2006-01-02"))
		suppressedLog := &model.LstepDeliveryTriggerLog{
			OwnerID:              ownerID,
			ClinicID:             clinicID,
			TriggerType:          triggerType,
			ScheduledAt:          asOf,
			Status:               model.TriggerStatusScheduled,
			SuppressedByPriority: true,
			SuppressionReason:    &suppressionReason,
		}
		if createErr := s.triggerLogRepo.Create(ctx, suppressedLog); createErr != nil {
			slog.ErrorContext(ctx, "delivery trigger: failed to create suppressed log", "owner_id", ownerID, "trigger", triggerType, "error", createErr)
			return false, apperrors.Wrap(createErr, "failed to create suppressed trigger log")
		}
		return false, nil
	}

	excluded, reason, err := s.checkExclusion(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: checkExclusion error", "owner_id", ownerID, "error", err)
		return false, err
	}

	logID, err := s.recordTrigger(ctx, clinicID, ownerID, triggerType, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: recordTrigger error", "owner_id", ownerID, "error", err)
		return false, err
	}

	if excluded {
		_ = s.triggerLogRepo.UpdateStatus(ctx, logID, model.TriggerStatusExcluded, nil, &reason)
		return false, nil
	}

	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: failed to find owner for lineUserID", "owner_id", ownerID, "error", err)
		return false, apperrors.Wrap(err, "failed to find owner")
	}

	if err := s.applyTagAndLog(ctx, client, *owner.LineUserID, tagName, logID); err != nil {
		return false, err
	}
	return true, nil
}

// ---- public trigger methods ----

func (s *lstepDeliveryTriggerService) TriggerFirstVisitFollowUp3D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	target := asOf.AddDate(0, 0, -3)
	ownerIDs, err := s.medRecordRepo.FindOwnersByFirstVisitDate(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger first_visit_followup_3d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by first visit date")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFirstVisitFollowUp3D, model.TriggerTypeFirstVisitFollowUp3D, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerFirstVisitFollowUp7D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	target := asOf.AddDate(0, 0, -7)
	ownerIDs, err := s.medRecordRepo.FindOwnersByFirstVisitDate(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger first_visit_followup_7d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by first visit date")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFirstVisitFollowUp7D, model.TriggerTypeFirstVisitFollowUp7D, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerNextVisitReminder(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.medRecordRepo.FindOwnersByNextVisitRecommended(ctx, clinicID, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger next_visit_reminder: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by next visit recommended date")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeNextVisitReminder, model.TriggerTypeNextVisitReminder, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerVaccineDeadline60(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	target := asOf.AddDate(0, 0, 60)
	ownerIDs, err := s.vaccinationRepo.FindOwnersByVaccineDeadline(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger vaccine_deadline_60d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by vaccine deadline")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeVaccineDeadline60, model.TriggerTypeVaccineDeadline60, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerVaccineDeadline30(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	target := asOf.AddDate(0, 0, 30)
	ownerIDs, err := s.vaccinationRepo.FindOwnersByVaccineDeadline(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger vaccine_deadline_30d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by vaccine deadline")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeVaccineDeadline30, model.TriggerTypeVaccineDeadline30, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerBirthdayMessage(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.petRepo.FindOwnersByPetBirthday(ctx, clinicID, int(asOf.Month()), asOf.Day())
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger birthday_message: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by pet birthday")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeBirthdayMessage, model.TriggerTypeBirthdayMessage, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerDormantPrevention180(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	thresholds, err := s.settingsSvc.GetDormantThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_180d: get thresholds error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to get dormant thresholds")}
	}
	ownerIDs, err := s.medRecordRepo.FindOwnersByLastVisitDays(ctx, clinicID, thresholds.Stage180, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_180d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by last visit days")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention180, model.TriggerTypeDormantPrevention180, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerDormantPrevention210(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	thresholds, err := s.settingsSvc.GetDormantThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_210d: get thresholds error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to get dormant thresholds")}
	}
	ownerIDs, err := s.medRecordRepo.FindOwnersByLastVisitDays(ctx, clinicID, thresholds.Stage210, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_210d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by last visit days")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention210, model.TriggerTypeDormantPrevention210, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerDormantPrevention240(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	thresholds, err := s.settingsSvc.GetDormantThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_240d: get thresholds error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to get dormant thresholds")}
	}
	ownerIDs, err := s.medRecordRepo.FindOwnersByLastVisitDays(ctx, clinicID, thresholds.Stage240, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_240d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by last visit days")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention240, model.TriggerTypeDormantPrevention240, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerDormantPrevention365(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	thresholds, err := s.settingsSvc.GetDormantThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_365d: get thresholds error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to get dormant thresholds")}
	}
	ownerIDs, err := s.medRecordRepo.FindOwnersByLastVisitDays(ctx, clinicID, thresholds.Stage365, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_365d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by last visit days")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention365, model.TriggerTypeDormantPrevention365, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerFilariaAlert(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.tagCacheRepo.FindOwnerIDsByTag(ctx, clinicID, tagFilariaAlert)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger filaria_alert: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by filaria tag")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFilariaAlert, model.TriggerTypeFilariaAlert, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerFleaTickAlert(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.tagCacheRepo.FindOwnerIDsByTag(ctx, clinicID, tagFleaTickAlert)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger flea_tick_alert: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by flea tick tag")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFleaTickAlert, model.TriggerTypeFleaTickAlert, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerFoodRefillReminder(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.tagCacheRepo.FindOwnerIDsByTag(ctx, clinicID, tagFoodRefill)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger food_refill_reminder: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by food refill tag")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFoodRefillReminder, model.TriggerTypeFoodRefillReminder, asOf)
}

// TriggerFirstVisitWelcome は初診完了直後に呼び出されるイベント駆動トリガー（FEAT-383 Phase 2）。
func (s *lstepDeliveryTriggerService) TriggerFirstVisitWelcome(ctx context.Context, clinicID, ownerID uint64) error {
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger first_visit_welcome: failed to build lstep client", "clinic_id", clinicID, "owner_id", ownerID, "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}
	_, err = s.processSingleOwner(ctx, client, clinicID, ownerID, model.TriggerTypeFirstVisitWelcome, model.TriggerTypeFirstVisitWelcome, time.Now())
	return err
}

// TriggerCheckupFollowUp は健診作成直後に呼び出されるイベント駆動トリガー（FEAT-383 Phase 2）。
func (s *lstepDeliveryTriggerService) TriggerCheckupFollowUp(ctx context.Context, clinicID, ownerID uint64) error {
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger checkup_followup: failed to build lstep client", "clinic_id", clinicID, "owner_id", ownerID, "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}
	_, err = s.processSingleOwner(ctx, client, clinicID, ownerID, model.TriggerTypeCheckupFollowUp, model.TriggerTypeCheckupFollowUp, time.Now())
	return err
}
