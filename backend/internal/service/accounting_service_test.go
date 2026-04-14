package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockAccountingRepository は AccountingRepository のテスト用モック実装
type mockAccountingRepository struct {
	findAllFn               func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	findByIDFn              func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	createFn                func(ctx context.Context, clinicID uint64, accounting *model.Billing) error
	updateFieldsFn          func(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error)
	deleteFn                func(ctx context.Context, clinicID, id uint64) error
	countItemsByBillingIDFn func(ctx context.Context, clinicID, billingID uint64) (int64, error)
}

func (m *mockAccountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (m *mockAccountingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockAccountingRepository) Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error {
	return m.createFn(ctx, clinicID, accounting)
}

func (m *mockAccountingRepository) UpdateFields(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error) {
	return m.updateFieldsFn(ctx, clinicID, billingID, fields)
}

func (m *mockAccountingRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockAccountingRepository) UpsertPayment(_ context.Context, _ *model.Payment) error {
	return nil
}

func (m *mockAccountingRepository) CountItemsByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error) {
	if m.countItemsByBillingIDFn == nil {
		return 0, nil
	}
	return m.countItemsByBillingIDFn(ctx, clinicID, billingID)
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

func TestAccountingService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		itemCount    int64
		countItemErr error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:         "deletes billing successfully when no items exist",
			clinicID:     1,
			id:           10,
			itemCount:    0,
			countItemErr: nil,
			repoErr:      nil,
			wantErr:      false,
		},
		{
			name:         "returns conflict error when billing has items",
			clinicID:     1,
			id:           10,
			itemCount:    3,
			countItemErr: nil,
			repoErr:      nil,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:         "returns error when item count check fails",
			clinicID:     1,
			id:           10,
			itemCount:    0,
			countItemErr: errors.New("db error"),
			repoErr:      nil,
			wantErr:      true,
		},
		{
			name:         "returns not found error when billing does not exist",
			clinicID:     1,
			id:           999,
			itemCount:    0,
			countItemErr: nil,
			repoErr:      apperrors.WrapNotFound("billing", "999"),
			wantErr:      true,
			wantNF:       true,
		},
		{
			name:         "returns error on repository failure",
			clinicID:     1,
			id:           10,
			itemCount:    0,
			countItemErr: nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNF:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountingRepository{
				countItemsByBillingIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.itemCount, tt.countItemErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewAccountingService(repo)

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

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
