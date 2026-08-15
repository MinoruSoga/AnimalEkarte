package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock BillingConfirmationService ----

type mockBillingConfirmationService struct {
	getOrCreateFn func(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error)
	confirmFn     func(ctx context.Context, clinicID, medicalRecordID uint64, input *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error)
	returnFn      func(ctx context.Context, clinicID, medicalRecordID uint64, input *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error)
}

func (m *mockBillingConfirmationService) GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error) {
	return m.getOrCreateFn(ctx, clinicID, medicalRecordID)
}

func (m *mockBillingConfirmationService) Confirm(ctx context.Context, clinicID, medicalRecordID uint64, input *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
	return m.confirmFn(ctx, clinicID, medicalRecordID, input)
}

func (m *mockBillingConfirmationService) Return(ctx context.Context, clinicID, medicalRecordID uint64, input *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
	return m.returnFn(ctx, clinicID, medicalRecordID, input)
}

// ---- test helper ----

func newHandlerWithBillingConfirmationSvc(svc BillingConfirmationService) *BillingConfirmationHandler {
	return NewBillingConfirmationHandler(svc, func(_, _ string) gin.HandlerFunc { return func(_ *gin.Context) {} })
}

func setBillingConfirmationIdentity(c *gin.Context) {
	setClinicID(c)
	setStaffID(c)
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

	validBody := map[string]any{"memo": "確認済み"}

	tests := []struct {
		name                 string
		paramID              string
		body                 any
		rawBody              string
		malformed            bool
		unknownContentLength bool
		contentType          string
		omitContentType      bool
		setupCtx             func(c *gin.Context)
		svc                  *mockBillingConfirmationService
		wantStatus           int
		wantBody             string
	}{
		{
			name:     "returns 200 when confirmed successfully",
			paramID:  "5",
			body:     map[string]any{"memo": "  確認済み \n"},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, clinicID, medicalRecordID uint64, input *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), medicalRecordID)
					assert.Equal(t, uint64(1), input.ConfirmedBy, "body confirmed_by must not override the authenticated staff")
					assert.Equal(t, "確認済み", input.Memo)
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
			setupCtx:   func(c *gin.Context) { setStaffID(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 401 when authenticated staff is missing even if body supplies confirmed_by",
			paramID:  "5",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, _ *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called without an authenticated staff")
					return nil, apperrors.WrapUnauthorized("missing staff")
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			body:       validBody,
			setupCtx:   func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when body is malformed",
			paramID:    "5",
			malformed:  true,
			setupCtx:   func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 400 when body contains an unknown field",
			paramID:  "5",
			body:     map[string]any{"memo": "x", "confirmed_by": uint64(999)},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, _ *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for an unknown request field")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"invalid billing confirmation request body"`,
		},
		{
			name:     "returns 400 when a field only matches case-insensitively",
			paramID:  "5",
			body:     map[string]any{"Memo": "x"},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, _ *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for a non-canonical field name")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"invalid billing confirmation request body"`,
		},
		{
			name:     "returns 400 when memo is null",
			paramID:  "5",
			body:     map[string]any{"memo": nil},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, _ *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for a null memo")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"invalid billing confirmation request body"`,
		},
		{
			name:     "returns 400 when body contains trailing JSON",
			paramID:  "5",
			rawBody:  `{"memo":"x"}{"memo":"y"}`,
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, _ *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for trailing JSON")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"invalid billing confirmation request body"`,
		},
		{
			name:     "returns 400 when body is null",
			paramID:  "5",
			rawBody:  `null`,
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, _ *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for a null request body")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"invalid billing confirmation request body"`,
		},
		{
			name:        "returns 415 when content type is not JSON",
			paramID:     "5",
			body:        map[string]any{"memo": "x"},
			contentType: "text/plain",
			setupCtx:    func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, _ *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for an unsupported content type")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusUnsupportedMediaType,
			wantBody:   `"error":"Content-Type must be application/json"`,
		},
		{
			name:        "accepts application JSON with charset",
			paramID:     "5",
			body:        map[string]any{"memo": "x"},
			contentType: "application/json; charset=utf-8",
			setupCtx:    func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, input *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					assert.Equal(t, "x", input.Memo)
					return &model.BillingConfirmation{Status: model.ConfirmationStatusConfirmed}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:                 "returns 413 for an oversized chunked body",
			paramID:              "5",
			body:                 map[string]any{"memo": strings.Repeat("x", int(billingConfirmationJSONBodyMaxBytes))},
			unknownContentLength: true,
			setupCtx:             func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, _ *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for an oversized request body")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody:   `"error":"request body exceeds size limit"`,
		},
		{
			name:     "returns 400 when memo exceeds 1000 characters",
			paramID:  "5",
			body:     map[string]any{"memo": strings.Repeat("m", billingConfirmationMemoMaxLength+1)},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, _ *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for an oversized memo")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "accepts memo-only body because actor comes from authenticated context",
			paramID:  "5",
			body:     map[string]any{"memo": "x"},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, input *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
					assert.Equal(t, uint64(1), input.ConfirmedBy)
					return &model.BillingConfirmation{ID: 1, MedicalRecordID: 5, Status: model.ConfirmationStatusConfirmed}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				confirmFn: func(_ context.Context, _, _ uint64, _ *ConfirmBillingConfirmationInput) (*model.BillingConfirmation, error) {
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
			if tt.rawBody != "" {
				bodyBytes = []byte(tt.rawBody)
			} else if !tt.malformed {
				b, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyBytes = b
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/medical-records/"+tt.paramID+"/billing-confirmation/confirm", bytes.NewReader(bodyBytes))
			if !tt.omitContentType {
				contentType := tt.contentType
				if contentType == "" {
					contentType = "application/json"
				}
				c.Request.Header.Set("Content-Type", contentType)
			}
			if tt.unknownContentLength {
				c.Request.ContentLength = -1
			}
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

	validBody := map[string]any{"return_reason": "金額確認必要", "memo": ""}

	tests := []struct {
		name            string
		paramID         string
		body            any
		malformed       bool
		contentType     string
		omitContentType bool
		setupCtx        func(c *gin.Context)
		svc             *mockBillingConfirmationService
		wantStatus      int
		wantBody        string
	}{
		{
			name:     "returns 200 when returned successfully",
			paramID:  "5",
			body:     map[string]any{"return_reason": "  金額確認必要 \n", "memo": "  補足  "},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, clinicID, medicalRecordID uint64, input *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), medicalRecordID)
					assert.Equal(t, uint64(1), input.ReturnedBy, "body returned_by must not override the authenticated staff")
					assert.Equal(t, "金額確認必要", input.ReturnReason)
					assert.Equal(t, "補足", input.Memo)
					return &model.BillingConfirmation{ID: 1, MedicalRecordID: 5, Status: model.ConfirmationStatusReturned}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"status":"returned"`,
		},
		{
			name:    "accepts four-byte UTF-8 fields at both character boundaries",
			paramID: "5",
			body: map[string]any{
				"return_reason": strings.Repeat("🩺", billingConfirmationReturnReasonMaxLength),
				"memo":          strings.Repeat("🩺", billingConfirmationMemoMaxLength),
			},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, _, _ uint64, input *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					assert.Equal(
						t,
						strings.Repeat("🩺", billingConfirmationReturnReasonMaxLength),
						input.ReturnReason,
					)
					assert.Equal(
						t,
						strings.Repeat("🩺", billingConfirmationMemoMaxLength),
						input.Memo,
					)
					return &model.BillingConfirmation{Status: model.ConfirmationStatusReturned}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "5",
			body:       validBody,
			setupCtx:   func(c *gin.Context) { setStaffID(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 401 when authenticated staff is missing even if body supplies returned_by",
			paramID:  "5",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, _, _ uint64, _ *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called without an authenticated staff")
					return nil, apperrors.WrapUnauthorized("missing staff")
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			body:       validBody,
			setupCtx:   func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when body is malformed",
			paramID:    "5",
			malformed:  true,
			setupCtx:   func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 400 when body contains an unknown field",
			paramID:  "5",
			body:     map[string]any{"return_reason": "reason", "returned_by": uint64(999)},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, _, _ uint64, _ *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for an unknown request field")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"invalid billing confirmation request body"`,
		},
		{
			name:     "returns 400 when return_reason is null",
			paramID:  "5",
			body:     map[string]any{"return_reason": nil},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, _, _ uint64, _ *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for a null return_reason")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"invalid billing confirmation request body"`,
		},
		{
			name:            "returns 415 when content type is missing",
			paramID:         "5",
			body:            map[string]any{"return_reason": "reason"},
			omitContentType: true,
			setupCtx:        func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, _, _ uint64, _ *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called without a JSON content type")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusUnsupportedMediaType,
			wantBody:   `"error":"Content-Type must be application/json"`,
		},
		{
			name:    "returns 413 for an oversized body with declared length",
			paramID: "5",
			body: map[string]any{
				"return_reason": "reason",
				"memo":          strings.Repeat("x", int(billingConfirmationJSONBodyMaxBytes)),
			},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, _, _ uint64, _ *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for an oversized request body")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody:   `"error":"request body exceeds size limit"`,
		},
		{
			name:       "returns 400 when return_reason is missing",
			paramID:    "5",
			body:       map[string]any{},
			setupCtx:   func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc:        &mockBillingConfirmationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 400 when return_reason is whitespace only",
			paramID:  "5",
			body:     map[string]any{"return_reason": " \t\n "},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, _, _ uint64, _ *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for an empty normalized return_reason")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 400 when return_reason exceeds 500 characters",
			paramID:  "5",
			body:     map[string]any{"return_reason": strings.Repeat("r", billingConfirmationReturnReasonMaxLength+1)},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, _, _ uint64, _ *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for an oversized return_reason")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 400 when memo exceeds 1000 characters",
			paramID: "5",
			body: map[string]any{
				"return_reason": "reason",
				"memo":          strings.Repeat("m", billingConfirmationMemoMaxLength+1),
			},
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, _, _ uint64, _ *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
					t.Error("service must not be called for an oversized memo")
					return &model.BillingConfirmation{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setBillingConfirmationIdentity(c) },
			svc: &mockBillingConfirmationService{
				returnFn: func(_ context.Context, _, _ uint64, _ *ReturnBillingConfirmationInput) (*model.BillingConfirmation, error) {
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
			if !tt.omitContentType {
				contentType := tt.contentType
				if contentType == "" {
					contentType = "application/json"
				}
				c.Request.Header.Set("Content-Type", contentType)
			}
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
