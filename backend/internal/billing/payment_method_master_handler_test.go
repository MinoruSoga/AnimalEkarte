package billing

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

// ---- mock PaymentMethodMasterService ----

type mockPaymentMethodMasterService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockPaymentMethodMasterService) List(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
	return m.listFn(ctx, clinicID)
}

func (m *mockPaymentMethodMasterService) GetByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockPaymentMethodMasterService) Create(ctx context.Context, clinicID uint64, input *CreatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockPaymentMethodMasterService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockPaymentMethodMasterService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockPaymentMethodMasterService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- test helper ----

func newHandlerWithPaymentMethodSvc(svc PaymentMethodMasterService) *PaymentMethodMasterHandler {
	return NewPaymentMethodMasterHandler(svc)
}

// ---- TestPaymentMethodMasterHandler_List ----

func TestPaymentMethodMasterHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockPaymentMethodMasterService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of payment methods",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPaymentMethodMasterService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.PaymentMethodMaster{
						{ID: 1, ClinicID: 1, Name: "現金"},
						{ID: 2, ClinicID: 1, Name: "クレジット"},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"現金"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPaymentMethodMasterService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPaymentMethodMasterService{
				listFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPaymentMethodSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListPaymentMethods(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- TestPaymentMethodMasterHandler_GetByID ----

func TestPaymentMethodMasterHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockPaymentMethodMasterService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns payment method for valid id",
			paramID:  "3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPaymentMethodMasterService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), id)
					return &model.PaymentMethodMaster{ID: 3, ClinicID: 1, Name: "電子マネー"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"電子マネー"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPaymentMethodMasterService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPaymentMethodMasterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when payment method not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPaymentMethodMasterService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.PaymentMethodMaster, error) {
					return nil, apperrors.WrapNotFound("payment_method", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPaymentMethodSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetPaymentMethod(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- TestPaymentMethodMasterHandler_Create ----

func TestPaymentMethodMasterHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":          "現金",
			"display_order": 1,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockPaymentMethodMasterService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates payment method successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPaymentMethodMasterService{
				createFn: func(_ context.Context, clinicID uint64, input *CreatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "現金", input.Name)
					assert.Equal(t, 1, input.DisplayOrder)
					return &model.PaymentMethodMaster{ID: 10, ClinicID: 1, Name: input.Name, DisplayOrder: input.DisplayOrder}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"現金"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPaymentMethodMasterService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"display_order": 1},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPaymentMethodMasterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPaymentMethodMasterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPaymentMethodMasterService{
				createFn: func(_ context.Context, _ uint64, _ *CreatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPaymentMethodSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreatePaymentMethod(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- TestPaymentMethodMasterHandler_Update ----

func TestPaymentMethodMasterHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockPaymentMethodMasterService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates payment method successfully",
			paramID:  "1",
			body:     map[string]any{"name": "クレジットカード"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPaymentMethodMasterService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *UpdatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.Name)
					assert.Equal(t, "クレジットカード", *input.Name)
					return &model.PaymentMethodMaster{ID: 1, ClinicID: 1, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"クレジットカード"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPaymentMethodMasterService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPaymentMethodMasterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPaymentMethodMasterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when payment method not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPaymentMethodMasterService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error) {
					return nil, apperrors.WrapNotFound("payment_method", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPaymentMethodSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdatePaymentMethod(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- TestPaymentMethodMasterHandler_Delete ----

func TestPaymentMethodMasterHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		paramID     string
		setupRouter func(svc PaymentMethodMasterService) *gin.Engine
		svc         *mockPaymentMethodMasterService
		wantStatus  int
	}{
		{
			name:    "deletes payment method successfully",
			paramID: "2",
			svc: &mockPaymentMethodMasterService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(2), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockPaymentMethodMasterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when payment method not found",
			paramID: "999",
			svc: &mockPaymentMethodMasterService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("payment_method", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 when payment method is in use",
			paramID: "1",
			svc: &mockPaymentMethodMasterService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この支払方法は使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			h := newHandlerWithPaymentMethodSvc(tt.svc)
			r.DELETE("/payment-methods/:id", func(c *gin.Context) {
				setClinicID(c)
				h.DeletePaymentMethod(c)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/payment-methods/"+tt.paramID, http.NoBody)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- TestPaymentMethodMasterHandler_Reorder ----

func newReorderPaymentMethodRouter(svc PaymentMethodMasterService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithPaymentMethodSvc(svc)
	r.PATCH("/payment-methods/reorder", func(c *gin.Context) {
		setClinicID(c)
		h.ReorderPaymentMethods(c)
	})
	return r
}

func TestPaymentMethodMasterHandler_Reorder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		svc        *mockPaymentMethodMasterService
		wantStatus int
	}{
		{
			name: "reorders payment methods successfully",
			body: map[string]any{"ids": []int{3, 1, 2}},
			svc: &mockPaymentMethodMasterService{
				reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, []uint64{3, 1, 2}, ids)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			svc:        &mockPaymentMethodMasterService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 on service error",
			body: map[string]any{"ids": []int{1, 2}},
			svc: &mockPaymentMethodMasterService{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newReorderPaymentMethodRouter(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/payment-methods/reorder", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
