package lstep

import (
	"context"
	"errors"
	"fmt"
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
	// CPMStageCounts は cpm_stage 絞り込み・ページネーション前の母集団全体の
	// CPM 区分人数 (EMR-334)。aggregationCPMStageKeys の全 7 キーを常に持つ。
	CPMStageCounts map[string]int
}

// aggregationCPMStageKeys は集計レスポンス cpm_stage_counts が必ず返すキー集合 (EMR-334)。
// 配信対象の allCPMStages (6区分) に、画面チップで辿れる未分類を加えた全7区分。
var aggregationCPMStageKeys = []CPMStage{
	CPMStageEncounter, CPMStageGrowing, CPMStageCore, CPMStageSpot,
	CPMStageNoah, CPMStageDormant, CPMStageUnclassified,
}

// AggregationService は顧客集計・一括タグ同期の業務ロジックインターフェース（BE-010）。
type AggregationService interface {
	ListOwnerAggregation(ctx context.Context, clinicID uint64, input *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error)
}

type aggregationService struct {
	repo        OwnerAggregationRepository
	settingsSvc aggregationSettingsService
	// queryTimeout は集計クエリ自身の deadline (既定 aggregationQueryTimeout)。
	// テストがタイマー失効を実時間待ちせず再現できるようにするための duration seam。
	queryTimeout time.Duration
}

type aggregationSettingsService interface {
	GetCPMV1Thresholds(ctx context.Context, clinicID uint64) (model.CPMV1Thresholds, error)
}

// NewAggregationService は AggregationService を初期化して返す。
func NewAggregationService(repo OwnerAggregationRepository, settingsSvc aggregationSettingsService) AggregationService {
	return &aggregationService{repo: repo, settingsSvc: settingsSvc, queryTimeout: aggregationQueryTimeout}
}

// aggregationQueryTimeout bounds the heavy LTV scan (BUG-012). Client should
// surface a retryable error instead of infinite "loading".
const aggregationQueryTimeout = 20 * time.Second

// errAggregationQueryTimeout は集計クエリ自身の deadline (aggregationQueryTimeout) が
// 失効したことを表す sentinel (EMR-334)。依存先が返す DeadlineExceeded や呼び出し元の
// 期限切れ・cancel と区別し、ハンドラ側の専用 503 (code=aggregation_timeout) 判定に使う。
var errAggregationQueryTimeout = errors.New("aggregation query timeout")

// aggregationError は集計クエリ自身のタイマー失効だけを sentinel errAggregationQueryTimeout
// に包み替え、それ以外は従来どおり apperrors で包む。
// context.WithTimeoutCause により context.Cause は、自身のタイマー失効時のみ sentinel を返す
// (親 ctx の期限切れ/cancel では親側の cause がそのまま返る)。
func aggregationError(ctx context.Context, err error, msg string) error {
	if errors.Is(context.Cause(ctx), errAggregationQueryTimeout) {
		return fmt.Errorf("%w: %w", errAggregationQueryTimeout, err)
	}
	return apperrors.Wrap(err, msg)
}

func (s *aggregationService) ListOwnerAggregation(ctx context.Context, clinicID uint64, input *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
	ctx, cancel := context.WithTimeoutCause(ctx, s.queryTimeout, errAggregationQueryTimeout)
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
		return nil, aggregationError(ctx, err, "failed to find owner aggregation")
	}
	// EMR334-1: ctx を見ずに期限後に成功を返す依存や、クエリ直後の失効を拾い、
	// 期限超過済みの集計を 200 として返さない。
	if err := ctx.Err(); err != nil {
		return nil, aggregationError(ctx, err, "owner aggregation aborted after ltv scan")
	}

	// CPMStage 絞り込み条件を正規化（"spot" → "cpm_spot" 等）。
	cpmStageFilter := normalizeCPMStageFilter(input.CPMStage)

	thresholds, err := s.settingsSvc.GetCPMV1Thresholds(ctx, clinicID)
	if err != nil {
		return nil, aggregationError(ctx, err, "failed to get cpm v1 thresholds")
	}

	// EMR-334: 区分人数は cpm_stage 絞り込み・ページネーション前の母集団全体から数える。
	// 一覧と同じ LTV スキャン結果を共有し、区分ごとの再走査は行わない。
	counts := make(map[string]int, len(aggregationCPMStageKeys))
	for _, stage := range aggregationCPMStageKeys {
		counts[string(stage)] = 0
	}

	var items []OwnerAggregationItem
	for i := range rows {
		// EMR334-1: 全行 CPM 分類・区分人数集計中の deadline/cancel 失効を
		// 各行の区切りで再確認し、処理を早期に中断する。
		if err := ctx.Err(); err != nil {
			return nil, aggregationError(ctx, err, "owner aggregation aborted during cpm classification")
		}
		row := &rows[i]
		stage := computeCPMStageFromLTVRow(row, thresholds)
		counts[stage]++
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

	// EMR334-1: 0 行やフィルタ全落ちでループを通らない場合、およびページング後の
	// 成功返却直前の失効を拾う最終確認。
	if err := ctx.Err(); err != nil {
		return nil, aggregationError(ctx, err, "owner aggregation aborted before response")
	}

	return &ListOwnerAggregationResult{
		Total:          total,
		Page:           page,
		PerPage:        perPage,
		Items:          items,
		CPMStageCounts: counts,
	}, nil
}
