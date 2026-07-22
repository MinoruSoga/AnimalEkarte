package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- LtvRepository モック（ISSUE-006） ----

type mockLtvRepository struct {
	rows []OwnerLTVRow
	err  error
}

func (m *mockLtvRepository) FindOwnerLTV(_ context.Context, _ *FindOwnerLTVParams) ([]OwnerLTVRow, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := make([]OwnerLTVRow, len(m.rows))
	copy(out, m.rows)
	return out, nil
}

// aggIntPtr / strPtr は *int / *string ヘルパ。
func aggIntPtr(v int) *int       { return &v }
func aggStrPtr(v string) *string { return &v }
func timePtr(t time.Time) *time.Time {
	return &t
}

// TestListOwnerAggregation_CPMStageParity は LTV/CPM 一覧側で計算される CPM ステージが
// タグ同期側 SyncCPMStageTag が呼ぶ CalculateCPMStage と同一の判定結果になることを保証する（ISSUE-006）。
// CPMData の入力フィールドが揃っていれば、リポジトリ層から渡る LTV 行から計算される CPMStage と
// CalculateCPMStage を直接呼び出した結果が一致しなければならない。
func TestListOwnerAggregation_CPMStageParity(t *testing.T) {
	now := time.Now()

	type row struct {
		name string
		// LTV 集計行の値
		totalAmount   int64
		totalVisits   int64
		annualVisits  int64
		daysSince     *int
		firstVisitAgo int // first_visit_date を now から何日前にするか（負値）
		maxSingle     int64
		hasFirstVisit bool
		expectedStage string
	}
	cases := []row{
		{
			name:          "spot: 単回30,000円超 + 90日超来院なし",
			totalAmount:   35_000,
			totalVisits:   1,
			annualVisits:  1,
			daysSince:     aggIntPtr(120),
			firstVisitAgo: -120,
			maxSingle:     35_000,
			hasFirstVisit: true,
			expectedStage: "cpm_spot",
		},
		{
			name:          "dormant: 240日超来院なし",
			totalAmount:   100_000,
			totalVisits:   5,
			annualVisits:  0,
			daysSince:     aggIntPtr(300),
			firstVisitAgo: -800,
			maxSingle:     50_000,
			hasFirstVisit: true,
			expectedStage: "cpm_dormant",
		},
		{
			name:          "noah: 在籍1年以上・年間3回以上・LTV 80,000円以上",
			totalAmount:   100_000,
			totalVisits:   10,
			annualVisits:  3,
			daysSince:     aggIntPtr(30),
			firstVisitAgo: -400,
			maxSingle:     20_000,
			hasFirstVisit: true,
			expectedStage: "cpm_noah",
		},
		{
			name:          "core: 在籍180日以上・年間2回以上・LTV 50,000円以上",
			totalAmount:   60_000,
			totalVisits:   4,
			annualVisits:  2,
			daysSince:     aggIntPtr(30),
			firstVisitAgo: -200,
			maxSingle:     20_000,
			hasFirstVisit: true,
			expectedStage: "cpm_core",
		},
		{
			name:          "growing: 初診から90日以内に再来院",
			totalAmount:   25_000,
			totalVisits:   2,
			annualVisits:  2,
			daysSince:     aggIntPtr(30),
			firstVisitAgo: -50,
			maxSingle:     15_000,
			hasFirstVisit: true,
			expectedStage: "cpm_growing",
		},
		{
			name:          "growing: 初診90日以内・直近来院あり (新ロジック該当)",
			totalAmount:   25_000,
			totalVisits:   2,
			annualVisits:  2,
			daysSince:     aggIntPtr(30),
			firstVisitAgo: -80,
			maxSingle:     15_000,
			hasFirstVisit: true,
			expectedStage: "cpm_growing",
		},
		{
			name:          "growing 非該当: 初診90日超 (旧ロジックでは該当、明示判定後 unclassified)",
			totalAmount:   25_000,
			totalVisits:   2,
			annualVisits:  2,
			daysSince:     aggIntPtr(30),
			firstVisitAgo: -200,
			maxSingle:     15_000,
			hasFirstVisit: true,
			expectedStage: "cpm_unclassified",
		},
		{
			name:          "growing 非該当: LTV 2万未満 (仕様書整合 LTV 条件、visit=2 で unclassified)",
			totalAmount:   15_000,
			totalVisits:   2,
			annualVisits:  2,
			daysSince:     aggIntPtr(30),
			firstVisitAgo: -50,
			maxSingle:     10_000,
			hasFirstVisit: true,
			expectedStage: "cpm_unclassified",
		},
		{
			name:          "growing 非該当: LTV 5万以上 (仕様書整合 LTV 条件、visit=2 で unclassified)",
			totalAmount:   60_000,
			totalVisits:   2,
			annualVisits:  2,
			daysSince:     aggIntPtr(30),
			firstVisitAgo: -50,
			maxSingle:     30_000,
			hasFirstVisit: true,
			expectedStage: "cpm_unclassified",
		},
		{
			name:          "encounter: 単回低額・初回来院",
			totalAmount:   5_000,
			totalVisits:   1,
			annualVisits:  1,
			daysSince:     aggIntPtr(10),
			firstVisitAgo: -10,
			maxSingle:     5_000,
			hasFirstVisit: true,
			expectedStage: "cpm_encounter",
		},
		{
			name:          "no_visit (DaysSinceLastVisit nil) → dormant",
			totalAmount:   0,
			totalVisits:   0,
			annualVisits:  0,
			daysSince:     nil,
			firstVisitAgo: 0,
			maxSingle:     0,
			hasFirstVisit: false,
			expectedStage: "cpm_dormant",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ltvRow := OwnerLTVRow{
				OwnerID:              1,
				OwnerName:            tc.name,
				TotalAmount:          tc.totalAmount,
				TotalVisitCount:      tc.totalVisits,
				AnnualVisitCount:     tc.annualVisits,
				DaysSinceLastVisit:   tc.daysSince,
				MaxSingleVisitAmount: tc.maxSingle,
				// include_zero / include_no_visit を回避
				AnnualAmount:    func() *int64 { v := int64(1); return &v }(),
				LastVisitBucket: aggStrPtr("over_3m"),
			}
			if tc.hasFirstVisit {
				ltvRow.FirstVisitDate = timePtr(now.AddDate(0, 0, tc.firstVisitAgo))
			}

			repoMock := &mockLtvRepository{rows: []OwnerLTVRow{ltvRow}}
			svc := NewAggregationService(repoMock, &mockLstepSettingsService{})

			result, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{
				IncludeZero:    true,
				IncludeNoVisit: true,
			})
			assert.NoError(t, err)
			assert.Len(t, result.Items, 1)

			// 一覧側で計算された CPM ステージ
			gotStage := result.Items[0].CPMStage

			// タグ同期側と同じ判定材料で CalculateCPMStage を直接呼ぶ
			daysSince := -1
			if tc.daysSince != nil {
				daysSince = *tc.daysSince
			}
			firstVisitDays := 0
			if tc.hasFirstVisit {
				firstVisitDays = int(time.Since(now.AddDate(0, 0, tc.firstVisitAgo)).Hours() / 24)
			}
			expected := string(CalculateCPMStage(CPMData{
				TotalVisitCount:      tc.totalVisits,
				AnnualVisitCount:     tc.annualVisits,
				DaysSinceVisit:       daysSince,
				LTVAmount:            tc.totalAmount,
				FirstVisitDaysSince:  firstVisitDays,
				MaxSingleVisitAmount: tc.maxSingle,
			}))

			assert.Equal(t, expected, gotStage, "一覧側とタグ同期側の CPM 判定が一致しない")
			assert.Equal(t, tc.expectedStage, gotStage, "期待ステージが一致しない")
		})
	}
}

