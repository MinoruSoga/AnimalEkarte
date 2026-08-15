package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// CPMStage は顧客ポートフォリオ管理ステージ（BE-004）。
// 仕様: docs/spec/line/lstep-integration.md Section 7.2
type CPMStage string

// タグプレフィックス定数 — lstep_auto_managed_prefixes (migration 006) の DB 登録値と一致させる。
// これらの値は Lステップ側のタグ命名プロトコルとして固定されているため変更不可。
// 検出ロジック（strings.HasPrefix）および生成ロジック（fmt.Sprintf/+）で使用する。
const (
	tagPrefixNextVisit   = "next_visit_"
	tagPrefixRefillDue   = "refill_due_"
	tagPrefixCheckupDone = "checkup_done_"
	tagPrefixNextCheckup = "next_checkup_"
	tagPrefixChronic     = "chronic_"
)

// TagPrefixCheckupDone is shared with the remaining lifecycle service until L④.
const TagPrefixCheckupDone = tagPrefixCheckupDone

const (
	// FEAT-375: 連続失敗 N 回で付与するエラー除外タグ
	lstepErrorTag           = "EXCL_カルテ連携エラー"
	lstepSyncErrorThreshold = 5
	exclTagDeliveryStop     = "EXCL_配信停止"
	exclTagDeliveryCaution  = "EXCL_配信注意"

	CPMStageEncounter CPMStage = "cpm_encounter" // 来院1回・LTV 20,000円未満（仕様書 §3 明示）
	CPMStageGrowing   CPMStage = "cpm_growing"   // 2〜3回来院・90日以内・LTV 20,000〜50,000円
	CPMStageCore      CPMStage = "cpm_core"      // 在籍180日以上・年間2回以上・LTV 50,000円以上
	CPMStageSpot      CPMStage = "cpm_spot"      // 単回高額（30,000円以上）・90日超来院なし
	CPMStageNoah      CPMStage = "cpm_noah"      // 在籍1年以上・年間3回以上・LTV 80,000円以上
	CPMStageDormant   CPMStage = "cpm_dormant"   // 最終来院から240日超
	// CPMStageUnclassified は全6ステージのいずれにも該当しない異常データ検出用。配信対象外のため allCPMStages には含めない。
	CPMStageUnclassified CPMStage = "cpm_unclassified"
)

var allCPMStages = []CPMStage{
	CPMStageEncounter, CPMStageGrowing, CPMStageCore, CPMStageSpot, CPMStageNoah, CPMStageDormant,
}

// CPMStageV2 は来院累計回数ベースの 5 段階 CPM ステージ（Q19 確定 2026-05-08）。
type CPMStageV2 string

const (
	CPMStageV2Encounter CPMStageV2 = "CPM_01_出会い"   // 累計 0〜1 回
	CPMStageV2Coming    CPMStageV2 = "CPM_02_これから"  // 累計 2〜3 回
	CPMStageV2Good      CPMStageV2 = "CPM_03_いいかんじ" // 累計 4〜7 回
	CPMStageV2Family    CPMStageV2 = "CPM_04_ファミリー" // 累計 8〜12 回
	CPMStageV2Noah      CPMStageV2 = "CPM_05_ノア"    // 累計 13 回以上
)

var allCPMV2Stages = []CPMStageV2{
	CPMStageV2Encounter, CPMStageV2Coming, CPMStageV2Good, CPMStageV2Family, CPMStageV2Noah,
}

// VISIT dormant タグ定数 — 重複付与可（複数閾値を同時保持）。
const (
	visitTag120 = "VISIT_120日超"
	visitTag180 = "VISIT_180日超"
	visitTag220 = "VISIT_220日超"
	visitTag240 = "VISIT_240日超"
	ltvTop20Tag = "LTV_上位20"
)

// CPMStageV2Input は CalculateCPMStageV2 に渡す集計データ（Q19 確定 2026-05-08）。
type CPMStageV2Input struct {
	TotalVisitCount int64                 // 累計来院回数
	CPMV2Thresholds model.CPMV2Thresholds // クリニック単位閾値（0 以下はデフォルト補完）
}

