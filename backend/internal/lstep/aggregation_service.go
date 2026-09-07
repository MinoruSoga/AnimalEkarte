package lstep

import (
	"context"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// OwnerLTVRow is the lstep-owned projection consumed by customer aggregation.
type OwnerLTVRow struct {
	OwnerID              uint64
	OwnerName            string
	LineUserID           *string
	LstepOptOut          bool
	TotalAmount          int64
	TotalVisitCount      int64
	AnnualVisitCount     int64
	LastVisitDate        *time.Time
	FirstVisitDate       *time.Time
	AnnualAmount         *int64
	BillingCount         *int64
	PeriodVisitCount     *int64
	DaysSinceLastVisit   *int
	LastVisitBucket      *string
	MaxSingleVisitAmount int64
}

// FindOwnerLTVParams is the lstep-owned query contract for the legacy LTV
// persistence adapter. The adapter is removed when aggregation persistence
// leaves internal/repository.
type FindOwnerLTVParams struct {
	ClinicID        uint64
	Sort            string
	MinTotalAmount  *int64
	MaxTotalAmount  *int64
	MinVisitCount   *int64
	LineLinked      bool
	Year            *int
	From            *string
	To              *string
	AmountBasis     string
	IncludeZero     bool
	Search          string
	PeriodPreset    string
	MaxVisitCount   *int64
	LastVisitBucket string
	IncludeNoVisit  bool
	Order           string
}

// OwnerAggregationRepository is the narrow persistence view consumed by the
// LSTEP customer aggregation use case.
type OwnerAggregationRepository interface {
	FindOwnerLTV(ctx context.Context, params *FindOwnerLTVParams) ([]OwnerLTVRow, error)
}

// computeCPMStageFromLTVRow は LTV 集計行から CPM ステージを判定する（ISSUE-006）。
// タグ同期側 SyncCPMStageTag と同じ CalculateCPMStage を呼び、判定材料も
// AccountingRepository.MaxSingleVisitAmountByOwner と同一の集計範囲に揃える。
// 一覧側の DaysSinceLastVisit は SQL 側で EXTRACT(DAY FROM NOW() - MAX(mr.date)) として
// 計算されているため、来院なし（NULL）は -1 にマップする。
// caller must pass non-nil row; panics on nil.
//
//nolint:gocritic // hugeParam: thresholds は CalculateCPMStage 側で値型を要求するため統一
func computeCPMStageFromLTVRow(row *OwnerLTVRow, thresholds model.CPMV1Thresholds) string {
	daysSince := -1
	if row.DaysSinceLastVisit != nil {
		daysSince = *row.DaysSinceLastVisit
	}
	firstVisitDays := 0
	if row.FirstVisitDate != nil {
		firstVisitDays = int(time.Since(*row.FirstVisitDate).Hours() / 24)
	}
	return string(CalculateCPMStage(CPMData{
		TotalVisitCount:      row.TotalVisitCount,
		AnnualVisitCount:     row.AnnualVisitCount,
		DaysSinceVisit:       daysSince,
		LTVAmount:            row.TotalAmount,
		FirstVisitDaysSince:  firstVisitDays,
		MaxSingleVisitAmount: row.MaxSingleVisitAmount,
		Thresholds:           thresholds,
	}))
}

// normalizeCPMStageFilter は cpm_stage クエリパラメータを内部表現に正規化する（ISSUE-006）。
// "spot" / "cpm_spot" のいずれも受け付ける。空文字列は絞り込みなし。
func normalizeCPMStageFilter(s string) string {
	v := strings.TrimSpace(strings.ToLower(s))
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "cpm_") {
		return v
	}
	return "cpm_" + v
}

// OwnerAggregationItem は顧客集計の1件。
type OwnerAggregationItem struct {
	OwnerID            uint64
	OwnerName          string
	TotalAmount        int64
	TotalVisitCount    int64
	AnnualVisitCount   int64
	LastVisitDate      *time.Time
	FirstVisitDate     *time.Time
	AnnualAmount       *int64  // 期間内年間売上（AGG-BE-001）
	BillingCount       *int64  // 期間内会計件数（AGG-BE-001）
	PeriodVisitCount   *int64  // 期間内来院回数（AGG-BE-001/002）
	DaysSinceLastVisit *int    // 最終来院からの経過日数（AGG-BE-003）
	LastVisitBucket    *string // 最終来院分類: within_3m/over_3m/over_6m/over_1y/no_visit（AGG-BE-003）
	// CPMStage はタグ同期側 SyncCPMStageTag と同一ロジックで判定した CPM ステージ（ISSUE-006）。
	// "cpm_encounter" / "cpm_growing" / "cpm_core" / "cpm_spot" / "cpm_noah" / "cpm_dormant"
	CPMStage string
	// MaxSingleVisitAmount は完了済み請求の単一最大額（CPMスポット判定材料、ISSUE-006）。
	MaxSingleVisitAmount int64
}