// TestListOwnerAggregation_CPMStageFilter は cpm_stage クエリフィルタが
// CPM ステージ判定結果と一致する飼い主のみを抽出することを検証する（ISSUE-006）。
// "spot" / "cpm_spot" の両表記を受け付ける。
func TestListOwnerAggregation_CPMStageFilter(t *testing.T) {
	now := time.Now()

	rows := []OwnerLTVRow{
		{
			OwnerID:              1,
			OwnerName:            "spot owner",
			TotalAmount:          35_000,
			TotalVisitCount:      1,
			AnnualVisitCount:     1,
			DaysSinceLastVisit:   aggIntPtr(120),
			FirstVisitDate:       timePtr(now.AddDate(0, 0, -120)),
			MaxSingleVisitAmount: 35_000,
			AnnualAmount:         func() *int64 { v := int64(35_000); return &v }(),
			LastVisitBucket:      aggStrPtr("over_3m"),
		},
		{
			OwnerID:              2,
			OwnerName:            "encounter owner",
			TotalAmount:          5_000,
			TotalVisitCount:      1,
			AnnualVisitCount:     1,
			DaysSinceLastVisit:   aggIntPtr(10),
			FirstVisitDate:       timePtr(now.AddDate(0, 0, -10)),
			MaxSingleVisitAmount: 5_000,
			AnnualAmount:         func() *int64 { v := int64(5_000); return &v }(),
			LastVisitBucket:      aggStrPtr("within_3m"),
		},
	}

	t.Run("cpm_stage=spot は spot のみ抽出", func(t *testing.T) {
		repoMock := &mockLtvRepository{rows: rows}
		svc := NewAggregationService(repoMock, &mockLstepSettingsService{})
		result, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{
			CPMStage: "spot",
		})
		assert.NoError(t, err)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "spot owner", result.Items[0].OwnerName)
		assert.Equal(t, "cpm_spot", result.Items[0].CPMStage)
	})

	t.Run("cpm_stage=cpm_spot もエイリアスとして受け付ける", func(t *testing.T) {
		repoMock := &mockLtvRepository{rows: rows}
		svc := NewAggregationService(repoMock, &mockLstepSettingsService{})
		result, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{
			CPMStage: "cpm_spot",
		})
		assert.NoError(t, err)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "spot owner", result.Items[0].OwnerName)
	})

	t.Run("cpm_stage 未指定は全件返す", func(t *testing.T) {
		repoMock := &mockLtvRepository{rows: rows}
		svc := NewAggregationService(repoMock, &mockLstepSettingsService{})
		result, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{})
		assert.NoError(t, err)
		assert.Len(t, result.Items, 2)
	})
}

