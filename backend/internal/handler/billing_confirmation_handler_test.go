package handler

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

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock BillingConfirmationService ----

type mockBillingConfirmationService struct {
	getOrCreateFn func(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error)
	confirmFn     func(ctx context.Context, clinicID, medicalRecordID uint64, input *service.ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error)
	returnFn      func(ctx context.Context, clinicID, medicalRecordID uint64, input *service.ReturnBillingConfirmationInput) (*model.BillingConfirmation, error)
}

func (m *mockBillingConfirmationService) GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error) {
	return m.getOrCreateFn(ctx, clinicID, medicalRecordID)
}

func (m *mockBillingConfirmationService) Confirm(ctx context.Context, clinicID, medicalRecordID uint64, input *service.ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
	return m.confirmFn(ctx, clinicID, medicalRecordID, input)
}

func (m *mockBillingConfirmationService) Return(ctx context.Context, clinicID, medicalRecordID uint64, input *service.ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
	return m.returnFn(ctx, clinicID, medicalRecordID, input)
}

// ---- test helper ----

func newHandlerWithBillingConfirmationSvc(svc service.BillingConfirmationService) *Handler {
	return &Handler{svc: &service.Services{BillingConfirmation: svc}}
}

// ---- GetBillingConfirmation ----

func TestGetBillingConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockBillingConfirmationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 with billing confirmation",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingConfirmationService{
				getOrCreateFn: func(_ context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), medicalRecordID)
					return &model.BillingConfirmation{ID: 1, MedicalRecordID: 5, Status: model.ConfirmationStatusPending}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"status":"pending"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "5",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingConfirmationService{
				getOrCreateFn: func(_ context.Context, _, _ uint64) (*model.BillingConfirmation, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithBillingConfirmationSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/"+tt.paramID+"/billing-confirmation", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetBillingConfirmation(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ConfirmBillingConfirmation ----

func TestConfirmBillingConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := confirmBillingConfirmationRequest{ConfirmedBy: 3, Memo: "確認済み"}

	tests := []struct {
		name       string
		paramID    string
		body       any
		malformed  bool
		setupCtx   func(c *gin.Context)
		svc        *mockBillingConfirmationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 when confirmed successfully",
			paramID:  "5",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, clinicID, medicalRecordID uint64, input *service.ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), medicalRecordID)
					assert.Equal(t, uint64(3), input.ConfirmedBy)
					return &model.BillingConfirmation{ID: 1, MedicalRecordID: 5, Status: model.ConfirmationStatusConfirmed}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"status":"confirmed"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "5",
			body:       validBody,
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			body:       validBody,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when body is malformed",
			paramID:    "5",
			malformed:  true,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when confirmed_by is missing",
			paramID:    "5",
			body:       confirmBillingConfirmationRequest{Memo: "x"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, _ *service.ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					return nil, apperrors.WrapConflict("already confirmed")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithBillingConfirmationSvc(tt.svc)

			bodyBytes := []byte("{invalid")
			if !tt.malformed {
				b, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyBytes = b
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/medical-records/"+tt.paramID+"/billing-confirmation/confirm", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.ConfirmBillingConfirmation(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ReturnBillingConfirmation ----

func TestReturnBillingConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := returnBillingConfirmationRequest{ReturnedBy: 3, ReturnReason: "金額確認必要", Memo: ""}

	tests := []struct {
		name       string
		paramID    string
		body       any
		malformed  bool
		setupCtx   func(c *gin.Context)
		svc        *mockBillingConfirmationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 when returned successfully",
			paramID:  "5",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, clinicID, medicalRecordID uint64, input *service.ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), medicalRecordID)
					assert.Equal(t, uint64(3), input.ReturnedBy)
					assert.Equal(t, "金額確認必要", input.ReturnReason)
					return &model.BillingConfirmation{ID: 1, MedicalRecordID: 5, Status: model.ConfirmationStatusReturned}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"status":"returned"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "5",
			body:       validBody,
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			body:       validBody,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when body is malformed",
			paramID:    "5",
			malformed:  true,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when returned_by is missing",
			paramID:    "5",
			body:       returnBillingConfirmationRequest{ReturnReason: "reason"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, _, _ uint64, _ *service.ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithBillingConfirmationSvc(tt.svc)

			bodyBytes := []byte("{invalid")
			if !tt.malformed {
				b, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyBytes = b
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/medical-records/"+tt.paramID+"/billing-confirmation/return", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.ReturnBillingConfirmation(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
