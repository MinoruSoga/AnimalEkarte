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

// CPMStage は顧客ポートフォリオ管理ステージ（BE-011）。
type CPMStage string

const (
	CPMStageNew       CPMStage = "cpm_new"        // 初来院・1回のみ
	CPMStageStep      CPMStage = "cpm_step"       // 2〜3回、90日以内
	CPMStageRegular   CPMStage = "cpm_regular"    // 4回以上、90日以内
	CPMStageLoyalHigh CPMStage = "cpm_loyal_high" // 高LTV、90日以内
	CPMStageAtRisk    CPMStage = "cpm_at_risk"    // 90〜180日間来院なし
	CPMStageDormant   CPMStage = "cpm_dormant"    // 180日超来院なし
)

var allCPMStages = []CPMStage{
	CPMStageNew, CPMStageStep, CPMStageRegular, CPMStageLoyalHigh, CPMStageAtRisk, CPMStageDormant,
}

// CPMData は CPM ステージ計算に必要な集計データ。
type CPMData struct {
	TotalVisitCount  int64
	AnnualVisitCount int64
	DaysSinceVisit   int   // 最終来院からの経過日数（来院なし = -1）
	LTVAmount        int64 // 支払済み累計金額（円）
}

// CalculateCPMStage は純粋関数として CPM ステージを計算する（BE-011）。
func CalculateCPMStage(d CPMData) CPMStage {
	if d.DaysSinceVisit < 0 {
		return CPMStageDormant
	}
	if d.DaysSinceVisit > 180 {
		return CPMStageDormant
	}
	if d.DaysSinceVisit > 90 {
		return CPMStageAtRisk
	}
	// 90日以内来院
	switch {
	case d.LTVAmount >= 100_000 || d.AnnualVisitCount >= 6:
		return CPMStageLoyalHigh
	case d.TotalVisitCount >= 4:
		return CPMStageRegular
	case d.TotalVisitCount >= 2:
		return CPMStageStep
	default:
		return CPMStageNew
	}
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
	// SyncReservationTag は予約登録・変更時に reserved_YYYY-MM-DD タグを更新する（BE-007）。
	// 旧 reserved_* / canceled_visit / no_show_* タグを解除してから新タグを付与する。
	SyncReservationTag(ctx context.Context, clinicID, ownerID uint64, reservationDate time.Time) error
	// SyncCancellationTag は予約キャンセル時に canceled_visit タグを付与し reserved_* を解除する（BE-007）。
	SyncCancellationTag(ctx context.Context, clinicID, ownerID uint64, canceledDate time.Time) error
	// SyncCheckupTag は健診記録の作成・更新時に checkup_*/next_checkup_* タグを同期する（BE-008）。
	// checkupDate タグは累積付与。next_checkup_* は最新1件のみ（旧タグ削除→新タグ付与）。
	SyncCheckupTag(ctx context.Context, clinicID, ownerID uint64, checkupDate time.Time, nextDate *time.Time) error
	// SyncPrescriptionTag は飼い主の全アクティブ処方を取得し、補充推奨日が最も遅い処方に基づいて
	// refill_due_* タグを更新する（BE-009）。処方記録の追加・更新・削除後に呼び出すこと。
	SyncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) error
	// SyncChronicConditionTags は慢性疾患フラグに基づき chronic_* タグを差分同期する（BE-012）。
	// activeConditionCodes は飼い主の全生存ペットのアクティブ疾患コード一覧。
	SyncChronicConditionTags(ctx context.Context, clinicID, ownerID uint64, activeConditionCodes []string) error
	// SyncNoShowTag は予約ノーショウ時に no_show_YYYY-MM-DD タグを付与し reserved_* を解除する（BE-014）。
	SyncNoShowTag(ctx context.Context, clinicID, ownerID uint64, reservationDate time.Time) error
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
	}
}

