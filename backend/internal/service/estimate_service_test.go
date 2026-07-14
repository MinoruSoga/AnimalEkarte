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

// ---- Estimate モック ----

const estimateTestCreatedByStaffID = uint64(1)

func estimateTestMembershipCounter() staffClinicMembershipCounter {
	return &mockStaffClinicMembershipCounter{}
}

type mockEstimateRepository struct {
	findAllFn              func(ctx context.Context, clinicID uint64, ownerID, medicalRecordID *uint64, status *string, page, limit int) ([]model.Estimate, int64, error)
	findByIDFn             func(ctx context.Context, clinicID, id uint64) (*model.Estimate, error)
	createFn               func(ctx context.Context, estimate *model.Estimate) error
	updateFn               func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn               func(ctx context.Context, clinicID, id uint64) error
	countItemsByEstimateID func(ctx context.Context, estimateID uint64) (int64, error)
}

func (m *mockEstimateRepository) FindAll(ctx context.Context, clinicID uint64, ownerID, medicalRecordID *uint64, status *string, page, limit int) ([]model.Estimate, int64, error) {
	return m.findAllFn(ctx, clinicID, ownerID, medicalRecordID, status, page, limit)
}

func (m *mockEstimateRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockEstimateRepository) Create(ctx context.Context, estimate *model.Estimate) error {
	return m.createFn(ctx, estimate)
}

func (m *mockEstimateRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockEstimateRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockEstimateRepository) CountItemsByEstimateID(ctx context.Context, _, estimateID uint64) (int64, error) {
	if m.countItemsByEstimateID != nil {
		return m.countItemsByEstimateID(ctx, estimateID)
	}
	return 0, nil
}

// ---- Tests ----

func TestEstimateService_List(t *testing.T) {
	ownerID5 := uint64(5)

	tests := []struct {
		name            string
		ownerID         *uint64
		medicalRecordID *uint64
		status          *string
		page            int
		limit           int
		repoEstimates   []model.Estimate
		repoTotal       int64
		repoErr         error
		wantLen         int
		wantTotal       int64
		wantErr         bool
	}{
		{
			name:    "returns estimates with pagination",
			ownerID: nil,
			status:  nil,
			page:    1,
			limit:   10,
			repoEstimates: []model.Estimate{
				{ID: 1, ClinicID: 1, Title: "見積1", TotalAmount: 10000},
				{ID: 2, ClinicID: 1, Title: "見積2", TotalAmount: 20000},
			},
			repoTotal: 25,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 25,
			wantErr:   false,
		},
		{
			name:    "filters by owner_id",
			ownerID: &ownerID5,
			status:  nil,
			page:    1,
			limit:   10,
			repoEstimates: []model.Estimate{
				{ID: 1, ClinicID: 1, OwnerID: &ownerID5, Title: "見積1", TotalAmount: 10000},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:          "returns empty list",
			page:          1,
			limit:         10,
			repoEstimates: []model.Estimate{},
			repoTotal:     0,
			repoErr:       nil,
			wantLen:       0,
			wantTotal:     0,
			wantErr:       false,
		},
		{
			name:    "propagates repository error",
			page:    1,
			limit:   10,
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEstimateRepository{
				findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _ *string, _ int, _ int) ([]model.Estimate, int64, error) {
					return tt.repoEstimates, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewEstimateService(repo, nil, nil, nil)

			estimates, total, err := svc.List(context.Background(), 1, tt.ownerID, tt.medicalRecordID, tt.status, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, estimates, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestEstimateService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoEstimate *model.Estimate
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name: "returns estimate when found",
			id:   1,
			repoEstimate: &model.Estimate{
				ID:          1,
				ClinicID:    1,
				Title:       "見積1",
				TotalAmount: 10000,
				Status:      model.EstimateStatusDraft,
			},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error",
			id:           999,
			repoEstimate: nil,
			repoErr:      apperrors.WrapNotFound("estimate", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoEstimate: nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEstimateRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					return tt.repoEstimate, tt.repoErr
				},
			}
			svc := NewEstimateService(repo, nil, nil, nil)

			estimate, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, estimate)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoEstimate, estimate)
			}
		})
	}
}

