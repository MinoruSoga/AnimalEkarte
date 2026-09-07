package billing

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
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
	lockEditableByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.Estimate, error)
	createFn               func(ctx context.Context, estimate *model.Estimate) error
	updateFn               func(ctx context.Context, clinicID, id uint64, cmd UpdateEstimateInput) error
	updateIfNotLockedFn    func(ctx context.Context, clinicID, id uint64, cmd UpdateEstimateInput) (*model.Estimate, error)
	deleteIfNotLockedFn    func(ctx context.Context, clinicID, id uint64) error
	countItemsByEstimateID func(ctx context.Context, estimateID uint64) (int64, error)
	allocateNextEstimateNo func(ctx context.Context, clinicID uint64) (string, error)
	replaceItemsFn         func(ctx context.Context, clinicID, estimateID uint64, items []model.EstimateItem) error
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

func (m *mockEstimateRepository) LockEditableByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error) {
	if m.lockEditableByIDFn != nil {
		return m.lockEditableByIDFn(ctx, clinicID, id)
	}
	estimate, err := m.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if estimate != nil && isEstimateLocked(estimate.Status) {
		return nil, apperrors.WrapConflict("承認済みまたは却下済みの見積書は編集できません")
	}
	return estimate, nil
}

func (m *mockEstimateRepository) Create(ctx context.Context, estimate *model.Estimate) error {
	return m.createFn(ctx, estimate)
}

func (m *mockEstimateRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateEstimateInput) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, cmd)
	}
	return nil
}

func (m *mockEstimateRepository) UpdateIfNotLocked(ctx context.Context, clinicID, id uint64, cmd UpdateEstimateInput) (*model.Estimate, error) {
	if m.updateIfNotLockedFn != nil {
		return m.updateIfNotLockedFn(ctx, clinicID, id, cmd)
	}
	return nil, nil
}

func (m *mockEstimateRepository) DeleteIfNotLocked(ctx context.Context, clinicID, id uint64) error {
	if m.deleteIfNotLockedFn != nil {
		return m.deleteIfNotLockedFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockEstimateRepository) CountItemsByEstimateID(ctx context.Context, _, estimateID uint64) (int64, error) {
	if m.countItemsByEstimateID != nil {
		return m.countItemsByEstimateID(ctx, estimateID)
	}
	return 0, nil
}

func (m *mockEstimateRepository) ReplaceItems(ctx context.Context, clinicID, estimateID uint64, items []model.EstimateItem) error {
	if m.replaceItemsFn != nil {
		return m.replaceItemsFn(ctx, clinicID, estimateID, items)
	}
	return nil
}

func (m *mockEstimateRepository) AllocateNextEstimateNo(ctx context.Context, clinicID uint64) (string, error) {
	if m.allocateNextEstimateNo != nil {
		return m.allocateNextEstimateNo(ctx, clinicID)
	}
	return "EST-1", nil
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
			svc := NewEstimateService(repo, nil, nil, nil, nil, noopTransactor{})

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
			svc := NewEstimateService(repo, nil, nil, nil, nil, noopTransactor{})

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
		wantConflict bool
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
			name: "returns conflict when status is approved",
			input: &CreateEstimateInput{
				Title:       "承認済み直作成",
				Status:      model.EstimateStatusApproved,
				TotalAmount: 10000,
				CreatedBy:   ptrU64(estimateTestCreatedByStaffID),
			},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name: "returns conflict when status is rejected",
			input: &CreateEstimateInput{
				Title:       "却下済み直作成",
				Status:      model.EstimateStatusRejected,
				TotalAmount: 10000,
				CreatedBy:   ptrU64(estimateTestCreatedByStaffID),
			},
			wantErr:      true,
			wantConflict: true,
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
			createCalled := false
			repo := &mockEstimateRepository{
				createFn: func(_ context.Context, _ *model.Estimate) error {
					createCalled = true
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return tt.repoEstimate, nil
				},
			}
			svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})

			estimate, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
					assert.False(t, createCalled, "repo.Create must not be called for locked status")
				}
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
		wantConflict bool
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
		{
			name: "returns conflict when estimate is approved",
			input: &UpdateEstimateInput{
				Title: &newTitle,
			},
			repoEstimate: &model.Estimate{
				ID:     1,
				Status: model.EstimateStatusApproved,
			},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name: "returns conflict when estimate is rejected",
			input: &UpdateEstimateInput{
				Title: &newTitle,
			},
			repoEstimate: &model.Estimate{
				ID:     1,
				Status: model.EstimateStatusRejected,
			},
			wantErr:      true,
			wantConflict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalled := false
			repo := &mockEstimateRepository{
				updateIfNotLockedFn: func(_ context.Context, _, _ uint64, _ UpdateEstimateInput) (*model.Estimate, error) {
					updateCalled = true
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					if tt.repoEstimate != nil {
						return tt.repoEstimate, nil
					}
					return &model.Estimate{ID: 1, Status: model.EstimateStatusDraft}, nil
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					if tt.repoErr != nil && apperrors.IsNotFound(tt.repoErr) {
						return nil, tt.repoErr
					}
					if tt.repoEstimate != nil {
						return tt.repoEstimate, nil
					}
					return &model.Estimate{ID: 1, Status: model.EstimateStatusDraft}, nil
				},
			}
			svc := NewEstimateService(repo, nil, nil, nil, nil, noopTransactor{})

			estimate, err := svc.Update(context.Background(), 1, 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
					assert.False(t, updateCalled, "repo.UpdateIfNotLocked must not be called for locked estimate")
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, estimate)
			}
		})
	}
}

