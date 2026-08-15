package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock MerchandiseItemService ----

type mockMerchandiseItemService struct {
	listFn    func(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateMerchandiseItemInput) (*model.MerchandiseItem, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockMerchandiseItemService) List(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error) {
	return m.listFn(ctx, clinicID, category)
}

func (m *mockMerchandiseItemService) GetByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockMerchandiseItemService) Create(ctx context.Context, clinicID uint64, input *CreateMerchandiseItemInput) (*model.MerchandiseItem, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.MerchandiseItem{}, nil
}

func (m *mockMerchandiseItemService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockMerchandiseItemService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMerchandiseItemService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- test helper ----

func newHandlerWithMerchandiseItemSvc(svc MerchandiseItemService) *Handler {
	return NewHandler(nil, svc, allowInventoryTestPermission)
}

// ---- ListMerchandiseItems ----

func TestListMerchandiseItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockMerchandiseItemService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns all merchandise items",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				listFn: func(_ context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "", category)
					return []model.MerchandiseItem{{ID: 1, Name: "ドッグフード"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"ドッグフード"`,
		},
		{
			name:     "passes category filter to service",
			query:    "category=food",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				listFn: func(_ context.Context, _ uint64, category string) ([]model.MerchandiseItem, error) {
					assert.Equal(t, "food", category)
					return []model.MerchandiseItem{}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "returns empty list when no items",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				listFn: func(_ context.Context, _ uint64, _ string) ([]model.MerchandiseItem, error) {
					return []model.MerchandiseItem{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   "[]",
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				listFn: func(_ context.Context, _ uint64, _ string) ([]model.MerchandiseItem, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMerchandiseItemSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListMerchandiseItems(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetMerchandiseItem ----

func TestGetMerchandiseItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockMerchandiseItemService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns merchandise item for valid id",
			paramID:  "3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), id)
					return &model.MerchandiseItem{ID: 3, Name: "キャットフード"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"キャットフード"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when item not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MerchandiseItem, error) {
					return nil, apperrors.WrapNotFound("merchandise_item", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMerchandiseItemSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetMerchandiseItem(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateMerchandiseItem ----

func TestCreateMerchandiseItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":       "プレミアムフード",
			"category":   "food",
			"unit_price": 1200,
			"tax_type":   "excluded",
			"tax_rate":   0.10,
			"is_active":  true,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockMerchandiseItemService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates merchandise item successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateMerchandiseItemInput) (*model.MerchandiseItem, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "プレミアムフード", input.Name)
					assert.Equal(t, "food", input.Category)
					return &model.MerchandiseItem{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"プレミアムフード"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required name is missing",
			body:       map[string]any{"category": "food", "unit_price": 100, "tax_type": "excluded"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when required category is missing",
			body:       map[string]any{"name": "テスト商品", "unit_price": 100, "tax_type": "excluded"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when required tax_type is missing",
			body:       map[string]any{"name": "テスト商品", "category": "food"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when category value is invalid",
			body:       map[string]any{"name": "テスト", "category": "invalid", "unit_price": 100, "tax_type": "excluded"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				createFn: func(_ context.Context, _ uint64, _ *CreateMerchandiseItemInput) (*model.MerchandiseItem, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMerchandiseItemSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateMerchandiseItem(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateMerchandiseItem ----

func TestUpdateMerchandiseItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockMerchandiseItemService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates merchandise item successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新済み商品"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新済み商品", *input.Name)
					return &model.MerchandiseItem{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"更新済み商品"`,
		},
		{
			name:     "updates unit_price (non-negative validation from BUG-380)",
			paramID:  "1",
			body:     map[string]any{"unit_price": 500},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
					require.NotNil(t, input.UnitPrice)
					assert.Equal(t, int64(500), *input.UnitPrice)
					return &model.MerchandiseItem{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when item not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
					return nil, apperrors.WrapNotFound("merchandise_item", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 for empty update body (BUG-397)",
			paramID:  "1",
			body:     map[string]any{},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMerchandiseItemService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
					return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMerchandiseItemSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateMerchandiseItem(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteMerchandiseItem ----

func newDeleteMerchandiseItemRouter(svc MerchandiseItemService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithMerchandiseItemSvc(svc)
	r.DELETE("/merchandise-items/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteMerchandiseItem)
	return r
}

func TestDeleteMerchandiseItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockMerchandiseItemService
		wantStatus int
	}{
		{
			name:    "deletes merchandise item successfully",
			paramID: "1",
			svc: &mockMerchandiseItemService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when item not found",
			paramID: "999",
			svc: &mockMerchandiseItemService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("merchandise_item", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when item is in use",
			paramID: "5",
			svc: &mockMerchandiseItemService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この物販品目は請求・見積データで使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteMerchandiseItemRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/merchandise-items/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithMerchandiseItemSvc(&mockMerchandiseItemService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteMerchandiseItem(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderMerchandiseItems ----

func newReorderMerchandiseItemsRouter(svc MerchandiseItemService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithMerchandiseItemSvc(svc)
	r.PUT("/merchandise-items/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderMerchandiseItems)
	return r
}

func TestReorderMerchandiseItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		svc        *mockMerchandiseItemService
		wantStatus int
	}{
		{
			name: "reorders merchandise items successfully",
			body: map[string]any{"ids": []int{2, 3, 1}},
			svc: &mockMerchandiseItemService{
				reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, []uint64{2, 3, 1}, ids)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for missing ids field",
			body:       map[string]any{},
			svc:        &mockMerchandiseItemService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newReorderMerchandiseItemsRouter(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/merchandise-items/reorder", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithMerchandiseItemSvc(&mockMerchandiseItemService{})

		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderMerchandiseItems(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