// CalculateCPMStageV2 は累計来院回数ベースの V2 CPM ステージを計算する（Q19 確定 2026-05-08）。
// 閾値は CPMV2Thresholds.WithDefaults() で補完される。
func CalculateCPMStageV2(d CPMStageV2Input) CPMStageV2 {
	t := d.CPMV2Thresholds.WithDefaults()
	switch {
	case d.TotalVisitCount >= int64(t.Noah):
		return CPMStageV2Noah
	case d.TotalVisitCount >= int64(t.Family):
		return CPMStageV2Family
	case d.TotalVisitCount >= int64(t.Good):
		return CPMStageV2Good
	case d.TotalVisitCount >= int64(t.Coming):
		return CPMStageV2Coming
	default: // 0 または 1 回
		return CPMStageV2Encounter
	}
}

// CPMData は CPM ステージ計算に必要な集計データ。
type CPMData struct {
	TotalVisitCount      int64
	AnnualVisitCount     int64
	DaysSinceVisit       int                   // 最終来院からの経過日数（来院なし = -1）
	LTVAmount            int64                 // 支払済み累計金額（円）
	FirstVisitDaysSince  int                   // 初来院からの経過日数＝在籍期間（来院なし = -1）
	MaxSingleVisitAmount int64                 // 単回最大支払い額（cpm_spot 判定用）
	Thresholds           model.CPMV1Thresholds // P2: 設定駆動閾値（ゼロ値は WithDefaults() で補完）
}

// CalculateCPMStage は純粋関数として CPM ステージを計算する（BE-004）。
// 仕様: docs/spec/line/lstep-integration.md Section 7.2
//
//nolint:gocritic // hugeParam: 純粋関数 API として immutable な値型で受け取る設計
func CalculateCPMStage(d CPMData) CPMStage {
	t := d.Thresholds.WithDefaults()
	// cpm_dormant: 最終来院から DormantDays 日超または来院なし（最優先）
	if d.DaysSinceVisit < 0 || d.DaysSinceVisit >= t.DormantDays {
		return CPMStageDormant
	}
	// cpm_noah: 在籍 NoahDays 日以上、年間 NoahAnnualVisits 回以上、LTV NoahLTV 円以上
	// V1 cpm_noah は簡略 3 条件 (LTV/visit/在籍) で判定する。
	// 仕様書 §3 の 5 条件記述は V2 詳細化対象であり、V1 では参照しない。
	// PO-QA Q29 (2026-05-08 設計確定) — V1 簡略判定の役割分担を明記。
	// 詳細条件追加は V2 (CalculateCPMStageV2) で対応すること。
	if d.FirstVisitDaysSince >= t.NoahDays && d.AnnualVisitCount >= int64(t.NoahAnnualVisits) && d.LTVAmount >= t.NoahLTV {
		return CPMStageNoah
	}
	// cpm_core: 在籍 CoreDays 日以上、年間 CoreAnnualVisits 回以上、LTV CoreLTV 円以上
	if d.FirstVisitDaysSince >= t.CoreDays && d.AnnualVisitCount >= int64(t.CoreAnnualVisits) && d.LTVAmount >= t.CoreLTV {
		return CPMStageCore
	}
	// cpm_spot: 単回高額（SpotMinAmount 円以上）かつ SpotInactiveDays 日超来院なし
	if d.MaxSingleVisitAmount >= t.SpotMinAmount && d.DaysSinceVisit > t.SpotInactiveDays {
		return CPMStageSpot
	}
	// cpm_growing: 初診から GrowingMaxDays 日以内 AND GrowingMinVisits〜GrowingMaxVisits 回来院 AND LTV LTVBreakLow〜CoreLTV 円未満
	if d.FirstVisitDaysSince >= 0 && d.FirstVisitDaysSince <= t.GrowingMaxDays &&
		d.TotalVisitCount >= int64(t.GrowingMinVisits) && d.TotalVisitCount <= int64(t.GrowingMaxVisits) &&
		d.LTVAmount >= t.LTVBreakLow && d.LTVAmount < t.CoreLTV {
		return CPMStageGrowing
	}
	// cpm_encounter: 来院1回 AND LTV LTVBreakLow 円未満（仕様書 §3 明示判定）
	if d.TotalVisitCount == 1 && d.LTVAmount < t.LTVBreakLow {
		return CPMStageEncounter
	}
	// cpm_unclassified: 全6ステージのいずれにも該当しない異常データ（配信対象外）
	return CPMStageUnclassified
}

