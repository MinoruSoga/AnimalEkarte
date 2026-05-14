package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CPMStage は顧客ポートフォリオ管理ステージ（BE-004）。
// 仕様: docs/line/lstep-integration.md Section 7.2
type CPMStage string

const (
	// FEAT-375: 連続失敗 N 回で付与するエラー除外タグ
	lstepErrorTag           = "EXCL_カルテ連携エラー"
	lstepSyncErrorThreshold = 5

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
	TotalVisitCount int64 // 累計来院回数
}

// CalculateCPMStageV2 は累計来院回数ベースの V2 CPM ステージを計算する（Q19 確定 2026-05-08）。
// 仕様: 1回→出会い / 2-3回→これから / 4-7回→いいかんじ / 8-12回→ファミリー / 13回以上→ノア
func CalculateCPMStageV2(d CPMStageV2Input) CPMStageV2 {
	switch {
	case d.TotalVisitCount >= 13:
		return CPMStageV2Noah
	case d.TotalVisitCount >= 8:
		return CPMStageV2Family
	case d.TotalVisitCount >= 4:
		return CPMStageV2Good
	case d.TotalVisitCount >= 2:
		return CPMStageV2Coming
	default: // 0 または 1 回
		return CPMStageV2Encounter
	}
}

// CPMData は CPM ステージ計算に必要な集計データ。
type CPMData struct {
	TotalVisitCount      int64
	AnnualVisitCount     int64
	DaysSinceVisit       int   // 最終来院からの経過日数（来院なし = -1）
	LTVAmount            int64 // 支払済み累計金額（円）
	FirstVisitDaysSince  int   // 初来院からの経過日数＝在籍期間（来院なし = -1）
	MaxSingleVisitAmount int64 // 単回最大支払い額（cpm_spot 判定用）
}

