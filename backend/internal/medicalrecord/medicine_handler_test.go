package medicalrecord

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

// ---- mock MedicineService ----

type mockMedicineService struct {
	listFn    func(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateMedicineInput) (*model.Medicine, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockMedicineService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	return m.listFn(ctx, clinicID, page, limit)
}

func (m *mockMedicineService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockMedicineService) Create(ctx context.Context, clinicID uint64, input *CreateMedicineInput) (*model.Medicine, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.Medicine{}, nil
}

func (m *mockMedicineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockMedicineService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMedicineService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- test helper ----

func newHandlerWithMedicineSvc(svc MedicineService) *MedicineHandler {
	return NewMedicineHandler(svc)
}

// ---- ListMedicines ----

func TestListMedicines(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// TASK-232: ListMedicines は全件返却（page=1, limit=10000固定）。ページネーション引数は無視。
	// レスポンスは JSON 配列（PaginatedResponse ラッパーなし）。
	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockMedicineService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns all medicines as array",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				listFn: func(_ context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, 1, page)
					assert.Equal(t, medicineListMaxLimit, limit)
					return []model.Medicine{{ID: 1, Name: "アモキシシリン"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"アモキシシリン"`,
		},
		{
			name:     "returns empty array when no medicines",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				listFn: func(_ context.Context, _ uint64, _, _ int) ([]model.Medicine, int64, error) {
					return []model.Medicine{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMedicineService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				listFn: func(_ context.Context, _ uint64, _, _ int) ([]model.Medicine, int64, error) {
					return nil, 0, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicineSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListMedicines(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetMedicine ----

func TestGetMedicine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockMedicineService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns medicine for valid id",
			paramID:  "7",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Medicine, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(7), id)
					return &model.Medicine{ID: 7, Name: "セファレキシン"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"セファレキシン"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMedicineService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicineService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when medicine not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
					return nil, apperrors.WrapNotFound("medicine", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicineSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetMedicine(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateMedicine ----

func TestCreateMedicine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":      "テスト薬",
			"is_active": true,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockMedicineService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates medicine successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateMedicineInput) (*model.Medicine, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "テスト薬", input.Name)
					return &model.Medicine{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"テスト薬"`,
		},
		{
			name:     "creates medicine with optional inventory_id",
			body:     map[string]any{"name": "在庫連携薬", "inventory_id": 42},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				createFn: func(_ context.Context, _ uint64, input *CreateMedicineInput) (*model.Medicine, error) {
					require.NotNil(t, input.InventoryID)
					assert.Equal(t, uint64(42), *input.InventoryID)
					return &model.Medicine{ID: 2, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMedicineService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required name field is missing",
			body:       map[string]any{"is_active": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicineService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				createFn: func(_ context.Context, _ uint64, _ *CreateMedicineInput) (*model.Medicine, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicineSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateMedicine(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateMedicine ----

func TestUpdateMedicine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockMedicineService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates medicine name successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新済み薬"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateMedicineInput) (*model.Medicine, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新済み薬", *input.Name)
					return &model.Medicine{ID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"更新済み薬"`,
		},
		{
			name:     "updates medicine with inventory_id link",
			paramID:  "1",
			body:     map[string]any{"inventory_id": 10},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateMedicineInput) (*model.Medicine, error) {
					require.NotNil(t, input.InventoryID)
					assert.Equal(t, uint64(10), *input.InventoryID)
					return &model.Medicine{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMedicineService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicineService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when medicine not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateMedicineInput) (*model.Medicine, error) {
					return nil, apperrors.WrapNotFound("medicine", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 for empty update body (BUG-397)",
			paramID:  "1",
			body:     map[string]any{},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicineService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateMedicineInput) (*model.Medicine, error) {
					return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicineSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateMedicine(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteMedicine ----

func newDeleteMedicineRouter(svc MedicineService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithMedicineSvc(svc)
	r.DELETE("/medicines/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteMedicine)
	return r
}

func TestDeleteMedicine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockMedicineService
		wantStatus int
	}{
		{
			name:    "deletes medicine successfully",
			paramID: "1",
			svc: &mockMedicineService{
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
			svc:        &mockMedicineService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when medicine not found",
			paramID: "999",
			svc: &mockMedicineService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("medicine", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when medicine category has children",
			paramID: "5",
			svc: &mockMedicineService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("このカテゴリには3件の薬剤が含まれています。先に薬剤を移動または削除してください")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteMedicineRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/medicines/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithMedicineSvc(&mockMedicineService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteMedicine(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderMedicines ----

func newReorderMedicinesRouter(svc MedicineService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithMedicineSvc(svc)
	r.PUT("/medicines/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderMedicines)
	return r
}

func TestReorderMedicines(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		svc        *mockMedicineService
		wantStatus int
	}{
		{
			name: "reorders medicines successfully",
			body: map[string]any{"ids": []int{3, 1, 2}},
			svc: &mockMedicineService{
				reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, []uint64{3, 1, 2}, ids)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for missing ids field",
			body:       map[string]any{},
			svc:        &mockMedicineService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newReorderMedicinesRouter(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/medicines/reorder", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithMedicineSvc(&mockMedicineService{})

		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderMedicines(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