// LstepTagSyncService は Lステップタグ同期の業務ロジックインターフェース（BE-003, BE-004, BE-005, BE-011）。
type LstepTagSyncService interface {
	// SyncVaccineTag はワクチン接種記録に基づきタグを同期する（BE-003）。
	SyncVaccineTag(ctx context.Context, clinicID, ownerID, vaccinationID uint64) error
	// SyncVisitCompletionTags は診療完了時の来院・LTV タグを同期する（BE-004）。
	SyncVisitCompletionTags(ctx context.Context, clinicID, ownerID uint64) error
	// SyncOwnerAnimalClassificationTags は飼い主の動物分類タグ（has_dog/has_cat/has_both）を同期する（BE-005）。
	SyncOwnerAnimalClassificationTags(ctx context.Context, clinicID, ownerID uint64) error
	// SyncPetBasicInfoTags は全生存ペットの基本情報タグ（品種・性別・誕生日・去勢）を同期する（BE-005）。
	SyncPetBasicInfoTags(ctx context.Context, clinicID, ownerID uint64) error
	// SyncCPMStageTag は CPM ステージタグを同期する（BE-011）。
	SyncCPMStageTag(ctx context.Context, clinicID, ownerID uint64) error
	// SyncNextVisitTag は次回来院推奨日タグ（next_visit_YYYY-MM-DD）を同期する（BE-006）。
	// 最新カルテの next_visit_recommended_date を参照し、古い next_visit_* タグを差し替える。
	SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error
	// SyncCheckupTag は健診記録の作成・更新時に checkup_done_{typeID}_{YYYY-MM}/next_checkup_* タグを同期する（BE-008）。
	// 同一健診種別の古い checkup_done タグを解除してから新タグを付与する。next_checkup_* は最新1件のみ。
	SyncCheckupTag(ctx context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error
	// SyncPrescriptionTag は飼い主の全アクティブ処方を取得し、補充推奨日が最も遅い処方に基づいて
	// refill_due_* タグを更新する（BE-009）。処方記録の追加・更新・削除後に呼び出すこと。
	SyncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) error
	// SyncChronicConditionTags は慢性疾患フラグに基づき chronic_* タグを差分同期する（BE-012）。
	// activeConditionCodes は飼い主の全生存ペットのアクティブ疾患コード一覧。
	SyncChronicConditionTags(ctx context.Context, clinicID, ownerID uint64, activeConditionCodes []string) error
	// SyncDormantTagsWithThresholds は事前取得済みの閾値を使って dormant タグを同期する（N+1 解消用 PERF-2）。
	// DetectDormantOwners バッチがループ外で閾値を 1 回取得し、各オーナーに渡す。
	SyncDormantTagsWithThresholds(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int, thresholds model.DormantThresholds) error
	// ResyncOwnerVaccineTags は飼い主の生存ワクチン記録から vaccine_* タグを再構築する（ISSUE-004）。
	// 種別ごとに最新の接種日のみタグを保持する。レコードが0件の場合は全 vaccine_* タグを解除する。
	ResyncOwnerVaccineTags(ctx context.Context, clinicID, ownerID uint64) error
	// ResyncOwnerCheckupTags は飼い主の生存健診記録から checkup_done_* / next_checkup_* タグを再構築する（ISSUE-004）。
	// 種別ごとに最新の検査日のみ保持。next_checkup は最新の next_date 1件のみ保持。
	ResyncOwnerCheckupTags(ctx context.Context, clinicID, ownerID uint64) error
	// SyncLTVTopPercent は LTV 上位 20% の飼い主に LTV_上位20 タグを付与し、
	// それ以外の飼い主から解除する（FEAT-377）。
	// 処理件数と個別エラーのスライスを返す（全体は失敗しない）。
	SyncLTVTopPercent(ctx context.Context, clinicID uint64) (int, []error)
	// SyncVisitDormantTags は最終来院経過日数に基づき VISIT_* タグを差分同期する（FEAT-377）。
	// VISIT タグは重複付与可（複数閾値を同時保持）。daysSinceLastVisit < 0 は来院なしを表す。
	SyncVisitDormantTags(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int) error
	// SyncExclusionTags は配信停止条件（opt-out / 会員ステータス / 全ペット死亡）に基づき
	// EXCL_配信停止 タグを同期する（FEAT-377）。
	// 注: checkOptOut は呼ばない（このメソッド自体が opt-out 判定の実装）。
	SyncExclusionTags(ctx context.Context, clinicID, ownerID uint64) error
	// SyncHealthPreventionTagsForClinic は指定クリニックの全飼い主に対して
	// 健診・予防・物販タグを一括同期する（FEAT-379 バッチエントリポイント）。
	// 処理件数と個別エラーのスライスを返す（全体は失敗しない）。
	SyncHealthPreventionTagsForClinic(ctx context.Context, clinicID uint64) (int, []error)
}

