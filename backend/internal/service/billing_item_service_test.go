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

// ---- BillingItem モック ----

type mockBillingItemRepository struct {
	findByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error)
	findByBillingIDFn             func(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error)
	createFn                      func(ctx context.Context, item *model.BillingItem) error
	updateFieldsFn                func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
	updateBillingTotals           func(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error
	hasItemByOwnerSinceFn         func(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
	hasFoodPurchaseByOwnerSinceFn func(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
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
func (m *mockBillingItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
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
func (m *mockBillingItemRepository) HasItemByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	if m.hasItemByOwnerSinceFn != nil {
		return m.hasItemByOwnerSinceFn(ctx, clinicID, ownerID, since, names)
	}
	return false, nil
}
func (m *mockBillingItemRepository) HasFoodPurchaseByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	if m.hasFoodPurchaseByOwnerSinceFn != nil {
		return m.hasFoodPurchaseByOwnerSinceFn(ctx, clinicID, ownerID, since, names)
	}
	return false, nil
}
func (m *mockBillingItemRepository) FindOwnersByCategoryPurchaseDate(_ context.Context, _ uint64, _ string, _ time.Time) ([]uint64, error) {
	return nil, nil
}

func defaultMockBillingItemRepo() *mockBillingItemRepository {
	return &mockBillingItemRepository{
		findByBillingIDFn: func(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
			return []model.BillingItem{}, nil
		},
	}
}

type mockBillingItemRepositoryWithTrimming struct {
	*mockBillingItemRepository
	findUnbilledTrimmingItemsByPetIDFn func(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error)
}

func (m *mockBillingItemRepositoryWithTrimming) FindUnbilledTrimmingItemsByPetID(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error) {
	if m.findUnbilledTrimmingItemsByPetIDFn != nil {
		return m.findUnbilledTrimmingItemsByPetIDFn(ctx, clinicID, petID)
	}
	return nil, nil
}

type mockTreatmentRepositoryForBilling struct {
	findUnbilledByPetIDFn func(ctx context.Context, clinicID, petID uint64) ([]model.Treatment, error)
}

func (m *mockTreatmentRepositoryForBilling) FindUnbilledByPetID(ctx context.Context, clinicID, petID uint64) ([]model.Treatment, error) {
	if m.findUnbilledByPetIDFn != nil {
		return m.findUnbilledByPetIDFn(ctx, clinicID, petID)
	}
	return nil, nil
}
func (m *mockTreatmentRepositoryForBilling) FindByMedicalRecordID(_ context.Context, _, _ uint64) ([]model.Treatment, error) {
	return nil, nil
}
func (m *mockTreatmentRepositoryForBilling) FindByID(_ context.Context, _, _ uint64) (*model.Treatment, error) {
	return nil, nil
}
func (m *mockTreatmentRepositoryForBilling) Create(_ context.Context, _ *model.Treatment) error {
	return nil
}
func (m *mockTreatmentRepositoryForBilling) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mockTreatmentRepositoryForBilling) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockTreatmentRepositoryForBilling) BulkUpdateSortOrder(_ context.Context, _ []repository.TreatmentSortUpdate) error {
	return nil
}

func defaultMockTreatmentRepo() *mockTreatmentRepositoryForBilling {
	return &mockTreatmentRepositoryForBilling{}
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
	treatmentID := uint64(100)
	appointmentID := uint64(200)
	trimmingCourseID := uint64(300)
	trimmingOptionID := uint64(400)

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
				ClinicID:         1,
				BillingID:        10,
				Category:         string(model.ItemCategoryMedicine),
				Name:             "薬剤料",
				UnitPrice:        500,
				Quantity:         2,
				TaxType:          string(model.TaxTypeIncluded),
				TaxRate:          0.08,
				Source:           string(model.ItemSourceMedicalRecord),
				TreatmentID:      &treatmentID,
				AppointmentID:    &appointmentID,
				TrimmingCourseID: &trimmingCourseID,
				TrimmingOptionID: &trimmingOptionID,
			},
			wantErr: false,
			checkDefaults: func(t *testing.T, item *model.BillingItem) {
				assert.Equal(t, model.TaxTypeIncluded, item.TaxType)
				assert.Equal(t, 0.08, item.TaxRate)
				assert.Equal(t, model.ItemSourceMedicalRecord, item.Source)
				assert.Equal(t, &treatmentID, item.TreatmentID)
				assert.Equal(t, &appointmentID, item.AppointmentID)
				assert.Equal(t, &trimmingCourseID, item.TrimmingCourseID)
				assert.Equal(t, &trimmingOptionID, item.TrimmingOptionID)
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
			svc := NewBillingItemService(repo, billingRepo, defaultMockTreatmentRepo())
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
			name:      "propagates Update error",
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
			svc := NewBillingItemService(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo())
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
			svc := NewBillingItemService(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo())
			err := svc.DeleteItem(context.Background(), 1, 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBillingItemService_GetUnbilledItems_IncludesMedicalAndTrimming(t *testing.T) {
	appointmentID := uint64(30)
	repo := &mockBillingItemRepositoryWithTrimming{
		mockBillingItemRepository: defaultMockBillingItemRepo(),
		findUnbilledTrimmingItemsByPetIDFn: func(_ context.Context, clinicID, petID uint64) ([]model.BillingItem, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(20), petID)
			return []model.BillingItem{
				{
					ID:            30000000001,
					Category:      model.ItemCategoryTrimming,
					Name:          "シャンプーコース",
					UnitPrice:     5000,
					Quantity:      1,
					TaxType:       model.TaxTypeExcluded,
					TaxRate:       0.10,
					Source:        model.ItemSourceTrimming,
					AppointmentID: &appointmentID,
				},
			}, nil
		},
	}
	treatmentRepo := &mockTreatmentRepositoryForBilling{
		findUnbilledByPetIDFn: func(_ context.Context, clinicID, petID uint64) ([]model.Treatment, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(20), petID)
			return []model.Treatment{
				{
					ID:          10,
					ItemType:    model.TreatmentItemTypeProcedure,
					Content:     "処置",
					UnitPrice:   1000,
					Quantity:    1,
					IsInsurance: true,
					SortOrder:   2,
				},
			}, nil
		},
	}
	svc := NewBillingItemService(repo, defaultMockBillingRepo(), treatmentRepo)

	items, err := svc.GetUnbilledItems(context.Background(), 1, 20)

	assert.NoError(t, err)
	if assert.Len(t, items, 2) {
		assert.Equal(t, model.ItemSourceMedicalRecord, items[0].Source)
		assert.Equal(t, model.ItemCategoryProcedure, items[0].Category)
		assert.Equal(t, ptrUint64(10), items[0].TreatmentID)
		assert.Equal(t, model.ItemSourceTrimming, items[1].Source)
		assert.Equal(t, model.ItemCategoryTrimming, items[1].Category)
		assert.Equal(t, &appointmentID, items[1].AppointmentID)
	}
}
