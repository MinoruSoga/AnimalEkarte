package billing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestAccountingService_Update_RejectsMoveToClosedDestinationWithoutReason(t *testing.T) {
	source := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	dest := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	billing := &model.Billing{
		ID: 7, ClinicID: 1, Status: model.BillingStatusWaiting, ScheduledDate: source, TotalAmount: 1000,
	}
	var locked []string
	closeRepo := &mockCashRegisterCloseRepository{
		lockCloseBoundaryFn: func(_ context.Context, _ uint64, date time.Time) error {
			locked = append(locked, date.Format(time.DateOnly))
			return nil
		},
		hasCloseOnDateFn: func(_ context.Context, _ uint64, date time.Time) (bool, error) {
			return date.Format(time.DateOnly) == dest.Format(time.DateOnly), nil
		},
		findByDateAndPeriodFn: func(_ context.Context, _ uint64, date time.Time, period string) (*model.CashRegisterClose, error) {
			if period == "am" && date.Format(time.DateOnly) == dest.Format(time.DateOnly) {
				return &model.CashRegisterClose{ID: 88, ClinicID: 1, CloseDate: dest, Period: "am"}, nil
			}
			return nil, nil
		},
	}
	updated := false
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
			updated = true
			return billing, nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{},
		WithCashRegisterCloseRepository(closeRepo))

	memo := "move into closed day"
	_, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID: 7, ClinicID: 1, ScheduledDate: &dest, Memo: &memo, IsPostClose: false,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "got %v", err)
	assert.ErrorContains(t, err, "post_close_reason")
	assert.False(t, updated)
	assert.Equal(t, []string{"2026-06-01", "2026-06-10"}, locked)
}

func TestAccountingService_Update_ChecksSourceAndDestinationClosePeriods(t *testing.T) {
	source := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	dest := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	memo := "date change"

	newSvc := func(closedDays ...string) (AccountingService, *mockAccountingRepository, *[]string) {
		closed := make(map[string]struct{}, len(closedDays))
		for _, day := range closedDays {
			closed[day] = struct{}{}
		}
		var locked []string
		closeRepo := &mockCashRegisterCloseRepository{
			lockCloseBoundaryFn: func(_ context.Context, _ uint64, date time.Time) error {
				locked = append(locked, date.Format(time.DateOnly))
				return nil
			},
			hasCloseOnDateFn: func(_ context.Context, _ uint64, date time.Time) (bool, error) {
				_, ok := closed[date.Format(time.DateOnly)]
				return ok, nil
			},
			findByDateAndPeriodFn: func(_ context.Context, _ uint64, date time.Time, period string) (*model.CashRegisterClose, error) {
				if period != "am" {
					return nil, nil
				}
				if _, ok := closed[date.Format(time.DateOnly)]; !ok {
					return nil, nil
				}
				return &model.CashRegisterClose{ID: 1, ClinicID: 1, CloseDate: date, Period: "am"}, nil
			},
		}
		billing := &model.Billing{
			ID: 7, ClinicID: 1, Status: model.BillingStatusWaiting, ScheduledDate: source, TotalAmount: 1000,
		}
		repo := &mockAccountingRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
				return billing, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
				copied := *billing
				copied.ScheduledDate = dest
				billing = &copied
				return billing, nil
			},
		}
		svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{},
			WithCashRegisterCloseRepository(closeRepo))
		return svc, repo, &locked
	}

	t.Run("source closed dest open still requires post-close reason", func(t *testing.T) {
		svc, _, locked := newSvc("2026-06-01")
		_, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID: 7, ClinicID: 1, ScheduledDate: &dest, Memo: &memo, IsPostClose: false,
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Equal(t, []string{"2026-06-01", "2026-06-10"}, *locked)
	})

	t.Run("source open dest closed requires post-close reason", func(t *testing.T) {
		svc, _, locked := newSvc("2026-06-10")
		_, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID: 7, ClinicID: 1, ScheduledDate: &dest, Memo: &memo, IsPostClose: false,
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Equal(t, []string{"2026-06-01", "2026-06-10"}, *locked)
	})

	t.Run("both open allows date change without post-close reason", func(t *testing.T) {
		svc, _, locked := newSvc()
		got, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID: 7, ClinicID: 1, ScheduledDate: &dest, Memo: &memo, IsPostClose: false,
		})
		assert.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, dest, got.ScheduledDate)
		assert.Equal(t, []string{"2026-06-01", "2026-06-10"}, *locked)
	})
}

func TestAccountingService_Update_CompletedStatusNoOpAllowsMemoCorrection(t *testing.T) {
	completed := model.BillingStatusCompleted
	memo := "post-complete memo"
	var captured map[string]any
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: clinicID, Status: model.BillingStatusCompleted}, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, fields map[string]any) (*model.Billing, error) {
			captured = fields
			return &model.Billing{ID: id, ClinicID: clinicID, Status: model.BillingStatusCompleted, Memo: memo}, nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

	got, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID: 30, ClinicID: 1, Status: &completed, Memo: &memo,
	})
	assert.NoError(t, err)
	require.NotNil(t, got)
	assert.NotContains(t, captured, "status")
	assert.Equal(t, memo, captured["memo"])
}