// TestNormalizeCPMStageFilter は cpm_stage クエリパラメータの正規化を検証する（ISSUE-006）。
func TestNormalizeCPMStageFilter(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"   ", ""},
		{"spot", "cpm_spot"},
		{"SPOT", "cpm_spot"},
		{"cpm_spot", "cpm_spot"},
		{"CPM_SPOT", "cpm_spot"},
		{"noah", "cpm_noah"},
		{"dormant", "cpm_dormant"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, normalizeCPMStageFilter(tc.input))
		})
	}
}

// ISSUE-005 ====================================================================
// ListOwnerAggregation のレスポンス互換維持・ページング・エラー処理のテスト
// =============================================================================

// TestListOwnerAggregation_FieldMapping は OwnerLTVRow → OwnerAggregationItem の
// フィールドマッピングが完全であることを保証する。仕様変更時の回帰検知用。
func TestListOwnerAggregation_FieldMapping(t *testing.T) {
	last := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	first := time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC)

	row := OwnerLTVRow{
		OwnerID:              42,
		OwnerName:            "山田 太郎",
		LineUserID:           aggStrPtr("U-line-id"),
		LstepOptOut:          false,
		TotalAmount:          156_000,
		TotalVisitCount:      12,
		AnnualVisitCount:     8,
		LastVisitDate:        timePtr(last),
		FirstVisitDate:       timePtr(first),
		AnnualAmount:         func() *int64 { v := int64(156_000); return &v }(),
		BillingCount:         func() *int64 { v := int64(12); return &v }(),
		PeriodVisitCount:     func() *int64 { v := int64(4); return &v }(),
		DaysSinceLastVisit:   aggIntPtr(47),
		LastVisitBucket:      aggStrPtr("over_3m"),
		MaxSingleVisitAmount: 35_000,
	}
	repoMock := &mockLtvRepository{rows: []OwnerLTVRow{row}}
	svc := NewAggregationService(repoMock, &mockLstepSettingsService{})

	result, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{})
	require.NoError(t, err)
	require.Len(t, result.Items, 1)

	got := result.Items[0]
	assert.Equal(t, uint64(42), got.OwnerID)
	assert.Equal(t, "山田 太郎", got.OwnerName)
	assert.Equal(t, int64(156_000), got.TotalAmount)
	assert.Equal(t, int64(12), got.TotalVisitCount)
	assert.Equal(t, int64(8), got.AnnualVisitCount)
	require.NotNil(t, got.LastVisitDate)
	assert.Equal(t, last, *got.LastVisitDate)
	require.NotNil(t, got.FirstVisitDate)
	assert.Equal(t, first, *got.FirstVisitDate)
	require.NotNil(t, got.AnnualAmount)
	assert.Equal(t, int64(156_000), *got.AnnualAmount)
	require.NotNil(t, got.BillingCount)
	assert.Equal(t, int64(12), *got.BillingCount)
	require.NotNil(t, got.PeriodVisitCount)
	assert.Equal(t, int64(4), *got.PeriodVisitCount)
	require.NotNil(t, got.DaysSinceLastVisit)
	assert.Equal(t, 47, *got.DaysSinceLastVisit)
	require.NotNil(t, got.LastVisitBucket)
	assert.Equal(t, "over_3m", *got.LastVisitBucket)
	assert.Equal(t, int64(35_000), got.MaxSingleVisitAmount)
	assert.NotEmpty(t, got.CPMStage, "CPMStage はLTV行から派生計算される")
}

