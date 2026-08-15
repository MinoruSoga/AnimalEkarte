package lstep

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ISSUE-005: aggregation API レスポンス互換維持・テスト整備
//
// このテストファイルは以下を保証する:
//  1. レスポンスの正規項目固定 (annual_amount / period_visit_count / last_visit_bucket /
//     days_since_last_visit / total_visit_count / first_visit_date / last_visit_date)
//  2. 既存互換項目 (total_fee = total_amount, annual_visit_count, cpm_stage) の維持
//  3. クエリパラメータ → ListOwnerAggregationInput への変換
//  4. ソート別名解決 (resolveAggregationSort)
//  5. バリデーションエラーの 400 / 認証エラーの 401

// ---- mock AggregationService ----

type mockAggregationService struct {
	listFn func(ctx context.Context, clinicID uint64, input *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error)
}

func (m *mockAggregationService) ListOwnerAggregation(ctx context.Context, clinicID uint64, input *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID, input)
	}
	return &ListOwnerAggregationResult{}, nil
}

func newHandlerWithAggregationSvc(svc AggregationService) *Handler {
	return &Handler{aggregation: svc}
}

// intPtrLocal / int64PtrLocal / strPtrLocal / timePtrLocal はテストヘルパ。
// owner_handler_test.go や ltv_repository_test.go と命名衝突を避けるため Local サフィックスを付ける。
func intPtrLocal(v int) *int              { return &v }
func int64PtrLocal(v int64) *int64        { return &v }
func strPtrLocal(v string) *string        { return &v }
func timePtrLocal(t time.Time) *time.Time { return &t }

// ---- ResolveAggregationSort ----

// TestResolveAggregationSort はFEのsortパラメータ→repo sortフィールド変換を網羅する。
// FE が送る別名(total_fee, visit_count, last_visit, period_visit_count) が
// 既存ソートフィールド(total_amount, visit_count, last_visit_date, visit_count) に解決されること。
func TestResolveAggregationSort(t *testing.T) {
	cases := []struct {
		input  string
		expect string
	}{
		{"total_fee", "total_amount"},    // FE 互換別名
		{"total_amount", "total_amount"}, // 正規
		{"annual_amount", "annual_amount"},
		{"annual_visit_count", "annual_visit_count"},
		{"visit_count", "visit_count"},        // 別名（period_visit_count と同義）
		{"period_visit_count", "visit_count"}, // 仕様書: period_visit_count → repo: visit_count
		{"total_visit_count", "total_visit_count"},
		{"last_visit_date", "last_visit_date"},
		{"last_visit", "last_visit_date"}, // FE 互換別名
		{"days_since_last_visit", "days_since_last_visit"},
		{"owner_name", "owner_name"},
		{"", "total_amount"},              // デフォルト
		{"unknown_field", "total_amount"}, // 不正値もデフォルトに落ちる
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expect, resolveAggregationSort(tc.input))
		})
	}
}

// ---- ListOwnerAggregation: レスポンス項目固定 ----