type lstepTagSyncService struct {
	settingsSvc      LstepSettingsService
	ownerRepo        tagSyncOwnerRepo
	vacRepo          tagSyncVaccinationRepo
	medRecordRepo    tagSyncMedicalRecordRepo
	accountRepo      tagSyncAccountingRepo
	tagCacheRepo     LstepTagCacheRepository
	petRepo          tagSyncPetRepo
	prescriptionRepo tagSyncPrescriptionRepo
	checkupRepo      tagSyncCheckupRepo
	errorCounterRepo LstepSyncErrorCounterRepository
	// FEAT-379
	tagCodeRepo     LstepTagCodeMappingRepository
	billingItemRepo tagSyncBillingItemRepo
	// 動的タグ設定 (B/C1/C2/C3)
	tagConfigRepo LstepTagConfigRepository
	// buildClientFn はテスト時にモック Client を注入するためのフック（FEAT-381-2）。
	buildClientFn func(ctx context.Context, clinicID uint64) (lstep.Client, error)
}

// NewLstepTagSyncService は LstepTagSyncService を初期化して返す。
// tagConfigRepo が nil の場合は動的プレフィックス/条件マッピングを使用しない。
func NewLstepTagSyncService(
	settingsSvc LstepSettingsService,
	ownerRepo tagSyncOwnerRepo,
	vacRepo tagSyncVaccinationRepo,
	medRecordRepo tagSyncMedicalRecordRepo,
	accountRepo tagSyncAccountingRepo,
	tagCacheRepo LstepTagCacheRepository,
	petRepo tagSyncPetRepo,
	prescriptionRepo tagSyncPrescriptionRepo,
	checkupRepo tagSyncCheckupRepo,
	errorCounterRepo LstepSyncErrorCounterRepository,
	tagCodeRepo LstepTagCodeMappingRepository,
	billingItemRepo tagSyncBillingItemRepo,
	tagConfigRepo LstepTagConfigRepository,
) LstepTagSyncService {
	return &lstepTagSyncService{
		settingsSvc:      settingsSvc,
		ownerRepo:        ownerRepo,
		vacRepo:          vacRepo,
		medRecordRepo:    medRecordRepo,
		accountRepo:      accountRepo,
		tagCacheRepo:     tagCacheRepo,
		petRepo:          petRepo,
		prescriptionRepo: prescriptionRepo,
		checkupRepo:      checkupRepo,
		errorCounterRepo: errorCounterRepo,
		tagCodeRepo:      tagCodeRepo,
		billingItemRepo:  billingItemRepo,
		tagConfigRepo:    tagConfigRepo,
	}
}

// buildClient はクリニック設定から lstep.Client を構築する。
// 同期無効（is_sync_enabled=false）または API キー未設定の場合は nil, nil を返す（スキップ）。
func (s *lstepTagSyncService) buildClient(ctx context.Context, clinicID uint64) (lstep.Client, error) {
	if s.buildClientFn != nil {
		return s.buildClientFn(ctx, clinicID)
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

func (s *lstepTagSyncService) shouldSkipSync(ctx context.Context, clinicID uint64) (bool, error) {
	if s.settingsSvc == nil {
		return true, nil
	}
	enabled, err := s.settingsSvc.IsSyncEnabled(ctx, clinicID)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to check lstep sync enabled")
	}
	if !enabled {
		slog.InfoContext(ctx, "lstep sync skipped: clinic sync disabled", "clinic_id", clinicID)
		return true, nil
	}
	return false, nil
}

// checkOptOut は飼い主のオプトアウト状態を確認する。
// オプトアウト済みの場合は true を返す（呼び出し元でスキップする）。
func (s *lstepTagSyncService) checkOptOut(ctx context.Context, clinicID, ownerID uint64) (bool, *model.Owner, error) {
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		return false, nil, apperrors.Wrap(err, "failed to find owner")
	}
	if owner.LstepOptOut {
		slog.InfoContext(ctx, "lstep sync skipped: owner opted out", "clinic_id", clinicID, "owner_id", ownerID)
		return true, nil, nil
	}
	return false, owner, nil
}
