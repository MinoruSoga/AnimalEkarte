package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ISSUE-005: 除外理由（優先度順）。複数該当する場合は上位を採用する。
const (
	exclusionReasonOptOut       = "Lステップ配信停止中"
	exclusionReasonNoLivingPet  = "生存ペットなし"
	exclusionReasonLineUnlinked = "LINE未連携"
)

// deriveExclusionReason は除外理由を優先度順に判定する（ISSUE-005）。
// 全条件 OK のとき nil を返す（送信可能）。
func deriveExclusionReason(hasLine, isOptOut, hasLivingPet bool) *string {
	switch {
	case isOptOut:
		s := exclusionReasonOptOut
		return &s
	case !hasLivingPet:
		s := exclusionReasonNoLivingPet
		return &s
	case !hasLine:
		s := exclusionReasonLineUnlinked
		return &s
	default:
		return nil
	}
}

// CheckupSyncPreviewOwner はプレビュー一覧の1件。
type CheckupSyncPreviewOwner struct {
	OwnerID         uint64
	OwnerName       string
	PetNames        []string
	LastVisitDate   *time.Time
	HasLine         bool
	IsOptOut        bool
	HasLivingPet    bool    // ISSUE-005: 生存ペットの有無
	ExclusionReason *string // ISSUE-005: 対象外理由（送信可能なら nil）
	CurrentTags     []string

	// ISSUE-009: 抽出条件確認用に追加した表示フィールド（additive、既存契約は変更しない）
	MinPetAgeYears      *int       // 生存ペットの最小年齢（years）
	MaxPetAgeYears      *int       // 生存ペットの最大年齢（years）
	HasChronicCondition bool       // 慢性疾患（アクティブ）の有無
	CPMStage            string     // CPM ステージ（cpm_encounter / cpm_growing / cpm_core / cpm_spot / cpm_noah / cpm_dormant）
	TotalAmount         int64      // 累計診療費（円、completed billings 合計）
	AnnualVisitCount    int64      // 年間来院回数（過去365日 distinct visit）
	LastCheckupDate     *time.Time // 最終健診実施日
}

// PreviewCheckupSyncInput はPreviewCheckupSyncの入力パラメータ。
type PreviewCheckupSyncInput struct {
	CheckupType     string
	Species         string
	LastVisitBefore *time.Time
	LastVisitAfter  *time.Time

	// ISSUE-009: 追加フィルタ
	MinAgeYears         *int       // 生存ペットの最小年齢（years 以上）
	MaxAgeYears         *int       // 生存ペットの最大年齢（years 以下）
	HasChronicCondition *bool      // 慢性疾患フラグ（true: あり / false: なし / nil: 絞らない）
	CPMStage            string     // CPM ステージ絞り込み（空文字なら絞らない）
	MinTotalAmount      *int64     // 累計診療費（円）以上
	MinAnnualVisitCount *int64     // 年間来院回数（過去365日）以上
	LastCheckupBefore   *time.Time // 最終健診実施日 <= この日
	LastCheckupAfter    *time.Time // 最終健診実施日 >= この日
}

// PreviewCheckupSyncResult はPreviewCheckupSyncの結果。
// ISSUE-005: EligibleCount は誤配信防止のため、opt-out / 生存ペットなし / LINE未連携 を除いた送信可能件数。
type PreviewCheckupSyncResult struct {
	Owners           []CheckupSyncPreviewOwner
	TotalCount       int
	EligibleCount    int // ISSUE-005: 送信可能件数（has_line && !opt_out && has_living_pet）
	LineLinkedCount  int
	OptOutCount      int // ISSUE-005: opt-out 中の件数
	NoLivingPetCount int // ISSUE-005: 生存ペットなしの件数（死亡ペットのみの飼い主）
}

// CreateCheckupSyncInput はCreateCheckupSyncの入力パラメータ。
type CreateCheckupSyncInput struct {
	CheckupType string
	OwnerIDs    []uint64
	TagName     string
}

