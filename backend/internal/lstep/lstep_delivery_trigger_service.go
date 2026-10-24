package lstep

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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
	// TriggerNextVisitTentativeReminder3D は tentative 予約の intended_date 3日前
	// リマインダー配信をトリガーする（EMR-308。stage は trigger_type 文字列が担う）。
	TriggerNextVisitTentativeReminder3D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
	// TriggerNextVisitTentativeReminder0D は tentative 予約の intended_date 当日
	// リマインダー配信をトリガーする（EMR-308）。
	TriggerNextVisitTentativeReminder0D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error)
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
	// petID は初診を起票したイベントペット（EMR-330: 送信前に生存・所有・clinic 整合を再読する）。
	TriggerFirstVisitWelcome(ctx context.Context, clinicID, ownerID, petID uint64) error
	// TriggerCheckupFollowUp はイベント駆動（健診後）のフォローアップ配信トリガー（スタブ）。
	// EMR-330 R1: petID は健診イベント由来ペット — claim/write 前に生存・所有・医院を再読する。
	TriggerCheckupFollowUp(ctx context.Context, clinicID, ownerID, petID uint64) error
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
	appointmentRepo deliveryAppointmentRepository                                    // EMR-308: tentative reminder の source of truth。nil → tentative stages no-op
	clientBuilderFn func(ctx context.Context, clinicID uint64) (lstep.Client, error) // overridden in tests
	// EMR-504: 紐付け済み LINE 配信先（owner_line_binding_states）の reader。
	recipients lineRecipientReader
	// EMR-491 D6: ウェルカム配信のイベントペットの生死 reader（delivery owner lock 内の tx で読む）。
	petLife sharedkernel.PetLifeStateReader
}

// NewLstepDeliveryTriggerService は LstepDeliveryTriggerService を初期化して返す。
// EMR-308: appointmentRepo は tentative リマインダーの candidate 抽出と
// 送信直前再検証に使う appointment 参照（reservation store を composition で注入）。
func NewLstepDeliveryTriggerService(
	ownerRepo deliveryOwnerRepository,
	medRecordRepo deliveryMedicalRecordRepository,
	vaccinationRepo deliveryVaccinationRepository,
	petRepo deliveryPetRepository,
	tagCacheRepo deliveryTagCacheRepository,
	triggerLogRepo LstepDeliveryTriggerLogRepository,
	settingsSvc LstepSettingsService,
	prioritySvc LstepTriggerPriorityService,
	appointmentRepo deliveryAppointmentRepository,
	recipients lineRecipientReader,
	petLife sharedkernel.PetLifeStateReader,
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
		appointmentRepo: appointmentRepo,
		// EMR-504 / EMR-491 D6: 配信先 reader と死亡 reader は constructor 注入する。
		recipients: recipients,
		petLife:    petLife,
	}
}