func TestAccountingService_Update_RejectsPaymentFinalizationOnOpenBilling(t *testing.T) {
	amount := int64(1000)
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: clinicID, Status: model.BillingStatusWaiting}, nil
		},
		savePaymentSplitsFn: func(_ context.Context, _ []model.PaymentSplit) error {
			t.Fatal("payments must not persist via generic PATCH on waiting billing")
			return nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, seededPayMethodMock())

	_, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID:            1,
		ClinicID:      1,
		BillingAmount: &amount,
		PaymentSplits: []PaymentSplitInput{{Method: model.PaymentMethodCash, Amount: 1000, ReceivedAmount: 1000}},
	})
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.ErrorContains(t, err, "POST /accountings/complete")
}

func TestAccountingService_Update_RejectsClientCompletedAt(t *testing.T) {
	completedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	memo := "x"
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: clinicID, Status: model.BillingStatusCompleted}, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Billing, error) {
			t.Fatal("completed_at must not persist via generic PATCH")
			return &model.Billing{ID: id, ClinicID: clinicID}, nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})

	_, err := svc.Update(context.Background(), &UpdateAccountingInput{
		ID: 1, ClinicID: 1, Memo: &memo, CompletedAt: &completedAt,
	})
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.ErrorContains(t, err, "completed_at")
}

func TestUpdateAccounting_PatchCancelledForbiddenWithoutCancelPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	existing := &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusWaiting, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Billing, error) {
			t.Fatal("PATCH must not persist cancelled")
			return existing, nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})
	h := NewAccountingHandler(
		svc,
		&stubCashRegisterIsClosed{isDateClosedFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) { return false, nil }},
		permCheckerFromRules(func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
			return []model.PermissionGroupRule{{Resource: string(model.ResourceAccounting), CanEdit: true}}, nil
		}),
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setNonSystemAdmin(c)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v1/accountings/1", strings.NewReader(`{"status":"cancelled"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.UpdateAccounting(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "POST /accountings/:id/cancel")
}

func TestUpdateAccounting_PatchCompletedFromWaitingRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	existing := &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusWaiting, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Billing, error) {
			t.Fatal("PATCH must not persist completed")
			return existing, nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})
	h := NewAccountingHandler(
		svc,
		&stubCashRegisterIsClosed{isDateClosedFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) { return false, nil }},
		func(_ *gin.Context, _, _ string) bool { return true },
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setNonSystemAdmin(c)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v1/accountings/1", strings.NewReader(`{"status":"completed"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.UpdateAccounting(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "POST /accountings/complete")
}

func TestCreateAccounting_RejectsClientCompletedAt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("legacy create must not persist client completed_at")
			return nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{})
	h := NewAccountingHandler(svc, nil, func(_ *gin.Context, _, _ string) bool { return true })

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setClinicID(c)
	body := `{"owner_id":1,"pet_id":2,"subtotal":1000,"tax_total":100,"total_amount":1100,"scheduled_date":"2026-06-01T00:00:00Z","completed_at":"2026-06-01T10:00:00Z"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/accountings", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.CreateAccounting(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "completed_at")
}

func TestCashRegisterCloseRepository_LockCloseBoundary_RequiresAmbientTx(t *testing.T) {
	repo := NewCashRegisterCloseRepository(nil)
	err := repo.LockCloseBoundary(context.Background(), 1, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "active transaction")
}

func TestAccountingService_CreateAndCancel_LockCloseBoundary(t *testing.T) {
	scheduled := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	var locked []string
	closeRepo := &mockCashRegisterCloseRepository{
		lockCloseBoundaryFn: func(_ context.Context, _ uint64, date time.Time) error {
			locked = append(locked, date.Format(time.DateOnly))
			return nil
		},
	}
	created := false
	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, billing *model.Billing) error {
			created = true
			billing.ID = 9
			return nil
		},
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			return &model.Billing{
				ID: id, ClinicID: clinicID, Status: model.BillingStatusWaiting, ScheduledDate: scheduled,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: clinicID, Status: model.BillingStatusCancelled, ScheduledDate: scheduled}, nil
		},
	}
	svc := NewAccountingService(
		repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{},
		WithCashRegisterCloseRepository(closeRepo),
	)

	billing, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID: 1, ScheduledDate: scheduled, Status: model.BillingStatusWaiting,
	})
	assert.NoError(t, err)
	require.NotNil(t, billing)
	assert.True(t, created)
	assert.Equal(t, []string{"2026-06-01"}, locked)

	locked = nil
	actor := uint64(1)
	assert.NoError(t, svc.Cancel(context.Background(), 1, 9, &actor))
	assert.Equal(t, []string{"2026-06-01"}, locked)
}