// buildClient はクリニック設定から lstep.Client を構築する。
// API キーが未設定の場合は nil, nil を返す（設定前はスキップ）。
func (s *lstepTagSyncService) buildClient(ctx context.Context, clinicID uint64) (lstep.Client, error) {
	apiKey, baseURL, _, err := s.settingsSvc.GetRawCredentials(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get lstep credentials")
	}
	if apiKey == "" {
		return nil, nil
	}
	return lstep.NewClient(apiKey, baseURL), nil
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

	for _, tag := range tags {
		if addErr := client.AddTag(ctx, *owner.LineUserID, tag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add vaccine tag", "error", addErr, "tag", tag)
			return apperrors.Wrap(addErr, fmt.Sprintf("failed to add vaccine tag %s", tag))
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto"); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert tag cache", "error", cacheErr, "tag", tag)
		}
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
	for _, p := range pets {
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
	for _, old := range []string{"has_dog", "has_cat", "has_both"} {
		if old == newTag {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, old); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove old classification tag", "error", delErr, "tag", old)
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, old)
		}
	}

	if newTag == "" {
		return nil
	}

	if addErr := client.AddTag(ctx, lineUserID, newTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add classification tag", "error", addErr, "tag", newTag)
		return apperrors.Wrap(addErr, fmt.Sprintf("failed to add classification tag %s", newTag))
	}
	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, newTag, "auto"); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert classification tag cache", "error", cacheErr, "tag", newTag)
	}
	return nil
}

// SyncPetBasicInfoTags は全生存ペットの基本情報タグを同期する（BE-005）。
func (s *lstepTagSyncService) SyncPetBasicInfoTags(ctx context.Context, clinicID, ownerID uint64) error {
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
	for old := range oldSet {
		if _, keep := newSet[old]; keep {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, old); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove stale pet basic info tag", "error", delErr, "tag", old)
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
			return apperrors.Wrap(addErr, fmt.Sprintf("failed to add pet basic info tag %s", tag))
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto"); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert pet basic info tag cache", "error", cacheErr, "tag", tag)
		}
	}
	return nil
}

// buildPetBasicInfoTags は生存ペット一覧から基本情報タグ一覧を生成する。
func buildPetBasicInfoTags(pets []model.Pet) []string {
	tagSet := make(map[string]struct{})
	var hasNeutered, hasIntact bool

	for _, p := range pets {
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
	for _, tag := range tags {
		if addErr := client.AddTag(ctx, lineUserID, tag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add visit tag", "error", addErr, "tag", tag)
			return apperrors.Wrap(addErr, fmt.Sprintf("failed to add visit tag %s", tag))
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto"); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert tag cache", "error", cacheErr, "tag", tag)
		}
	}

	// 来院完了でキャッシュ経由の予約関連タグ（reserved_* / canceled_visit / no_show_*）を削除
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for visit completion cleanup", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	for _, c := range cached {
		if isReservationRelatedTag(c.TagName) {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove reservation tag on visit completion", "error", delErr, "tag", c.TagName)
			} else {
				_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName)
			}
		}
	}
	// レガシータグ（名前変更前との互換）
	for _, staleTag := range []string{"dormant", "noshow", "reserved"} {
		if delErr := client.RemoveTag(ctx, lineUserID, staleTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove stale legacy tag", "error", delErr, "tag", staleTag)
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, staleTag)
		}
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
	tags = append(tags, ltvBracketTag(ltv))
	tags = append(tags, visitCountAnnualTag(summary.AnnualCount))
	return tags
}

func ltvBracketTag(ltv int64) string {
	switch {
	case ltv >= 500_000:
		return "ltv_amount_500000plus"
	case ltv >= 200_000:
		return "ltv_amount_200000to500000"
	case ltv >= 100_000:
		return "ltv_amount_100000to200000"
	case ltv >= 50_000:
		return "ltv_amount_50000to100000"
	case ltv >= 10_000:
		return "ltv_amount_10000to50000"
	default:
		return "ltv_amount_under10000"
	}
}

func visitCountAnnualTag(count int64) string {
	switch {
	case count >= 12:
		return "visit_count_annual_12plus"
	case count >= 6:
		return "visit_count_annual_6to12"
	case count >= 3:
		return "visit_count_annual_3to6"
	default:
		return fmt.Sprintf("visit_count_annual_%d", count)
	}
}

// SyncNextVisitTag は次回来院推奨日タグを同期する（BE-006）。
// 最新カルテの next_visit_recommended_date を参照し、古い next_visit_* タグを差し替える。
// date が nil の場合はすべての next_visit_* タグを削除する。
func (s *lstepTagSyncService) SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error {
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for next visit tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil {
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
	for _, c := range cached {
		if strings.HasPrefix(c.TagName, "next_visit_") {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove next_visit tag", "error", delErr, "tag", c.TagName)
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete next_visit tag cache", "error", delErr, "tag", c.TagName)
			}
		}
	}

	// 新しい next_visit_YYYY-MM-DD タグを付与（日付が設定されている場合のみ）
	if latest == nil || latest.NextVisitRecommendedDate == nil {
		return nil
	}
	newTag := "next_visit_" + latest.NextVisitRecommendedDate.Format("2006-01-02")
	if addErr := client.AddTag(ctx, lineUserID, newTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add next_visit tag", "error", addErr, "tag", newTag)
		return apperrors.Wrap(addErr, "failed to add next_visit tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, newTag, "auto"); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert next_visit tag cache", "error", upsertErr)
	}
	return nil
}