// CalculateCPMStage は純粋関数として CPM ステージを計算する（BE-004）。
// 仕様: docs/line/lstep-integration.md Section 7.2
func CalculateCPMStage(d CPMData) CPMStage {
	// cpm_dormant: 最終来院から240日超または来院なし（最優先）
	if d.DaysSinceVisit < 0 || d.DaysSinceVisit >= 240 {
		return CPMStageDormant
	}
	// cpm_noah: 在籍1年以上、年間3回以上、LTV 80,000円以上
	// V1 cpm_noah は簡略 3 条件 (LTV/visit/在籍) で判定する。
	// 仕様書 §3 の 5 条件記述は V2 詳細化対象であり、V1 では参照しない。
	// PO-QA Q29 (2026-05-08 設計確定) — V1 簡略判定の役割分担を明記。
	// 詳細条件追加は V2 (CalculateCPMStageV2) で対応すること。
	if d.FirstVisitDaysSince >= 365 && d.AnnualVisitCount >= 3 && d.LTVAmount >= 80_000 {
		return CPMStageNoah
	}
	// cpm_core: 在籍180日以上、年間2回以上、LTV 50,000円以上
	if d.FirstVisitDaysSince >= 180 && d.AnnualVisitCount >= 2 && d.LTVAmount >= 50_000 {
		return CPMStageCore
	}
	// cpm_spot: 単回高額（30,000円以上）かつ90日超来院なし
	if d.MaxSingleVisitAmount >= 30_000 && d.DaysSinceVisit > 90 {
		return CPMStageSpot
	}
	// cpm_growing: 初診から90日以内 AND 2〜3回来院 AND LTV 20,000〜50,000円未満 (仕様書整合)
	if d.FirstVisitDaysSince >= 0 && d.FirstVisitDaysSince <= 90 &&
		d.TotalVisitCount >= 2 && d.TotalVisitCount <= 3 &&
		d.LTVAmount >= 20_000 && d.LTVAmount < 50_000 {
		return CPMStageGrowing
	}
	// cpm_encounter: 来院1回 AND LTV 20,000円未満（仕様書 §3 明示判定）
	if d.TotalVisitCount == 1 && d.LTVAmount < 20_000 {
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
	// SyncDormantTags は最終来院からの経過日数に基づき dormant_* タグを差分同期する（BE-005）。
	// daysSinceLastVisit < 0 は来院なしを表す。
	SyncDormantTags(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int) error
	// ResyncOwnerVaccineTags は飼い主の生存ワクチン記録から vaccine_* タグを再構築する（ISSUE-004）。
	// 種別ごとに最新の接種日のみタグを保持する。レコードが0件の場合は全 vaccine_* タグを解除する。
	ResyncOwnerVaccineTags(ctx context.Context, clinicID, ownerID uint64) error
	// ResyncOwnerCheckupTags は飼い主の生存健診記録から checkup_done_* / next_checkup_* タグを再構築する（ISSUE-004）。
	// 種別ごとに最新の検査日のみ保持。next_checkup は最新の next_date 1件のみ保持。
	ResyncOwnerCheckupTags(ctx context.Context, clinicID, ownerID uint64) error
	// SyncCPMStageTagV2 は来院回数ベース V2 CPM ステージタグを同期する（FEAT-377）。
	SyncCPMStageTagV2(ctx context.Context, clinicID, ownerID uint64) error
	// SyncLTVTopPercent は LTV 上位 20% の飼い主に LTV_上位20 タグを付与し、
	// それ以外の飼い主から解除する（FEAT-377）。
	// 処理件数と個別エラーのスライスを返す（全体は失敗しない）。
	SyncLTVTopPercent(ctx context.Context, clinicID uint64) (int, []error)
	// SyncVisitDormantTags は最終来院経過日数に基づき VISIT_* タグを差分同期する（FEAT-377）。
	// VISIT タグは重複付与可（複数閾値を同時保持）。daysSinceLastVisit < 0 は来院なしを表す。
	SyncVisitDormantTags(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int) error
	// SyncPetSpeciesTags は飼い主の生存ペット種別タグ（PET_犬あり / PET_猫あり）を同期する（FEAT-377）。
	SyncPetSpeciesTags(ctx context.Context, clinicID, ownerID uint64) error
	// SyncSeniorTag は飼い主の生存ペットに 7 歳以上の犬猫がいる場合 PET_シニア対象 タグを付与する（FEAT-377）。
	SyncSeniorTag(ctx context.Context, clinicID, ownerID uint64) error
	// SyncExclusionTags は配信停止条件（opt-out / 会員ステータス / 全ペット死亡）に基づき
	// EXCL_配信停止 タグを同期する（FEAT-377）。
	// 注: checkOptOut は呼ばない（このメソッド自体が opt-out 判定の実装）。
	SyncExclusionTags(ctx context.Context, clinicID, ownerID uint64) error
	// SyncHealthcheckTags は健診履歴に基づき HLTH_健診あり / HLTH_健診未受診 タグを同期する（FEAT-379）。
	// SPEC-002 Q5 確定待ち: HealthCheckupCodes が空の場合は noop。
	SyncHealthcheckTags(ctx context.Context, clinicID, ownerID uint64) error
	// SyncAnnual4CheckupTag は年2回以上来院かつ健診履歴がある飼い主に HLTH_年4回候補 タグを付与する（FEAT-379）。
	// SPEC-002 Q5 確定待ち: HealthCheckupCodes が空の場合は noop。
	SyncAnnual4CheckupTag(ctx context.Context, clinicID, ownerID uint64) error
	// SyncVaccineDeadlineTag はワクチン次回予定日が VaccineDeadlineDays 以内に迫っている場合
	// PREV_ワクチン期限 タグを付与し、それ以外は解除する（FEAT-379）。
	SyncVaccineDeadlineTag(ctx context.Context, clinicID, ownerID uint64) error
	// SyncFilariaTag はフィラリア検査・予防薬処方履歴に基づき PREV_フィラリア未完了 タグを同期する（FEAT-379）。
	// SPEC-002 Q5 確定待ち: FilariaTestCodes / FilariaPrescriptionCodes が共に空の場合は noop。
	SyncFilariaTag(ctx context.Context, clinicID, ownerID uint64) error
	// SyncFleaTickTag はノミ・マダニ駆除薬処方履歴に基づき PREV_ノミダニ対象 タグを同期する（FEAT-379）。
	// SPEC-002 Q5 確定待ち: FleaTickPrescriptionCodes が空の場合は noop。
	SyncFleaTickTag(ctx context.Context, clinicID, ownerID uint64) error
	// SyncFoodPurchaseTag はフード購入履歴に基づき LTV_フード購入あり タグを同期する（FEAT-379）。
	// SPEC-002 Q5 確定待ち: FoodPurchaseCodes が空の場合は noop。
	SyncFoodPurchaseTag(ctx context.Context, clinicID, ownerID uint64) error
	// SyncHealthPreventionTagsForClinic は指定クリニックの全飼い主に対して
	// 健診・予防・物販タグを一括同期する（FEAT-379 バッチエントリポイント）。
	// 処理件数と個別エラーのスライスを返す（全体は失敗しない）。
	SyncHealthPreventionTagsForClinic(ctx context.Context, clinicID uint64) (int, []error)
}

type lstepTagSyncService struct {
	settingsSvc      LstepSettingsService
	ownerRepo        repository.OwnerRepository
	vacRepo          repository.VaccinationRepository
	medRecordRepo    repository.MedicalRecordRepository
	accountRepo      repository.AccountingRepository
	tagCacheRepo     repository.LstepTagCacheRepository
	petRepo          repository.PetRepository
	prescriptionRepo repository.PrescriptionRepository
	checkupRepo      repository.CheckupRepository
	reservationRepo  repository.ReservationCRUDRepository
	errorCounterRepo repository.LstepSyncErrorCounterRepository
	// FEAT-379
	tagCodeRepo     repository.LstepTagCodeMappingRepository
	billingItemRepo repository.BillingItemRepository
	// buildClientFn はテスト時にモック Client を注入するためのフック（FEAT-381-2）。
	buildClientFn func(ctx context.Context, clinicID uint64) (lstep.Client, error)
}

// NewLstepTagSyncService は LstepTagSyncService を初期化して返す。
func NewLstepTagSyncService(
	settingsSvc LstepSettingsService,
	ownerRepo repository.OwnerRepository,
	vacRepo repository.VaccinationRepository,
	medRecordRepo repository.MedicalRecordRepository,
	accountRepo repository.AccountingRepository,
	tagCacheRepo repository.LstepTagCacheRepository,
	petRepo repository.PetRepository,
	prescriptionRepo repository.PrescriptionRepository,
	checkupRepo repository.CheckupRepository,
	reservationRepo repository.ReservationCRUDRepository,
	errorCounterRepo repository.LstepSyncErrorCounterRepository,
	tagCodeRepo repository.LstepTagCodeMappingRepository,
	billingItemRepo repository.BillingItemRepository,
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
		reservationRepo:  reservationRepo,
		errorCounterRepo: errorCounterRepo,
		tagCodeRepo:      tagCodeRepo,
		billingItemRepo:  billingItemRepo,
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

// SyncVaccineTag はワクチン接種記録からタグを同期する（BE-003）。
// 接種種別（dog/cat）とラビーズを date 付きタグとして付与する。
func (s *lstepTagSyncService) SyncVaccineTag(ctx context.Context, clinicID, ownerID, vaccinationID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for vaccine tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	vac, err := s.vacRepo.FindByID(ctx, clinicID, vaccinationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find vaccination for tag sync", "error", err)
		return apperrors.Wrap(err, "failed to find vaccination")
	}

	tags := vaccineTagNames(vac)
	if len(tags) == 0 {
		return nil
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID

	// 同一カテゴリの古い日付タグを解除してから新タグを付与（ISSUE-006）
	newTagSet := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		newTagSet[t] = struct{}{}
	}
	apiFailed := s.removeStaleTagsByPrefixes(ctx, client, clinicID, ownerID, lineUserID,
		[]string{"vaccine_dog_", "vaccine_cat_", "vaccine_rabies_"}, newTagSet)

	for _, tag := range tags {
		if addErr := client.AddTag(ctx, lineUserID, tag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add vaccine tag", "error", addErr, "tag", tag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, fmt.Sprintf("failed to add vaccine tag %s", tag))
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto", ""); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert tag cache", "error", cacheErr, "tag", tag)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// vaccineTagNames はワクチン接種記録から付与すべきタグ名一覧を返す。
func vaccineTagNames(vac *model.Vaccination) []string {
	if vac.Vaccine == nil {
		return nil
	}
	date := vac.Date.Format("2006-01-02")
	var tags []string

	species := vac.Vaccine.Species
	if species != nil {
		switch *species {
		case model.VaccineSpeciesDog:
			tags = append(tags, "vaccine_dog_"+date)
		case model.VaccineSpeciesCat:
			tags = append(tags, "vaccine_cat_"+date)
		case model.VaccineSpeciesBoth:
			tags = append(tags, "vaccine_dog_"+date, "vaccine_cat_"+date)
		}
	}

	if isRabiesVaccine(vac.Vaccine.Name) {
		tags = append(tags, "vaccine_rabies_"+date)
	}
	return tags
}

func isRabiesVaccine(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "rabies") || strings.Contains(name, "狂犬病")
}

// SyncOwnerAnimalClassificationTags は飼い主の動物分類タグを同期する（BE-005）。
func (s *lstepTagSyncService) SyncOwnerAnimalClassificationTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for animal classification tags", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	pets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find living pets for classification tags", "error", err)
		return apperrors.Wrap(err, "failed to find living pets")
	}

	var hasDog, hasCat bool
	for i := range pets {
		p := &pets[i]
		if p.AnimalSpecies == nil {
			continue
		}
		if strings.Contains(p.AnimalSpecies.Name, "犬") {
			hasDog = true
		}
		if strings.Contains(p.AnimalSpecies.Name, "猫") {
			hasCat = true
		}
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID

	var newTag string
	switch {
	case hasDog && hasCat:
		newTag = "has_both"
	case hasDog:
		newTag = "has_dog"
	case hasCat:
		newTag = "has_cat"
	}

	// 旧分類タグを全削除してから新タグを付与
	apiFailed := false
	for _, old := range []string{"has_dog", "has_cat", "has_both"} {
		if old == newTag {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, old); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove old classification tag", "error", delErr, "tag", old)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, old)
		}
	}

	if newTag == "" {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}

	if addErr := client.AddTag(ctx, lineUserID, newTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add classification tag", "error", addErr, "tag", newTag)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, fmt.Sprintf("failed to add classification tag %s", newTag))
	}
	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, newTag, "auto", ""); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert classification tag cache", "error", cacheErr, "tag", newTag)
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncPetBasicInfoTags は全生存ペットの基本情報タグを同期する（BE-005）。
func (s *lstepTagSyncService) SyncPetBasicInfoTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for pet basic info tags", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	pets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find living pets for basic info tags", "error", err)
		return apperrors.Wrap(err, "failed to find living pets")
	}

	newTags := buildPetBasicInfoTags(pets)

	cachedTags, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag cache for pet basic info sync", "error", err)
		return apperrors.Wrap(err, "failed to find tag cache")
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID

	oldSet := make(map[string]struct{})
	for _, c := range cachedTags {
		if isPetBasicInfoTag(c.TagName) {
			oldSet[c.TagName] = struct{}{}
		}
	}

	newSet := make(map[string]struct{}, len(newTags))
	for _, t := range newTags {
		newSet[t] = struct{}{}
	}

	// 不要になったタグを削除
	apiFailed := false
	for old := range oldSet {
		if _, keep := newSet[old]; keep {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, old); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove stale pet basic info tag", "error", delErr, "tag", old)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, old)
		}
	}

	// 新規タグを追加
	for _, tag := range newTags {
		if _, exists := oldSet[tag]; exists {
			continue
		}
		if addErr := client.AddTag(ctx, lineUserID, tag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add pet basic info tag", "error", addErr, "tag", tag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, fmt.Sprintf("failed to add pet basic info tag %s", tag))
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto", ""); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert pet basic info tag cache", "error", cacheErr, "tag", tag)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// buildPetBasicInfoTags は生存ペット一覧から基本情報タグ一覧を生成する。
func buildPetBasicInfoTags(pets []model.Pet) []string {
	tagSet := make(map[string]struct{})
	var hasNeutered, hasIntact bool

	for i := range pets {
		p := &pets[i]
		fallback := "breed_mix_other"
		if p.AnimalSpecies != nil {
			if strings.Contains(p.AnimalSpecies.Name, "犬") {
				fallback = "breed_mix_dog"
			} else if strings.Contains(p.AnimalSpecies.Name, "猫") {
				fallback = "breed_mix_cat"
			}
		}
		tagSet[lstep.BreedTagName(p.Breed, fallback)] = struct{}{}

		switch p.Gender {
		case model.PetGenderMale:
			tagSet["sex_male"] = struct{}{}
		case model.PetGenderFemale:
			tagSet["sex_female"] = struct{}{}
		default:
			tagSet["sex_unknown"] = struct{}{}
		}

		if p.BirthDate != nil {
			tagSet["pet_birthday_"+p.BirthDate.Format("01-02")] = struct{}{}
			tagSet["birth_year_"+p.BirthDate.Format("2006")] = struct{}{}
		}

		if p.NeuteredDate != nil {
			hasNeutered = true
		} else {
			hasIntact = true
		}
	}

	if hasNeutered {
		tagSet["spay_neutered"] = struct{}{}
	}
	if hasIntact {
		tagSet["intact"] = struct{}{}
	}

	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}
	return tags
}