func TestEstimateService_Update_TotalsOnlyPatchDerivesTotalsFromExistingItems(t *testing.T) {
	requestedSubtotal := int64(1)
	requestedTaxTotal := int64(2)
	requestedTotalAmount := int64(3)
	existing := &model.Estimate{
		ID:              1,
		ClinicID:        1,
		Status:          model.EstimateStatusDraft,
		InsuranceAmount: 700,
		DiscountAmount:  300,
		Items: []model.EstimateItem{
			{
				EstimateID:     1,
				UnitPrice:      1000,
				Quantity:       2,
				TaxType:        model.TaxTypeExcluded,
				TaxRate:        0.10,
				DiscountAmount: 200,
			},
		},
	}

	var updatedCmd UpdateEstimateInput
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
			return existing, nil
		},
		updateIfNotLockedFn: func(_ context.Context, _, _ uint64, cmd UpdateEstimateInput) (*model.Estimate, error) {
			updatedCmd = cmd
			return existing, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, nil, nil, noopTransactor{})

	_, err := svc.Update(context.Background(), 1, 1, &UpdateEstimateInput{
		Subtotal:    &requestedSubtotal,
		TaxTotal:    &requestedTaxTotal,
		TotalAmount: &requestedTotalAmount,
	})

	require.NoError(t, err)
	require.NotNil(t, updatedCmd.Subtotal)
	require.NotNil(t, updatedCmd.TaxTotal)
	require.NotNil(t, updatedCmd.TotalAmount)
	assert.Equal(t, int64(1800), *updatedCmd.Subtotal)
	assert.Equal(t, int64(180), *updatedCmd.TaxTotal)
	assert.Equal(t, int64(1980), *updatedCmd.TotalAmount)
	assert.Nil(t, updatedCmd.InsuranceAmount)
	assert.Nil(t, updatedCmd.DiscountAmount)
}

// TestEstimateService_Update_TOCTOU_LockedAfterFind は FindByID 時点では draft でも、
// UpdateIfNotLocked が Conflict を返した場合に編集を拒否することを検証する（TOCTOU 回帰）。
func TestEstimateService_Update_TOCTOU_LockedAfterFind(t *testing.T) {
	newTitle := "TOCTOU改ざん"
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
			return &model.Estimate{
				ID:     1,
				Status: model.EstimateStatusDraft,
				Title:  "旧タイトル",
			}, nil
		},
		updateIfNotLockedFn: func(_ context.Context, _, _ uint64, _ UpdateEstimateInput) (*model.Estimate, error) {
			return nil, apperrors.WrapConflict("承認済みまたは却下済みの見積書は編集できません")
		},
	}
	svc := NewEstimateService(repo, nil, nil, nil, nil, noopTransactor{})

	_, err := svc.Update(context.Background(), 1, 1, &UpdateEstimateInput{Title: &newTitle})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
}

func TestEstimateService_Delete(t *testing.T) {
	tests := []struct {
		name           string
		id             uint64
		existingStatus model.EstimateStatus
		itemCount      int64
		countErr       error
		repoErr        error
		wantErr        bool
		wantNF         bool
		wantConflict   bool
	}{
		{
			name:           "deletes estimate successfully when no items",
			id:             1,
			existingStatus: model.EstimateStatusDraft,
			itemCount:      0,
			repoErr:        nil,
			wantErr:        false,
		},
		{
			name:           "deletes sent estimate successfully when no items",
			id:             1,
			existingStatus: model.EstimateStatusSent,
			itemCount:      0,
			repoErr:        nil,
			wantErr:        false,
		},
		{
			name:           "returns conflict error when estimate has items",
			id:             2,
			existingStatus: model.EstimateStatusDraft,
			itemCount:      3,
			wantErr:        true,
			wantConflict:   true,
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
			name:           "returns error on repository delete failure",
			id:             1,
			existingStatus: model.EstimateStatusDraft,
			itemCount:      0,
			repoErr:        errors.New("db error"),
			wantErr:        true,
		},
		{
			name:           "returns conflict when estimate is approved",
			id:             3,
			existingStatus: model.EstimateStatusApproved,
			itemCount:      0,
			wantErr:        true,
			wantConflict:   true,
		},
		{
			name:           "returns conflict when estimate is rejected",
			id:             4,
			existingStatus: model.EstimateStatusRejected,
			itemCount:      0,
			wantErr:        true,
			wantConflict:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteIfNotLockedCalled := false
			repo := &mockEstimateRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
					if tt.repoErr != nil && apperrors.IsNotFound(tt.repoErr) {
						return nil, tt.repoErr
					}
					return &model.Estimate{
						ID:     id,
						Status: tt.existingStatus,
					}, nil
				},
				countItemsByEstimateID: func(_ context.Context, _ uint64) (int64, error) {
					return tt.itemCount, tt.countErr
				},
				deleteIfNotLockedFn: func(_ context.Context, _, _ uint64) error {
					deleteIfNotLockedCalled = true
					return tt.repoErr
				},
			}
			svc := NewEstimateService(repo, nil, nil, nil, nil, noopTransactor{})

			err := svc.Delete(context.Background(), 1, tt.id, nil)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
					if tt.existingStatus == model.EstimateStatusApproved || tt.existingStatus == model.EstimateStatusRejected {
						assert.False(t, deleteIfNotLockedCalled, "repo.DeleteIfNotLocked must not be called for locked estimate")
					}
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, deleteIfNotLockedCalled, "repo.DeleteIfNotLocked must be called on success")
			}
		})
	}
}