// TestListOwnerAggregation_ResponseShape はレスポンス JSON の正規項目と互換項目が
// すべて含まれることを保証する。仕様変更時の回帰検知用。
func TestListOwnerAggregation_ResponseShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	last := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	first := time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC)

	svc := &mockAggregationService{
		listFn: func(_ context.Context, _ uint64, _ *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
			return &ListOwnerAggregationResult{
				Total:   1,
				Page:    1,
				PerPage: 50,
				Items: []OwnerAggregationItem{
					{
						OwnerID:            1,
						OwnerName:          "山田 太郎",
						TotalAmount:        156000,
						TotalVisitCount:    12,
						AnnualVisitCount:   8,
						LastVisitDate:      timePtrLocal(last),
						FirstVisitDate:     timePtrLocal(first),
						AnnualAmount:       int64PtrLocal(156000),
						BillingCount:       int64PtrLocal(12),
						PeriodVisitCount:   int64PtrLocal(4),
						DaysSinceLastVisit: intPtrLocal(47),
						LastVisitBucket:    strPtrLocal("over_3m"),
						CPMStage:           "cpm_core",
					},
				},
			}, nil
		},
	}

	h := newHandlerWithAggregationSvc(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&per_page=50", http.NoBody)
	setClinicID(c)

	h.ListOwnerAggregation(c)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// 一覧形メタ
	assert.EqualValues(t, 1, resp["total"])
	assert.EqualValues(t, 1, resp["page"])
	assert.EqualValues(t, 50, resp["per_page"])

	owners, ok := resp["owners"].([]any)
	require.True(t, ok)
	require.Len(t, owners, 1)
	owner := owners[0].(map[string]any)

	// === 正規項目 (仕様書 §2.2 / §5.3 で固定) ===
	assert.Equal(t, "1", owner["owner_id"], "owner_id は string 型で固定")
	assert.Equal(t, "山田 太郎", owner["owner_name"])
	assert.EqualValues(t, 156000, owner["annual_amount"], "annual_amount は正規項目")
	assert.EqualValues(t, 12, owner["billing_count"], "billing_count は正規項目")
	assert.EqualValues(t, 4, owner["period_visit_count"], "period_visit_count は正規項目")
	assert.EqualValues(t, 12, owner["total_visit_count"], "total_visit_count は正規項目")
	assert.EqualValues(t, 47, owner["days_since_last_visit"], "days_since_last_visit は正規項目")
	assert.Equal(t, "over_3m", owner["last_visit_bucket"], "last_visit_bucket は正規項目")
	assert.Equal(t, "2026-03-10", owner["last_visit_date"], "last_visit_date は YYYY-MM-DD")
	assert.Equal(t, "2022-05-01", owner["first_visit_date"], "first_visit_date は YYYY-MM-DD")

	// === 互換項目 (既存ダッシュボードが依存) ===
	assert.EqualValues(t, 156000, owner["total_amount"], "total_amount は互換維持")
	assert.EqualValues(t, 156000, owner["total_fee"], "total_fee は total_amount のFE別名で同値")
	assert.EqualValues(t, 8, owner["annual_visit_count"], "annual_visit_count は互換維持")
	assert.Equal(t, "cpm_core", owner["cpm_stage"], "cpm_stage はISSUE-006互換")
}

