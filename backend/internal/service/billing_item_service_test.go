package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- BillingItem モック ----

type mockBillingItemRepository struct {
	findByIDFn          func(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error)
	findByBillingIDFn   func(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error)
	createFn            func(ctx context.Context, item *model.BillingItem) error
	updateFieldsFn      func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn            func(ctx context.Context, clinicID, id uint64) error
	updateBillingTotals func(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error
}

func (m *mockBillingItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error) {
	return m.findByIDFn(ctx, clinicID, id)
}
func (m *mockBillingItemRepository) FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error) {
	return m.findByBillingIDFn(ctx, clinicID, billingID)
}
func (m *mockBillingItemRepository) Create(ctx context.Context, item *model.BillingItem) error {
	return m.createFn(ctx, item)
}
func (m *mockBillingItemRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}
func (m *mockBillingItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}
func (m *mockBillingItemRepository) UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error {
	if m.updateBillingTotals != nil {
		return m.updateBillingTotals(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount)
	}
	return nil
}

func defaultMockBillingItemRepo() *mockBillingItemRepository {
	return &mockBillingItemRepository{
		findByBillingIDFn: func(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
			return []model.BillingItem{}, nil
		},
	}
}

func defaultMockBillingRepo() *mockAccountingRepository {
	return &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return &model.Billing{ID: 10}, nil
		},
	}
}

// ---- Tests ----

func TestBillingItemService_CreateItem(t *testing.T) {
	tests := []struct {
		name          string
		input         *CreateBillingItemInput
		billingFindFn func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
		createErr     error
		wantErr       bool
		checkDefaults func(t *testing.T, item *model.BillingItem)
	}{
		{
			name: "creates item successfully with defaults",
			input: &CreateBillingItemInput{
				ClinicID:  1,
				BillingID: 10,
				Category:  string(model.ItemCategoryExamination),
				Name:      "診察料",
				UnitPrice: 3000,
				Quantity:  1,
			},
			wantErr: false,
			checkDefaults: func(t *testing.T, item *model.BillingItem) {
				assert.Equal(t, model.TaxTypeExcluded, item.TaxType)
				assert.Equal(t, 0.10, item.TaxRate)
				assert.Equal(t, model.ItemSourceManual, item.Source)
			},
		},
		{
			name: "creates item with explicit tax_type and source",
			input: &CreateBillingItemInput{
				ClinicID:  1,
				BillingID: 10,
				Category:  string(model.ItemCategoryMedicine),
				Name:      "薬剤料",
				UnitPrice: 500,
				Quantity:  2,
				TaxType:   string(model.TaxTypeIncluded),
				TaxRate:   0.08,
				Source:    string(model.ItemSourceMedicalRecord),
			},
			wantErr: false,
			checkDefaults: func(t *testing.T, item *model.BillingItem) {
				assert.Equal(t, model.TaxTypeIncluded, item.TaxType)
				assert.Equal(t, 0.08, item.TaxRate)
				assert.Equal(t, model.ItemSourceMedicalRecord, item.Source)
			},
		},
		{
			name:    "returns error for empty name",
			input:   &CreateBillingItemInput{ClinicID: 1, BillingID: 10, Name: ""},
			wantErr: true,
		},
		{
			name:    "returns error for zero billing_id",
			input:   &CreateBillingItemInput{ClinicID: 1, BillingID: 0, Name: "診察料"},
			wantErr: true,
		},
		{
			name:    "returns error for negative unit_price",
			input:   &CreateBillingItemInput{ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryExamination), Name: "診察料", UnitPrice: -1},
			wantErr: true,
		},
		{
			name: "returns error when billing not found (tenant check)",
			input: &CreateBillingItemInput{
				ClinicID:  1,
				BillingID: 10,
				Category:  string(model.ItemCategoryExamination),
				Name:      "診察料",
				UnitPrice: 3000,
			},
			billingFindFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
				return nil, apperrors.WrapNotFound("billing", "10")
			},
			wantErr: true,
		},
		{
			name: "returns error for invalid category",
			input: &CreateBillingItemInput{
				ClinicID:  1,
				BillingID: 10,
				Category:  "invalid_category",
				Name:      "診察料",
				UnitPrice: 3000,
			},
			wantErr: true,
		},
		{
			name: "returns error for invalid tax_type",
			input: &CreateBillingItemInput{
				ClinicID:  1,
				BillingID: 10,
				Category:  string(model.ItemCategoryExamination),
				Name:      "診察料",
				UnitPrice: 3000,
				TaxType:   "invalid_tax",
			},
			wantErr: true,
		},
		{
			name: "returns error for invalid source",
			input: &CreateBillingItemInput{
				ClinicID:  1,
				BillingID: 10,
				Category:  string(model.ItemCategoryExamination),
				Name:      "診察料",
				UnitPrice: 3000,
				Source:    "invalid_source",
			},
			wantErr: true,
		},
		{
			name: "propagates repository create error",
			input: &CreateBillingItemInput{
				ClinicID:  1,
				BillingID: 10,
				Category:  string(model.ItemCategoryExamination),
				Name:      "診察料",
				UnitPrice: 3000,
			},
			createErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := defaultMockBillingItemRepo()
			repo.createFn = func(_ context.Context, item *model.BillingItem) error {
				item.ID = 1
				return tt.createErr
			}
			billingRepo := defaultMockBillingRepo()
			if tt.billingFindFn != nil {
				billingRepo.findByIDFn = tt.billingFindFn
			}
			svc := NewBillingItemService(repo, billingRepo)
			result, err := svc.CreateItem(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Name, result.Name)
				if tt.checkDefaults != nil {
					tt.checkDefaults(t, result)
				}
			}
		})
	}
}