// ListOwnerAggregationInput はListOwnerAggregation の入力パラメータ。
type ListOwnerAggregationInput struct {
	// 既存フィールド
	Sort           string
	MinTotalAmount *int64
	MaxTotalAmount *int64
	MinVisitCount  *int64
	LineLinked     bool
	Page           int
	PerPage        int
	// AGG-BE-001 年間売上ランキング用
	Year        *int
	From        *string // YYYY-MM-DD format
	To          *string // YYYY-MM-DD format
	AmountBasis string  // gross_total_amount (default) / paid_amount / net_paid_amount
	IncludeZero bool
	Search      string // 飼い主名部分一致検索
	// AGG-BE-002 来院回数用
	PeriodPreset  string // last_3_months / last_6_months / last_12_months / calendar_year
	MaxVisitCount *int64
	// AGG-BE-003 最終来院分類用
	LastVisitBucket string // within_3m / over_3m / over_6m / over_1y / no_visit or ""
	IncludeNoVisit  bool
	Order           string // asc / desc
	// AGG-BE-004 メトリクス選択用
	Metric string // annual_sales (default) / visit_count / last_visit
	// ISSUE-006 CPMステージ絞り込み（"cpm_spot" / "spot" / "" のいずれか）。
	// タグ同期側 SyncCPMStageTag と同一ロジックの判定結果で絞り込む。
	CPMStage string
}

// ListOwnerAggregationResult はListOwnerAggregation の結果。
type ListOwnerAggregationResult struct {
	Total   int
	Page    int
	PerPage int
	Items   []OwnerAggregationItem
}

// AggregationService は顧客集計・一括タグ同期の業務ロジックインターフェース（BE-010）。
type AggregationService interface {
	ListOwnerAggregation(ctx context.Context, clinicID uint64, input *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error)
}

type aggregationService struct {
	repo        OwnerAggregationRepository
	settingsSvc aggregationSettingsService
}

type aggregationSettingsService interface {
	GetCPMV1Thresholds(ctx context.Context, clinicID uint64) (model.CPMV1Thresholds, error)
}

// NewAggregationService は AggregationService を初期化して返す。
func NewAggregationService(repo OwnerAggregationRepository, settingsSvc aggregationSettingsService) AggregationService {
	return &aggregationService{repo: repo, settingsSvc: settingsSvc}
}

// aggregationQueryTimeout bounds the heavy LTV scan (BUG-012). Client should
// surface a retryable error instead of infinite "loading".
const aggregationQueryTimeout = 20 * time.Second

func (s *aggregationService) ListOwnerAggregation(ctx context.Context, clinicID uint64, input *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
	ctx, cancel := context.WithTimeout(ctx, aggregationQueryTimeout)
	defer cancel()

	rows, err := s.repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:        clinicID,
		Sort:            input.Sort,
		MinTotalAmount:  input.MinTotalAmount,
		MaxTotalAmount:  input.MaxTotalAmount,
		MinVisitCount:   input.MinVisitCount,
		LineLinked:      input.LineLinked,
		Year:            input.Year,
		From:            input.From,
		To:              input.To,
		AmountBasis:     input.AmountBasis,
		IncludeZero:     input.IncludeZero,
		Search:          input.Search,
		PeriodPreset:    input.PeriodPreset,
		MaxVisitCount:   input.MaxVisitCount,
		LastVisitBucket: input.LastVisitBucket,
		IncludeNoVisit:  input.IncludeNoVisit,
		Order:           input.Order,
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner aggregation")
	}

	// CPMStage 絞り込み条件を正規化（"spot" → "cpm_spot" 等）。
	cpmStageFilter := normalizeCPMStageFilter(input.CPMStage)

	thresholds, err := s.settingsSvc.GetCPMV1Thresholds(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get cpm v1 thresholds")
	}

	var items []OwnerAggregationItem
	for i := range rows {
		row := &rows[i]
		stage := computeCPMStageFromLTVRow(row, thresholds)
		if cpmStageFilter != "" && stage != cpmStageFilter {
			continue
		}
		item := OwnerAggregationItem{
			OwnerID:              row.OwnerID,
			OwnerName:            row.OwnerName,
			TotalAmount:          row.TotalAmount,
			TotalVisitCount:      row.TotalVisitCount,
			AnnualVisitCount:     row.AnnualVisitCount,
			LastVisitDate:        row.LastVisitDate,
			FirstVisitDate:       row.FirstVisitDate,
			AnnualAmount:         row.AnnualAmount,
			BillingCount:         row.BillingCount,
			PeriodVisitCount:     row.PeriodVisitCount,
			DaysSinceLastVisit:   row.DaysSinceLastVisit,
			LastVisitBucket:      row.LastVisitBucket,
			CPMStage:             stage,
			MaxSingleVisitAmount: row.MaxSingleVisitAmount,
		}
		items = append(items, item)
	}

	total := len(items)

	// Go-side pagination（C-7: aggregation は per_page 上限が無かったため maxPerPage=100 を新設）
	page, perPage, offset := normalizePagination(input.Page, input.PerPage, 50, 100)
	if offset >= total {
		items = nil
	} else {
		end := min(offset+perPage, total)
		items = items[offset:end]
	}

	return &ListOwnerAggregationResult{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Items:   items,
	}, nil
}
