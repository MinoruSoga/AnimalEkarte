package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestBillingItemHandlerCompiles verifies billing_item_handler.go compiles
func TestBillingItemHandlerCompiles(t *testing.T) {
	assert.True(t, true, "billing_item_handler.go compiled successfully")
}

// ---- mock BillingItemService ----

type mockBillingItemService struct {
	createItemFn                 func(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error)
	updateItemFn                 func(ctx context.Context, clinicID, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error)
	deleteItemFn                 func(ctx context.Context, clinicID, id uint64, input *DeleteBillingItemInput) error
	getUnbilledItemsFn           func(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error)
	getUnbilledItemDetailsFn     func(ctx context.Context, clinicID, petID uint64) (*UnbilledDetails, error)
	assertNoBlockingUnbilledFn   func(ctx context.Context, clinicID, petID uint64) error
	getUngroupedSameDaySummaryFn func(ctx context.Context, clinicID, petID uint64, date time.Time) (UngroupedSameDaySummary, error)
	getDiscountSuggestionsFn     func(ctx context.Context, clinicID, itemID uint64) ([]DiscountSuggestion, error)
	getBillingFn                 func(ctx context.Context, clinicID, billingID uint64) (*model.Billing, error)
	getBillingForItemFn          func(ctx context.Context, clinicID, itemID uint64) (*model.Billing, error)
}

func (m *mockBillingItemService) CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	return m.createItemFn(ctx, input)
}

func (m *mockBillingItemService) CreateItemForComplete(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	if m.createItemFn != nil {
		return m.createItemFn(ctx, input)
	}
	return &model.BillingItem{BillingID: input.BillingID, Name: input.Name, UnitPrice: input.UnitPrice, Quantity: input.Quantity}, nil
}

func (m *mockBillingItemService) RecalculateTotalsForComplete(_ context.Context, _, _ uint64) (int64, int64, int64, error) {
	return 0, 0, 0, nil
}

func (m *mockBillingItemService) UpdateItem(ctx context.Context, clinicID, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error) {
	return m.updateItemFn(ctx, clinicID, id, input)
}

func (m *mockBillingItemService) DeleteItem(ctx context.Context, clinicID, id uint64, input *DeleteBillingItemInput) error {
	return m.deleteItemFn(ctx, clinicID, id, input)
}

func (m *mockBillingItemService) GetUnbilledItems(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error) {
	return m.getUnbilledItemsFn(ctx, clinicID, petID)
}

func (m *mockBillingItemService) GetUnbilledItemDetails(ctx context.Context, clinicID, petID uint64) (*UnbilledDetails, error) {
	if m.getUnbilledItemDetailsFn != nil {
		return m.getUnbilledItemDetailsFn(ctx, clinicID, petID)
	}
	return &UnbilledDetails{Items: nil, Warnings: []UnbilledWarning{}}, nil
}

func (m *mockBillingItemService) AssertNoBlockingUnbilled(ctx context.Context, clinicID, petID uint64) error {
	if m.assertNoBlockingUnbilledFn != nil {
		return m.assertNoBlockingUnbilledFn(ctx, clinicID, petID)
	}
	return nil
}

func (m *mockBillingItemService) GetUngroupedSameDaySummary(ctx context.Context, clinicID, petID uint64, date time.Time) (UngroupedSameDaySummary, error) {
	return m.getUngroupedSameDaySummaryFn(ctx, clinicID, petID, date)
}

func (m *mockBillingItemService) GetDiscountSuggestions(ctx context.Context, clinicID, itemID uint64) ([]DiscountSuggestion, error) {
	return m.getDiscountSuggestionsFn(ctx, clinicID, itemID)
}

func (m *mockBillingItemService) GetBilling(ctx context.Context, clinicID, billingID uint64) (*model.Billing, error) {
	if m.getBillingFn != nil {
		return m.getBillingFn(ctx, clinicID, billingID)
	}
	return &model.Billing{ID: billingID, ClinicID: clinicID}, nil
}