// TestEstimateService_Delete_TOCTOU_LockedAfterFind は FindByID 時点では draft でも、
// DeleteIfNotLocked が Conflict を返した場合に削除を拒否することを検証する（TOCTOU 回帰）。
func TestEstimateService_Delete_TOCTOU_LockedAfterFind(t *testing.T) {
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
			return &model.Estimate{
				ID:     1,
				Status: model.EstimateStatusDraft,
			}, nil
		},
		countItemsByEstimateID: func(_ context.Context, _ uint64) (int64, error) {
			return 0, nil
		},
		deleteIfNotLockedFn: func(_ context.Context, _, _ uint64) error {
			return apperrors.WrapConflict("承認済みまたは却下済みの見積書は削除できません")
		},
	}
	svc := NewEstimateService(repo, nil, nil, nil, nil, noopTransactor{})

	err := svc.Delete(context.Background(), 1, 1, nil)

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
}

// TestEstimateService_Delete_TOCTOU_ItemsAddedAfterCount は CountItems==0 確認後に
// 明細が入り DeleteIfNotLocked が Conflict を返した場合、削除を拒否することを検証する。
func TestEstimateService_Delete_TOCTOU_ItemsAddedAfterCount(t *testing.T) {
	deleteIfNotLockedCalled := false
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
			return &model.Estimate{
				ID:     1,
				Status: model.EstimateStatusDraft,
			}, nil
		},
		countItemsByEstimateID: func(_ context.Context, _ uint64) (int64, error) {
			return 0, nil
		},
		deleteIfNotLockedFn: func(_ context.Context, _, _ uint64) error {
			deleteIfNotLockedCalled = true
			return apperrors.WrapConflict("この見積書には明細が登録されているため削除できません")
		},
	}
	svc := NewEstimateService(repo, nil, nil, nil, nil, noopTransactor{})

	err := svc.Delete(context.Background(), 1, 1, nil)

	require.Error(t, err)
	assert.True(t, deleteIfNotLockedCalled, "repo.DeleteIfNotLocked must be called after CountItems==0")
	assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	assert.Contains(t, err.Error(), "明細")
}

// estimateAuditActions は mockAuditService.entries から Resource="estimate" の Action を収集する。
func estimateAuditActions(auditSvc *mockAuditService) []string {
	actions := make([]string, 0, len(auditSvc.entries))
	for _, e := range auditSvc.entries {
		if e != nil && e.Resource == "estimate" {
			actions = append(actions, e.Action)
		}
	}
	return actions
}

// TestEstimateService_Create_AuditLog は見積作成時に audit "create" が LogEntry で記録されることを確認する。
func TestEstimateService_Create_AuditLog(t *testing.T) {
	auditSvc := &mockAuditService{}
	const estimateID = uint64(42)
	repo := &mockEstimateRepository{
		createFn: func(_ context.Context, e *model.Estimate) error {
			e.ID = estimateID
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{
				ID:          id,
				ClinicID:    1,
				Title:       "新規見積",
				TotalAmount: 11000,
				Status:      model.EstimateStatusDraft,
				CreatedBy:   ptrU64(estimateTestCreatedByStaffID),
			}, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), auditSvc, noopTransactor{})

	created, err := svc.Create(context.Background(), 1, &CreateEstimateInput{
		Title:       "新規見積",
		Subtotal:    10000,
		TaxTotal:    1000,
		TotalAmount: 11000,
		CreatedBy:   ptrU64(estimateTestCreatedByStaffID),
	})

	assert.NoError(t, err)
	assert.NotNil(t, created)
	actions := estimateAuditActions(auditSvc)
	assert.Contains(t, actions, "create", "create 操作が audit に記録されること")
	require.NotNil(t, auditSvc.lastLogEntry)
	assert.Equal(t, "estimate", auditSvc.lastLogEntry.Resource)
	assert.Equal(t, estimateTestCreatedByStaffID, *auditSvc.lastLogEntry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, auditSvc.lastLogEntry.ActorType)
	require.NotNil(t, auditSvc.lastLogEntry.ResourceID)
	assert.Equal(t, estimateID, *auditSvc.lastLogEntry.ResourceID)
}

// TestEstimateService_Update_AuditLog は見積更新時に audit "update" が LogEntry で記録されることを確認する。
func TestEstimateService_Update_AuditLog(t *testing.T) {
	auditSvc := &mockAuditService{}
	newTitle := "更新見積"
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
			return &model.Estimate{
				ID:       10,
				ClinicID: 1,
				Status:   model.EstimateStatusDraft,
				Title:    "旧タイトル",
			}, nil
		},
		updateIfNotLockedFn: func(_ context.Context, _, _ uint64, _ UpdateEstimateInput) (*model.Estimate, error) {
			return &model.Estimate{
				ID:       10,
				ClinicID: 1,
				Status:   model.EstimateStatusDraft,
				Title:    newTitle,
			}, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, nil, auditSvc, noopTransactor{})

	actorID := uint64(7)
	updated, err := svc.Update(context.Background(), 1, 10, &UpdateEstimateInput{
		Title:   &newTitle,
		ActorID: &actorID,
	})

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	actions := estimateAuditActions(auditSvc)
	assert.Contains(t, actions, "update", "update 操作が audit に記録されること")
	assert.NotContains(t, actions, "approve")
	assert.NotContains(t, actions, "reject")
	require.NotNil(t, auditSvc.lastLogEntry)
	assert.Equal(t, "estimate", auditSvc.lastLogEntry.Resource)
	assert.Equal(t, actorID, *auditSvc.lastLogEntry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, auditSvc.lastLogEntry.ActorType)
}