// TestListOwnerAggregation_Pagination はGo側ページング (offset/limit) を検証する。
// repo は全件返し、service が page/per_page でスライスする実装になっている。
func TestListOwnerAggregation_Pagination(t *testing.T) {
	// 5 件のオーナー (annual_amount は include_zero 回避用に正値)
	rows := make([]OwnerLTVRow, 5)
	for i := range rows {
		v := int64(100 + i)
		rows[i] = OwnerLTVRow{
			OwnerID:         uint64(i + 1),
			OwnerName:       "owner",
			AnnualAmount:    &v,
			LastVisitBucket: aggStrPtr("within_3m"),
		}
	}

	t.Run("page=1, per_page=2 returns first 2", func(t *testing.T) {
		repoMock := &mockLtvRepository{rows: rows}
		svc := NewAggregationService(repoMock, &mockLstepSettingsService{})
		r, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{Page: 1, PerPage: 2})
		require.NoError(t, err)
		assert.Equal(t, 5, r.Total, "Total は全件数")
		assert.Equal(t, 1, r.Page)
		assert.Equal(t, 2, r.PerPage)
		assert.Len(t, r.Items, 2)
		assert.Equal(t, uint64(1), r.Items[0].OwnerID)
		assert.Equal(t, uint64(2), r.Items[1].OwnerID)
	})

	t.Run("page=3, per_page=2 returns last item", func(t *testing.T) {
		repoMock := &mockLtvRepository{rows: rows}
		svc := NewAggregationService(repoMock, &mockLstepSettingsService{})
		r, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{Page: 3, PerPage: 2})
		require.NoError(t, err)
		assert.Equal(t, 5, r.Total)
		assert.Len(t, r.Items, 1)
		assert.Equal(t, uint64(5), r.Items[0].OwnerID)
	})

	t.Run("page beyond total returns empty items but keeps Total", func(t *testing.T) {
		repoMock := &mockLtvRepository{rows: rows}
		svc := NewAggregationService(repoMock, &mockLstepSettingsService{})
		r, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{Page: 99, PerPage: 10})
		require.NoError(t, err)
		assert.Equal(t, 5, r.Total)
		assert.Len(t, r.Items, 0)
	})

	t.Run("zero page/per_page falls back to defaults", func(t *testing.T) {
		repoMock := &mockLtvRepository{rows: rows}
		svc := NewAggregationService(repoMock, &mockLstepSettingsService{})
		r, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{Page: 0, PerPage: 0})
		require.NoError(t, err)
		assert.Equal(t, 1, r.Page)
		assert.Equal(t, 50, r.PerPage)
		assert.Len(t, r.Items, 5, "全件 (5件) が1ページに収まる")
	})

	t.Run("per_page over 100 is capped at 100 (C-7)", func(t *testing.T) {
		repoMock := &mockLtvRepository{rows: rows}
		svc := NewAggregationService(repoMock, &mockLstepSettingsService{})
		r, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{Page: 1, PerPage: 500})
		require.NoError(t, err)
		assert.Equal(t, 100, r.PerPage, "maxPerPage=100 に切り詰められる")
		assert.Len(t, r.Items, 5, "元データが5件なので全件返るが PerPage 表示は100")
	})

	t.Run("CPM filter applied before pagination affects Total", func(t *testing.T) {
		// CPMStage 絞り込みは pagination より先に適用されるため Total に反映される。
		// 5件のうち 1件だけ encounter になるよう仕立てる。
		now := time.Now()
		rows := []OwnerLTVRow{
			// encounter: 単回低額・初回来院
			{
				OwnerID: 1, OwnerName: "encounter",
				TotalAmount: 5_000, TotalVisitCount: 1, AnnualVisitCount: 1,
				DaysSinceLastVisit:   aggIntPtr(10),
				FirstVisitDate:       timePtr(now.AddDate(0, 0, -10)),
				MaxSingleVisitAmount: 5_000,
				AnnualAmount:         func() *int64 { v := int64(5_000); return &v }(),
				LastVisitBucket:      aggStrPtr("within_3m"),
			},
			// 残りは spot ステージ判定にする
			{
				OwnerID: 2, OwnerName: "spot a",
				TotalAmount: 35_000, TotalVisitCount: 1, AnnualVisitCount: 1,
				DaysSinceLastVisit:   aggIntPtr(120),
				FirstVisitDate:       timePtr(now.AddDate(0, 0, -120)),
				MaxSingleVisitAmount: 35_000,
				AnnualAmount:         func() *int64 { v := int64(35_000); return &v }(),
				LastVisitBucket:      aggStrPtr("over_3m"),
			},
		}
		repoMock := &mockLtvRepository{rows: rows}
		svc := NewAggregationService(repoMock, &mockLstepSettingsService{})
		r, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{
			CPMStage: "encounter",
			Page:     1, PerPage: 10,
		})
		require.NoError(t, err)
		// CPM フィルタ通過後は 1 件のみ
		assert.Equal(t, 1, r.Total)
		require.Len(t, r.Items, 1)
		assert.Equal(t, "encounter", r.Items[0].OwnerName)
	})
}