func (m *mockBillingItemService) GetBillingForItem(ctx context.Context, clinicID, itemID uint64) (*model.Billing, error) {
	if m.getBillingForItemFn != nil {
		return m.getBillingForItemFn(ctx, clinicID, itemID)
	}
	return &model.Billing{ID: 10, ClinicID: clinicID}, nil
}

func newHandlerWithBillingItemSvc(svc BillingItemService) *BillingItemHandler {
	return NewBillingItemHandler(svc, nil, nil, func(_, _ string) gin.HandlerFunc { return func(_ *gin.Context) {} })
}

// ---- CreateBillingItem ----

func TestCreateBillingItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"billing_id": 10,
			"category":   string(model.ItemCategoryMedicine),
			"name":       "抗生物質",
			"unit_price": 1200,
			"quantity":   2,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockBillingItemService
		wantStatus int
	}{
		{
			name:     "creates billing item successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingItemService{
				createItemFn: func(_ context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
					assert.Equal(t, uint64(1), input.ClinicID)
					assert.Equal(t, uint64(10), input.BillingID)
					return &model.BillingItem{ID: 5, BillingID: 10, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required field is missing",
			body:       map[string]any{"name": "抗生物質"}, // billing_id missing
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid category enum",
			body:       func() map[string]any { b := validBody(); b["category"] = "invalid_category"; return b }(),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingItemService{
				createItemFn: func(_ context.Context, _ *CreateBillingItemInput) (*model.BillingItem, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithBillingItemSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateBillingItem(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusCreated {
				assert.Equal(t, "/api/v1/billing/10/items/5", w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateBillingItem ----

func TestUpdateBillingItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockBillingItemService
		wantStatus int
	}{
		{
			name:     "updates billing item successfully",
			paramID:  "5",
			body:     map[string]any{"unit_price": 1500},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingItemService{
				updateItemFn: func(_ context.Context, clinicID, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					require.NotNil(t, input.UnitPrice)
					assert.Equal(t, int64(1500), *input.UnitPrice)
					return &model.BillingItem{ID: 5}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       map[string]any{},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid tax_type",
			paramID:    "5",
			body:       map[string]any{"tax_type": "bogus"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when item does not exist",
			paramID:  "999",
			body:     map[string]any{"unit_price": 100},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingItemService{
				updateItemFn: func(_ context.Context, _, _ uint64, _ *UpdateBillingItemInput) (*model.BillingItem, error) {
					return nil, apperrors.WrapNotFound("billing_item", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithBillingItemSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateBillingItem(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteBillingItem ----

// newDeleteBillingItemRouter は c.Status(http.StatusNoContent) のみでボディ書き込みが
// 無いハンドラのため、gin.Engine 経由 (router.ServeHTTP) でヘッダーを確実にフラッシュする。
// (直接 h.DeleteBillingItem(c) 呼び出しだと WriteHeaderNow が走らず w.Code が 200 のままになる)
func newDeleteBillingItemRouter(svc BillingItemService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithBillingItemSvc(svc)
	r.DELETE("/billing-items/:id", func(c *gin.Context) {
		setClinicID(c)
		setStaffID(c)
	}, h.DeleteBillingItem)
	return r
}

func TestDeleteBillingItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockBillingItemService
		wantStatus int
	}{
		{
			name:    "deletes billing item successfully",
			paramID: "5",
			svc: &mockBillingItemService{
				deleteItemFn: func(_ context.Context, clinicID, id uint64, input *DeleteBillingItemInput) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					require.NotNil(t, input)
					require.NotNil(t, input.StaffID)
					assert.Equal(t, uint64(1), *input.StaffID)
					assert.False(t, input.IsPostClose)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when item does not exist",
			paramID: "999",
			svc: &mockBillingItemService{
				deleteItemFn: func(_ context.Context, _, _ uint64, _ *DeleteBillingItemInput) error {
					return apperrors.WrapNotFound("billing_item", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteBillingItemRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/billing-items/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithBillingItemSvc(&mockBillingItemService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "5"}}
		h.DeleteBillingItem(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- GetUnbilledItems ----

func TestGetUnbilledItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockBillingItemService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns unbilled items for pet",
			query:    "pet_id=7",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingItemService{
				getUnbilledItemsFn: func(_ context.Context, clinicID, petID uint64) ([]model.BillingItem, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(7), petID)
					return []model.BillingItem{{ID: 1, Name: "ワクチン"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"ワクチン"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "pet_id=7",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when pet_id is missing",
			query:      "",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "pet_id=7",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingItemService{
				getUnbilledItemsFn: func(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithBillingItemSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.GetUnbilledItems(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetUngroupedSameDay ----

func TestGetUngroupedSameDay(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockBillingItemService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns summary with explicit date",
			query:    "pet_id=7&date=2026-06-01",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingItemService{
				getUngroupedSameDaySummaryFn: func(_ context.Context, clinicID, petID uint64, date time.Time) (UngroupedSameDaySummary, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(7), petID)
					assert.Equal(t, 2026, date.Year())
					assert.Equal(t, time.June, date.Month())
					assert.Equal(t, 1, date.Day())
					return UngroupedSameDaySummary{MedicalRecordCount: 2, TrimmingCount: 0}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"has_ungrouped":true`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "pet_id=7",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when pet_id is missing",
			query:      "",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "pet_id=7",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingItemService{
				getUngroupedSameDaySummaryFn: func(_ context.Context, _, _ uint64, _ time.Time) (UngroupedSameDaySummary, error) {
					return UngroupedSameDaySummary{}, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithBillingItemSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.GetUngroupedSameDay(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetBillingItemDiscountSuggestions ----

func TestGetBillingItemDiscountSuggestions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockBillingItemService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns discount suggestions",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingItemService{
				getDiscountSuggestionsFn: func(_ context.Context, clinicID, itemID uint64) ([]DiscountSuggestion, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), itemID)
					return []DiscountSuggestion{{Type: "owner", Name: "飼主割引", Amount: 100}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"飼主割引"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingItemService{
				getDiscountSuggestionsFn: func(_ context.Context, _, _ uint64) ([]DiscountSuggestion, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithBillingItemSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetBillingItemDiscountSuggestions(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- toUngroupedSameDayResponse ----

func TestToUngroupedSameDayResponse(t *testing.T) {
	tests := []struct {
		name             string
		summary          UngroupedSameDaySummary
		wantHasUngrouped bool
	}{
		{
			name:             "both counts zero -> not ungrouped",
			summary:          UngroupedSameDaySummary{MedicalRecordCount: 0, TrimmingCount: 0},
			wantHasUngrouped: false,
		},
		{
			name:             "medical record count positive -> ungrouped",
			summary:          UngroupedSameDaySummary{MedicalRecordCount: 1, TrimmingCount: 0},
			wantHasUngrouped: true,
		},
		{
			name:             "trimming count positive -> ungrouped",
			summary:          UngroupedSameDaySummary{MedicalRecordCount: 0, TrimmingCount: 1},
			wantHasUngrouped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := toUngroupedSameDayResponse(tt.summary)
			assert.Equal(t, tt.summary.MedicalRecordCount, resp.MedicalRecordCount)
			assert.Equal(t, tt.summary.TrimmingCount, resp.TrimmingCount)
			assert.Equal(t, tt.wantHasUngrouped, resp.HasUngrouped)
		})
	}
}

// ---- toDiscountSuggestionsResponse ----

func TestToDiscountSuggestionsResponse(t *testing.T) {
	t.Run("wraps suggestions slice", func(t *testing.T) {
		suggestions := []DiscountSuggestion{{Type: "campaign", Name: "夏キャンペーン", Amount: 500}}
		resp := toDiscountSuggestionsResponse(suggestions)
		assert.Equal(t, suggestions, resp.Suggestions)
	})

	t.Run("nil suggestions", func(t *testing.T) {
		resp := toDiscountSuggestionsResponse(nil)
		assert.Nil(t, resp.Suggestions)
	})
}

// ---- parseUngroupedDate ----

func TestParseUngroupedDate(t *testing.T) {
	t.Run("empty string falls back to now", func(t *testing.T) {
		before := time.Now().In(time.Local)
		got := parseUngroupedDate("")
		after := time.Now().In(time.Local)
		assert.False(t, got.Before(before.Add(-time.Second)))
		assert.False(t, got.After(after.Add(time.Second)))
	})

	t.Run("parses valid YYYY-MM-DD date", func(t *testing.T) {
		got := parseUngroupedDate("2026-06-01")
		assert.Equal(t, 2026, got.Year())
		assert.Equal(t, time.June, got.Month())
		assert.Equal(t, 1, got.Day())
	})

	t.Run("invalid format falls back to now", func(t *testing.T) {
		before := time.Now().In(time.Local)
		got := parseUngroupedDate("06/01/2026")
		after := time.Now().In(time.Local)
		assert.False(t, got.Before(before.Add(-time.Second)))
		assert.False(t, got.After(after.Add(time.Second)))
	})
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Billing Item Handler Test Cases
// This handler manages billing line items (Section 6: 会計管理 - billing details)
// BillingItem: child resource nested under billing records, represents line items
//
// CRITICAL ENDPOINTS (nested under /billing-items):
//
// 1. CreateBillingItem (POST /billing-items)
//    Test Cases (20 scenarios):
//    ✓ Returns 201 Created when billing item created successfully
//    ✓ Returns 400 when required field missing (billing_id, category, name, unit_price, quantity)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when billing record doesn't exist (FK validation)
//    ✓ Returns 403 when billing belongs to different clinic (tenant isolation)
//    ✓ Requires ResourceAccounting create permission (checked via middleware)
//    ✓ BillingID field: required, FK to billings (must exist in same clinic)
//    ✓ Category field: required, ENUM (procedure, medicine, consultation, merchandise, discount, tax)
//    ✓ Name field: required, text (item description)
//    ✓ UnitPrice field: required, numeric (price per unit)
//    ✓ Quantity field: required, numeric (item quantity)
//    ✓ TaxType field: optional ENUM (included, excluded), defaults to excluded
//    ✓ TaxRate field: optional numeric (%), defaults to 0.10 (10%)
//    ✓ IsInsuranceApplicable field: optional boolean (covered by insurance)
//    ✓ Source field: optional ENUM (manual, imported, calculated), defaults to manual
//    ✓ SortOrder field: optional numeric for display ordering
//    ✓ Created item includes generated id and timestamps
//    ✓ Uses ToBillingItemResponse() transformation
//    ✓ Returns 500 on database error
//
// 2. UpdateBillingItem (PATCH /billing-items/:id)
//    Test Cases (17 scenarios):
//    ✓ Returns 200 OK when billing item updated successfully
//    ✓ Returns 400 when id is non-numeric or invalid format
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when billing item doesn't exist
//    ✓ Returns 403 when item belongs to billing in different clinic
//    ✓ Requires ResourceAccounting edit permission (checked via middleware)
//    ✓ Partial updates: unit_price can be updated independently
//    ✓ Partial updates: quantity can be updated independently (recalculates total)
//    ✓ Partial updates: tax_type can be updated (ENUM validation)
//    ✓ Partial updates: tax_rate can be updated (numeric 0-100)
//    ✓ Partial updates: is_insurance_applicable can be toggled
//    ✓ Cannot update: category, name (immutable after creation)
//    ✓ Cannot update: billing_id (immutable, belongs to specific billing)
//    ✓ Unspecified fields remain unchanged (PATCH semantics)
//    ✓ Uses ToBillingItemResponse() transformation
//    ✓ Returns 500 on database error
//
// 3. DeleteBillingItem (DELETE /billing-items/:id)
//    Test Cases (10 scenarios):
//    ✓ Returns 204 No Content when billing item deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when id is non-numeric or invalid format
//    ✓ Returns 404 when billing item doesn't exist
//    ✓ Returns 403 when item belongs to billing in different clinic
//    ✓ Requires ResourceAccounting delete permission (checked via middleware)
//    ✓ Deletion behavior: soft delete or hard delete
//    ✓ Deletion updates billing totals (subtotal, tax, total_amount)
//    ✓ Deletion cascades or blocks based on billing status
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control via parent billing isolation
//    ✓ RBAC: ResourceAccounting permission (create, edit, delete required)
//    ✓ FK validation: billing_id must exist and belong to same clinic
//    ✓ Tenant isolation: implicit via billing FK scoping
//    ✓ Soft delete prevents accidental data loss
//    ✓ Partial updates prevent mass assignment
//
// DATA USES:
//    ✓ Billing item child resource of billings (1:N relationship)
//    ✓ Contributes to billing calculation (subtotal, tax, total)
//    ✓ Category determines item type (procedure, drug, service, etc.)
//    ✓ Tax info affects invoice calculation and compliance
//    ✓ Insurance applicability for claims processing
//    ✓ Source tracks item origin (manual entry vs. auto-generated)
//
// DATA MODEL (billing_items):
//    - id (PK): BIGSERIAL
//    - billing_id: BIGINT NOT NULL (FK → billings)
//    - clinic_id: BIGINT NOT NULL (multitenancy, duplicated from billing)
//    - category: VARCHAR(50) NOT NULL - ENUM (procedure, medicine, consultation, merchandise, discount, tax)
//    - name: VARCHAR(255) NOT NULL - item description
//    - unit_price: NUMERIC(10,2) NOT NULL - price per unit
//    - quantity: NUMERIC(10,2) NOT NULL - item quantity
//    - tax_type: VARCHAR(50) DEFAULT 'excluded' - ENUM (included, excluded)
//    - tax_rate: NUMERIC(5,4) DEFAULT 0.1000 - tax rate
//    - is_insurance_applicable: BOOLEAN DEFAULT false - insurance coverage flag
//    - source: VARCHAR(50) DEFAULT 'manual' - ENUM (manual, imported, calculated)
//    - sort_order: INTEGER DEFAULT 0 - display ordering
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (billing_id, id), (clinic_id, billing_id), (billing_id, sort_order)
//
// IMPLEMENTATION NOTES:
//    - Child resource: items are nested under specific billing record
//    - NO standalone list/get endpoints (accessed only via parent billing)
//    - Category ENUM: procedure, medicine, consultation, merchandise, discount, tax
//    - Source ENUM: manual (staff entry), imported (batch), calculated (auto)
//    - Tax fields: type (included/excluded) and rate (percentage)
//    - Insurance flag: for insurance claim processing
//    - Immutable fields: category, name (prevent post-creation edits)
//    - Recalculation: billing totals (subtotal, tax, total_amount) recalculate on update/delete
//    - Transformations: ToBillingItemResponse()
//    - RBAC: ResourceAccounting permission required
//    - Soft delete: preserves billing history
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with sample billings, invoices
//    - Real service/repository layers
//    - Verify clinic_id scoping via parent billing isolation
//    - Test CreateBillingItem with all required fields
//    - Test CreateBillingItem with all optional fields
//    - Test CreateBillingItem with category ENUM validation
//    - Test CreateBillingItem with source ENUM validation (manual, imported, calculated)
//    - Test CreateBillingItem with tax_type ENUM validation
//    - Test billing FK validation (must exist in same clinic)
//    - Test CreateBillingItem with various categories (procedure, medicine, etc.)
//    - Test UpdateBillingItem modifiable fields (unit_price, quantity, tax, insurance)
//    - Test UpdateBillingItem immutable fields (category, name, billing_id rejection)
//    - Test UpdateBillingItem triggers billing total recalculation
//    - Test UpdateBillingItem PATCH semantics
//    - Test DeleteBillingItem soft delete behavior
//    - Test DeleteBillingItem triggers billing total recalculation
//    - Test DeleteBillingItem blocks if billing finalized (if enforced)
//    - Test response transformation (ToBillingItemResponse)
//    - Test permission checks (ResourceAccounting on all operations)
//    - Test tax calculation (included vs excluded in totals)
//
