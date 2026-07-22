package lstep

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
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
	repo          CheckupSyncRepository
	ownerRepo     checkupSyncOwnerRepo
	petRepo       checkupSyncPetRepo
	tagCacheRepo  checkupSyncTagCacheRepo
	settingsSvc   checkupSyncSettingsService
	auditSvc      checkupSyncAuditLogger
	buildClientFn func(ctx context.Context, clinicID uint64) (lstepapi.Client, error) // テスト注入用
}

// NewCheckupSyncService は CheckupSyncService を初期化して返す。
// ISSUE-007: petRepo は CreateCheckupSync で生存ペット数の二重防御チェックに使用する。
func NewCheckupSyncService(
	repo CheckupSyncRepository,
	ownerRepo checkupSyncOwnerRepo,
	petRepo checkupSyncPetRepo,
	tagCacheRepo checkupSyncTagCacheRepo,
	settingsSvc checkupSyncSettingsService,
	auditSvc checkupSyncAuditLogger,
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

func (s *checkupSyncService) buildClient(ctx context.Context, clinicID uint64) (lstepapi.Client, error) {
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
	return lstepapi.NewClient(apiKey, baseURL), nil
}

// computeCPMStageFromRow は preview 行から CPM ステージを計算する（ISSUE-009）。
// CPM 判定は LTV/CPM サービス側 (lstep_tag_sync_service.CalculateCPMStage) と同一の純粋関数を再利用し、
// 判定基準のドリフトを防ぐ。
// caller must pass non-nil row; panics on nil.
//
//nolint:gocritic // hugeParam: thresholds は CalculateCPMStage 側で値型を要求するため統一
func computeCPMStageFromRow(row *CheckupSyncPreviewRow, thresholds model.CPMV1Thresholds) CPMStage {
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
