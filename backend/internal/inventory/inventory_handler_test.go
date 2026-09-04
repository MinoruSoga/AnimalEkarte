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
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock InventoryService ----

type mockInventoryService struct {
	listFn    func(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateInventoryInput) (*model.InventoryItem, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateInventoryInput) (*model.InventoryItem, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockInventoryService) List(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error) {
	return m.listFn(ctx, clinicID, category, status, page, limit)
}

func (m *mockInventoryService) GetByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockInventoryService) Create(ctx context.Context, clinicID uint64, input *CreateInventoryInput) (*model.InventoryItem, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.InventoryItem{}, nil
}

func (m *mockInventoryService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInventoryInput) (*model.InventoryItem, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockInventoryService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

// ---- test helper ----

func newHandlerWithInventorySvc(svc InventoryService) *Handler {
	return NewHandler(svc, nil, allowInventoryTestPermission)
}

// ---- ListInventory ----

func TestListInventory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockInventoryService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated inventory items",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				listFn: func(_ context.Context, _ uint64, category, status *string, _, _ int) ([]model.InventoryItem, int64, error) {
					assert.Nil(t, category)
					assert.Nil(t, status)
					return []model.InventoryItem{{ID: 1, Name: "生理食塩水"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"生理食塩水"`,
		},
		{
			name:     "passes category filter to service",
			query:    "page=1&limit=10&category=medicine",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				listFn: func(_ context.Context, _ uint64, category, _ *string, _, _ int) ([]model.InventoryItem, int64, error) {
					require.NotNil(t, category)
					assert.Equal(t, "medicine", *category)
					return []model.InventoryItem{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "page=1&limit=10",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInventoryService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid pagination",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInventoryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				listFn: func(_ context.Context, _ uint64, _, _ *string, _, _ int) ([]model.InventoryItem, int64, error) {
					return nil, 0, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInventorySvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListInventory(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetInventory ----

func TestGetInventory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockInventoryService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns inventory item for valid id",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					return &model.InventoryItem{ID: 5, Name: "注射器"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"注射器"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInventoryService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInventoryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when item not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.InventoryItem, error) {
					return nil, apperrors.WrapNotFound("inventory", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInventorySvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetInventory(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateInventory ----

func TestCreateInventory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":     "アルコール消毒液",
			"category": "consumable",
			"unit":     "本",
			"quantity": 10,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockInventoryService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates inventory item successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateInventoryInput) (*model.InventoryItem, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "アルコール消毒液", input.Name)
					return &model.InventoryItem{Name: input.Name, Category: model.InventoryCategory(input.Category)}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"アルコール消毒液"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInventoryService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required fields are missing",
			body:       map[string]any{"quantity": 5}, // name, category, unit が missing
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInventoryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			// BUG-466: quantity は非負のみ許可（0 は在庫切れ作成として許容）
			name: "returns 400 when quantity is negative",
			body: map[string]any{
				"name": "負数在庫", "category": "consumable", "unit": "本", "quantity": -1,
			},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInventoryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			// BUG-466: quantity=0 は許容する（min=0）
			name: "creates inventory item with zero quantity",
			body: map[string]any{
				"name": "ゼロ在庫", "category": "consumable", "unit": "本", "quantity": 0,
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				createFn: func(_ context.Context, _ uint64, input *CreateInventoryInput) (*model.InventoryItem, error) {
					assert.Equal(t, 0, input.Quantity)
					return &model.InventoryItem{Name: input.Name, Category: model.InventoryCategory(input.Category), Quantity: 0}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"ゼロ在庫"`,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				createFn: func(_ context.Context, _ uint64, _ *CreateInventoryInput) (*model.InventoryItem, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			// BE-RC-006: toServiceInput は既に AppError を返す。WrapInvalidInput(err.Error())
			// すると Message に ": invalid input" が二重に載る。
			name: "returns 400 for invalid expiry_date without wrapping AppError.Error",
			body: map[string]any{
				"name": "日付不正", "category": "consumable", "unit": "本", "quantity": 1,
				"expiry_date": "not-a-date",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				createFn: func(_ context.Context, _ uint64, _ *CreateInventoryInput) (*model.InventoryItem, error) {
					t.Fatal("Create must not be called when toServiceInput fails")
					return nil, nil
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   httpapi.FlexibleDateInvalidInputMsg,
		},
		{
			name: "returns 400 for invalid last_restocked without wrapping AppError.Error",
			body: map[string]any{
				"name": "入荷日不正", "category": "consumable", "unit": "本", "quantity": 1,
				"last_restocked": "not-a-date",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				createFn: func(_ context.Context, _ uint64, _ *CreateInventoryInput) (*model.InventoryItem, error) {
					t.Fatal("Create must not be called when toServiceInput fails")
					return nil, nil
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   httpapi.FlexibleDateInvalidInputMsg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInventorySvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateInventory(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusBadRequest && tt.wantBody == httpapi.FlexibleDateInvalidInputMsg {
				assert.NotContains(t, w.Body.String(), "invalid input")
				assert.NotContains(t, w.Body.String(), "not-a-date")
			}
		})
	}
}

// ---- UpdateInventory ----

func TestUpdateInventory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockInventoryService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates inventory item successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新済み消毒液", "quantity": 20},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateInventoryInput) (*model.InventoryItem, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新済み消毒液", *input.Name)
					require.NotNil(t, input.Quantity)
					assert.Equal(t, 20, *input.Quantity)
					return &model.InventoryItem{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockInventoryService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInventoryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when item not found",
			paramID:  "999",
			body:     map[string]any{"quantity": 5},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateInventoryInput) (*model.InventoryItem, error) {
					return nil, apperrors.WrapNotFound("inventory", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			// BUG-466: 更新時も quantity は非負のみ許可
			name:       "returns 400 when quantity is negative",
			paramID:    "1",
			body:       map[string]any{"quantity": -1},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockInventoryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			// BE-RC-006: toServiceInput は既に AppError を返す。
			name:     "returns 400 for invalid expiry_date without wrapping AppError.Error",
			paramID:  "1",
			body:     map[string]any{"expiry_date": "not-a-date"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateInventoryInput) (*model.InventoryItem, error) {
					t.Fatal("Update must not be called when toServiceInput fails")
					return nil, nil
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   httpapi.FlexibleDateInvalidInputMsg,
		},
		{
			name:     "returns 400 for invalid last_restocked without wrapping AppError.Error",
			paramID:  "1",
			body:     map[string]any{"last_restocked": "not-a-date"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockInventoryService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateInventoryInput) (*model.InventoryItem, error) {
					t.Fatal("Update must not be called when toServiceInput fails")
					return nil, nil
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   httpapi.FlexibleDateInvalidInputMsg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithInventorySvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateInventory(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusBadRequest && tt.wantBody == httpapi.FlexibleDateInvalidInputMsg {
				assert.NotContains(t, w.Body.String(), "invalid input")
				assert.NotContains(t, w.Body.String(), "not-a-date")
			}
		})
	}
}

// ---- DeleteInventory ----

func newDeleteInventoryRouter(svc InventoryService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithInventorySvc(svc)
	r.DELETE("/inventory/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteInventory)
	return r
}

func TestDeleteInventory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockInventoryService
		wantStatus int
	}{
		{
			name:    "deletes inventory item successfully",
			paramID: "1",
			svc: &mockInventoryService{
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
			svc:        &mockInventoryService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when item not found",
			paramID: "999",
			svc: &mockInventoryService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("inventory", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteInventoryRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/inventory/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithInventorySvc(&mockInventoryService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteInventory(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