// CreateCheckupSyncResult はCreateCheckupSyncの結果。
type CreateCheckupSyncResult struct {
	SuccessCount   int
	SkippedCount   int
	FailedCount    int
	FailedOwnerIDs []uint64
}

// CheckupSyncService は健診対象者抽出・一括タグ連携の業務ロジックインターフェース（BE-004）。
// ISSUE-010: PreviewCheckupSync は actorID を受け取り audit_logs にメタデータと共に永続化する。
//   - actorID=nil はシステム実行扱い、それ以外は staff 実行扱いとして actor_type が決まる。
type CheckupSyncService interface {
	PreviewCheckupSync(ctx context.Context, clinicID uint64, input *PreviewCheckupSyncInput, actorID *uint64) (*PreviewCheckupSyncResult, error)
	CreateCheckupSync(ctx context.Context, clinicID uint64, input CreateCheckupSyncInput, actorID *uint64) (*CreateCheckupSyncResult, error)
}

type checkupSyncService struct {
	repo         repository.CheckupSyncRepository
	ownerRepo    repository.OwnerRepository
	petRepo      repository.PetRepository
	tagCacheRepo repository.LstepTagCacheRepository
	settingsSvc  LstepSettingsService
	auditSvc     AuditService
}

// NewCheckupSyncService は CheckupSyncService を初期化して返す。
// ISSUE-007: petRepo は CreateCheckupSync で生存ペット数の二重防御チェックに使用する。
func NewCheckupSyncService(
	repo repository.CheckupSyncRepository,
	ownerRepo repository.OwnerRepository,
	petRepo repository.PetRepository,
	tagCacheRepo repository.LstepTagCacheRepository,
	settingsSvc LstepSettingsService,
	auditSvc AuditService,
) CheckupSyncService {
	return &checkupSyncService{
		repo:         repo,
		ownerRepo:    ownerRepo,
		petRepo:      petRepo,
		tagCacheRepo: tagCacheRepo,
		settingsSvc:  settingsSvc,
		auditSvc:     auditSvc,
	}
}

