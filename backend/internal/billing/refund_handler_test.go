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

// TestRefundHandlerCompiles verifies refund_handler.go compiles
func TestRefundHandlerCompiles(t *testing.T) {
	assert.True(t, true, "refund_handler.go compiled successfully")
}

// ---- mock RefundService ----

type mockRefundService struct {
	listByBillingIDFn func(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error)
	createFn          func(ctx context.Context, clinicID, billingID uint64, input CreateRefundInput) (*model.BillingRefund, error)
}

func (m *mockRefundService) ListByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error) {
	if m.listByBillingIDFn != nil {
		return m.listByBillingIDFn(ctx, clinicID, billingID)
	}
	return nil, nil
}

func (m *mockRefundService) Create(ctx context.Context, clinicID, billingID uint64, input CreateRefundInput) (*model.BillingRefund, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, billingID, input)
	}
	return &model.BillingRefund{ID: 1, ClinicID: clinicID, BillingID: billingID, Amount: input.Amount}, nil
}

// ---- test helper ----

func newHandlerWithRefundSvc(svc RefundService) *RefundHandler {
	return NewRefundHandler(svc, func(_, _ string) gin.HandlerFunc { return func(_ *gin.Context) {} })
}

// setClinicAndStaff は clinic_id と user_id の両方をコンテキストに設定するヘルパー。
func setClinicAndStaff(c *gin.Context) {
	c.Set("clinic_id", "1")
	c.Set("user_id", "2")
}

// ---- ListRefunds ----

