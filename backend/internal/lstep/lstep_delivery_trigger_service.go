package lstep

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
)

const (
	tagFilariaAlert  = PrevFilariaTag     // FEAT-379 新命名 (旧: HEALTH_フィラリア対策中)
	tagFleaTickAlert = PrevFleaTickTag    // FEAT-379 新命名 (旧: HEALTH_ノミダニ予防中)
	tagFoodRefill    = LtvFoodPurchaseTag // FEAT-379 新命名 (旧: PROD_フード購入)
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
	ownerRepo       deliveryOwnerRepository
	medRecordRepo   deliveryMedicalRecordRepository
	vaccinationRepo deliveryVaccinationRepository
	petRepo         deliveryPetRepository
	tagCacheRepo    deliveryTagCacheRepository
	triggerLogRepo  LstepDeliveryTriggerLogRepository
	settingsSvc     LstepSettingsService
	prioritySvc     LstepTriggerPriorityService                                      // nil → suppression skip (Q23)
	clientBuilderFn func(ctx context.Context, clinicID uint64) (lstep.Client, error) // overridden in tests
}

// NewLstepDeliveryTriggerService は LstepDeliveryTriggerService を初期化して返す。
func NewLstepDeliveryTriggerService(
	ownerRepo deliveryOwnerRepository,
	medRecordRepo deliveryMedicalRecordRepository,
	vaccinationRepo deliveryVaccinationRepository,
	petRepo deliveryPetRepository,
	tagCacheRepo deliveryTagCacheRepository,
	triggerLogRepo LstepDeliveryTriggerLogRepository,
	settingsSvc LstepSettingsService,
	prioritySvc LstepTriggerPriorityService,
) LstepDeliveryTriggerService {
	return &lstepDeliveryTriggerService{
		ownerRepo:       ownerRepo,
		medRecordRepo:   medRecordRepo,
		vaccinationRepo: vaccinationRepo,
		petRepo:         petRepo,
		tagCacheRepo:    tagCacheRepo,
		triggerLogRepo:  triggerLogRepo,
		settingsSvc:     settingsSvc,
		prioritySvc:     prioritySvc,
	}
}