// TestEstimateService_Update_ApproveAuditLog は draft|sent→approved 遷移で "approve" が追加記録されることを確認する。
func TestEstimateService_Update_ApproveAuditLog(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus model.EstimateStatus
	}{
		{name: "draft to approved", initialStatus: model.EstimateStatusDraft},
		{name: "sent to approved", initialStatus: model.EstimateStatusSent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditSvc := &mockAuditService{}
			approved := model.EstimateStatusApproved
			repo := &mockEstimateRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					return &model.Estimate{
						ID:       10,
						ClinicID: 1,
						Status:   tt.initialStatus,
					}, nil
				},
				updateIfNotLockedFn: func(_ context.Context, _, _ uint64, _ UpdateEstimateInput) (*model.Estimate, error) {
					return &model.Estimate{
						ID:       10,
						ClinicID: 1,
						Status:   model.EstimateStatusApproved,
					}, nil
				},
			}
			svc := NewEstimateService(repo, nil, nil, nil, auditSvc, noopTransactor{})

			actorID := uint64(7)
			_, err := svc.Update(context.Background(), 1, 10, &UpdateEstimateInput{
				Status:  &approved,
				ActorID: &actorID,
			})

			assert.NoError(t, err)
			actions := estimateAuditActions(auditSvc)
			assert.Contains(t, actions, "update", "status 変更でも update が記録されること")
			assert.Contains(t, actions, "approve", "draft|sent→approved で approve が追加記録されること")
		})
	}
}

// TestEstimateService_Update_RejectAuditLog は →rejected 遷移で "reject" が追加記録されることを確認する。
func TestEstimateService_Update_RejectAuditLog(t *testing.T) {
	auditSvc := &mockAuditService{}
	rejected := model.EstimateStatusRejected
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
			return &model.Estimate{
				ID:       10,
				ClinicID: 1,
				Status:   model.EstimateStatusSent,
			}, nil
		},
		updateIfNotLockedFn: func(_ context.Context, _, _ uint64, _ UpdateEstimateInput) (*model.Estimate, error) {
			return &model.Estimate{
				ID:       10,
				ClinicID: 1,
				Status:   model.EstimateStatusRejected,
			}, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, nil, auditSvc, noopTransactor{})

	actorID := uint64(7)
	_, err := svc.Update(context.Background(), 1, 10, &UpdateEstimateInput{
		Status:  &rejected,
		ActorID: &actorID,
	})

	assert.NoError(t, err)
	actions := estimateAuditActions(auditSvc)
	assert.Contains(t, actions, "update", "status 変更でも update が記録されること")
	assert.Contains(t, actions, "reject", "→rejected で reject が追加記録されること")
}

// TestEstimateService_Delete_AuditLog は見積削除時に audit "delete" が LogEntry で記録されることを確認する。
func TestEstimateService_Delete_AuditLog(t *testing.T) {
	t.Run("staff actor from ActorID", func(t *testing.T) {
		auditSvc := &mockAuditService{}
		repo := &mockEstimateRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
				return &model.Estimate{
					ID:     id,
					Status: model.EstimateStatusDraft,
				}, nil
			},
			countItemsByEstimateID: func(_ context.Context, _ uint64) (int64, error) {
				return 0, nil
			},
			deleteIfNotLockedFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
		}
		svc := NewEstimateService(repo, nil, nil, nil, auditSvc, noopTransactor{})

		actorID := uint64(9)
		err := svc.Delete(context.Background(), 1, 10, &actorID)

		assert.NoError(t, err)
		actions := estimateAuditActions(auditSvc)
		assert.Contains(t, actions, "delete", "delete 操作が audit に記録されること")
		require.NotNil(t, auditSvc.lastLogEntry)
		assert.Equal(t, "estimate", auditSvc.lastLogEntry.Resource)
		assert.Equal(t, actorID, *auditSvc.lastLogEntry.ActorID)
		assert.Equal(t, model.AuditActorTypeStaff, auditSvc.lastLogEntry.ActorType)
	})

	t.Run("system actor when ActorID is nil", func(t *testing.T) {
		auditSvc := &mockAuditService{}
		repo := &mockEstimateRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
				return &model.Estimate{
					ID:     id,
					Status: model.EstimateStatusSent,
				}, nil
			},
			countItemsByEstimateID: func(_ context.Context, _ uint64) (int64, error) {
				return 0, nil
			},
			deleteIfNotLockedFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
		}
		svc := NewEstimateService(repo, nil, nil, nil, auditSvc, noopTransactor{})

		err := svc.Delete(context.Background(), 1, 11, nil)

		assert.NoError(t, err)
		actions := estimateAuditActions(auditSvc)
		assert.Contains(t, actions, "delete", "ActorID nil でも delete audit が記録されること")
		require.NotNil(t, auditSvc.lastLogEntry)
		assert.Nil(t, auditSvc.lastLogEntry.ActorID)
		assert.Equal(t, model.AuditActorTypeSystem, auditSvc.lastLogEntry.ActorType)
	})
}