// TestListOwnerAggregation_RepoError は repository エラーが apperrors でラップされて
// 返ることを保証する。
func TestListOwnerAggregation_RepoError(t *testing.T) {
	wantErr := errors.New("connection refused")
	repoMock := &mockLtvRepository{err: wantErr}
	svc := NewAggregationService(repoMock, &mockLstepSettingsService{})
	r, err := svc.ListOwnerAggregation(context.Background(), 1, &ListOwnerAggregationInput{})
	require.Error(t, err)
	assert.Nil(t, r)
	// apperrors.Wrap でラップされる (P8)
	assert.Contains(t, err.Error(), "failed to find owner aggregation")
	// 元エラーが errors.Is で取得可能
	assert.ErrorIs(t, err, wantErr)
}

// TestListOwnerAggregation_InputForwarding はservice入力がrepoパラメータに
// そのまま転記されることを保証する。検索条件・期間・ソートが渡らないと回帰につながる。
func TestListOwnerAggregation_InputForwarding(t *testing.T) {
	var captured *FindOwnerLTVParams
	repoMock := &mockLtvRepositoryCapture{
		fn: func(_ context.Context, p *FindOwnerLTVParams) ([]OwnerLTVRow, error) {
			captured = p
			return nil, nil
		},
	}
	svc := NewAggregationService(repoMock, &mockLstepSettingsService{})

	year := 2026
	from := "2026-01-01"
	to := "2026-12-31"
	minAmt := int64(1000)
	maxAmt := int64(99999)
	minVC := int64(1)
	maxVC := int64(20)
	in := &ListOwnerAggregationInput{
		Sort:            "annual_amount",
		Order:           "asc",
		MinTotalAmount:  &minAmt,
		MaxTotalAmount:  &maxAmt,
		MinVisitCount:   &minVC,
		MaxVisitCount:   &maxVC,
		LineLinked:      true,
		Page:            1,
		PerPage:         50,
		Year:            &year,
		From:            &from,
		To:              &to,
		AmountBasis:     "net_paid_amount",
		IncludeZero:     true,
		Search:          "山田",
		PeriodPreset:    "last_6_months",
		LastVisitBucket: "over_6m",
		IncludeNoVisit:  true,
		Metric:          "last_visit",
	}
	_, err := svc.ListOwnerAggregation(context.Background(), 7, in)
	require.NoError(t, err)
	require.NotNil(t, captured)

	assert.Equal(t, uint64(7), captured.ClinicID, "clinic_id 強制")
	assert.Equal(t, "annual_amount", captured.Sort)
	assert.Equal(t, "asc", captured.Order)
	require.NotNil(t, captured.MinTotalAmount)
	assert.Equal(t, int64(1000), *captured.MinTotalAmount)
	require.NotNil(t, captured.MaxTotalAmount)
	assert.Equal(t, int64(99999), *captured.MaxTotalAmount)
	require.NotNil(t, captured.MinVisitCount)
	assert.Equal(t, int64(1), *captured.MinVisitCount)
	require.NotNil(t, captured.MaxVisitCount)
	assert.Equal(t, int64(20), *captured.MaxVisitCount)
	assert.True(t, captured.LineLinked)
	require.NotNil(t, captured.Year)
	assert.Equal(t, 2026, *captured.Year)
	require.NotNil(t, captured.From)
	assert.Equal(t, "2026-01-01", *captured.From)
	require.NotNil(t, captured.To)
	assert.Equal(t, "2026-12-31", *captured.To)
	assert.Equal(t, "net_paid_amount", captured.AmountBasis)
	assert.True(t, captured.IncludeZero)
	assert.Equal(t, "山田", captured.Search)
	assert.Equal(t, "last_6_months", captured.PeriodPreset)
	assert.Equal(t, "over_6m", captured.LastVisitBucket)
	assert.True(t, captured.IncludeNoVisit)
	// Metric は service が repo に渡さない (Metric は表示用)
}

// ---- 補助モック ----

// mockLtvRepositoryCapture は FindOwnerLTV 引数キャプチャ用。
type mockLtvRepositoryCapture struct {
	fn func(ctx context.Context, p *FindOwnerLTVParams) ([]OwnerLTVRow, error)
}

func (m *mockLtvRepositoryCapture) FindOwnerLTV(ctx context.Context, p *FindOwnerLTVParams) ([]OwnerLTVRow, error) {
	if m.fn != nil {
		return m.fn(ctx, p)
	}
	return nil, nil
}