// isPetBasicInfoTag は BE-005 ペット基本情報カテゴリのタグかを判定する。
func isPetBasicInfoTag(tag string) bool {
	return strings.HasPrefix(tag, "breed_") ||
		strings.HasPrefix(tag, "sex_") ||
		strings.HasPrefix(tag, "pet_birthday_") ||
		strings.HasPrefix(tag, "birth_year_") ||
		tag == "spay_neutered" ||
		tag == "intact"
}

// SyncVisitCompletionTags は診療完了時の来院・LTV タグを同期する（BE-004）。
func (s *lstepTagSyncService) SyncVisitCompletionTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for visit tags sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	summary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary", "error", err)
		return apperrors.Wrap(err, "failed to find visit summary")
	}

	ltv, err := s.accountRepo.SumPaidByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to sum paid amount", "error", err)
		return apperrors.Wrap(err, "failed to sum paid amount")
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID
	tags := buildVisitTags(summary, ltv)

	// 同一カテゴリ1タグ保持: 古い ltv_amount_* / visit_count_annual_* / first_visit_* / last_visit_* を解除（ISSUE-006）
	newTagSet := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		newTagSet[t] = struct{}{}
	}
	apiFailed := s.removeStaleTagsByPrefixes(ctx, client, clinicID, ownerID, lineUserID,
		[]string{"ltv_amount_", "visit_count_annual_", "first_visit_", "last_visit_"}, newTagSet)

	for _, tag := range tags {
		if addErr := client.AddTag(ctx, lineUserID, tag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add visit tag", "error", addErr, "tag", tag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, fmt.Sprintf("failed to add visit tag %s", tag))
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto", ""); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert tag cache", "error", cacheErr, "tag", tag)
		}
	}

	// 来院完了でキャッシュ経由の予約関連タグ（reserved_* / canceled_visit / no_show_*）とdormantタグを削除
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for visit completion cleanup", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	for _, c := range cached {
		if isDormantTag(c.TagName) || isVisitDormantTag(c.TagName) || c.TagName == "cpm_dormant" {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove reservation/dormant tag on visit completion", "error", delErr, "tag", c.TagName)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
			} else {
				_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName)
			}
		}
	}
	// レガシータグ（名前変更前との互換）
	for _, staleTag := range []string{"dormant", "noshow", "reserved"} {
		if delErr := client.RemoveTag(ctx, lineUserID, staleTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove stale legacy tag", "error", delErr, "tag", staleTag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, staleTag)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// buildVisitTags は来院サマリーから付与するタグ一覧を生成する。
func buildVisitTags(summary *repository.OwnerVisitSummary, ltv int64) []string {
	var tags []string
	if summary.FirstVisitAt != nil {
		tags = append(tags, "first_visit_"+summary.FirstVisitAt.Format("2006-01-02"))
	}
	if summary.LastVisitAt != nil {
		tags = append(tags, "last_visit_"+summary.LastVisitAt.Format("2006-01-02"))
	}
	tags = append(tags, ltvBracketTag(ltv), visitCountAnnualTag(summary.AnnualCount))
	return tags
}

func ltvBracketTag(ltv int64) string {
	switch {
	case ltv >= 80_000:
		return "ltv_amount_8"
	case ltv >= 50_000:
		return "ltv_amount_5"
	case ltv >= 20_000:
		return "ltv_amount_2"
	default:
		return "ltv_amount_0"
	}
}

func visitCountAnnualTag(count int64) string {
	switch {
	case count >= 10:
		return "visit_count_annual_10"
	case count >= 5:
		return "visit_count_annual_5"
	case count >= 3:
		return "visit_count_annual_3"
	case count >= 2:
		return "visit_count_annual_2"
	default:
		return "visit_count_annual_1"
	}
}

// isDormantTag は dormant 系タグかどうかを判定する。
func isDormantTag(tag string) bool {
	return tag == "dormant_180d" || tag == "dormant_210d" || tag == "dormant_240d" || tag == "dormant_365d"
}

// isVisitDormantTag は VISIT_* 系タグかどうかを判定する（FEAT-377）。
func isVisitDormantTag(tag string) bool {
	return tag == visitTag120 || tag == visitTag180 || tag == visitTag220 || tag == visitTag240
}

// removeStaleTagsByPrefixes はキャッシュ内で指定プレフィックスに一致する古いタグを Lステップから解除する（ISSUE-006）。
// 同一カテゴリ1タグ保持ルールのための前処理として呼ぶ。
func (s *lstepTagSyncService) removeStaleTagsByPrefixes(ctx context.Context, client lstep.Client, clinicID, ownerID uint64, lineUserID string, prefixes []string, skipTags map[string]struct{}) bool {
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for stale cleanup", "error", err)
		return false
	}
	apiFailed := false
	for _, c := range cached {
		if _, skip := skipTags[c.TagName]; skip {
			continue
		}
		for _, pfx := range prefixes {
			if strings.HasPrefix(c.TagName, pfx) {
				if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
					slog.ErrorContext(ctx, "failed to remove stale tag", "error", delErr, "tag", c.TagName)
					s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
					apiFailed = true
				} else {
					_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName)
				}
				break
			}
		}
	}
	return apiFailed
}