// TestEstimateService_AuditFailure_NonFatal は監査ログ書き込み失敗が Create のメイン処理を止めないことを確認する。
func TestEstimateService_AuditFailure_NonFatal(t *testing.T) {
	auditSvc := &mockAuditService{
		logEntryErr: errors.New("audit db down"),
	}
	repo := &mockEstimateRepository{
		createFn: func(_ context.Context, e *model.Estimate) error {
			e.ID = 1
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{
				ID:          id,
				ClinicID:    1,
				Title:       "新規見積",
				TotalAmount: 11000,
				Status:      model.EstimateStatusDraft,
				CreatedBy:   ptrU64(estimateTestCreatedByStaffID),
			}, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), auditSvc, noopTransactor{})

	created, err := svc.Create(context.Background(), 1, &CreateEstimateInput{
		Title:       "新規見積",
		TotalAmount: 11000,
		CreatedBy:   ptrU64(estimateTestCreatedByStaffID),
	})

	assert.NoError(t, err, "監査ログ失敗はメイン処理のエラーを返さない（best-effort）")
	assert.NotNil(t, created)
	assert.True(t, auditSvc.logEntryCalled, "audit LogEntry は呼ばれること")
}

// TestEstimateService_LockedStatus_DoesNotAudit は locked 見積の update/delete 拒否時に audit が呼ばれないことを確認する。
func TestEstimateService_LockedStatus_DoesNotAudit(t *testing.T) {
	t.Run("update on approved estimate", func(t *testing.T) {
		auditSvc := &mockAuditService{}
		newTitle := "変更不可"
		repo := &mockEstimateRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
				return &model.Estimate{
					ID:     1,
					Status: model.EstimateStatusApproved,
				}, nil
			},
		}
		svc := NewEstimateService(repo, nil, nil, nil, auditSvc, noopTransactor{})

		_, err := svc.Update(context.Background(), 1, 1, &UpdateEstimateInput{
			Title:   &newTitle,
			ActorID: ptrU64(7),
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Empty(t, estimateAuditActions(auditSvc), "locked 見積の update 拒否時は audit しない")
	})

	t.Run("delete on rejected estimate", func(t *testing.T) {
		auditSvc := &mockAuditService{}
		repo := &mockEstimateRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
				return &model.Estimate{
					ID:     id,
					Status: model.EstimateStatusRejected,
				}, nil
			},
			countItemsByEstimateID: func(_ context.Context, _ uint64) (int64, error) {
				return 0, nil
			},
		}
		svc := NewEstimateService(repo, nil, nil, nil, auditSvc, noopTransactor{})

		err := svc.Delete(context.Background(), 1, 2, ptrU64(7))

		assert.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Empty(t, estimateAuditActions(auditSvc), "locked 見積の delete 拒否時は audit しない")
	})
}