func (s *checkupSyncService) buildClient(ctx context.Context, clinicID uint64) (lstep.Client, error) {
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

// computeCPMStageFromRow は preview 行から CPM ステージを計算する（ISSUE-009）。
// CPM 判定は LTV/CPM サービス側 (lstep_tag_sync_service.CalculateCPMStage) と同一の純粋関数を再利用し、
// 判定基準のドリフトを防ぐ。
// caller must pass non-nil row; panics on nil.
func computeCPMStageFromRow(row *repository.CheckupSyncPreviewRow, thresholds model.CPMV1Thresholds) CPMStage {
	daysSince := -1
	if row.LastVisitDate != nil {
		daysSince = int(time.Since(*row.LastVisitDate).Hours() / 24)
	}
	firstVisitDaysSince := 0
	if row.FirstVisitDate != nil {
		firstVisitDaysSince = int(time.Since(*row.FirstVisitDate).Hours() / 24)
	}
	return CalculateCPMStage(CPMData{
		TotalVisitCount:      row.TotalVisitCount,
		AnnualVisitCount:     row.AnnualVisitCount,
		DaysSinceVisit:       daysSince,
		LTVAmount:            row.TotalAmount,
		FirstVisitDaysSince:  firstVisitDaysSince,
		MaxSingleVisitAmount: row.MaxSingleVisitAmount,
		Thresholds:           thresholds,
	})
}

func (s *checkupSyncService) PreviewCheckupSync(ctx context.Context, clinicID uint64, input *PreviewCheckupSyncInput, actorID *uint64) (*PreviewCheckupSyncResult, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input is nil")
	}
	rows, err := s.repo.FindCheckupSyncPreview(ctx, &repository.FindCheckupSyncPreviewParams{
		ClinicID:            clinicID,
		Species:             input.Species,
		LastVisitBefore:     input.LastVisitBefore,
		LastVisitAfter:      input.LastVisitAfter,
		MinAgeYears:         input.MinAgeYears,
		MaxAgeYears:         input.MaxAgeYears,
		HasChronicCondition: input.HasChronicCondition,
		MinTotalAmount:      input.MinTotalAmount,
		MinAnnualVisitCount: input.MinAnnualVisitCount,
		LastCheckupBefore:   input.LastCheckupBefore,
		LastCheckupAfter:    input.LastCheckupAfter,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to find checkup sync preview", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to find checkup sync preview")
	}

	thresholds, err := s.settingsSvc.GetCPMV1Thresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get cpm v1 thresholds for checkup preview", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get cpm v1 thresholds")
	}

	owners := make([]CheckupSyncPreviewOwner, 0, len(rows))
	lineLinkedCount := 0
	optOutCount := 0
	noLivingPetCount := 0
	eligibleCount := 0

	for i := range rows {
		row := &rows[i]
		// ISSUE-009: CPM ステージは集計値ベースの後段フィルタとして適用する。
		// SQL ではなく service 層で判定することで、タグ同期側 CalculateCPMStage と同じロジックを共有する。
		cpmStage := computeCPMStageFromRow(row, thresholds)
		if input.CPMStage != "" && string(cpmStage) != input.CPMStage {
			continue
		}

		hasLine := row.LineUserID != nil && *row.LineUserID != ""
		hasLivingPet := row.LivingPetCount > 0
		if hasLine {
			lineLinkedCount++
		}
		if row.LstepOptOut {
			optOutCount++
		}
		if !hasLivingPet {
			noLivingPetCount++
		}

		var petNames []string
		if row.PetNamesCSV != "" {
			petNames = strings.Split(row.PetNamesCSV, ",")
		} else {
			petNames = []string{}
		}

		var currentTags []string
		if hasLine {
			cached, cacheErr := s.tagCacheRepo.FindByOwner(ctx, clinicID, row.OwnerID)
			if cacheErr != nil {
				slog.ErrorContext(ctx, "failed to load tag cache for preview", "error", cacheErr, "owner_id", row.OwnerID)
				currentTags = []string{}
			} else {
				currentTags = make([]string, 0, len(cached))
				for _, c := range cached {
					currentTags = append(currentTags, c.TagName)
				}
			}
		} else {
			currentTags = []string{}
		}

		exclusionReason := deriveExclusionReason(hasLine, row.LstepOptOut, hasLivingPet)
		if exclusionReason == nil {
			eligibleCount++
		}

		owners = append(owners, CheckupSyncPreviewOwner{
			OwnerID:             row.OwnerID,
			OwnerName:           row.OwnerName,
			PetNames:            petNames,
			LastVisitDate:       row.LastVisitDate,
			HasLine:             hasLine,
			IsOptOut:            row.LstepOptOut,
			HasLivingPet:        hasLivingPet,
			ExclusionReason:     exclusionReason,
			CurrentTags:         currentTags,
			MinPetAgeYears:      row.MinPetAgeYears,
			MaxPetAgeYears:      row.MaxPetAgeYears,
			HasChronicCondition: row.HasChronicCondition,
			CPMStage:            string(cpmStage),
			TotalAmount:         row.TotalAmount,
			AnnualVisitCount:    row.AnnualVisitCount,
			LastCheckupDate:     row.LastCheckupDate,
		})
	}

	// ISSUE-005 / ISSUE-009: 抽出条件と除外理由ごとの件数を残す（監査ログ要件）。
	slog.InfoContext(ctx, "checkup sync preview extracted",
		"clinic_id", clinicID,
		"checkup_type", input.CheckupType,
		"species", input.Species,
		"last_visit_before", input.LastVisitBefore,
		"last_visit_after", input.LastVisitAfter,
		"min_age_years", input.MinAgeYears,
		"max_age_years", input.MaxAgeYears,
		"has_chronic_condition", input.HasChronicCondition,
		"cpm_stage", input.CPMStage,
		"min_total_amount", input.MinTotalAmount,
		"min_annual_visit_count", input.MinAnnualVisitCount,
		"last_checkup_before", input.LastCheckupBefore,
		"last_checkup_after", input.LastCheckupAfter,
		"total_count", len(owners),
		"eligible_count", eligibleCount,
		"line_linked_count", lineLinkedCount,
		"opt_out_count", optOutCount,
		"no_living_pet_count", noLivingPetCount,
	)

	// ISSUE-010: 抽出条件と件数を audit_logs.metadata に永続化する（slog の構造化ログを DB にも残す）。
	//   resource_id は対象者リストが多次元のため使えない（個別 owner_id 単一参照は意味を持たない）。
	//   よって metadata に集計値・条件をまとめて格納する。
	_ = s.auditSvc.LogLstepOperationWithMetadata(ctx, clinicID, actorID,
		"checkup_sync_preview", "owner", nil,
		buildCheckupSyncPreviewMetadata(input, len(owners), eligibleCount, lineLinkedCount, optOutCount, noLivingPetCount),
	)

	return &PreviewCheckupSyncResult{
		Owners:           owners,
		TotalCount:       len(owners),
		EligibleCount:    eligibleCount,
		LineLinkedCount:  lineLinkedCount,
		OptOutCount:      optOutCount,
		NoLivingPetCount: noLivingPetCount,
	}, nil
}