// TestListOwnerAggregation_ResponseShapeOmitEmpty は派生フィールドが nil の場合に
// JSON から omitempty で除外されることを保証する。
// FE は optional 型 (annual_amount?: number) として扱い、nil ↔ 未指定の双方を許容する。
func TestListOwnerAggregation_ResponseShapeOmitEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockAggregationService{
		listFn: func(_ context.Context, _ uint64, _ *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
			return &ListOwnerAggregationResult{
				Total: 1, Page: 1, PerPage: 50,
				Items: []OwnerAggregationItem{
					{
						OwnerID:         1,
						OwnerName:       "no derived",
						TotalAmount:     0,
						TotalVisitCount: 0,
						// 派生値はすべて nil
					},
				},
			}, nil
		},
	}
	h := newHandlerWithAggregationSvc(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	setClinicID(c)
	h.ListOwnerAggregation(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Owners []map[string]any `json:"owners"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Owners, 1)
	o := resp.Owners[0]

	// nil → omitempty で除外
	_, hasAnnual := o["annual_amount"]
	_, hasBilling := o["billing_count"]
	_, hasPeriodVC := o["period_visit_count"]
	_, hasDays := o["days_since_last_visit"]
	_, hasBucket := o["last_visit_bucket"]
	_, hasLast := o["last_visit_date"]
	_, hasFirst := o["first_visit_date"]
	assert.False(t, hasAnnual, "nil annual_amount は omitempty で除外")
	assert.False(t, hasBilling, "nil billing_count は omitempty で除外")
	assert.False(t, hasPeriodVC, "nil period_visit_count は omitempty で除外")
	assert.False(t, hasDays, "nil days_since_last_visit は omitempty で除外")
	assert.False(t, hasBucket, "nil last_visit_bucket は omitempty で除外")
	assert.False(t, hasLast, "nil last_visit_date は omitempty で除外")
	assert.False(t, hasFirst, "nil first_visit_date は omitempty で除外")

	// 必須項目は常に存在
	assert.Contains(t, o, "owner_id")
	assert.Contains(t, o, "owner_name")
	assert.Contains(t, o, "total_amount")
	assert.Contains(t, o, "total_fee")
	assert.Contains(t, o, "total_visit_count")
	assert.Contains(t, o, "annual_visit_count")
}

// ---- ListOwnerAggregation: クエリ→Input 変換 ----

// TestListOwnerAggregation_QueryToInput はすべての検索・フィルタ・ソートクエリパラメータが
// ListOwnerAggregationInput に正しくマップされることを検証する。
func TestListOwnerAggregation_QueryToInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("all parameters mapped", func(t *testing.T) {
		var captured *ListOwnerAggregationInput
		svc := &mockAggregationService{
			listFn: func(_ context.Context, clinicID uint64, in *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
				assert.Equal(t, uint64(1), clinicID)
				captured = in
				return &ListOwnerAggregationResult{}, nil
			},
		}
		q := "page=2&per_page=100" +
			"&min_total_amount=1000&max_total_amount=99999" +
			"&min_visit_count=1&max_visit_count=20" +
			"&line_linked=true" +
			"&sort=annual_amount&order=asc" +
			"&year=2026&from=2026-01-01&to=2026-12-31" +
			"&amount_basis=net_paid_amount" +
			"&include_zero=true&include_no_visit=true" +
			"&search=%E5%B1%B1%E7%94%B0" + // "山田"
			"&period_preset=last_6_months" +
			"&last_visit_bucket=over_6m" +
			"&metric=last_visit" +
			"&cpm_stage=spot"

		h := newHandlerWithAggregationSvc(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?"+q, http.NoBody)
		setClinicID(c)
		h.ListOwnerAggregation(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.NotNil(t, captured)
		assert.Equal(t, 2, captured.Page)
		assert.Equal(t, 100, captured.PerPage)
		require.NotNil(t, captured.MinTotalAmount)
		assert.Equal(t, int64(1000), *captured.MinTotalAmount)
		require.NotNil(t, captured.MaxTotalAmount)
		assert.Equal(t, int64(99999), *captured.MaxTotalAmount)
		require.NotNil(t, captured.MinVisitCount)
		assert.Equal(t, int64(1), *captured.MinVisitCount)
		require.NotNil(t, captured.MaxVisitCount)
		assert.Equal(t, int64(20), *captured.MaxVisitCount)
		assert.True(t, captured.LineLinked)
		assert.Equal(t, "annual_amount", captured.Sort)
		assert.Equal(t, "asc", captured.Order)
		require.NotNil(t, captured.Year)
		assert.Equal(t, 2026, *captured.Year)
		require.NotNil(t, captured.From)
		assert.Equal(t, "2026-01-01", *captured.From)
		require.NotNil(t, captured.To)
		assert.Equal(t, "2026-12-31", *captured.To)
		assert.Equal(t, "net_paid_amount", captured.AmountBasis)
		assert.True(t, captured.IncludeZero)
		assert.True(t, captured.IncludeNoVisit)
		assert.Equal(t, "山田", captured.Search)
		assert.Equal(t, "last_6_months", captured.PeriodPreset)
		assert.Equal(t, "over_6m", captured.LastVisitBucket)
		assert.Equal(t, "last_visit", captured.Metric)
		assert.Equal(t, "spot", captured.CPMStage)
	})

	t.Run("default values when params omitted", func(t *testing.T) {
		var captured *ListOwnerAggregationInput
		svc := &mockAggregationService{
			listFn: func(_ context.Context, _ uint64, in *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
				captured = in
				return &ListOwnerAggregationResult{}, nil
			},
		}
		h := newHandlerWithAggregationSvc(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		setClinicID(c)
		h.ListOwnerAggregation(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.NotNil(t, captured)
		// デフォルト
		assert.Equal(t, 1, captured.Page)
		assert.Equal(t, 50, captured.PerPage)
		assert.Equal(t, "total_amount", captured.Sort) // resolveAggregationSort のデフォルト
		assert.Equal(t, "desc", captured.Order)
		assert.Equal(t, "gross_total_amount", captured.AmountBasis)
		assert.Equal(t, "annual_sales", captured.Metric)
		assert.False(t, captured.IncludeZero)
		assert.False(t, captured.IncludeNoVisit)
		assert.False(t, captured.LineLinked)
		assert.Nil(t, captured.MinTotalAmount)
		assert.Nil(t, captured.MaxTotalAmount)
		assert.Nil(t, captured.Year)
		assert.Empty(t, captured.Search)
	})

	t.Run("canonical min_amount / max_amount maps to MinTotalAmount / MaxTotalAmount", func(t *testing.T) {
		// 仕様書 §4.1 の正規名 `min_amount` / `max_amount` を BE が受け付けることを保証する。
		// FE (AggregationFilterPanel) はこの正規名で送信している。
		var captured *ListOwnerAggregationInput
		svc := &mockAggregationService{
			listFn: func(_ context.Context, _ uint64, in *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
				captured = in
				return &ListOwnerAggregationResult{}, nil
			},
		}
		h := newHandlerWithAggregationSvc(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?min_amount=1000&max_amount=99999", http.NoBody)
		setClinicID(c)
		h.ListOwnerAggregation(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.NotNil(t, captured.MinTotalAmount)
		assert.Equal(t, int64(1000), *captured.MinTotalAmount, "min_amount → MinTotalAmount")
		require.NotNil(t, captured.MaxTotalAmount)
		assert.Equal(t, int64(99999), *captured.MaxTotalAmount, "max_amount → MaxTotalAmount")
	})

	t.Run("min_amount takes precedence over min_total_amount and min_total_fee", func(t *testing.T) {
		// 優先度: min_amount > min_total_amount > min_total_fee
		var captured *ListOwnerAggregationInput
		svc := &mockAggregationService{
			listFn: func(_ context.Context, _ uint64, in *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
				captured = in
				return &ListOwnerAggregationResult{}, nil
			},
		}
		h := newHandlerWithAggregationSvc(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?min_amount=100&min_total_amount=200&min_total_fee=300", http.NoBody)
		setClinicID(c)
		h.ListOwnerAggregation(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.NotNil(t, captured.MinTotalAmount)
		assert.Equal(t, int64(100), *captured.MinTotalAmount, "min_amount が最優先")
	})

	t.Run("FE alias min_total_fee maps to MinTotalAmount when min_total_amount absent", func(t *testing.T) {
		var captured *ListOwnerAggregationInput
		svc := &mockAggregationService{
			listFn: func(_ context.Context, _ uint64, in *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
				captured = in
				return &ListOwnerAggregationResult{}, nil
			},
		}
		h := newHandlerWithAggregationSvc(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?min_total_fee=500&max_total_fee=10000", http.NoBody)
		setClinicID(c)
		h.ListOwnerAggregation(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.NotNil(t, captured)
		require.NotNil(t, captured.MinTotalAmount)
		assert.Equal(t, int64(500), *captured.MinTotalAmount, "min_total_fee → MinTotalAmount")
		require.NotNil(t, captured.MaxTotalAmount)
		assert.Equal(t, int64(10000), *captured.MaxTotalAmount, "max_total_fee → MaxTotalAmount")
	})

	t.Run("min_total_amount takes precedence over min_total_fee", func(t *testing.T) {
		var captured *ListOwnerAggregationInput
		svc := &mockAggregationService{
			listFn: func(_ context.Context, _ uint64, in *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
				captured = in
				return &ListOwnerAggregationResult{}, nil
			},
		}
		h := newHandlerWithAggregationSvc(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?min_total_amount=1234&min_total_fee=999", http.NoBody)
		setClinicID(c)
		h.ListOwnerAggregation(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.NotNil(t, captured.MinTotalAmount)
		assert.Equal(t, int64(1234), *captured.MinTotalAmount)
	})

	t.Run("FE alias has_line maps to LineLinked", func(t *testing.T) {
		var captured *ListOwnerAggregationInput
		svc := &mockAggregationService{
			listFn: func(_ context.Context, _ uint64, in *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
				captured = in
				return &ListOwnerAggregationResult{}, nil
			},
		}
		h := newHandlerWithAggregationSvc(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?has_line=true", http.NoBody)
		setClinicID(c)
		h.ListOwnerAggregation(c)

		require.Equal(t, http.StatusOK, w.Code)
		assert.True(t, captured.LineLinked)
	})

	t.Run("sort=total_fee resolves to total_amount", func(t *testing.T) {
		var captured *ListOwnerAggregationInput
		svc := &mockAggregationService{
			listFn: func(_ context.Context, _ uint64, in *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
				captured = in
				return &ListOwnerAggregationResult{}, nil
			},
		}
		h := newHandlerWithAggregationSvc(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?sort=total_fee", http.NoBody)
		setClinicID(c)
		h.ListOwnerAggregation(c)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "total_amount", captured.Sort)
	})
}

// ---- ListOwnerAggregation: バリデーション ----

// TestListOwnerAggregation_Validation は不正クエリパラメータが400を返すことを保証する。
func TestListOwnerAggregation_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name  string
		query string
	}{
		{"page=0 (must be >=1)", "page=0"},
		{"page=-1", "page=-1"},
		{"page=abc (non-numeric)", "page=abc"},
		{"per_page=0", "per_page=0"},
		{"per_page=201 (over 200 max)", "per_page=201"},
		{"per_page=abc", "per_page=abc"},
		{"min_amount=-1 (canonical name negative)", "min_amount=-1"},
		{"min_amount=abc", "min_amount=abc"},
		{"max_amount=-1", "max_amount=-1"},
		{"min_total_amount=-1", "min_total_amount=-1"},
		{"min_total_amount=abc", "min_total_amount=abc"},
		{"max_total_amount=-1", "max_total_amount=-1"},
		{"min_total_fee=-1", "min_total_fee=-1"},
		{"max_total_fee=-1", "max_total_fee=-1"},
		{"min_visit_count=-1", "min_visit_count=-1"},
		{"min_visit_count=abc", "min_visit_count=abc"},
		{"max_visit_count=-1", "max_visit_count=-1"},
		{"year=1899 (below min)", "year=1899"},
		{"year=2101 (above max)", "year=2101"},
		{"year=abc", "year=abc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandlerWithAggregationSvc(&mockAggregationService{})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tc.query, http.NoBody)
			setClinicID(c)
			h.ListOwnerAggregation(c)
			assert.Equal(t, http.StatusBadRequest, w.Code, "%s should return 400", tc.name)
		})
	}
}

// TestListOwnerAggregation_AcceptsBoundaryValues は境界値を受け付けることを保証する。
func TestListOwnerAggregation_AcceptsBoundaryValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name  string
		query string
	}{
		{"page=1 (min)", "page=1"},
		{"per_page=1 (min)", "per_page=1"},
		{"per_page=200 (max)", "per_page=200"},
		{"year=1900 (min)", "year=1900"},
		{"year=2100 (max)", "year=2100"},
		{"min_total_amount=0 (zero allowed)", "min_total_amount=0"},
		{"min_visit_count=0", "min_visit_count=0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockAggregationService{}
			h := newHandlerWithAggregationSvc(svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tc.query, http.NoBody)
			setClinicID(c)
			h.ListOwnerAggregation(c)
			assert.Equal(t, http.StatusOK, w.Code, "%s should return 200", tc.name)
		})
	}
}

// ---- ListOwnerAggregation: 認証 / エラー伝播 ----

// TestListOwnerAggregation_MissingClinicID はclinic_id未設定で401を返すことを保証する。
func TestListOwnerAggregation_MissingClinicID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithAggregationSvc(&mockAggregationService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	// clinic_id を設定しない
	h.ListOwnerAggregation(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestListOwnerAggregation_ServiceError はservice層のエラーを5xxに伝播することを保証する。
func TestListOwnerAggregation_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("generic error → 500", func(t *testing.T) {
		svc := &mockAggregationService{
			listFn: func(_ context.Context, _ uint64, _ *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
				return nil, errors.New("db connection lost")
			},
		}
		h := newHandlerWithAggregationSvc(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		setClinicID(c)
		h.ListOwnerAggregation(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid input error → 400", func(t *testing.T) {
		svc := &mockAggregationService{
			listFn: func(_ context.Context, _ uint64, _ *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
				return nil, apperrors.WrapInvalidInput("bad date range")
			},
		}
		h := newHandlerWithAggregationSvc(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		setClinicID(c)
		h.ListOwnerAggregation(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ---- ListOwnerAggregation: 空リスト ----

// TestListOwnerAggregation_EmptyResult は0件結果が空配列で返ることを保証する。
func TestListOwnerAggregation_EmptyResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockAggregationService{
		listFn: func(_ context.Context, _ uint64, _ *ListOwnerAggregationInput) (*ListOwnerAggregationResult, error) {
			return &ListOwnerAggregationResult{Total: 0, Page: 1, PerPage: 50, Items: nil}, nil
		},
	}
	h := newHandlerWithAggregationSvc(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	setClinicID(c)
	h.ListOwnerAggregation(c)
	require.Equal(t, http.StatusOK, w.Code)

	var resp ownerAggregationListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 50, resp.PerPage)
	assert.NotNil(t, resp.Owners, "owners は null でなく [] であるべき (FE が map で iterate するため)")
	assert.Len(t, resp.Owners, 0)
}