// TestEstimateService_Create_RejectsFinalizedParentMedicalRecord は確定済みカルテに紐付く
// 見積書作成が Conflict(409) で拒否され、repo.Create が呼ばれないことを検証する
// （SD-2 系ガード監査: estimates は docs/architecture/erd.md で「カルテ配下データ」に分類）。
func TestEstimateService_Create_RejectsFinalizedParentMedicalRecord(t *testing.T) {
	createCalled := false
	repo := &mockEstimateRepository{
		createFn: func(_ context.Context, _ *model.Estimate) error {
			createCalled = true
			return nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewEstimateService(repo, mrRepo, nil, &mockStaffClinicMembershipCounter{}, nil, noopTransactor{})

	in := estimateCreateBaseInput()
	in.MedicalRecordID = ptrU64(1)
	out, err := svc.Create(context.Background(), 1, in)

	assert.Error(t, err)
	assert.Nil(t, out)
	assert.True(t, apperrors.IsConflict(err), "確定済みカルテへの見積書作成は Conflict(409) であるべき: %v", err)
	assert.False(t, createCalled, "確定済みカルテに見積書が作成されてはならない")
}

// TestEstimateService_Create_AllowsNoParentMedicalRecord は medical_record_id 未指定（独立見積）の
// 作成がカルテ確定ガードの影響を受けないことを検証する（回帰: nil ガードの取りこぼし防止）。
func TestEstimateService_Create_AllowsNoParentMedicalRecord(t *testing.T) {
	createCalled := false
	repo := &mockEstimateRepository{
		createFn: func(_ context.Context, e *model.Estimate) error {
			createCalled = true
			e.ID = 1
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{ID: id}, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, &mockStaffClinicMembershipCounter{}, nil, noopTransactor{})

	out, err := svc.Create(context.Background(), 1, estimateCreateBaseInput())

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.True(t, createCalled)
}

// TestEstimateService_Update_RejectsFinalizedParentMedicalRecord は確定済みカルテに紐付く
// 見積書更新が Conflict(409) で拒否され、repo.UpdateIfNotLocked が呼ばれないことを検証する。
func TestEstimateService_Update_RejectsFinalizedParentMedicalRecord(t *testing.T) {
	updateCalled := false
	mrID := uint64(1)
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{ID: id, Status: model.EstimateStatusDraft, MedicalRecordID: &mrID}, nil
		},
		updateIfNotLockedFn: func(_ context.Context, _, _ uint64, _ UpdateEstimateInput) (*model.Estimate, error) {
			updateCalled = true
			return nil, nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewEstimateService(repo, mrRepo, nil, nil, nil, noopTransactor{})

	newTitle := "更新後タイトル"
	out, err := svc.Update(context.Background(), 1, 1, &UpdateEstimateInput{Title: &newTitle})

	assert.Error(t, err)
	assert.Nil(t, out)
	assert.True(t, apperrors.IsConflict(err), "確定済みカルテの見積書更新は Conflict(409) であるべき: %v", err)
	assert.False(t, updateCalled, "確定済みカルテの見積書は更新されてはならない")
}

// TestEstimateService_Delete_RejectsFinalizedParentMedicalRecord は確定済みカルテに紐付く
// 見積書削除が Conflict(409) で拒否され、repo.DeleteIfNotLocked が呼ばれないことを検証する。
func TestEstimateService_Delete_RejectsFinalizedParentMedicalRecord(t *testing.T) {
	deleteCalled := false
	mrID := uint64(1)
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{ID: id, Status: model.EstimateStatusDraft, MedicalRecordID: &mrID}, nil
		},
		deleteIfNotLockedFn: func(_ context.Context, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewEstimateService(repo, mrRepo, nil, nil, nil, noopTransactor{})

	err := svc.Delete(context.Background(), 1, 1, ptrU64(7))

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "確定済みカルテの見積書削除は Conflict(409) であるべき: %v", err)
	assert.False(t, deleteCalled, "確定済みカルテの見積書は削除されてはならない")
}

// ---- CreateSuccessor (TASK-012 FINAL B) ----

func TestEstimateService_CreateSuccessor_LockedOK(t *testing.T) {
	const clinicID, originalID, actorID = uint64(1), uint64(10), uint64(7)
	ownerID := uint64(5)
	mrID := uint64(3)
	original := &model.Estimate{
		ID:              originalID,
		ClinicID:        clinicID,
		EstimateNo:      "EST-1",
		MedicalRecordID: &mrID,
		Title:           "確定見積",
		OwnerID:         &ownerID,
		Status:          model.EstimateStatusApproved,
		Subtotal:        10000,
		TaxTotal:        1000,
		TotalAmount:     11000,
		InsuranceAmount: 500,
		DiscountAmount:  200,
		Comment:         "飼主向け",
		Notes:           "社内メモ",
	}

	var created *model.Estimate
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, cID, id uint64) (*model.Estimate, error) {
			assert.Equal(t, clinicID, cID)
			if id == originalID {
				return original, nil
			}
			if created != nil && id == created.ID {
				return created, nil
			}
			return nil, apperrors.WrapNotFound("estimate", fmt.Sprintf("%d", id))
		},
		allocateNextEstimateNo: func(_ context.Context, cID uint64) (string, error) {
			assert.Equal(t, clinicID, cID)
			return "EST-2", nil
		},
		createFn: func(_ context.Context, e *model.Estimate) error {
			e.ID = 99
			// copy for post-create FindByID
			cp := *e
			created = &cp
			return nil
		},
	}
	audit := &capturingAuditTxLogger{}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{}, WithEstimateAuditTx(audit))

	got, err := svc.CreateSuccessor(context.Background(), clinicID, originalID, &CreateSuccessorInput{
		Reason:  "金額訂正",
		ActorID: actorID,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, uint64(99), got.ID)
	assert.Equal(t, "EST-2", got.EstimateNo)
	assert.Equal(t, model.EstimateStatusDraft, got.Status)
	require.NotNil(t, got.SupersedesEstimateID)
	assert.Equal(t, originalID, *got.SupersedesEstimateID)
	assert.Equal(t, original.Title, got.Title)
	assert.Equal(t, original.TotalAmount, got.TotalAmount)
	assert.Equal(t, original.MedicalRecordID, got.MedicalRecordID)
	assert.Equal(t, original.OwnerID, got.OwnerID)
	require.NotNil(t, got.CreatedBy)
	assert.Equal(t, actorID, *got.CreatedBy)

	// original は不変（mock 上の参照も status のまま）
	assert.Equal(t, model.EstimateStatusApproved, original.Status)

	require.NotNil(t, audit.lastInput)
	assert.Equal(t, "supersede", audit.lastInput.Action)
	assert.Equal(t, "estimate", audit.lastInput.Resource)
	require.NotNil(t, audit.lastInput.ResourceID)
	assert.Equal(t, uint64(99), *audit.lastInput.ResourceID)
	newVal, ok := audit.lastInput.NewValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, originalID, newVal["original_id"])
	assert.Equal(t, uint64(99), newVal["successor_id"])
	assert.Equal(t, "金額訂正", newVal["reason"])
	assert.Equal(t, "EST-2", newVal["estimate_no"])
}

func TestEstimateService_CreateSuccessor_RejectsUnlocked(t *testing.T) {
	createCalled := false
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{ID: id, Status: model.EstimateStatusDraft, Title: "下書き"}, nil
		},
		createFn: func(_ context.Context, _ *model.Estimate) error {
			createCalled = true
			return nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{}, WithEstimateAuditTx(&noopAuditTxLogger{}))

	_, err := svc.CreateSuccessor(context.Background(), 1, 10, &CreateSuccessorInput{
		Reason:  "訂正",
		ActorID: 1,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "unlocked estimate must Conflict: %v", err)
	assert.False(t, createCalled)
}

func TestEstimateService_CreateSuccessor_ReasonRequired(t *testing.T) {
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{ID: id, Status: model.EstimateStatusApproved}, nil
		},
		createFn: func(_ context.Context, _ *model.Estimate) error {
			t.Fatal("Create must not be called when reason is empty")
			return nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{}, WithEstimateAuditTx(&noopAuditTxLogger{}))

	for _, reason := range []string{"", "   "} {
		_, err := svc.CreateSuccessor(context.Background(), 1, 10, &CreateSuccessorInput{
			Reason:  reason,
			ActorID: 1,
		})
		require.Error(t, err, "reason=%q", reason)
		assert.True(t, apperrors.IsInvalidInput(err), "reason=%q: %v", reason, err)
	}
}