// buildCheckupSyncPreviewMetadata は PreviewCheckupSync の監査メタデータを生成する（ISSUE-010）。
// 受入条件の必須キー（total_count / eligible_count / line_linked_count / opt_out_count /
// no_living_pet_count / species / last_visit_before / last_visit_after）に加え、
// ISSUE-009 で追加された拡張フィルタ条件も含めて永続化する。
// 日付は YYYY-MM-DD 文字列に正規化する（タイムゾーンによる比較揺れを防ぐ）。
func buildCheckupSyncPreviewMetadata(
	input *PreviewCheckupSyncInput,
	totalCount, eligibleCount, lineLinkedCount, optOutCount, noLivingPetCount int,
) map[string]any {
	filter := map[string]any{
		"checkup_type": input.CheckupType,
		"species":      input.Species,
	}
	if input.LastVisitBefore != nil {
		filter["last_visit_before"] = input.LastVisitBefore.Format("2006-01-02")
	}
	if input.LastVisitAfter != nil {
		filter["last_visit_after"] = input.LastVisitAfter.Format("2006-01-02")
	}
	if input.MinAgeYears != nil {
		filter["min_age_years"] = *input.MinAgeYears
	}
	if input.MaxAgeYears != nil {
		filter["max_age_years"] = *input.MaxAgeYears
	}
	if input.HasChronicCondition != nil {
		filter["has_chronic_condition"] = *input.HasChronicCondition
	}
	if input.CPMStage != "" {
		filter["cpm_stage"] = input.CPMStage
	}
	if input.MinTotalAmount != nil {
		filter["min_total_amount"] = *input.MinTotalAmount
	}
	if input.MinAnnualVisitCount != nil {
		filter["min_annual_visit_count"] = *input.MinAnnualVisitCount
	}
	if input.LastCheckupBefore != nil {
		filter["last_checkup_before"] = input.LastCheckupBefore.Format("2006-01-02")
	}
	if input.LastCheckupAfter != nil {
		filter["last_checkup_after"] = input.LastCheckupAfter.Format("2006-01-02")
	}
	return map[string]any{
		"operation":           "checkup_sync_preview",
		"filter":              filter,
		"total_count":         totalCount,
		"eligible_count":      eligibleCount,
		"line_linked_count":   lineLinkedCount,
		"opt_out_count":       optOutCount,
		"no_living_pet_count": noLivingPetCount,
	}
}