func TestEstimateService_Create(t *testing.T) {
	validUntil := time.Now().AddDate(0, 1, 0)

	tests := []struct {
		name         string
		input        *CreateEstimateInput
		repoErr      error
		findByIDErr  error
		repoEstimate *model.Estimate
		wantErr      bool
	}{
		{
			name: "creates estimate successfully with default status",
			input: &CreateEstimateInput{
				Title:       "新規見積",
				Subtotal:    10000,
				TaxTotal:    1000,
				TotalAmount: 11000,
				CreatedBy:   ptrU64(estimateTestCreatedByStaffID),
			},
			repoErr: nil,
			repoEstimate: &model.Estimate{
				ID:          1,
				ClinicID:    1,
				Title:       "新規見積",
				TotalAmount: 11000,
				Status:      model.EstimateStatusDraft,
			},
			wantErr: false,
		},
		{
			name: "creates estimate with custom status",
			input: &CreateEstimateInput{
				Title:       "カスタム見積",
				Status:      model.EstimateStatusSent,
				TotalAmount: 50000,
				CreatedBy:   ptrU64(estimateTestCreatedByStaffID),
			},
			repoErr: nil,
			repoEstimate: &model.Estimate{
				ID:          2,
				ClinicID:    1,
				Title:       "カスタム見積",
				TotalAmount: 50000,
				Status:      model.EstimateStatusSent,
			},
			wantErr: false,
		},
		{
			name: "returns error when title is empty",
			input: &CreateEstimateInput{
				Title:       "",
				TotalAmount: 10000,
			},
			wantErr: true,
		},
		{
			name: "returns error when subtotal is negative",
			input: &CreateEstimateInput{
				Title:    "負の小計",
				Subtotal: -1,
			},
			wantErr: true,
		},
		{
			name: "returns error when tax_total is negative",
			input: &CreateEstimateInput{
				Title:    "負の税合計",
				TaxTotal: -500,
			},
			wantErr: true,
		},
		{
			name: "returns error when total_amount is negative",
			input: &CreateEstimateInput{
				Title:       "負の合計金額",
				TotalAmount: -100,
			},
			wantErr: true,
		},
		{
			name: "returns error when insurance_amount is negative",
			input: &CreateEstimateInput{
				Title:           "負の保険金額",
				InsuranceAmount: -200,
			},
			wantErr: true,
		},
		{
			name: "returns error when discount_amount is negative",
			input: &CreateEstimateInput{
				Title:          "負の割引金額",
				DiscountAmount: -300,
			},
			wantErr: true,
		},
		{
			name: "returns error when estimate already exists",
			input: &CreateEstimateInput{
				Title:       "既存見積",
				TotalAmount: 20000,
				CreatedBy:   ptrU64(estimateTestCreatedByStaffID),
			},
			repoErr: apperrors.WrapAlreadyExists("estimate", "既存見積"),
			wantErr: true,
		},
		{
			name: "creates estimate with valid_until",
			input: &CreateEstimateInput{
				Title:       "期限付き見積",
				ValidUntil:  &validUntil,
				TotalAmount: 15000,
				CreatedBy:   ptrU64(estimateTestCreatedByStaffID),
			},
			repoErr: nil,
			repoEstimate: &model.Estimate{
				ID:          3,
				ClinicID:    1,
				Title:       "期限付き見積",
				ValidUntil:  &validUntil,
				TotalAmount: 15000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEstimateRepository{
				createFn: func(_ context.Context, _ *model.Estimate) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return tt.repoEstimate, nil
				},
			}
			svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter())

			estimate, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, estimate)
			}
		})
	}
}

func TestEstimateService_Update(t *testing.T) {
	newTitle := "更新見積"
	newAmount := int64(25000)
	newStatus := model.EstimateStatusSent

	tests := []struct {
		name         string
		input        *UpdateEstimateInput
		repoErr      error
		repoEstimate *model.Estimate
		wantErr      bool
	}{
		{
			name: "updates estimate successfully",
			input: &UpdateEstimateInput{
				Title:       &newTitle,
				TotalAmount: &newAmount,
			},
			repoErr: nil,
			repoEstimate: &model.Estimate{
				ID:          1,
				Title:       "更新見積",
				TotalAmount: 25000,
			},
			wantErr: false,
		},
		{
			name: "updates status",
			input: &UpdateEstimateInput{
				Status: &newStatus,
			},
			repoErr: nil,
			repoEstimate: &model.Estimate{
				ID:     1,
				Status: model.EstimateStatusSent,
			},
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   &UpdateEstimateInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when subtotal is negative",
			input: &UpdateEstimateInput{
				Subtotal: func() *int64 { v := int64(-1); return &v }(),
			},
			wantErr: true,
		},
		{
			name: "returns error when total_amount is negative",
			input: &UpdateEstimateInput{
				TotalAmount: func() *int64 { v := int64(-100); return &v }(),
			},
			wantErr: true,
		},
		{
			name: "returns error when discount_amount is negative",
			input: &UpdateEstimateInput{
				DiscountAmount: func() *int64 { v := int64(-50); return &v }(),
			},
			wantErr: true,
		},
		{
			name: "returns not found error",
			input: &UpdateEstimateInput{
				Title: &newTitle,
			},
			repoErr: apperrors.WrapNotFound("estimate", "999"),
			wantErr: true,
		},
		{
			name: "clears valid_until when flag set",
			input: &UpdateEstimateInput{
				ClearValidUntil: true,
			},
			repoErr: nil,
			repoEstimate: &model.Estimate{
				ID:         1,
				ValidUntil: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEstimateRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					if tt.repoErr != nil && apperrors.IsNotFound(tt.repoErr) {
						return nil, tt.repoErr
					}
					return tt.repoEstimate, nil
				},
			}
			svc := NewEstimateService(repo, nil, nil, nil)

			estimate, err := svc.Update(context.Background(), 1, 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, estimate)
			}
		})
	}
}

func TestEstimateService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		itemCount    int64
		countErr     error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:      "deletes estimate successfully when no items",
			id:        1,
			itemCount: 0,
			repoErr:   nil,
			wantErr:   false,
		},
		{
			name:         "returns conflict error when estimate has items",
			id:           2,
			itemCount:    3,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:     "returns error when count check fails",
			id:       1,
			countErr: errors.New("db error"),
			wantErr:  true,
		},
		{
			name:      "returns not found error on delete",
			id:        999,
			itemCount: 0,
			repoErr:   apperrors.WrapNotFound("estimate", "999"),
			wantErr:   true,
			wantNF:    true,
		},
		{
			name:      "returns error on repository delete failure",
			id:        1,
			itemCount: 0,
			repoErr:   errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEstimateRepository{
				countItemsByEstimateID: func(_ context.Context, _ uint64) (int64, error) {
					return tt.itemCount, tt.countErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewEstimateService(repo, nil, nil, nil)

			err := svc.Delete(context.Background(), 1, tt.id)

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