func TestEstimateService_CreateSuccessor_AuditFailRollsBack(t *testing.T) {
	const clinicID, originalID = uint64(1), uint64(10)
	var persisted *model.Estimate
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{
				ID:          id,
				ClinicID:    clinicID,
				Status:      model.EstimateStatusRejected,
				Title:       "却下見積",
				TotalAmount: 5000,
			}, nil
		},
		allocateNextEstimateNo: func(_ context.Context, _ uint64) (string, error) {
			return "EST-9", nil
		},
		createFn: func(_ context.Context, e *model.Estimate) error {
			e.ID = 55
			cp := *e
			persisted = &cp
			return nil
		},
	}
	// TX がエラーを返したら mock 上の「永続化」を破棄して rollback を模擬する。
	tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		if err := fn(ctx); err != nil {
			persisted = nil
			return err
		}
		return nil
	}}
	audit := &capturingAuditTxLogger{err: errors.New("audit write failed")}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, tx, WithEstimateAuditTx(audit))

	_, err := svc.CreateSuccessor(context.Background(), clinicID, originalID, &CreateSuccessorInput{
		Reason:  "監査失敗注入",
		ActorID: 1,
	})
	require.Error(t, err)
	assert.Nil(t, persisted, "監査失敗時は successor が永続化されてはならない（TX rollback）")
	require.NotNil(t, audit.lastInput, "audit LogEntryTx should be attempted before fail-closed")
}

func TestEstimateService_Create_AssignsEstimateNo(t *testing.T) {
	var saved *model.Estimate
	repo := &mockEstimateRepository{
		allocateNextEstimateNo: func(_ context.Context, clinicID uint64) (string, error) {
			assert.Equal(t, uint64(1), clinicID)
			return "EST-42", nil
		},
		createFn: func(_ context.Context, e *model.Estimate) error {
			saved = e
			e.ID = 7
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			require.NotNil(t, saved)
			cp := *saved
			cp.ID = id
			return &cp, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})

	got, err := svc.Create(context.Background(), 1, &CreateEstimateInput{
		Title:     "番号採番見積",
		CreatedBy: ptrU64(estimateTestCreatedByStaffID),
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, saved)
	assert.Equal(t, "EST-42", saved.EstimateNo)
	assert.Equal(t, "EST-42", got.EstimateNo)
}

func TestEstimateService_Create_PersistsItemsInSameTx(t *testing.T) {
	var replaced []model.EstimateItem
	repo := &mockEstimateRepository{
		createFn: func(_ context.Context, e *model.Estimate) error {
			e.ID = 11
			return nil
		},
		replaceItemsFn: func(_ context.Context, clinicID, estimateID uint64, items []model.EstimateItem) error {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(11), estimateID)
			replaced = items
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{ID: id, Title: "明細付き", Items: replaced}, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})

	got, err := svc.Create(context.Background(), 1, &CreateEstimateInput{
		Title:     "明細付き",
		CreatedBy: ptrU64(estimateTestCreatedByStaffID),
		Items: []EstimateItemInput{{
			Name:      "診察料",
			UnitPrice: 1000,
			Quantity:  1,
		}},
	})
	require.NoError(t, err)
	require.Len(t, replaced, 1)
	assert.Equal(t, "診察料", replaced[0].Name)
	require.Len(t, got.Items, 1)
}

func TestEstimateService_Create_DerivesHeaderTotalsFromItems(t *testing.T) {
	var saved *model.Estimate
	repo := &mockEstimateRepository{
		createFn: func(_ context.Context, estimate *model.Estimate) error {
			estimate.ID = 12
			cp := *estimate
			saved = &cp
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			cp := *saved
			cp.ID = id
			return &cp, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})

	_, err := svc.Create(context.Background(), 1, &CreateEstimateInput{
		Title:       "明細合計",
		CreatedBy:   ptrU64(estimateTestCreatedByStaffID),
		Subtotal:    999999,
		TaxTotal:    999999,
		TotalAmount: 999999,
		Items: []EstimateItemInput{{
			Name: "処置", Category: model.ItemCategoryProcedure,
			UnitPrice: 1000, Quantity: 2, DiscountAmount: 100,
		}},
	})

	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, int64(1900), saved.Subtotal)
	assert.Equal(t, int64(190), saved.TaxTotal)
	assert.Equal(t, int64(2090), saved.TotalAmount)
}

func TestEstimateService_Create_RejectsInvalidItemValues(t *testing.T) {
	tests := []struct {
		name string
		item EstimateItemInput
	}{
		{name: "invalid category", item: EstimateItemInput{Name: "x", Category: model.ItemCategory("invalid"), Quantity: 1}},
		{name: "zero quantity", item: EstimateItemInput{Name: "x", Category: model.ItemCategoryOther, Quantity: 0}},
		{name: "negative discount amount", item: EstimateItemInput{Name: "x", Category: model.ItemCategoryOther, Quantity: 1, DiscountAmount: -1}},
		{name: "discount rate above 100", item: EstimateItemInput{Name: "x", Category: model.ItemCategoryOther, Quantity: 1, DiscountRate: 101}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createCalled := false
			repo := &mockEstimateRepository{createFn: func(_ context.Context, _ *model.Estimate) error { createCalled = true; return nil }}
			svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
			got, err := svc.Create(context.Background(), 1, &CreateEstimateInput{Title: "invalid", CreatedBy: ptrU64(estimateTestCreatedByStaffID), Items: []EstimateItemInput{tt.item}})
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input, got %v", err)
			assert.Nil(t, got)
			assert.False(t, createCalled)
		})
	}
}