func (s *checkupSyncService) CreateCheckupSync(ctx context.Context, clinicID uint64, input CreateCheckupSyncInput, actorID *uint64) (*CreateCheckupSyncResult, error) {
	if isSystemManagedTag(input.TagName) {
		return nil, apperrors.WrapInvalidInput("tag_name は自動管理タグのため使用できません")
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, apperrors.WrapInvalidInput("Lステップ API が設定されていません")
	}

	result := &CreateCheckupSyncResult{FailedOwnerIDs: []uint64{}}
	// ISSUE-007: 監査要件 — スキップ理由内訳をカウントしてログに残す。
	skippedOptOut := 0
	skippedNoLivingPet := 0
	skippedLineUnlinked := 0

	for _, ownerID := range input.OwnerIDs {
		owner, findErr := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
		if findErr != nil {
			slog.ErrorContext(ctx, "checkup sync: owner not found", "owner_id", ownerID)
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, ownerID)
			result.FailedCount++
			continue
		}

		// ISSUE-007: スキップ判定は preview 側の deriveExclusionReason と同じ優先度で行う。
		//   優先度: opt-out > 生存ペットなし > LINE未連携
		// API を直接叩かれた場合でも死亡ペットのみの飼い主を確実に除外し、誤配信を防ぐ。
		if owner.LstepOptOut {
			skippedOptOut++
			result.SkippedCount++
			continue
		}

		livingPetCount, countErr := s.petRepo.CountLivingByOwner(ctx, clinicID, ownerID)
		if countErr != nil {
			slog.ErrorContext(ctx, "checkup sync: failed to count living pets", "error", countErr, "owner_id", ownerID)
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, ownerID)
			result.FailedCount++
			continue
		}
		if livingPetCount == 0 {
			skippedNoLivingPet++
			result.SkippedCount++
			continue
		}

		if owner.LineUserID == nil || *owner.LineUserID == "" {
			skippedLineUnlinked++
			result.SkippedCount++
			continue
		}

		if addErr := client.AddTag(ctx, *owner.LineUserID, input.TagName); addErr != nil {
			slog.ErrorContext(ctx, "checkup sync: failed to add lstep tag", "error", addErr, "owner_id", ownerID)
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, ownerID)
			result.FailedCount++
			continue
		}

		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, input.TagName, "manual", ""); upsertErr != nil {
			slog.ErrorContext(ctx, "checkup sync: failed to upsert tag cache", "error", upsertErr, "owner_id", ownerID)
		}
		result.SuccessCount++
	}

	// ISSUE-007: スキップ理由内訳を監査ログとして残す（opt-out / 生存ペットなし / LINE未連携）。
	slog.InfoContext(ctx, "checkup sync executed",
		"clinic_id", clinicID,
		"checkup_type", input.CheckupType,
		"tag_name", input.TagName,
		"requested_count", len(input.OwnerIDs),
		"success_count", result.SuccessCount,
		"skipped_count", result.SkippedCount,
		"failed_count", result.FailedCount,
		"skipped_opt_out", skippedOptOut,
		"skipped_no_living_pet", skippedNoLivingPet,
		"skipped_line_unlinked", skippedLineUnlinked,
	)

	// ISSUE-010: 一括タグ付与の実行件数とスキップ理由内訳を audit_logs.metadata に永続化する。
	_ = s.auditSvc.LogLstepOperationWithMetadata(ctx, clinicID, actorID,
		"checkup_sync", "owner", nil,
		map[string]any{
			"operation":             "checkup_sync_execute",
			"checkup_type":          input.CheckupType,
			"tag_name":              input.TagName,
			"requested_count":       len(input.OwnerIDs),
			"success_count":         result.SuccessCount,
			"skipped_count":         result.SkippedCount,
			"failed_count":          result.FailedCount,
			"skipped_opt_out":       skippedOptOut,
			"skipped_no_living_pet": skippedNoLivingPet,
			"skipped_line_unlinked": skippedLineUnlinked,
		},
	)

	return result, nil
}