func TestBillingItemService_UpdateItem(t *testing.T) {
	existingItem := &model.BillingItem{ID: 1, BillingID: 10, UnitPrice: 3000}
	newPrice := int64(5000)
	negPrice := int64(-1)

	tests := []struct {
		name         string
		input        *UpdateBillingItemInput
		findErr      error
		updateErr    error
		findAfterErr error
		wantErr      bool
	}{
		{
			name:    "updates item successfully",
			input:   &UpdateBillingItemInput{UnitPrice: &newPrice},
			wantErr: false,
		},
		{
			name:    "returns existing item when no fields provided",
			input:   &UpdateBillingItemInput{},
			wantErr: false,
		},
		{
			name:    "returns error for negative unit_price",
			input:   &UpdateBillingItemInput{UnitPrice: &negPrice},
			wantErr: true,
		},
		{
			name:    "propagates FindByID error",
			input:   &UpdateBillingItemInput{UnitPrice: &newPrice},
			findErr: apperrors.WrapNotFound("billing_item", "1"),
			wantErr: true,
		},
		{
			name:      "propagates UpdateFields error",
			input:     &UpdateBillingItemInput{UnitPrice: &newPrice},
			updateErr: errors.New("update failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			repo := defaultMockBillingItemRepo()
			repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
				callCount++
				if callCount == 1 && tt.findErr != nil {
					return nil, tt.findErr
				}
				return existingItem, nil
			}
			repo.updateFieldsFn = func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return tt.updateErr
			}
			svc := NewBillingItemService(repo, defaultMockBillingRepo())
			result, err := svc.UpdateItem(context.Background(), 1, 1, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestBillingItemService_DeleteItem(t *testing.T) {
	existingItem := &model.BillingItem{ID: 1, BillingID: 10}

	tests := []struct {
		name      string
		findErr   error
		deleteErr error
		wantErr   bool
	}{
		{
			name:    "deletes item successfully",
			wantErr: false,
		},
		{
			name:    "propagates FindByID error",
			findErr: apperrors.WrapNotFound("billing_item", "1"),
			wantErr: true,
		},
		{
			name:      "propagates delete error",
			deleteErr: errors.New("delete failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := defaultMockBillingItemRepo()
			repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
				if tt.findErr != nil {
					return nil, tt.findErr
				}
				return existingItem, nil
			}
			repo.deleteFn = func(_ context.Context, _, _ uint64) error {
				return tt.deleteErr
			}
			svc := NewBillingItemService(repo, defaultMockBillingRepo())
			err := svc.DeleteItem(context.Background(), 1, 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