func TestEstimateService_Create_FindByIDFailureAbortsCreate(t *testing.T) {
	repo := &mockEstimateRepository{
		createFn: func(_ context.Context, e *model.Estimate) error {
			e.ID = 12
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
			return nil, apperrors.WrapInternalServerError("reload failed")
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})

	got, err := svc.Create(context.Background(), 1, &CreateEstimateInput{
		Title:     "再取得失敗",
		CreatedBy: ptrU64(estimateTestCreatedByStaffID),
	})
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestEstimateService_CreateSuccessor_CopiesItems(t *testing.T) {
	const clinicID, originalID, actorID = uint64(1), uint64(10), uint64(7)
	original := &model.Estimate{
		ID:         originalID,
		ClinicID:   clinicID,
		EstimateNo: "EST-1",
		Title:      "確定見積",
		Status:     model.EstimateStatusApproved,
		Items: []model.EstimateItem{{
			ID: 3, EstimateID: originalID, Name: "処置料", UnitPrice: 2000, Quantity: 1,
			Category: model.ItemCategoryOther, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
		}},
	}
	var copied []model.EstimateItem
	var created *model.Estimate
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _ uint64, id uint64) (*model.Estimate, error) {
			if id == originalID {
				return original, nil
			}
			if created != nil && id == created.ID {
				cp := *created
				cp.Items = copied
				return &cp, nil
			}
			return nil, apperrors.WrapNotFound("estimate", fmt.Sprintf("%d", id))
		},
		createFn: func(_ context.Context, e *model.Estimate) error {
			e.ID = 99
			cp := *e
			created = &cp
			return nil
		},
		replaceItemsFn: func(_ context.Context, _, estimateID uint64, items []model.EstimateItem) error {
			assert.Equal(t, uint64(99), estimateID)
			copied = items
			return nil
		},
	}
	audit := &capturingAuditTxLogger{}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{}, WithEstimateAuditTx(audit))

	got, err := svc.CreateSuccessor(context.Background(), clinicID, originalID, &CreateSuccessorInput{
		Reason:  "明細引き継ぎ",
		ActorID: actorID,
	})
	require.NoError(t, err)
	require.Len(t, copied, 1)
	assert.Equal(t, "処置料", copied[0].Name)
	assert.Zero(t, copied[0].ID)
	require.Len(t, got.Items, 1)
	assert.Equal(t, "処置料", got.Items[0].Name)
}

func TestEstimateService_Update_RejectsInvalidItemsBeforeWrite(t *testing.T) {
	updateCalled := false
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{ID: id, Status: model.EstimateStatusDraft}, nil
		},
		updateIfNotLockedFn: func(_ context.Context, _, _ uint64, _ UpdateEstimateInput) (*model.Estimate, error) {
			updateCalled = true
			return nil, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
	items := []EstimateItemInput{{Name: "薬", Category: model.ItemCategoryMedicine, Quantity: -1}}

	got, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{Items: &items})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input, got %v", err)
	assert.Nil(t, got)
	assert.False(t, updateCalled)
}

func TestEstimateService_Update_ItemsOnlyDerivesHeaderTotals(t *testing.T) {
	var updatedCmd UpdateEstimateInput
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{ID: id, Status: model.EstimateStatusDraft}, nil
		},
		updateIfNotLockedFn: func(_ context.Context, _, id uint64, cmd UpdateEstimateInput) (*model.Estimate, error) {
			updatedCmd = cmd
			return &model.Estimate{ID: id, Status: model.EstimateStatusDraft}, nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
	items := []EstimateItemInput{{Name: "薬", Category: model.ItemCategoryMedicine, UnitPrice: 500, Quantity: 3, DiscountAmount: 200}}

	_, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{Items: &items})

	require.NoError(t, err)
	require.NotNil(t, updatedCmd.Subtotal)
	require.NotNil(t, updatedCmd.TaxTotal)
	require.NotNil(t, updatedCmd.TotalAmount)
	assert.Equal(t, int64(1300), *updatedCmd.Subtotal)
	assert.Equal(t, int64(130), *updatedCmd.TaxTotal)
	assert.Equal(t, int64(1430), *updatedCmd.TotalAmount)
}

func TestEstimateService_Update_PersistsItemsInSameTx(t *testing.T) {
	title := "明細更新"
	var replaced []model.EstimateItem
	repo := &mockEstimateRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			return &model.Estimate{
				ID:     id,
				Title:  title,
				Status: model.EstimateStatusDraft,
				Items:  replaced,
			}, nil
		},
		updateIfNotLockedFn: func(_ context.Context, _, id uint64, _ UpdateEstimateInput) (*model.Estimate, error) {
			return &model.Estimate{ID: id, Title: title, Status: model.EstimateStatusDraft}, nil
		},
		replaceItemsFn: func(_ context.Context, clinicID, estimateID uint64, items []model.EstimateItem) error {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(3), estimateID)
			replaced = items
			return nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})

	items := []EstimateItemInput{{Name: "処置料", UnitPrice: 2000, Quantity: 1}}
	got, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{
		Title: &title,
		Items: &items,
	})
	require.NoError(t, err)
	require.Len(t, replaced, 1)
	assert.Equal(t, "処置料", replaced[0].Name)
	require.Len(t, got.Items, 1)

	empty := []EstimateItemInput{}
	got, err = svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{Items: &empty})
	require.NoError(t, err)
	require.Empty(t, replaced)
	require.Empty(t, got.Items)
}