// SyncCPMStageTag は CPM ステージタグを同期する（BE-011）。
func (s *lstepTagSyncService) SyncCPMStageTag(ctx context.Context, clinicID, ownerID uint64) error {
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

	daysSince := -1
	if summary.LastVisitAt != nil {
		daysSince = int(time.Since(*summary.LastVisitAt).Hours() / 24)
	}

	stage := CalculateCPMStage(CPMData{
		TotalVisitCount:  summary.TotalCount,
		AnnualVisitCount: summary.AnnualCount,
		DaysSinceVisit:   daysSince,
		LTVAmount:        ltv,
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
	for _, old := range allCPMStages {
		if string(old) == string(stage) {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, string(old)); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove old CPM stage tag", "error", delErr, "tag", old)
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, string(old))
		}
	}

	if addErr := client.AddTag(ctx, lineUserID, string(stage)); addErr != nil {
		slog.ErrorContext(ctx, "failed to add CPM stage tag", "error", addErr, "stage", stage)
		return apperrors.Wrap(addErr, "failed to add CPM stage tag")
	}
	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, string(stage), "auto"); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert CPM stage tag cache", "error", cacheErr)
	}

	return nil
}

// SyncReservationTag は予約登録・変更時に reserved_YYYY-MM-DD タグを更新する（BE-007）。
// 旧 reserved_* / canceled_visit / no_show_* タグを解除してから新タグを付与する。
func (s *lstepTagSyncService) SyncReservationTag(ctx context.Context, clinicID, ownerID uint64, reservationDate time.Time) error {
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for reservation tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil {
		return nil
	}
	lineUserID := *owner.LineUserID

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for reservation tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for reservation tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	for _, c := range cached {
		if isReservationRelatedTag(c.TagName) {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale reservation tag", "error", delErr, "tag", c.TagName)
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete stale reservation tag cache", "error", delErr, "tag", c.TagName)
			}
		}
	}

	newTag := "reserved_" + reservationDate.Format("2006-01-02")
	if addErr := client.AddTag(ctx, lineUserID, newTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add reservation tag", "error", addErr, "tag", newTag)
		return apperrors.Wrap(addErr, "failed to add reservation tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, newTag, "auto"); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert reservation tag cache", "error", upsertErr)
	}
	return nil
}

// SyncCancellationTag は予約キャンセル時に canceled_visit タグを付与し reserved_* を解除する（BE-007）。
func (s *lstepTagSyncService) SyncCancellationTag(ctx context.Context, clinicID, ownerID uint64, canceledDate time.Time) error {
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for cancellation tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil {
		return nil
	}
	lineUserID := *owner.LineUserID

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for cancellation tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for cancellation tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	for _, c := range cached {
		if strings.HasPrefix(c.TagName, "reserved_") {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove reserved tag on cancellation", "error", delErr, "tag", c.TagName)
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete reserved tag cache on cancellation", "error", delErr, "tag", c.TagName)
			}
		}
	}

	const canceledVisitTag = "canceled_visit"
	if addErr := client.AddTag(ctx, lineUserID, canceledVisitTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add canceled_visit tag", "error", addErr)
		return apperrors.Wrap(addErr, "failed to add canceled_visit tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, canceledVisitTag, "auto"); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert canceled_visit tag cache", "error", upsertErr)
	}
	_ = canceledDate // 将来的に canceled_visit_YYYY-MM-DD 形式に拡張する場合に使用
	return nil
}

// isReservationRelatedTag は BE-007 予約関連タグ（reserved_* / canceled_visit / no_show_*）かを判定する。
func isReservationRelatedTag(tag string) bool {
	return strings.HasPrefix(tag, "reserved_") ||
		tag == "canceled_visit" ||
		strings.HasPrefix(tag, "no_show_")
}

