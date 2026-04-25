package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// mockAccountingRepository は AccountingRepository のテスト用モック実装
type mockAccountingRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	createFn       func(ctx context.Context, clinicID uint64, accounting *model.Billing) error
	updateFieldsFn func(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error)
}

func (m *mockAccountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (m *mockAccountingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockAccountingRepository) Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error {
	return m.createFn(ctx, clinicID, accounting)
}

func (m *mockAccountingRepository) Update(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error) {
	return m.updateFieldsFn(ctx, clinicID, billingID, fields)
}

func (m *mockAccountingRepository) SavePayment(_ context.Context, _ *model.Payment) error {
	return nil
}

// BUG-370: 月末未納者一覧 repository メソッドの mock
func (m *mockAccountingRepository) FindUnpaidByBilling(_ context.Context, _ uint64, _ string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}

func (m *mockAccountingRepository) FindUnpaidByOwner(_ context.Context, _ uint64, _ string, _, _ int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	return nil, 0, repository.UnpaidSummary{}, nil
}

func (m *mockAccountingRepository) GetDailySummary(_ context.Context, _ uint64, _ time.Time) (*repository.DailySummaryResult, error) {
	return &repository.DailySummaryResult{PaymentTotals: []repository.PaymentMethodTotal{}, CategoryTotals: []repository.CategoryTotal{}}, nil
}

// FEAT-368: 集計・締め機能 mock スタブ
func (m *mockAccountingRepository) GetCloseAggregate(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
	return &repository.CloseAggregateResult{
		AggregateRows:  []repository.BillingAggregateRow{},
		BillingDetails: []repository.CloseBillingDetail{},
	}, nil
}

func (m *mockAccountingRepository) GetMonthlyReport(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
	return &repository.MonthlyReportResult{Rows: []repository.MonthlyReportRow{}}, nil
}

func (m *mockAccountingRepository) SumPaidByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func ptrString(v string) *string { return &v }

func TestAccountingService_List(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name         string
		clinicID     uint64
		petID        *uint64
		ownerID      *uint64
		status       *string
		page         int
		limit        int
		repoBillings []model.Billing
		repoTotal    int64
		repoErr      error
		wantLen      int
		wantTotal    int64
		wantErr      bool
	}{
		{
			name:     "returns all billings for clinic",
			clinicID: 1,
			petID:    nil,
			ownerID:  nil,
			status:   nil,
			page:     1,
			limit:    20,
			repoBillings: []model.Billing{
				{ID: 1, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusWaiting},
				{ID: 2, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusCompleted},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:     "filters by pet_id",
			clinicID: 1,
			petID:    ptrUint64(10),
			ownerID:  nil,
			status:   nil,
			page:     1,
			limit:    20,
			repoBillings: []model.Billing{
				{ID: 1, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusWaiting},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by owner_id",
			clinicID: 1,
			petID:    nil,
			ownerID:  ptrUint64(5),
			status:   nil,
			page:     1,
			limit:    20,
			repoBillings: []model.Billing{
				{ID: 2, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusCompleted},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by status",
			clinicID: 1,
			petID:    nil,
			ownerID:  nil,
			status:   ptrString("waiting"),
			page:     1,
			limit:    20,
			repoBillings: []model.Billing{
				{ID: 1, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusWaiting},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:         "returns empty list when no billings exist",
			clinicID:     1,
			petID:        nil,
			ownerID:      nil,
			status:       nil,
			page:         1,
			limit:        20,
			repoBillings: []model.Billing{},
			repoTotal:    0,
			repoErr:      nil,
			wantLen:      0,
			wantTotal:    0,
			wantErr:      false,
		},
		{
			name:         "propagates repository error",
			clinicID:     1,
			petID:        nil,
			ownerID:      nil,
			status:       nil,
			page:         1,
			limit:        20,
			repoBillings: nil,
			repoTotal:    0,
			repoErr:      errors.New("db connection error"),
			wantLen:      0,
			wantTotal:    0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _, _, _ *string, _, _ int) ([]model.Billing, int64, error) {
					return tt.repoBillings, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewAccountingService(repo)

			billings, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, tt.status, nil, nil, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, billings, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestAccountingService_GetByID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		clinicID    uint64
		id          uint64
		repoBilling *model.Billing
		repoErr     error
		wantBilling *model.Billing
		wantErr     error
	}{
		{
			name:        "returns billing when found",
			clinicID:    1,
			id:          10,
			repoBilling: &model.Billing{ID: 10, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusWaiting},
			repoErr:     nil,
			wantBilling: &model.Billing{ID: 10, ClinicID: 1, ScheduledDate: now, Status: model.BillingStatusWaiting},
			wantErr:     nil,
		},
		{
			name:        "returns not found error when billing does not exist",
			clinicID:    1,
			id:          999,
			repoBilling: nil,
			repoErr:     apperrors.WrapNotFound("billing", "999"),
			wantBilling: nil,
			wantErr:     apperrors.ErrNotFound,
		},
		{
			name:        "returns error on repository failure",
			clinicID:    1,
			id:          10,
			repoBilling: nil,
			repoErr:     errors.New("db error"),
			wantBilling: nil,
			wantErr:     errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
					return tt.repoBilling, tt.repoErr
				},
			}
			svc := NewAccountingService(repo)

			billing, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBilling, billing)
			}
		})
	}
}