// SyncNextVisitTag は次回来院推奨日タグを同期する（BE-006）。
// 最新カルテの next_visit_recommended_date を参照し、古い next_visit_* タグを差し替える。
// date が nil の場合はすべての next_visit_* タグを削除する。
func (s *lstepTagSyncService) SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for next visit tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for next visit tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	// 最新カルテの次回来院推奨日を取得
	latest, err := s.medRecordRepo.FindLatestByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find latest medical record for next visit tag", "error", err)
		return apperrors.Wrap(err, "failed to find latest medical record")
	}

	// 既存の next_visit_* タグをキャッシュから取得して削除
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for next visit tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	apiFailed := false
	for _, c := range cached {
		if strings.HasPrefix(c.TagName, "next_visit_") {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove next_visit tag", "error", delErr, "tag", c.TagName)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete next_visit tag cache", "error", delErr, "tag", c.TagName)
			}
		}
	}

	// 新しい next_visit_YYYY-MM-DD タグを付与（日付が設定されている場合のみ）
	if latest == nil || latest.NextVisitRecommendedDate == nil {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}
	newTag := "next_visit_" + latest.NextVisitRecommendedDate.Format("2006-01-02")
	if addErr := client.AddTag(ctx, lineUserID, newTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add next_visit tag", "error", addErr, "tag", newTag)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add next_visit tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, newTag, "auto", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert next_visit tag cache", "error", upsertErr)
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncCPMStageTag は CPM ステージタグを同期する（BE-011）。
// cpm_version が "v2" のクリニックは SyncCPMStageTagV2 に委譲する（Q19）。
func (s *lstepTagSyncService) SyncCPMStageTag(ctx context.Context, clinicID, ownerID uint64) error {
	version, vErr := s.settingsSvc.GetCPMVersion(ctx, clinicID)
	if vErr != nil {
		slog.ErrorContext(ctx, "failed to get cpm_version for CPM dispatch", "error", vErr)
		return apperrors.Wrap(vErr, "failed to get cpm_version")
	}
	if version == "v2" {
		return s.SyncCPMStageTagV2(ctx, clinicID, ownerID)
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for CPM tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	summary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary for CPM", "error", err)
		return apperrors.Wrap(err, "failed to find visit summary")
	}

	ltv, err := s.accountRepo.SumPaidByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to sum paid amount for CPM", "error", err)
		return apperrors.Wrap(err, "failed to sum paid amount")
	}

	maxSingle, err := s.accountRepo.MaxSingleVisitAmountByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get max single visit amount for CPM", "error", err)
		return apperrors.Wrap(err, "failed to get max single visit amount")
	}

	daysSince := -1
	if summary.LastVisitAt != nil {
		daysSince = int(time.Since(*summary.LastVisitAt).Hours() / 24)
	}
	firstVisitDaysSince := 0
	if summary.FirstVisitAt != nil {
		firstVisitDaysSince = int(time.Since(*summary.FirstVisitAt).Hours() / 24)
	}

	stage := CalculateCPMStage(CPMData{
		TotalVisitCount:      summary.TotalCount,
		AnnualVisitCount:     summary.AnnualCount,
		DaysSinceVisit:       daysSince,
		LTVAmount:            ltv,
		FirstVisitDaysSince:  firstVisitDaysSince,
		MaxSingleVisitAmount: maxSingle,
	})

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID

	// 旧ステージタグをすべて削除してから新ステージを付与
	apiFailed := false
	for _, old := range allCPMStages {
		if string(old) == string(stage) {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, string(old)); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove old CPM stage tag", "error", delErr, "tag", old)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, string(old))
		}
	}

	if addErr := client.AddTag(ctx, lineUserID, string(stage)); addErr != nil {
		slog.ErrorContext(ctx, "failed to add CPM stage tag", "error", addErr, "stage", stage)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add CPM stage tag")
	}
	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, string(stage), "auto", ""); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert CPM stage tag cache", "error", cacheErr)
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncPrescriptionTag は飼い主の全アクティブ処方を取得し、補充推奨日が最も遅い処方に基づいて
// refill_due_* タグを更新する（BE-009）。duration_days < 7 の場合は prescribed_at + 1 日を使用する。
// 最新の refill_due が現在日時を過ぎている場合は refill_due_* タグをすべて削除して終了する。
func (s *lstepTagSyncService) SyncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for prescription tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for prescription tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	prescriptions, err := s.prescriptionRepo.FindActiveByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load prescriptions for tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load prescriptions")
	}

	// 補充推奨日の計算: prescribed_at + duration_days - 7。duration_days < 7 なら prescribed_at + 1
	var latestRefillDue *time.Time
	for i := range prescriptions {
		p := &prescriptions[i]
		var refill time.Time
		if p.DurationDays < 7 {
			refill = p.PrescribedAt.AddDate(0, 0, 1)
		} else {
			refill = p.PrescribedAt.AddDate(0, 0, p.DurationDays-7)
		}
		if latestRefillDue == nil || refill.After(*latestRefillDue) {
			t := refill
			latestRefillDue = &t
		}
	}

	// 旧 refill_due_* タグをキャッシュ経由で削除
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for prescription tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	apiFailed := false
	for _, c := range cached {
		if strings.HasPrefix(c.TagName, "refill_due_") {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale refill_due tag", "error", delErr, "tag", c.TagName)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete refill_due tag cache", "error", delErr, "tag", c.TagName)
			}
		}
	}

	// 補充推奨日が未来にある場合のみ新タグを付与
	if latestRefillDue == nil || !latestRefillDue.After(time.Now()) {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}
	newTag := fmt.Sprintf("refill_due_%s", latestRefillDue.Format("2006-01-02"))
	if addErr := client.AddTag(ctx, lineUserID, newTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add refill_due tag", "error", addErr, "tag", newTag)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add refill_due tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, newTag, "auto", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert refill_due tag cache", "error", upsertErr)
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncCheckupTag は健診記録の作成・更新時に checkup_done_{typeID}_{YYYY-MM}/next_checkup_* タグを同期する（BE-008）。
// 同一健診種別の古い checkup_done タグを解除してから新タグを付与する。next_checkup_* は最新1件のみ。
func (s *lstepTagSyncService) SyncCheckupTag(ctx context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for checkup tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for checkup tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	// 同一健診種別の古い checkup_done タグ + next_checkup_* タグをキャッシュ経由で削除
	stalePrefix := fmt.Sprintf("checkup_done_%d_", checkupTypeID)
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for checkup tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	apiFailed := false
	for _, c := range cached {
		if strings.HasPrefix(c.TagName, stalePrefix) || strings.HasPrefix(c.TagName, "next_checkup_") {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale checkup tag", "error", delErr, "tag", c.TagName)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete checkup tag cache", "error", delErr, "tag", c.TagName)
			}
		}
	}

	// checkup_done_{typeID}_{YYYY-MM} タグを付与
	checkupTag := fmt.Sprintf("checkup_done_%d_%s", checkupTypeID, checkupDate.Format("2006-01"))
	if addErr := client.AddTag(ctx, lineUserID, checkupTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add checkup tag", "error", addErr, "tag", checkupTag)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add checkup tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, checkupTag, "auto", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert checkup tag cache", "error", upsertErr)
	}

	// next_checkup_YYYY-MM-DD タグを付与（設定時のみ）
	if nextDate != nil {
		nextTag := "next_checkup_" + nextDate.Format("2006-01-02")
		if addErr := client.AddTag(ctx, lineUserID, nextTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add next_checkup tag", "error", addErr, "tag", nextTag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add next_checkup tag")
		}
		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, nextTag, "auto", ""); upsertErr != nil {
			slog.ErrorContext(ctx, "failed to upsert next_checkup tag cache", "error", upsertErr)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// conditionTagMap は疾患コードとLステップタグ名のマッピング（BE-012）。
var conditionTagMap = map[string]string{
	"ckd":      "chronic_ckd",
	"heart":    "chronic_heart",
	"skin":     "chronic_skin",
	"diabetes": "chronic_diabetes",
	"liver":    "chronic_liver",
	"thyroid":  "chronic_thyroid",
	"other":    "chronic_other",
}

// SyncChronicConditionTags は慢性疾患タグを差分同期する（BE-012）。
func (s *lstepTagSyncService) SyncChronicConditionTags(ctx context.Context, clinicID, ownerID uint64, activeConditionCodes []string) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for chronic condition tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for chronic condition tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	// 目標タグセットを構築
	activeTags := make(map[string]bool, len(activeConditionCodes))
	for _, code := range activeConditionCodes {
		if tag, ok := conditionTagMap[code]; ok {
			activeTags[tag] = true
		}
	}

	// 現在の chronic_* キャッシュを取得
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find cached tags for chronic sync", "error", err)
		return apperrors.Wrap(err, "failed to find cached tags")
	}

	existingSet := make(map[string]bool, len(cached))
	apiFailed := false
	for _, t := range cached {
		if !strings.HasPrefix(t.TagName, "chronic_") {
			continue
		}
		existingSet[t.TagName] = true

		// アクティブでないタグを解除
		if activeTags[t.TagName] {
			continue
		}
		if rmErr := client.RemoveTag(ctx, lineUserID, t.TagName); rmErr != nil {
			slog.WarnContext(ctx, "failed to remove chronic tag via lstep api", "tag", t.TagName, "error", rmErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
			continue
		}
		if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, t.TagName); delErr != nil {
			slog.ErrorContext(ctx, "failed to delete chronic tag cache", "tag", t.TagName, "error", delErr)
		}
	}

	// 未付与のアクティブタグを付与
	for tagName := range activeTags {
		if existingSet[tagName] {
			continue
		}
		if addErr := client.AddTag(ctx, lineUserID, tagName); addErr != nil {
			slog.WarnContext(ctx, "failed to add chronic tag via lstep api", "tag", tagName, "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
			continue
		}
		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tagName, "auto", ""); upsertErr != nil {
			slog.ErrorContext(ctx, "failed to upsert chronic tag cache", "tag", tagName, "error", upsertErr)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// ResyncOwnerVaccineTags は飼い主の生存ワクチン記録から vaccine_* タグを再構築する（ISSUE-004）。
// 接種記録の更新・削除後に呼び出すこと。種別ごとに最新の接種日のみタグを保持する。
// 記録が0件の場合は全 vaccine_* タグを解除する。
func (s *lstepTagSyncService) ResyncOwnerVaccineTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for vaccine resync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for vaccine resync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	vaccinations, err := s.vacRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find vaccinations for resync", "error", err)
		return apperrors.Wrap(err, "failed to find vaccinations")
	}

	newTagSet := buildLatestVaccineTagSet(vaccinations)

	apiFailed := s.removeStaleTagsByPrefixes(ctx, client, clinicID, ownerID, lineUserID,
		[]string{"vaccine_dog_", "vaccine_cat_", "vaccine_rabies_"}, newTagSet)

	for tag := range newTagSet {
		if addErr := client.AddTag(ctx, lineUserID, tag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add vaccine tag on resync", "error", addErr, "tag", tag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, fmt.Sprintf("failed to add vaccine tag %s", tag))
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto", ""); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert vaccine tag cache on resync", "error", cacheErr, "tag", tag)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// buildLatestVaccineTagSet は接種記録一覧から「種別ごとの最新接種日」のタグ集合を返す。
// 同一種別に複数記録がある場合は最も新しい date のタグのみを採用する。
func buildLatestVaccineTagSet(vaccinations []model.Vaccination) map[string]struct{} {
	latestByPrefix := make(map[string]time.Time)
	updateLatest := func(prefix string, date time.Time) {
		if cur, ok := latestByPrefix[prefix]; !ok || date.After(cur) {
			latestByPrefix[prefix] = date
		}
	}
	for i := range vaccinations {
		v := &vaccinations[i]
		if v.Vaccine == nil {
			continue
		}
		species := v.Vaccine.Species
		if species != nil {
			switch *species {
			case model.VaccineSpeciesDog:
				updateLatest("vaccine_dog_", v.Date)
			case model.VaccineSpeciesCat:
				updateLatest("vaccine_cat_", v.Date)
			case model.VaccineSpeciesBoth:
				updateLatest("vaccine_dog_", v.Date)
				updateLatest("vaccine_cat_", v.Date)
			}
		}
		if isRabiesVaccine(v.Vaccine.Name) {
			updateLatest("vaccine_rabies_", v.Date)
		}
	}
	tagSet := make(map[string]struct{}, len(latestByPrefix))
	for prefix, date := range latestByPrefix {
		tagSet[prefix+date.Format("2006-01-02")] = struct{}{}
	}
	return tagSet
}

// ResyncOwnerCheckupTags は飼い主の生存健診記録から checkup_done_* / next_checkup_* タグを再構築する（ISSUE-004）。
// 健診の更新・削除後に呼び出すこと。同一健診種別では最新検査日のみ保持。next_checkup は最遠の next_date のみ保持。
func (s *lstepTagSyncService) ResyncOwnerCheckupTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for checkup resync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for checkup resync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	checkups, err := s.checkupRepo.FindByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find checkups for resync", "error", err)
		return apperrors.Wrap(err, "failed to find checkups")
	}

	newTagSet := buildLatestCheckupTagSet(checkups)

	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for checkup resync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	apiFailed := false
	for _, c := range cached {
		if !strings.HasPrefix(c.TagName, "checkup_done_") && !strings.HasPrefix(c.TagName, "next_checkup_") {
			continue
		}
		if _, keep := newTagSet[c.TagName]; keep {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove stale checkup tag", "error", delErr, "tag", c.TagName)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
			continue
		}
		if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
			slog.ErrorContext(ctx, "failed to delete stale checkup tag cache", "error", delErr, "tag", c.TagName)
		}
	}

	for tag := range newTagSet {
		if addErr := client.AddTag(ctx, lineUserID, tag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add checkup tag on resync", "error", addErr, "tag", tag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, fmt.Sprintf("failed to add checkup tag %s", tag))
		}
		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto", ""); upsertErr != nil {
			slog.ErrorContext(ctx, "failed to upsert checkup tag cache on resync", "error", upsertErr, "tag", tag)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// buildLatestCheckupTagSet は健診記録から「種別ごとの最新検査日」「最遠の next_date」のタグ集合を返す。
func buildLatestCheckupTagSet(checkups []model.Checkup) map[string]struct{} {
	latestByType := make(map[uint64]time.Time)
	var latestNext *time.Time
	for i := range checkups {
		c := &checkups[i]
		if cur, ok := latestByType[c.CheckupTypeID]; !ok || c.Date.After(cur) {
			latestByType[c.CheckupTypeID] = c.Date
		}
		if c.NextDate != nil {
			if latestNext == nil || c.NextDate.After(*latestNext) {
				t := *c.NextDate
				latestNext = &t
			}
		}
	}
	tagSet := make(map[string]struct{}, len(latestByType)+1)
	for typeID, date := range latestByType {
		tagSet[fmt.Sprintf("checkup_done_%d_%s", typeID, date.Format("2006-01"))] = struct{}{}
	}
	if latestNext != nil {
		tagSet["next_checkup_"+latestNext.Format("2006-01-02")] = struct{}{}
	}
	return tagSet
}

// SyncDormantTags は最終来院からの経過日数に基づき dormant_* タグを差分同期する（BE-005）。
func (s *lstepTagSyncService) SyncDormantTags(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for dormant tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for dormant tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	// 付与すべき dormant タグを決定（閾値は clinic_settings から取得、Q21）
	thresholds, tErr := s.settingsSvc.GetDormantThresholds(ctx, clinicID)
	if tErr != nil {
		slog.ErrorContext(ctx, "SyncDormantTags: failed to get dormant thresholds", "clinic_id", clinicID, "error", tErr)
		return apperrors.Wrap(tErr, "failed to get dormant thresholds")
	}
	var targetTag string
	switch {
	case daysSinceLastVisit < 0 || daysSinceLastVisit >= thresholds.Stage365:
		targetTag = "dormant_365d"
	case daysSinceLastVisit >= thresholds.Stage240:
		targetTag = "dormant_240d"
	case daysSinceLastVisit >= thresholds.Stage210:
		targetTag = "dormant_210d"
	case daysSinceLastVisit >= thresholds.Stage180:
		targetTag = "dormant_180d"
	}

	// 240日以上の休眠では CPM 休眠ステージタグ（cpm_dormant）も付与する
	needsCpmDormant := targetTag == "dormant_240d" || targetTag == "dormant_365d"

	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for dormant sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}

	apiFailed := false
	for _, c := range cached {
		if isDormantTag(c.TagName) && c.TagName != targetTag {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale dormant tag", "error", delErr, "tag", c.TagName)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
			} else {
				_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName)
			}
		}
		// cpm_dormant は 240d 未満の場合に削除する
		if c.TagName == "cpm_dormant" && !needsCpmDormant {
			if delErr := client.RemoveTag(ctx, lineUserID, "cpm_dormant"); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale cpm_dormant tag", "error", delErr)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
			} else {
				_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, "cpm_dormant")
			}
		}
	}

	if targetTag == "" {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}

	if addErr := client.AddTag(ctx, lineUserID, targetTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add dormant tag", "error", addErr, "tag", targetTag)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add dormant tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, targetTag, "auto", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert dormant tag cache", "error", upsertErr)
	}

	// 240日以上の場合は cpm_dormant も付与する
	if needsCpmDormant {
		if addErr := client.AddTag(ctx, lineUserID, "cpm_dormant"); addErr != nil {
			slog.ErrorContext(ctx, "failed to add cpm_dormant tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add cpm_dormant tag")
		} else if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, "cpm_dormant", "auto", ""); upsertErr != nil {
			slog.ErrorContext(ctx, "failed to upsert cpm_dormant tag cache", "error", upsertErr)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// notifyAPIFailure は Lステップ API 呼び出し失敗時にカウンターを更新し、
// 閾値（lstepSyncErrorThreshold）に達した場合は EXCL_カルテ連携エラー タグを付与する。
// エラーはログのみ — 呼び出し元に伝播させない（best-effort）。
func (s *lstepTagSyncService) notifyAPIFailure(ctx context.Context, client lstep.Client, clinicID, ownerID uint64, lineUserID string) {
	if s.errorCounterRepo == nil {
		return
	}
	newCount, err := s.errorCounterRepo.IncrementFailure(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to increment lstep error counter", "error", err, "clinic_id", clinicID, "owner_id", ownerID)
		return
	}
	if newCount < lstepSyncErrorThreshold {
		return
	}
	if addErr := client.AddTag(ctx, lineUserID, lstepErrorTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add EXCL error tag", "error", addErr, "owner_id", ownerID)
		return
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, lstepErrorTag, "auto", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert EXCL error tag cache", "error", upsertErr)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// FEAT-377: CPM V2 / LTV 上位 20% / VISIT / PET / 除外タグ同期
// ─────────────────────────────────────────────────────────────────────────────

// SyncCPMStageTagV2 は来院回数ベース V2 CPM ステージタグを同期する（FEAT-377）。
func (s *lstepTagSyncService) SyncCPMStageTagV2(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for CPM V2 tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	summary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary for CPM V2", "error", err)
		return apperrors.Wrap(err, "failed to find visit summary")
	}

	stage := CalculateCPMStageV2(CPMStageV2Input{
		TotalVisitCount: summary.TotalCount,
	})

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID
	apiFailed := false
	for _, old := range allCPMV2Stages {
		if old == stage {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, string(old)); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove old CPM V2 stage tag", "error", delErr, "tag", old)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, string(old))
		}
	}

	if addErr := client.AddTag(ctx, lineUserID, string(stage)); addErr != nil {
		slog.ErrorContext(ctx, "failed to add CPM V2 stage tag", "error", addErr, "stage", stage)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add CPM V2 stage tag")
	}
	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, string(stage), "auto", ""); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert CPM V2 stage tag cache", "error", cacheErr)
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncLTVTopPercent は LTV 上位 20% の飼い主に LTV_上位20 タグを付与し、
// それ以外から解除する（FEAT-377）。
func (s *lstepTagSyncService) SyncLTVTopPercent(ctx context.Context, clinicID uint64) (int, []error) {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return 0, []error{err}
	} else if skip {
		return 0, nil
	}

	revenues, err := s.accountRepo.FindOwnersByAnnualRevenue(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to find revenues", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by annual revenue")}
	}

	topN := 0
	if len(revenues) > 0 {
		topN = (len(revenues)*20 + 99) / 100
	}
	topOwnerIDs := make(map[uint64]struct{}, topN)
	for i := 0; i < topN; i++ {
		topOwnerIDs[revenues[i].OwnerID] = struct{}{}
	}

	owners, err := s.ownerRepo.FindAllWithLineUserID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to find owners", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners with line user id")}
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return 0, []error{err}
	}
	if client == nil {
		return 0, nil
	}

	var errs []error
	count := 0
	for i := range owners {
		owner := &owners[i]
		if owner.LineUserID == nil || *owner.LineUserID == "" {
			continue
		}
		lineUserID := *owner.LineUserID
		_, isTop := topOwnerIDs[owner.ID]
		if isTop {
			if addErr := client.AddTag(ctx, lineUserID, ltvTop20Tag); addErr != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to add tag", "owner_id", owner.ID, "error", addErr)
				s.notifyAPIFailure(ctx, client, clinicID, owner.ID, lineUserID)
				errs = append(errs, apperrors.Wrap(addErr, fmt.Sprintf("failed to add LTV top20 tag for owner %d", owner.ID)))
				continue
			}
			if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, owner.ID, ltvTop20Tag, "auto", ""); cacheErr != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to upsert tag cache", "owner_id", owner.ID, "error", cacheErr)
			}
		} else {
			cached, cacheErr := s.tagCacheRepo.FindByOwner(ctx, clinicID, owner.ID)
			if cacheErr != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to load cache", "owner_id", owner.ID, "error", cacheErr)
				errs = append(errs, apperrors.Wrap(cacheErr, fmt.Sprintf("failed to load tag cache for owner %d", owner.ID)))
				continue
			}
			hasTag := false
			for _, c := range cached {
				if c.TagName == ltvTop20Tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				s.notifyAPISuccess(ctx, client, clinicID, owner.ID, lineUserID)
				count++
				continue
			}
			if delErr := client.RemoveTag(ctx, lineUserID, ltvTop20Tag); delErr != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to remove tag", "owner_id", owner.ID, "error", delErr)
				s.notifyAPIFailure(ctx, client, clinicID, owner.ID, lineUserID)
				errs = append(errs, apperrors.Wrap(delErr, fmt.Sprintf("failed to remove LTV top20 tag for owner %d", owner.ID)))
				continue
			}
			if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, owner.ID, ltvTop20Tag); delCacheErr != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to delete tag cache", "owner_id", owner.ID, "error", delCacheErr)
			}
		}
		s.notifyAPISuccess(ctx, client, clinicID, owner.ID, lineUserID)
		count++
	}
	return count, errs
}