// SyncPrescriptionTag は飼い主の全アクティブ処方を取得し、補充推奨日が最も遅い処方に基づいて
// refill_due_* タグを更新する（BE-009）。duration_days < 7 の場合は prescribed_at + 1 日を使用する。
// 最新の refill_due が現在日時を過ぎている場合は refill_due_* タグをすべて削除して終了する。
func (s *lstepTagSyncService) SyncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) error {
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for prescription tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil {
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
	for _, p := range prescriptions {
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
	for _, c := range cached {
		if strings.HasPrefix(c.TagName, "refill_due_") {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale refill_due tag", "error", delErr, "tag", c.TagName)
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete refill_due tag cache", "error", delErr, "tag", c.TagName)
			}
		}
	}

	// 補充推奨日が未来にある場合のみ新タグを付与
	if latestRefillDue == nil || !latestRefillDue.After(time.Now()) {
		return nil
	}
	newTag := fmt.Sprintf("refill_due_%s", latestRefillDue.Format("2006-01-02"))
	if addErr := client.AddTag(ctx, lineUserID, newTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add refill_due tag", "error", addErr, "tag", newTag)
		return apperrors.Wrap(addErr, "failed to add refill_due tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, newTag, "auto"); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert refill_due tag cache", "error", upsertErr)
	}
	return nil
}

// SyncCheckupTag は健診記録の作成・更新時に checkup_*/next_checkup_* タグを同期する（BE-008）。
// checkupDate タグは累積付与（他の日付のタグは保持）。
// next_checkup_* タグは最新1件のみ（旧タグを削除してから新タグを付与）。
func (s *lstepTagSyncService) SyncCheckupTag(ctx context.Context, clinicID, ownerID uint64, checkupDate time.Time, nextDate *time.Time) error {
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for checkup tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil {
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

	// next_checkup_* は最新1件のみ — 旧タグをキャッシュ経由で削除
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for checkup tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	for _, c := range cached {
		if strings.HasPrefix(c.TagName, "next_checkup_") {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale next_checkup tag", "error", delErr, "tag", c.TagName)
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete next_checkup tag cache", "error", delErr, "tag", c.TagName)
			}
		}
	}

	// checkup_YYYY-MM-DD タグを累積付与
	checkupTag := "checkup_" + checkupDate.Format("2006-01-02")
	if addErr := client.AddTag(ctx, lineUserID, checkupTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add checkup tag", "error", addErr, "tag", checkupTag)
		return apperrors.Wrap(addErr, "failed to add checkup tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, checkupTag, "auto"); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert checkup tag cache", "error", upsertErr)
	}

	// next_checkup_YYYY-MM-DD タグを付与（設定時のみ）
	if nextDate != nil {
		nextTag := "next_checkup_" + nextDate.Format("2006-01-02")
		if addErr := client.AddTag(ctx, lineUserID, nextTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add next_checkup tag", "error", addErr, "tag", nextTag)
			return apperrors.Wrap(addErr, "failed to add next_checkup tag")
		}
		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, nextTag, "auto"); upsertErr != nil {
			slog.ErrorContext(ctx, "failed to upsert next_checkup tag cache", "error", upsertErr)
		}
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
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for chronic condition tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil {
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
		}
		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tagName, "auto"); upsertErr != nil {
			slog.ErrorContext(ctx, "failed to upsert chronic tag cache", "tag", tagName, "error", upsertErr)
		}
	}

	return nil
}

// SyncNoShowTag は予約ノーショウ時に no_show_YYYY-MM-DD タグを付与し reserved_* を解除する（BE-014）。
func (s *lstepTagSyncService) SyncNoShowTag(ctx context.Context, clinicID, ownerID uint64, reservationDate time.Time) error {
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for no-show tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil {
		return nil
	}
	lineUserID := *owner.LineUserID

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for no-show tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for no-show tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	for _, c := range cached {
		if strings.HasPrefix(c.TagName, "reserved_") {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove reserved tag on no-show", "error", delErr, "tag", c.TagName)
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete reserved tag cache on no-show", "error", delErr, "tag", c.TagName)
			}
		}
	}

	noShowTag := "no_show_" + reservationDate.Format("2006-01-02")
	if addErr := client.AddTag(ctx, lineUserID, noShowTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add no_show tag", "error", addErr, "tag", noShowTag)
		return apperrors.Wrap(addErr, "failed to add no_show tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, noShowTag, "auto"); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert no_show tag cache", "error", upsertErr)
	}
	return nil
}