func TestAccountingService_Create(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		input   CreateAccountingInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates billing successfully",
			input: CreateAccountingInput{
				ClinicID:      1,
				ScheduledDate: now,
				Status:        model.BillingStatusWaiting,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns validation error when scheduled_date is zero",
			input: CreateAccountingInput{
				ClinicID: 1,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: CreateAccountingInput{
				ClinicID:      1,
				ScheduledDate: now,
			},
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
					return tt.repoErr
				},
			}
			svc := NewAccountingService(repo)

			billing, err := svc.Create(context.Background(), &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, billing)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, billing)
			}
		})
	}
}

func TestAccountingService_Update(t *testing.T) {
	now := time.Now()
	status := model.BillingStatusCompleted
	tests := []struct {
		name    string
		input   UpdateAccountingInput
		repoRet *model.Billing
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name: "updates billing successfully",
			input: UpdateAccountingInput{
				ID:            1,
				ClinicID:      1,
				ScheduledDate: &now,
				Status:        &status,
			},
			repoRet: &model.Billing{ID: 1, ClinicID: 1, ScheduledDate: now, Status: status},
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name: "returns error when no fields to update",
			input: UpdateAccountingInput{
				ID:       1,
				ClinicID: 1,
				// 全フィールドが nil
			},
			repoRet: nil,
			repoErr: nil,
			wantErr: true,
			wantNF:  false,
		},
		{
			name: "returns not found error when billing does not exist",
			input: UpdateAccountingInput{
				ID:       999,
				ClinicID: 1,
				Status:   &status,
			},
			repoRet: nil,
			repoErr: apperrors.WrapNotFound("billing", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateAccountingInput{
				ID:       1,
				ClinicID: 1,
				Status:   &status,
			},
			repoRet: nil,
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
					return tt.repoRet, tt.repoErr
				},
				findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.Billing, error) {
					return tt.repoRet, tt.repoErr
				},
			}
			svc := NewAccountingService(repo)

			billing, err := svc.Update(context.Background(), &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, billing)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, billing)
			}
		})
	}
}

// TestAccountingService_Cancel は BUG-371: 論理削除 (status=cancelled) の挙動を検証する。
func TestAccountingService_Cancel(t *testing.T) {
	tests := []struct {
		name           string
		clinicID       uint64
		id             uint64
		findByIDResult *model.Billing
		findByIDErr    error
		updateErr      error
		wantErr        bool
		wantConflict   bool
		wantNF         bool
	}{
		{
			name:           "正常: waiting → cancelled に遷移する",
			clinicID:       1,
			id:             10,
			findByIDResult: &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusWaiting},
			wantErr:        false,
		},
		{
			name:           "正常: completed → cancelled に遷移する",
			clinicID:       1,
			id:             10,
			findByIDResult: &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusCompleted},
			wantErr:        false,
		},
		{
			name:           "異常: 既に cancelled の場合は ErrConflict",
			clinicID:       1,
			id:             10,
			findByIDResult: &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusCancelled},
			wantErr:        true,
			wantConflict:   true,
		},
		{
			name:        "異常: 存在しない場合は ErrNotFound 経由で error",
			clinicID:    1,
			id:          999,
			findByIDErr: apperrors.WrapNotFound("billing", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:           "異常: Update 失敗時はエラー伝播",
			clinicID:       1,
			id:             10,
			findByIDResult: &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusWaiting},
			updateErr:      errors.New("db error"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return tt.findByIDResult, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
					if tt.updateErr != nil {
						return nil, tt.updateErr
					}
					return tt.findByIDResult, nil
				},
			}
			svc := NewAccountingService(repo)

			err := svc.Cancel(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