// SyncVisitDormantTags は最終来院経過日数に基づき VISIT_* タグを差分同期する（FEAT-377）。
// VISIT タグは重複付与可（複数閾値を同時保持する）。
// visitDormantTagsForDays は経過日数に対して付与すべき VISIT_* タグのスライスを返す純粋関数。
func visitDormantTagsForDays(days int) []string {
	type th struct {
		tag string
		min int
	}
	thresholds := []th{
		{tag: visitTag120, min: 120},
		{tag: visitTag180, min: 180},
		{tag: visitTag220, min: 220},
		{tag: visitTag240, min: 240},
	}
	var tags []string
	for _, t := range thresholds {
		if days > t.min {
			tags = append(tags, t.tag)
		}
	}
	return tags
}

func (s *lstepTagSyncService) SyncVisitDormantTags(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for VISIT dormant tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for VISIT dormant tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	type visitThreshold struct {
		tag  string
		days int
	}
	thresholds := []visitThreshold{
		{tag: visitTag120, days: 120},
		{tag: visitTag180, days: 180},
		{tag: visitTag220, days: 220},
		{tag: visitTag240, days: 240},
	}

	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for VISIT dormant sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	cachedSet := make(map[string]struct{}, len(cached))
	for _, c := range cached {
		cachedSet[c.TagName] = struct{}{}
	}

	apiFailed := false
	for _, th := range thresholds {
		shouldHave := daysSinceLastVisit > th.days
		_, hasCached := cachedSet[th.tag]
		if shouldHave {
			if addErr := client.AddTag(ctx, lineUserID, th.tag); addErr != nil {
				slog.ErrorContext(ctx, "failed to add VISIT dormant tag", "error", addErr, "tag", th.tag)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, th.tag, "auto", ""); cacheErr != nil {
				slog.ErrorContext(ctx, "failed to upsert VISIT dormant tag cache", "error", cacheErr, "tag", th.tag)
			}
		} else if hasCached {
			if delErr := client.RemoveTag(ctx, lineUserID, th.tag); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove VISIT dormant tag", "error", delErr, "tag", th.tag)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, th.tag)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// petSpeciesFlags は生存ペットのスライスから犬/猫の有無を返す純粋関数（FEAT-377）。
func petSpeciesFlags(pets []model.Pet) (hasDog, hasCat bool) {
	for i := range pets {
		p := &pets[i]
		if p.AnimalSpecies == nil {
			continue
		}
		name := p.AnimalSpecies.Name
		if strings.Contains(name, "犬") {
			hasDog = true
		}
		if strings.Contains(name, "猫") {
			hasCat = true
		}
	}
	return hasDog, hasCat
}

// hasSeniorPet は生存ペットのスライスに now 時点で 7 歳以上の犬猫が存在するかを返す純粋関数（FEAT-377）。
func hasSeniorPet(pets []model.Pet, now time.Time) bool {
	const seniorAgeYears = 7
	for i := range pets {
		p := &pets[i]
		if p.BirthDate == nil || p.AnimalSpecies == nil {
			continue
		}
		name := p.AnimalSpecies.Name
		if !strings.Contains(name, "犬") && !strings.Contains(name, "猫") {
			continue
		}
		ageYears := now.Year() - p.BirthDate.Year()
		if now.Before(p.BirthDate.AddDate(ageYears, 0, 0)) {
			ageYears--
		}
		if ageYears >= seniorAgeYears {
			return true
		}
	}
	return false
}

// SyncPetSpeciesTags は飼い主の生存ペット種別タグ（PET_犬あり / PET_猫あり）を同期する（FEAT-377）。
func (s *lstepTagSyncService) SyncPetSpeciesTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for PET species tags sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	pets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find living pets for PET species tags", "error", err)
		return apperrors.Wrap(err, "failed to find living pets")
	}

	hasDog, hasCat := petSpeciesFlags(pets)

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	const (
		tagDog = "PET_犬あり"
		tagCat = "PET_猫あり"
	)
	apiFailed := false
	for _, entry := range []struct {
		tag  string
		have bool
	}{
		{tagDog, hasDog},
		{tagCat, hasCat},
	} {
		if entry.have {
			if addErr := client.AddTag(ctx, lineUserID, entry.tag); addErr != nil {
				slog.ErrorContext(ctx, "failed to add PET species tag", "error", addErr, "tag", entry.tag)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, entry.tag, "auto", ""); cacheErr != nil {
				slog.ErrorContext(ctx, "failed to upsert PET species tag cache", "error", cacheErr)
			}
		} else {
			if delErr := client.RemoveTag(ctx, lineUserID, entry.tag); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove PET species tag", "error", delErr, "tag", entry.tag)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, entry.tag)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncSeniorTag は飼い主の生存ペットに 7 歳以上の犬猫がいる場合 PET_シニア対象 タグを付与する（FEAT-377）。
func (s *lstepTagSyncService) SyncSeniorTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for senior tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	pets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find living pets for senior tag", "error", err)
		return apperrors.Wrap(err, "failed to find living pets")
	}

	isSenior := hasSeniorPet(pets, time.Now())

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	const seniorTag = "PET_シニア対象"
	apiFailed := false
	if isSenior {
		if addErr := client.AddTag(ctx, lineUserID, seniorTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add senior tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add senior tag")
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, seniorTag, "auto", ""); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert senior tag cache", "error", cacheErr)
		}
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, seniorTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove senior tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, seniorTag)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncExclusionTags は配信停止条件に基づき EXCL_配信停止 タグを同期する（FEAT-377/FEAT-381）。
// 配信停止条件: LstepOptOut=true / DeliveryExcluded=true / IsTransferred=true /
//
//	会員ステータスが deceased / transferred / 全ペット死亡。
//
// 注: checkOptOut は呼ばない（このメソッド自体が opt-out 判定の実装）。
func (s *lstepTagSyncService) SyncExclusionTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner for exclusion tags", "error", err)
		return apperrors.Wrap(err, "failed to find owner")
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	// 全ペット死亡判定
	totalPets, err := s.petRepo.CountByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count pets for exclusion tags", "error", err)
		return apperrors.Wrap(err, "failed to count pets")
	}
	livingPets, err := s.petRepo.CountLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count living pets for exclusion tags", "error", err)
		return apperrors.Wrap(err, "failed to count living pets")
	}
	allPetsDead := totalPets > 0 && livingPets == 0

	shouldExclude := owner.LstepOptOut ||
		owner.DeliveryExcluded ||
		owner.IsTransferred ||
		owner.MembershipType == model.MembershipTypeDeceased ||
		owner.MembershipType == model.MembershipTypeTransferred ||
		allPetsDead

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	const exclTag = "EXCL_配信停止"
	apiFailed := false
	if shouldExclude {
		if addErr := client.AddTag(ctx, lineUserID, exclTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add EXCL exclusion tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add EXCL exclusion tag")
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, exclTag, "auto", ""); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert EXCL exclusion tag cache", "error", cacheErr)
		}
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, exclTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove EXCL exclusion tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, exclTag)
		}
	}

	// FEAT-381-2: EXCL_配信注意 タグは delivery_caution フラグのみで独立して制御する。
	const cautionTag = "EXCL_配信注意"
	if owner.DeliveryCaution {
		if addErr := client.AddTag(ctx, lineUserID, cautionTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add EXCL caution tag", "error", addErr)
			apiFailed = true
		} else {
			if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, cautionTag, "auto", ""); cacheErr != nil {
				slog.ErrorContext(ctx, "failed to upsert EXCL caution tag cache", "error", cacheErr)
			}
		}
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, cautionTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove EXCL caution tag", "error", delErr)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, cautionTag)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// notifyAPISuccess は Lステップ API 呼び出し成功時にエラーカウンターをリセットし、
// EXCL_カルテ連携エラー タグを解除する（カウンターが 0 の場合は noop）。
// エラーはログのみ — 呼び出し元に伝播させない（best-effort）。
func (s *lstepTagSyncService) notifyAPISuccess(ctx context.Context, client lstep.Client, clinicID, ownerID uint64, lineUserID string) {
	if s.errorCounterRepo == nil {
		return
	}
	counter, err := s.errorCounterRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return
		}
		slog.ErrorContext(ctx, "failed to find lstep error counter", "error", err, "clinic_id", clinicID, "owner_id", ownerID)
		return
	}
	if counter.FailureCount == 0 {
		return
	}
	if removeErr := client.RemoveTag(ctx, lineUserID, lstepErrorTag); removeErr != nil {
		slog.ErrorContext(ctx, "failed to remove EXCL error tag", "error", removeErr, "owner_id", ownerID)
		return
	}
	if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, lstepErrorTag); delErr != nil {
		slog.ErrorContext(ctx, "failed to delete EXCL error tag cache", "error", delErr)
	}
	if resetErr := s.errorCounterRepo.ResetFailure(ctx, clinicID, ownerID); resetErr != nil {
		slog.ErrorContext(ctx, "failed to reset lstep error counter", "error", resetErr, "owner_id", ownerID)
	}
}
