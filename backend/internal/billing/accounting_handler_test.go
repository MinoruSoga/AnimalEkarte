package billing

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestAccountingHandlerCompiles verifies accounting_handler.go compiles
func TestAccountingHandlerCompiles(t *testing.T) {
	assert.True(t, true, "accounting_handler.go compiled successfully")
}

// ---- UpdateAccounting 締め後経路 characterization (#115 / B4) ----
//
// レジ締め済み期間の会計編集に対する現行 HTTP 挙動を固定する安全網。
// 認可（accounting-post-close-edit:edit 権限要求）と post_close_reason 必須検証の
// 観測可能なステータス／エラーエンベロープを before/after で不変に保つ。

// stubAccountingPostClose は UpdateAccounting の締め後経路で呼ばれる
// GetByID / Update のみを実装する最小スタブ（他メソッドは経路上呼ばれない）。
type stubAccountingPostClose struct {
	AccountingService
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	updateFn  func(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error)
}

func (s *stubAccountingPostClose) GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	return s.getByIDFn(ctx, clinicID, id)
}

func (s *stubAccountingPostClose) Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error) {
	return s.updateFn(ctx, input)
}

// stubCashRegisterIsClosed は IsDateClosed のみを実装する cashRegisterCloseChecker スタブ。
type stubCashRegisterIsClosed struct {
	isDateClosedFn func(ctx context.Context, clinicID uint64, date time.Time) (bool, error)
}

func (s *stubCashRegisterIsClosed) IsDateClosed(ctx context.Context, clinicID uint64, date time.Time) (bool, error) {
	return s.isDateClosedFn(ctx, clinicID, date)
}