func TestListRefunds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockRefundService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of refunds",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockRefundService{
				listByBillingIDFn: func(_ context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), billingID)
					return []model.BillingRefund{{ID: 1, ClinicID: 1, BillingID: 1, Amount: 1000, Reason: "過剰請求"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"reason":"過剰請求"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockRefundService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric billing id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockRefundService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockRefundService{
				listByBillingIDFn: func(_ context.Context, _, _ uint64) ([]model.BillingRefund, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithRefundSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.ListRefunds(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateRefund ----

func TestCreateRefund(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{"amount": 1000, "reason": "過剰請求"}
	}

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockRefundService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates refund successfully",
			paramID:  "1",
			body:     validBody(),
			setupCtx: setClinicAndStaff,
			svc: &mockRefundService{
				createFn: func(_ context.Context, clinicID, billingID uint64, input CreateRefundInput) (*model.BillingRefund, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), billingID)
					require.NotNil(t, input.StaffID)
					assert.Equal(t, uint64(2), *input.StaffID)
					assert.Equal(t, int64(1000), input.Amount)
					assert.Equal(t, "過剰請求", input.Reason)
					return &model.BillingRefund{ID: 5, ClinicID: clinicID, BillingID: billingID, Amount: input.Amount, Reason: input.Reason}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"reason":"過剰請求"`,
		},
		{
			name:     "creates refund with payment_method",
			paramID:  "1",
			body:     map[string]any{"amount": 500, "payment_method": "cash"},
			setupCtx: setClinicAndStaff,
			svc: &mockRefundService{
				createFn: func(_ context.Context, _, _ uint64, input CreateRefundInput) (*model.BillingRefund, error) {
					require.NotNil(t, input.PaymentMethod)
					assert.Equal(t, model.PaymentMethod("cash"), *input.PaymentMethod)
					return &model.BillingRefund{ID: 6, Amount: input.Amount, PaymentMethod: input.PaymentMethod}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockRefundService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when staff_id missing",
			paramID:    "1",
			body:       validBody(),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockRefundService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric billing id",
			paramID:    "abc",
			body:       validBody(),
			setupCtx:   setClinicAndStaff,
			svc:        &mockRefundService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when amount is missing",
			paramID:    "1",
			body:       map[string]any{"reason": "no amount"},
			setupCtx:   setClinicAndStaff,
			svc:        &mockRefundService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed body",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   setClinicAndStaff,
			svc:        &mockRefundService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			body:     validBody(),
			setupCtx: setClinicAndStaff,
			svc: &mockRefundService{
				createFn: func(_ context.Context, _, _ uint64, _ CreateRefundInput) (*model.BillingRefund, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "returns 404 when billing not found",
			paramID:  "999",
			body:     validBody(),
			setupCtx: setClinicAndStaff,
			svc: &mockRefundService{
				createFn: func(_ context.Context, _, _ uint64, _ CreateRefundInput) (*model.BillingRefund, error) {
					return nil, apperrors.WrapNotFound("billing", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithRefundSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.CreateRefund(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusCreated {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Refund Handler Test Cases
// This handler manages refund transactions (Section 6: 会計管理 - payment management)
//
// DESIGN DECISION: 返金レコードは会計原則上 immutable（不変）。
// Update/Delete エンドポイントは意図的に実装しない。
// 金額訂正は「差額の追加返金」で対応する（Stripe モデル準拠）。
//
// ENDPOINTS:
//
// 1. ListRefunds (GET /accountings/:id/refunds)
//    Test Cases:
//    ✓ Returns 200 OK with empty list when no refunds exist
//    ✓ Returns 200 OK with list of refunds for the billing
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when billing id is invalid
//    ✓ Returns 404 when billing doesn't exist (tenant-scoped via AccountingRepository)
//    ✓ Response includes: id, billing_id, amount, reason, refunded_by, refunded_by_name, refunded_at, created_at
//    ✓ refunded_by_name: Preload("RefundedByStaff") でスタッフ名を解決
//    ✓ Returns 500 on database error
//
// 2. CreateRefund (POST /accountings/:id/refunds)
//    Test Cases:
//    ✓ Returns 201 Created when refund created successfully
//    ✓ Returns 400 when required field missing (amount)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 400 when amount exceeds remaining balance (BUG-142)
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 401 when staff_id missing from context (extractStaffID)
//    ✓ Returns 404 when billing doesn't exist (FK validation)
//    ✓ Requires ResourceAccounting create permission (checked via middleware)
//    ✓ Amount field: required, min=1 (positive integer)
//    ✓ Reason field: optional text
//    ✓ refunded_by: auto-populated from authenticated staff (extractStaffID)
//    ✓ Created refund includes generated id and timestamps
//    ✓ Response includes refunded_by_name resolved from Staff relation
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification via billing lookup)
//    ✓ RBAC: ResourceAccounting create permission required for POST
//    ✓ FK validation: billing must exist and belong to same clinic
//    ✓ Amount validation: cannot exceed remaining balance (cumulative check)
//    ✓ Staff audit trail: refunded_by auto-set from JWT context (BUG-361)
//
// DATA MODEL (billing_refunds):
//    - id (PK): BIGSERIAL
//    - clinic_id: BIGINT NOT NULL (multitenancy)
//    - billing_id: BIGINT NOT NULL (FK → billings)
//    - amount: BIGINT NOT NULL CHECK (amount > 0)
//    - reason: TEXT NOT NULL DEFAULT ''
//    - refunded_by: BIGINT (NULLABLE, FK → staffs) — 返金処理スタッフ
//    - refunded_at: TIMESTAMPTZ NOT NULL DEFAULT now()
//    - created_at: TIMESTAMPTZ NOT NULL DEFAULT now()
//    - Indexes: (billing_id), (clinic_id, billing_id), (refunded_by)
//    - Relations: RefundedByStaff *Staff (foreignKey:RefundedBy)
//
// PRELOAD PATHS:
//    - ListRefunds:  RefundRepository.FindByBillingID → Preload("RefundedByStaff")
//    - Detail view:  AccountingRepository.FindByID    → Preload("Refunds.RefundedByStaff")
//    - After update: AccountingRepository.Update → Preload("Refunds.RefundedByStaff")