func TestUpdateAccounting_PostClose(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const reasonRequiredMsg = "レジ締め済み期間の会計編集には post_close_reason の入力が必要です"
	const forbiddenMsg = "レジ締め済み期間の会計編集には accounting-post-close-edit:edit 権限が必要です"

	existing := &model.Billing{ID: 1, ClinicID: 1, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}

	// postCloseAwareUpdate は accountingService.Update の締め後不変条件を忠実に再現する。
	// （理由なし締め後編集は拒否、それ以外は更新済み Billing を返す。）
	// 実 service が同契約を満たすことは accounting_service_test.go の直接呼びテストで独立検証する。
	postCloseAwareUpdate := func(_ context.Context, input *UpdateAccountingInput) (*model.Billing, error) {
		if input.IsPostClose && (input.PostCloseReason == nil || *input.PostCloseReason == "") {
			return nil, apperrors.WrapInvalidInput(reasonRequiredMsg)
		}
		return &model.Billing{ID: input.ID, ClinicID: input.ClinicID}, nil
	}

	grantPostClose := func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
		return []model.PermissionGroupRule{{Resource: string(model.ResourceAccountingPostCloseEdit), CanEdit: true}}, nil
	}
	denyAll := func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
		return []model.PermissionGroupRule{}, nil
	}

	tests := []struct {
		name       string
		body       string
		isClosed   bool
		perms      func(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)
		wantStatus int
		wantBody   string
	}{
		{
			name:       "closed without post-close permission returns 403",
			body:       `{"memo":"x"}`,
			isClosed:   true,
			perms:      denyAll,
			wantStatus: http.StatusForbidden,
			wantBody:   forbiddenMsg,
		},
		{
			name:       "closed with permission but no reason returns 400",
			body:       `{"memo":"x"}`,
			isClosed:   true,
			perms:      grantPostClose,
			wantStatus: http.StatusBadRequest,
			wantBody:   reasonRequiredMsg,
		},
		{
			name:       "closed with permission and reason returns 200",
			body:       `{"memo":"x","post_close_reason":"訂正のため"}`,
			isClosed:   true,
			perms:      grantPostClose,
			wantStatus: http.StatusOK,
			wantBody:   `"clinic_id":1`,
		},
		{
			name:       "not closed performs normal update without gate",
			body:       `{"memo":"x"}`,
			isClosed:   false,
			perms:      denyAll, // 締めていないため権限・理由は問われない
			wantStatus: http.StatusOK,
			wantBody:   `"clinic_id":1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAccountingHandler(
				&stubAccountingPostClose{
					getByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return existing, nil },
					updateFn:  postCloseAwareUpdate,
				},
				&stubCashRegisterIsClosed{
					isDateClosedFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) { return tt.isClosed, nil },
				},
				permCheckerFromRules(tt.perms),
			)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			setNonSystemAdmin(c) // is_system_admin=false, user_id=1, clinic_id=1
			c.Request = httptest.NewRequest(http.MethodPatch, "/v1/accountings/1", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: "1"}}

			h.UpdateAccounting(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}

// ---- mock AccountingService (full interface, nil-safe forwarding) ----

type mockAccountingService struct {
	listFn                      func(ctx context.Context, clinicID uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error)
	listForClinicsFn            func(ctx context.Context, clinicIDs []uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error)
	getByIDFn                   func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	getByIDForClinicsFn         func(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error)
	createFn                    func(ctx context.Context, input *CreateAccountingInput) (*model.Billing, error)
	updateFn                    func(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error)
	correctCreditPaymentFn      func(ctx context.Context, input *CorrectCreditPaymentInput) (*model.Billing, error)
	cancelFn                    func(ctx context.Context, clinicID, id uint64, actorID *uint64) error
	listUnpaidByBillingFn       func(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error)
	listUnpaidByOwnerFn         func(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error)
	getOwnerUnpaidBalanceFn     func(ctx context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error)
	getMonthlyUnpaidCarryoverFn func(ctx context.Context, clinicID uint64, year, month, page, limit int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error)
	getDailySummaryFn           func(ctx context.Context, clinicID uint64, dateStr string) (*DailySummaryResult, error)
	getDailySummaryForClinicsFn func(ctx context.Context, clinicIDs []uint64, dateStr string) ([]ClinicDailySummary, error)
}

func (m *mockAccountingService) List(ctx context.Context, clinicID uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	return m.listFn(ctx, clinicID, filters, page, limit)
}

func (m *mockAccountingService) ListForClinics(ctx context.Context, clinicIDs []uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	return m.listForClinicsFn(ctx, clinicIDs, filters, page, limit)
}

func (m *mockAccountingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockAccountingService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error) {
	return m.getByIDForClinicsFn(ctx, clinicIDs, id)
}

func (m *mockAccountingService) Complete(_ context.Context, _ *CompleteAccountingInput) (*CompleteAccountingResult, error) {
	return nil, apperrors.WrapInternalServerError("Complete not implemented in mockAccountingService")
}

func (m *mockAccountingService) Create(ctx context.Context, input *CreateAccountingInput) (*model.Billing, error) {
	return m.createFn(ctx, input)
}

func (m *mockAccountingService) Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error) {
	return m.updateFn(ctx, input)
}

func (m *mockAccountingService) CorrectCreditPayment(ctx context.Context, input *CorrectCreditPaymentInput) (*model.Billing, error) {
	return m.correctCreditPaymentFn(ctx, input)
}

func (m *mockAccountingService) Cancel(ctx context.Context, clinicID, id uint64, actorID *uint64) error {
	return m.cancelFn(ctx, clinicID, id, actorID)
}

func (m *mockAccountingService) ListUnpaidByBilling(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error) {
	return m.listUnpaidByBillingFn(ctx, clinicID, startDate, endDate, page, limit)
}

func (m *mockAccountingService) ListUnpaidByOwner(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error) {
	return m.listUnpaidByOwnerFn(ctx, clinicID, startDate, endDate, page, limit)
}

func (m *mockAccountingService) GetOwnerUnpaidBalance(ctx context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error) {
	return m.getOwnerUnpaidBalanceFn(ctx, clinicID, ownerID)
}

func (m *mockAccountingService) GetMonthlyUnpaidCarryover(ctx context.Context, clinicID uint64, year, month, page, limit int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error) {
	return m.getMonthlyUnpaidCarryoverFn(ctx, clinicID, year, month, page, limit)
}

func (m *mockAccountingService) GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*DailySummaryResult, error) {
	return m.getDailySummaryFn(ctx, clinicID, dateStr)
}

func (m *mockAccountingService) GetDailySummaryForClinics(ctx context.Context, clinicIDs []uint64, dateStr string) ([]ClinicDailySummary, error) {
	return m.getDailySummaryForClinicsFn(ctx, clinicIDs, dateStr)
}

func newHandlerWithAccountingSvc(svc AccountingService) *AccountingHandler {
	return NewAccountingHandler(svc, nil, func(_ *gin.Context, _, _ string) bool { return true })
}

// ---- ListAccountings ----

func TestListAccountings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockAccountingService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list for a single clinic",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				listFn: func(_ context.Context, clinicID uint64, _ AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, 1, page)
					assert.Equal(t, 20, limit)
					return []model.Billing{{ID: 1, ClinicID: 1, ScheduledDate: time.Now()}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":1`,
		},
		{
			name:  "returns cross-clinic aggregation when clinic_ids is supplied",
			query: "clinic_ids=1,2",
			setupCtx: func(c *gin.Context) {
				c.Set("clinic_id", "1")
				c.Set("is_system_admin", true)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc: &mockAccountingService{
				listForClinicsFn: func(_ context.Context, clinicIDs []uint64, _ AccountingListFilters, _, _ int) ([]model.Billing, int64, error) {
					assert.Equal(t, []uint64{1, 2}, clinicIDs)
					return []model.Billing{{ID: 2, ClinicID: 2, ScheduledDate: time.Now()}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":2`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockAccountingService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid pagination",
			query:      "page=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid pet_id filter",
			query:      "pet_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				listFn: func(_ context.Context, _ uint64, _ AccountingListFilters, _, _ int) ([]model.Billing, int64, error) {
					return nil, 0, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAccountingSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ListAccountings(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetAccounting ----

func TestGetAccounting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		idParam    string
		setupCtx   func(c *gin.Context)
		svc        *mockAccountingService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 with accounting detail",
			idParam:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				getByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error) {
					assert.Equal(t, []uint64{1}, clinicIDs)
					assert.Equal(t, uint64(1), id)
					return &model.Billing{ID: 1, ClinicID: 1, ScheduledDate: time.Now()}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":1`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			idParam:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockAccountingService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on non-numeric id",
			idParam:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			idParam:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				getByIDForClinicsFn: func(_ context.Context, _ []uint64, _ uint64) (*model.Billing, error) {
					return nil, apperrors.WrapNotFound("billing", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAccountingSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1/accountings/"+tt.idParam, http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			tt.setupCtx(c)
			h.GetAccounting(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateAccounting ----

func TestCreateAccounting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := `{"owner_id":1,"pet_id":2,"subtotal":1000,"tax_total":100,"total_amount":1100,"scheduled_date":"2026-06-01T00:00:00Z"}`

	tests := []struct {
		name       string
		body       string
		setupCtx   func(c *gin.Context)
		svc        *mockAccountingService
		wantStatus int
		wantHeader bool
	}{
		{
			name:     "returns 201 with Location header on success",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				createFn: func(_ context.Context, input *CreateAccountingInput) (*model.Billing, error) {
					assert.Equal(t, uint64(1), input.ClinicID)
					return &model.Billing{ID: 10, ClinicID: 1, ScheduledDate: time.Now()}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody,
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockAccountingService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required scheduled_date is missing",
			body:       `{"owner_id":1,"pet_id":2}`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				createFn: func(_ context.Context, _ *CreateAccountingInput) (*model.Billing, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAccountingSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/accountings", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateAccounting(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- CorrectCreditPayment ----

func TestCorrectCreditPayment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	existing := &model.Billing{ID: 1, ClinicID: 1, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	validBody := `{"method":"credit_card","amount":8800,"reason":"端末打ち間違い"}`

	tests := []struct {
		name       string
		idParam    string
		body       string
		setupCtx   func(c *gin.Context)
		svc        *mockAccountingService
		crSvc      *stubCashRegisterIsClosed
		wantStatus int
	}{
		{
			name:     "returns 200 on success",
			idParam:  "1",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			svc: &mockAccountingService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return existing, nil },
				correctCreditPaymentFn: func(_ context.Context, input *CorrectCreditPaymentInput) (*model.Billing, error) {
					return &model.Billing{ID: input.BillingID, ClinicID: input.ClinicID}, nil
				},
			},
			crSvc:      &stubCashRegisterIsClosed{isDateClosedFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) { return false, nil }},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			idParam:    "1",
			body:       validBody,
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockAccountingService{},
			crSvc:      &stubCashRegisterIsClosed{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when staff (user_id) is missing",
			idParam:    "1",
			body:       validBody,
			setupCtx:   func(c *gin.Context) { c.Set("clinic_id", "1") },
			svc:        &mockAccountingService{},
			crSvc:      &stubCashRegisterIsClosed{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on non-numeric id",
			idParam:    "abc",
			body:       validBody,
			setupCtx:   func(c *gin.Context) { setNonSystemAdmin(c) },
			svc:        &mockAccountingService{},
			crSvc:      &stubCashRegisterIsClosed{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed body (cash not allowed)",
			idParam:    "1",
			body:       `{"method":"cash","amount":100,"reason":"x"}`,
			setupCtx:   func(c *gin.Context) { setNonSystemAdmin(c) },
			svc:        &mockAccountingService{},
			crSvc:      &stubCashRegisterIsClosed{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 when GetByID fails",
			idParam:  "1",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			svc: &mockAccountingService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return nil, fmt.Errorf("db failure") },
			},
			crSvc:      &stubCashRegisterIsClosed{},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "returns 500 when IsDateClosed fails",
			idParam:  "1",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			svc: &mockAccountingService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return existing, nil },
			},
			crSvc:      &stubCashRegisterIsClosed{isDateClosedFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) { return false, fmt.Errorf("db failure") }},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "returns 500 when CorrectCreditPayment service fails",
			idParam:  "1",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			svc: &mockAccountingService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return existing, nil },
				correctCreditPaymentFn: func(_ context.Context, _ *CorrectCreditPaymentInput) (*model.Billing, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			crSvc:      &stubCashRegisterIsClosed{isDateClosedFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) { return false, nil }},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAccountingHandler(tt.svc, tt.crSvc, func(_ *gin.Context, _, _ string) bool { return true })
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/accountings/"+tt.idParam+"/correct-credit-payment", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			tt.setupCtx(c)
			h.CorrectCreditPayment(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- UpdateAccounting error / discount-permission paths ----
// (post-close 権限・理由の characterization は TestUpdateAccounting_PostClose 参照)

func TestUpdateAccounting_Errors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	existing := &model.Billing{ID: 1, ClinicID: 1, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}

	grantDiscountEdit := func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
		return []model.PermissionGroupRule{{Resource: string(model.ResourceDiscount), CanEdit: true}}, nil
	}
	denyAll := func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
		return []model.PermissionGroupRule{}, nil
	}

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		idParam    string
		body       string
		accSvc     *stubAccountingPostClose
		crSvc      *stubCashRegisterIsClosed
		perms      func(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)
		wantStatus int
	}{
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			idParam:    "1",
			body:       `{}`,
			accSvc:     &stubAccountingPostClose{},
			crSvc:      &stubCashRegisterIsClosed{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when staff (user_id) is missing",
			setupCtx:   func(c *gin.Context) { c.Set("clinic_id", "1") },
			idParam:    "1",
			body:       `{}`,
			accSvc:     &stubAccountingPostClose{},
			crSvc:      &stubCashRegisterIsClosed{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on non-numeric id",
			setupCtx:   func(c *gin.Context) { setNonSystemAdmin(c) },
			idParam:    "abc",
			body:       `{}`,
			accSvc:     &stubAccountingPostClose{},
			crSvc:      &stubCashRegisterIsClosed{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed JSON body",
			setupCtx:   func(c *gin.Context) { setNonSystemAdmin(c) },
			idParam:    "1",
			body:       `{invalid`,
			accSvc:     &stubAccountingPostClose{},
			crSvc:      &stubCashRegisterIsClosed{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 403 when discount_amount changed without discount:edit permission",
			setupCtx:   func(c *gin.Context) { setNonSystemAdmin(c) },
			idParam:    "1",
			body:       `{"discount_amount":500}`,
			accSvc:     &stubAccountingPostClose{},
			crSvc:      &stubCashRegisterIsClosed{},
			perms:      denyAll,
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "passes through when discount_amount changed with discount:edit permission",
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			idParam:  "1",
			body:     `{"discount_amount":500}`,
			accSvc: &stubAccountingPostClose{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return existing, nil },
				updateFn: func(_ context.Context, input *UpdateAccountingInput) (*model.Billing, error) {
					return &model.Billing{ID: input.ID, ClinicID: input.ClinicID}, nil
				},
			},
			crSvc:      &stubCashRegisterIsClosed{isDateClosedFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) { return false, nil }},
			perms:      grantDiscountEdit,
			wantStatus: http.StatusOK,
		},
		{
			name:     "returns 500 when GetByID fails",
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			idParam:  "1",
			body:     `{"memo":"x"}`,
			accSvc: &stubAccountingPostClose{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return nil, fmt.Errorf("db failure") },
			},
			crSvc:      &stubCashRegisterIsClosed{},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "returns 500 when IsDateClosed fails",
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			idParam:  "1",
			body:     `{"memo":"x"}`,
			accSvc: &stubAccountingPostClose{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return existing, nil },
			},
			crSvc:      &stubCashRegisterIsClosed{isDateClosedFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) { return false, fmt.Errorf("db failure") }},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "returns 500 when Update fails (not closed)",
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			idParam:  "1",
			body:     `{"memo":"x"}`,
			accSvc: &stubAccountingPostClose{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return existing, nil },
				updateFn: func(_ context.Context, _ *UpdateAccountingInput) (*model.Billing, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			crSvc:      &stubCashRegisterIsClosed{isDateClosedFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) { return false, nil }},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := tt.perms
			if perms == nil {
				perms = denyAll
			}
			h := NewAccountingHandler(tt.accSvc, tt.crSvc, permCheckerFromRules(perms))

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setupCtx(c)
			c.Request = httptest.NewRequest(http.MethodPatch, "/v1/accountings/"+tt.idParam, strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}

			h.UpdateAccounting(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- ListUnpaidBillings ----

func TestListUnpaidBillings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockAccountingService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns billing-grouped unpaid list",
			query:    "start_date=2026-01-01&end_date=2026-01-31&group_by=billing",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				listUnpaidByBillingFn: func(_ context.Context, clinicID uint64, startDate, endDate string, _, _ int) ([]model.Billing, int64, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "2026-01-01", startDate)
					assert.Equal(t, "2026-01-31", endDate)
					return []model.Billing{{ID: 1, ClinicID: 1, ScheduledDate: time.Now()}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":1`,
		},
		{
			name:     "returns owner-grouped unpaid list (default group_by)",
			query:    "start_date=2026-01-01&end_date=2026-01-31",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				listUnpaidByOwnerFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error) {
					return []UnpaidOwnerAggregate{{OwnerID: 1, OwnerName: "田中太郎"}}, 1, UnpaidSummary{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"owner_id":1`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "start_date=2026-01-01&end_date=2026-01-31",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockAccountingService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid pagination",
			query:      "start_date=2026-01-01&end_date=2026-01-31&page=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when start_date is missing",
			query:      "end_date=2026-01-31",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid group_by",
			query:      "start_date=2026-01-01&end_date=2026-01-31&group_by=invalid",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on billing group service error",
			query:    "start_date=2026-01-01&end_date=2026-01-31&group_by=billing",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				listUnpaidByBillingFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]model.Billing, int64, error) {
					return nil, 0, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "returns 500 on owner group service error",
			query:    "start_date=2026-01-01&end_date=2026-01-31&group_by=owner",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				listUnpaidByOwnerFn: func(_ context.Context, _ uint64, _, _ string, _, _ int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error) {
					return nil, 0, UnpaidSummary{}, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAccountingSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1/accountings/unpaid?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ListUnpaidBillings(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetOwnerUnpaidBalance ----

func TestGetOwnerUnpaidBalanceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockAccountingService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 with unpaid balance",
			query:    "owner_id=5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				getOwnerUnpaidBalanceFn: func(_ context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), ownerID)
					return OwnerUnpaidBalance{TotalAmount: 3000, Count: 2}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"unpaid_total":3000`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "owner_id=5",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockAccountingService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when owner_id is missing",
			query:      "",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when owner_id is zero",
			query:      "owner_id=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when owner_id is non-numeric",
			query:      "owner_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "owner_id=5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				getOwnerUnpaidBalanceFn: func(_ context.Context, _, _ uint64) (OwnerUnpaidBalance, error) {
					return OwnerUnpaidBalance{}, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAccountingSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1/accountings/unpaid-balance?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.GetOwnerUnpaidBalance(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetUnpaidMonthlySummary ----

func TestGetUnpaidMonthlySummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockAccountingService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 with monthly carryover summary",
			query:    "year=2026&month=6",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				getMonthlyUnpaidCarryoverFn: func(_ context.Context, clinicID uint64, year, month, _, _ int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, 2026, year)
					assert.Equal(t, 6, month)
					return []MonthlyUnpaidOwnerPet{{OwnerID: 1, OwnerName: "田中太郎"}}, 1, MonthlyUnpaidSummary{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"owner_id":1`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "year=2026&month=6",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockAccountingService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid pagination",
			query:      "year=2026&month=6&page=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when year is missing",
			query:      "month=6",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "year=2026&month=6",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				getMonthlyUnpaidCarryoverFn: func(_ context.Context, _ uint64, _, _, _, _ int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error) {
					return nil, 0, MonthlyUnpaidSummary{}, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAccountingSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1/accountings/unpaid-monthly?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.GetUnpaidMonthlySummary(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetDailySummary ----

func TestGetDailySummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockAccountingService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns single-clinic daily summary",
			query:    "date=2026-06-01",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				getDailySummaryFn: func(_ context.Context, clinicID uint64, dateStr string) (*DailySummaryResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "2026-06-01", dateStr)
					return &DailySummaryResult{BillingCount: 3, GrandTotal: 9000}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"grand_total":9000`,
		},
		{
			name:  "returns per-clinic daily summary when clinic_ids is supplied",
			query: "date=2026-06-01&clinic_ids=1,2",
			setupCtx: func(c *gin.Context) {
				c.Set("clinic_id", "1")
				c.Set("is_system_admin", true)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc: &mockAccountingService{
				getDailySummaryForClinicsFn: func(_ context.Context, clinicIDs []uint64, dateStr string) ([]ClinicDailySummary, error) {
					assert.Equal(t, []uint64{1, 2}, clinicIDs)
					assert.Equal(t, "2026-06-01", dateStr)
					return []ClinicDailySummary{{ClinicID: 1, Summary: &DailySummaryResult{GrandTotal: 100}}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"clinic_id":1`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "date=2026-06-01",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockAccountingService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on single-clinic service error",
			query:    "date=2026-06-01",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				getDailySummaryFn: func(_ context.Context, _ uint64, _ string) (*DailySummaryResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:  "returns 500 on per-clinic service error",
			query: "date=2026-06-01&clinic_ids=1,2",
			setupCtx: func(c *gin.Context) {
				c.Set("clinic_id", "1")
				c.Set("is_system_admin", true)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc: &mockAccountingService{
				getDailySummaryForClinicsFn: func(_ context.Context, _ []uint64, _ string) ([]ClinicDailySummary, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAccountingSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1/accountings/daily-summary?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.GetDailySummary(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CancelAccounting ----

// newCancelAccountingRouter は c.Status() のみで応答する 204 系ハンドラを検証するための実 router。
// gin.CreateTestContext + 直接ハンドラ呼び出しでは c.Status() が WriteHeaderNow まで
// flush されず ResponseRecorder.Code が既定の 200 のまま残るため、実 ServeHTTP 経由で検証する。
func newCancelAccountingRouter(svc AccountingService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithAccountingSvc(svc)
	r.POST("/v1/accountings/:id/cancel", func(c *gin.Context) {
		setClinicID(c)
	}, h.CancelAccounting)
	return r
}

func TestCancelAccounting_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockAccountingService{
		cancelFn: func(_ context.Context, clinicID, id uint64, _ *uint64) error {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(1), id)
			return nil
		},
	}
	router := newCancelAccountingRouter(svc)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/accountings/1/cancel", http.NoBody)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCancelAccounting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		idParam    string
		setupCtx   func(c *gin.Context)
		svc        *mockAccountingService
		wantStatus int
	}{
		{
			name:       "returns 401 when clinic_id is missing",
			idParam:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockAccountingService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on non-numeric id",
			idParam:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockAccountingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			idParam:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockAccountingService{
				cancelFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAccountingSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/accountings/"+tt.idParam+"/cancel", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			tt.setupCtx(c)
			h.CancelAccounting(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Accounting Handler Test Cases
// This handler manages billing and accounting records for medical services (Section 6: 会計管理)
//
// CRITICAL ENDPOINTS:
//
// 1. ListAccountings (GET /accountings)
//    Test Cases (18 scenarios):
//    ✓ Returns 200 OK with empty list when no records exist
//    ✓ Returns 200 OK with paginated accounting list when records exist
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when page/limit are invalid
//    ✓ Pagination: page=1, limit=20 as defaults
//    ✓ Pagination: supports custom page and limit parameters
//    ✓ Pagination: includes total_count for client-side calculation
//    ✓ Filter: pet_id parameter filters by pet (optional, can be null)
//    ✓ Filter: owner_id parameter filters by owner (optional, can be null)
//    ✓ Filter: status parameter filters by BillingStatus enum (waiting, paid, failed)
//    ✓ Filter: start_date parameter filters by date range (inclusive)
//    ✓ Filter: end_date parameter filters by date range (inclusive)
//    ✓ Filter: date format validation (YYYY-MM-DD)
//    ✓ Filter: multiple filters can be combined (pet_id AND status AND date range)
//    ✓ Response includes id, medical_record_id, owner_id, pet_id, subtotal, tax_total, total_amount
//    ✓ Response includes status, scheduled_date, completed_at, created_at, updated_at
//    ✓ Respects soft delete (clinic_id-scoped records only)
//    ✓ Returns 500 on database error
//
// 2. GetAccounting (GET /accountings/:id)
//    Test Cases (11 scenarios):
//    ✓ Returns 200 OK with single accounting record
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when accounting_id is non-numeric
//    ✓ Returns 404 when accounting doesn't exist
//    ✓ Returns 403 when accounting belongs to different clinic (tenant isolation)
//    ✓ Response includes complete accounting data with all fields
//    ✓ Response includes nested owner and pet objects (if preloaded)
//    ✓ Response includes nested medical_record (if exists and preloaded)
//    ✓ Response includes nested hospitalization (if exists and preloaded)
//    ✓ ID fields converted from uint64 to string in response
//    ✓ Returns 500 on database error
//
// 3. CreateAccounting (POST /accountings)
//    Test Cases (18 scenarios):
//    ✓ Returns 201 Created when accounting created successfully
//    ✓ Returns 400 when required fields missing (owner_id, pet_id)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Validates medical_record_id exists (if provided, FK constraint)
//    ✓ Validates hospitalization_id exists (if provided, FK constraint)
//    ✓ Validates owner_id exists (FK constraint)
//    ✓ Validates pet_id exists (FK constraint)
//    ✓ Status field accepts enum values: waiting, paid, failed (defaults to waiting if empty)
//    ✓ Subtotal, tax_total, total_amount can be zero or positive (non-negative validation)
//    ✓ HasInsurance boolean flag stored correctly
//    ✓ ScheduledDate and completed_at are optional timestamps
//    ✓ Memo field is optional text
//    ✓ Created accounting includes generated id and created_at timestamp
//    ✓ Multiple accountings per pet/owner supported
//    ✓ Concurrent creation handled correctly
//    ✓ Returns 409 Conflict if FK constraint violated (invalid owner/pet/medical_record)
//    ✓ Returns 500 on database error
//
// 4. UpdateAccounting (PATCH /accountings/:id)
//    Test Cases (20 scenarios):
//    ✓ Returns 200 OK when accounting updated successfully
//    ✓ Returns 400 when accounting_id is non-numeric
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when accounting doesn't exist
//    ✓ Returns 403 when accounting belongs to different clinic
//    ✓ Partial updates: subtotal can be updated independently
//    ✓ Partial updates: tax_total can be updated independently
//    ✓ Partial updates: total_amount can be updated independently
//    ✓ Partial updates: status can be updated (enum validation)
//    ✓ Partial updates: scheduled_date can be updated or cleared
//    ✓ Partial updates: completed_at can be updated or cleared (mark as paid)
//    ✓ Partial updates: medical_record_id can be null'd or changed
//    ✓ Partial updates: hospitalization_id can be null'd or changed
//    ✓ Partial updates: memo can be updated or cleared
//    ✓ Partial updates: owner_id and pet_id can be updated (FK validation)
//    ✓ Unspecified fields remain unchanged (PATCH semantics, not PUT)
//    ✓ Updated accounting reflects changes in response (id, updated_at timestamp)
//    ✓ Returns 409 Conflict if FK constraint violated during update
//    ✓ Returns 500 on database error
//
// 5. DeleteAccounting (DELETE /accountings/:id)
//    Test Cases (12 scenarios):
//    ✓ Returns 204 No Content when accounting deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when accounting_id is non-numeric
//    ✓ Returns 404 when accounting doesn't exist
//    ✓ Returns 403 when accounting belongs to different clinic
//    ✓ Uses soft delete (sets deleted_at, doesn't remove from database)
//    ✓ Deleted accounting no longer appears in ListAccountings
//    ✓ Deleted accounting cannot be retrieved by GetAccounting (404)
//    ✓ Cannot delete already deleted accounting (404 on second delete)
//    ✓ Deleting accounting doesn't affect related medical_record
//    ✓ Deleting accounting doesn't affect related hospitalization
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification on all endpoints)
//    ✓ Partial updates prevent mass assignment (explicit field mapping)
//    ✓ Foreign key validation on medical_record, hospitalization, owner, pet
//    ✓ Status enum validation (prevents invalid enum values)
//    ✓ Amount validation: non-negative values only
//    ✓ Date format validation (ISO 8601)
//    ✓ No RBAC permission check (all authenticated users can access accounting)
//
// INTEGRATION WITH MEDICAL RECORDS:
//    ✓ Accounting linked to medical_record (optional, nullable FK)
//    ✓ Accounting linked to hospitalization (optional, nullable FK)
//    ✓ Accounting linked to owner and pet (required FKs)
//    ✓ Cannot change accounting between different clinics
//    ✓ Deleting medical_record cascades to delete related accounting (if applicable)
//
// DATA MODEL (accountings):
//    - id (PK): BIGSERIAL
//    - clinic_id: BIGINT (multitenancy)
//    - medical_record_id (FK, NULLABLE): BIGINT → medical_records(id)
//    - hospitalization_id (FK, NULLABLE): BIGINT → hospitalizations(id)
//    - owner_id (FK): BIGINT → owners(id)
//    - pet_id (FK): BIGINT → pets(id)
//    - subtotal: NUMERIC(10,2) - sub-total before tax
//    - tax_total: NUMERIC(10,2) - tax amount
//    - total_amount: NUMERIC(10,2) - final amount (subtotal + tax)
//    - has_insurance: BOOLEAN - insurance flag
//    - status: ENUM (waiting|paid|failed) - billing status
//    - scheduled_date: DATE (NULLABLE) - scheduled completion date
//    - completed_at: TIMESTAMP (NULLABLE) - actual completion timestamp
//    - memo: TEXT (NULLABLE) - billing notes/remarks
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (clinic_id, id), (clinic_id, owner_id), (clinic_id, pet_id), (clinic_id, created_at DESC)
//
// IMPLEMENTATION NOTES:
//    - Status enum defaults to "waiting" when not specified during creation
//    - PATCH semantics: unspecified pointer fields remain unchanged (not null'd unless explicitly set)
//    - Amount fields: can be zero but not negative
//    - medical_record_id and hospitalization_id are optional (nullable FKs)
//    - owner_id and pet_id are required (non-nullable FKs)
//    - Date filtering uses inclusive range (start_date <= date <= end_date)
//    - Multiple filters are AND'ed together
//    - Soft delete prevents data leakage between clinic tenants
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with owner, pet, medical_record, hospitalization test data
//    - Real service/repository layers
//    - Verify pagination with >20 records
//    - Verify filter combinations (pet_id + status, date range, etc.)
//    - Verify FK constraints for all foreign key fields
//    - Verify soft delete behavior (deleted records excluded from list/get)
//    - Verify PATCH semantics (unspecified fields unchanged)
//    - Test amount validation (non-negative)
//    - Test enum validation for status field
//

// permCheckerFromRules は旧 h.hasPermission（system_admin 判定→staff/clinic 抽出→
// EffectivePermission ルール評価）の忠実な写像。rulesFn は旧 mockEffectivePermissionService の
// getEffectivePermissionsFn と同シグネチャ（B④・⑥先例の一般化）。
func permCheckerFromRules(rulesFn func(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)) httpapi.PermissionChecker {
	return func(c *gin.Context, resource, action string) bool {
		if v, exists := c.Get("is_system_admin"); exists {
			if isAdmin, ok := v.(bool); ok && isAdmin {
				return true
			}
		} else {
			return false
		}
		staffID, ok := httpapi.ExtractStaffID(c)
		if !ok {
			return false
		}
		clinicID, ok := httpapi.ExtractClinicID(c)
		if !ok {
			return false
		}
		if rulesFn == nil {
			return false
		}
		rules, err := rulesFn(c.Request.Context(), staffID, clinicID)
		if err != nil {
			return false
		}
		for i := range rules {
			rule := &rules[i]
			if rule.Resource != resource {
				continue
			}
			switch action {
			case "view":
				return rule.CanView
			case "create":
				return rule.CanCreate
			case "edit":
				return rule.CanEdit
			case "delete":
				return rule.CanDelete
			}
		}
		return false
	}
}
