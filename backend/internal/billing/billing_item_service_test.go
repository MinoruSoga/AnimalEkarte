package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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
	validateCreateReferencesFn    func(ctx context.Context, clinicID, billingID uint64, merchandiseItemID, treatmentID, appointmentID, trimmingCourseID, trimmingOptionID *uint64) (model.ItemCategory, error)
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
func (m *mockBillingItemRepository) FindUnbilledTrimmingItemsByPetID(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
	return nil, nil
}
func (m *mockBillingItemRepository) FindUnbilledVaccinationItemsByPetID(_ context.Context, _, _ uint64) ([]model.BillingItem, int, error) {
	return nil, 0, nil
}
func (m *mockBillingItemRepository) CountNonAccountingTrimmingByPetAndDate(_ context.Context, _, _ uint64, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *mockBillingItemRepository) ValidateCreateReferences(ctx context.Context, clinicID, billingID uint64, merchandiseItemID, treatmentID, appointmentID, trimmingCourseID, trimmingOptionID *uint64) (model.ItemCategory, error) {
	if m.validateCreateReferencesFn == nil {
		return "", nil
	}
	return m.validateCreateReferencesFn(ctx, clinicID, billingID, merchandiseItemID, treatmentID, appointmentID, trimmingCourseID, trimmingOptionID)
}
func (m *mockBillingItemRepository) ValidateVaccinationCreateReference(
	_ context.Context,
	_, _, _ uint64,
) (*vaccinationBillingValues, error) {
	return nil, apperrors.WrapInternalServerError("unexpected vaccination validation in mock")
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
	countFinalizedFn      func(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error)
}

func (m *mockTreatmentRepositoryForBilling) FindUnbilledByPetID(ctx context.Context, clinicID, petID uint64) ([]model.Treatment, error) {
	if m.findUnbilledByPetIDFn != nil {
		return m.findUnbilledByPetIDFn(ctx, clinicID, petID)
	}
	return nil, nil
}
func (m *mockTreatmentRepositoryForBilling) CountFinalizedUnconfirmedByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error) {
	if m.countFinalizedFn != nil {
		return m.countFinalizedFn(ctx, clinicID, petID, date)
	}
	return 0, nil
}

// #77: 同日同ペットの未会計対象化サマリ(診察 count)を service が返すこと。
func TestBillingItemService_GetUngroupedSameDaySummary(t *testing.T) {
	treatmentRepo := &mockTreatmentRepositoryForBilling{
		countFinalizedFn: func(_ context.Context, _, _ uint64, _ time.Time) (int64, error) { return 2, nil },
	}
	svc := NewBillingItemServiceWithCampaign(defaultMockBillingItemRepo(), defaultMockBillingRepo(), treatmentRepo, &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)

	summary, err := svc.GetUngroupedSameDaySummary(context.Background(), 1, 20, time.Now())

	assert.NoError(t, err)
	assert.Equal(t, int64(2), summary.MedicalRecordCount)
	// defaultMockBillingItemRepo は ungroupedTrimmingCounter 未実装のため trimming は 0(型アサーション skip)
	assert.Equal(t, int64(0), summary.TrimmingCount)
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
	merchandiseItemID := uint64(50)
	treatmentID := uint64(100)
	appointmentID := uint64(200)
	trimmingCourseID := uint64(300)
	trimmingOptionID := uint64(400)
	foodCategory := model.ItemCategoryFood

	tests := []struct {
		name                 string
		input                *CreateBillingItemInput
		billingFindFn        func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
		resolvedCategory     model.ItemCategory
		wantCampaignCategory *model.ItemCategory
		createErr            error
		wantErr              bool
		checkDefaults        func(t *testing.T, item *model.BillingItem)
	}{
		{
			name:    "returns error for nil input",
			input:   nil,
			wantErr: true,
		},
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
				assert.Equal(t, model.ItemCategoryExamination, item.Category)
			},
		},
		{
			name: "creates item with explicit tax_type and source",
			input: &CreateBillingItemInput{
				ClinicID:          1,
				BillingID:         10,
				Category:          string(model.ItemCategoryMedicine),
				Name:              "薬剤料",
				UnitPrice:         500,
				Quantity:          2,
				TaxType:           string(model.TaxTypeIncluded),
				TaxRate:           0.08,
				Source:            string(model.ItemSourceMedicalRecord),
				MerchandiseItemID: &merchandiseItemID,
				TreatmentID:       &treatmentID,
				AppointmentID:     &appointmentID,
				TrimmingCourseID:  &trimmingCourseID,
				TrimmingOptionID:  &trimmingOptionID,
			},
			resolvedCategory:     model.ItemCategoryFood,
			wantCampaignCategory: &foodCategory,
			wantErr:              false,
			checkDefaults: func(t *testing.T, item *model.BillingItem) {
				assert.Equal(t, model.TaxTypeIncluded, item.TaxType)
				assert.Equal(t, 0.08, item.TaxRate)
				assert.Equal(t, model.ItemSourceMedicalRecord, item.Source)
				assert.Equal(t, &merchandiseItemID, item.MerchandiseItemID)
				assert.Equal(t, &treatmentID, item.TreatmentID)
				assert.Equal(t, &appointmentID, item.AppointmentID)
				assert.Equal(t, &trimmingCourseID, item.TrimmingCourseID)
				assert.Equal(t, &trimmingOptionID, item.TrimmingOptionID)
				assert.Equal(t, model.ItemCategoryFood, item.Category)
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
			repo.validateCreateReferencesFn = func(_ context.Context, _, _ uint64, _, _, _, _, _ *uint64) (model.ItemCategory, error) {
				return tt.resolvedCategory, nil
			}
			repo.createFn = func(_ context.Context, item *model.BillingItem) error {
				item.ID = 1
				return tt.createErr
			}
			billingRepo := defaultMockBillingRepo()
			if tt.billingFindFn != nil {
				billingRepo.findByIDFn = tt.billingFindFn
			}
			var gotCampaignCategory model.ItemCategory
			var campaignRepo CampaignRepository
			if tt.wantCampaignCategory != nil {
				campaignRepo = &mockCampaignRepository{
					findApplicableForItemFn: func(_ context.Context, _ uint64, _ time.Time, category model.ItemCategory, _ *uint64) (*model.Campaign, error) {
						gotCampaignCategory = category
						return nil, nil
					},
				}
			}
			svc := NewBillingItemServiceWithCampaign(repo, billingRepo, defaultMockTreatmentRepo(), &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), campaignRepo, nil)
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
				if tt.wantCampaignCategory != nil {
					assert.Equal(t, *tt.wantCampaignCategory, gotCampaignCategory)
					assert.Equal(t, result.Category, gotCampaignCategory)
					assert.NotEqual(t, model.ItemCategory(tt.input.Category), gotCampaignCategory)
				}
			}
		})
	}
}

func TestBillingItemService_CreateItem_FailsClosedWhenTransactionalReferenceValidationFails(t *testing.T) {
	treatmentID := uint64(100)
	appointmentID := uint64(200)
	created := false
	validated := false
	repo := defaultMockBillingItemRepo()
	repo.validateCreateReferencesFn = func(_ context.Context, clinicID, billingID uint64, _, gotTreatmentID, gotAppointmentID, _, _ *uint64) (model.ItemCategory, error) {
		validated = true
		assert.Equal(t, uint64(1), clinicID)
		assert.Equal(t, uint64(10), billingID)
		assert.Equal(t, &treatmentID, gotTreatmentID)
		assert.Equal(t, &appointmentID, gotAppointmentID)
		return "", apperrors.WrapNotFound("billing_item_reference", "cross-clinic")
	}
	repo.createFn = func(_ context.Context, _ *model.BillingItem) error {
		created = true
		return nil
	}
	svc := NewBillingItemServiceWithCampaign(
		repo,
		defaultMockBillingRepo(),
		defaultMockTreatmentRepo(),
		&mockTransactor{},
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		nil,
		nil,
	)

	item, err := svc.CreateItem(context.Background(), &CreateBillingItemInput{
		ClinicID:      1,
		BillingID:     10,
		Category:      string(model.ItemCategoryMedicine),
		Name:          "薬剤料",
		UnitPrice:     500,
		Quantity:      1,
		Source:        string(model.ItemSourceMedicalRecord),
		TreatmentID:   &treatmentID,
		AppointmentID: &appointmentID,
	})

	assert.Error(t, err)
	assert.Nil(t, item)
	assert.True(t, validated, "request-derived references must be validated inside the write transaction")
	assert.False(t, created, "invalid references must be rejected before persistence")
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
			svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
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
			svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
			err := svc.DeleteItem(context.Background(), 1, 1, nil)
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
					ID:              10,
					MedicalRecordID: 55,
					ItemType:        model.TreatmentItemTypeProcedure,
					Content:         "処置",
					UnitPrice:       1000,
					Quantity:        1,
					IsInsurance:     true,
					SortOrder:       2,
					Procedure:       &model.Procedure{IsSurgery: true},
				},
			}, nil
		},
	}
	svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), treatmentRepo, &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)

	items, err := svc.GetUnbilledItems(context.Background(), 1, 20)

	assert.NoError(t, err)
	if assert.Len(t, items, 2) {
		assert.Equal(t, model.ItemSourceMedicalRecord, items[0].Source)
		assert.Equal(t, model.ItemCategorySurgery, items[0].Category)
		assert.Equal(t, ptrUint64(10), items[0].TreatmentID)
		// BUG-011: unbilled treatment 候補は親 medical_record_id を返す
		assert.Equal(t, ptrUint64(55), items[0].MedicalRecordID)
		assert.Equal(t, model.ItemSourceTrimming, items[1].Source)
		assert.Equal(t, model.ItemCategoryTrimming, items[1].Category)
		assert.Equal(t, &appointmentID, items[1].AppointmentID)
	}
}

// ---- mock: CampaignRepository / OwnerRepository (billing_item_service 用) ----
// mockCampaignRepository は campaign_service_test.go、mockOwnerRepository は owner_service_test.go に既存定義。

// mockBillingItemRepositoryWithTrimmingCounter は ungroupedTrimmingCounter を実装する追加モック。
type mockBillingItemRepositoryWithTrimmingCounter struct {
	*mockBillingItemRepository
	countNonAccountingTrimmingByPetAndDateFn func(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error)
}

func (m *mockBillingItemRepositoryWithTrimmingCounter) CountNonAccountingTrimmingByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error) {
	if m.countNonAccountingTrimmingByPetAndDateFn != nil {
		return m.countNonAccountingTrimmingByPetAndDateFn(ctx, clinicID, petID, date)
	}
	return 0, nil
}

// ---- buildBillingItemUpdate (pure function) ----

func TestBuildBillingItemUpdate(t *testing.T) {
	taxType := model.TaxTypeIncluded

	tests := []struct {
		name    string
		input   *UpdateBillingItemInput
		wantLen int
	}{
		{
			name:    "no fields set returns empty map",
			input:   &UpdateBillingItemInput{},
			wantLen: 0,
		},
		{
			name: "all fields set returns all keys",
			input: &UpdateBillingItemInput{
				UnitPrice:             ptrInt64(1000),
				Quantity:              ptrFloat64(2),
				DiscountRate:          ptrFloat64(0.1),
				DiscountAmount:        ptrInt64(100),
				TaxType:               &taxType,
				TaxRate:               ptrFloat64(0.08),
				IsInsuranceApplicable: ptrBool(true),
			},
			wantLen: 7,
		},
		{
			name: "single field set returns single key",
			input: &UpdateBillingItemInput{
				UnitPrice: ptrInt64(500),
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := buildBillingItemUpdate(tt.input)
			assert.Len(t, fields, tt.wantLen)
		})
	}
}

// ---- treatmentTypeToItemCategory (resolver delegation) ----

func TestTreatmentTypeToItemCategory(t *testing.T) {
	tests := []struct {
		name string
		in   *model.Treatment
		want model.ItemCategory
	}{
		{
			"consultation maps to examination",
			&model.Treatment{ItemType: model.TreatmentItemTypeConsultation},
			model.ItemCategoryExamination,
		},
		{
			"procedure without loaded master maps to procedure",
			&model.Treatment{ItemType: model.TreatmentItemTypeProcedure},
			model.ItemCategoryProcedure,
		},
		{
			"surgical procedure maps to surgery",
			&model.Treatment{
				ItemType:  model.TreatmentItemTypeProcedure,
				Procedure: &model.Procedure{IsSurgery: true},
			},
			model.ItemCategorySurgery,
		},
		{
			"medicine maps to medicine",
			&model.Treatment{ItemType: model.TreatmentItemTypeMedicine},
			model.ItemCategoryMedicine,
		},
		{
			"other maps to other",
			&model.Treatment{ItemType: model.TreatmentItemTypeOther},
			model.ItemCategoryOther,
		},
		{
			"unknown type falls back to other",
			&model.Treatment{ItemType: model.TreatmentItemType("unknown")},
			model.ItemCategoryOther,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, treatmentTypeToItemCategory(tt.in))
		})
	}
}

// ---- resolveOwnerDiscountRate ----

func TestBillingItemService_ResolveOwnerDiscountRate(t *testing.T) {
	ownerID := uint64(5)

	tests := []struct {
		name      string
		ownerID   *uint64
		ownerRepo billingOwnerReader // interface type: keeps the "nil ownerRepo" case a true nil interface, not a typed-nil *mockOwnerRepository
		want      float64
	}{
		{
			name:    "nil ownerID returns zero",
			ownerID: nil,
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					t.Fatal("FindByID should not be called when ownerID is nil")
					return nil, nil
				},
			},
			want: 0,
		},
		{
			name:      "nil ownerRepo returns zero",
			ownerID:   &ownerID,
			ownerRepo: nil,
			want:      0,
		},
		{
			name:    "FindByID error returns zero",
			ownerID: &ownerID,
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return nil, errors.New("not found")
				},
			},
			want: 0,
		},
		{
			name:    "nil owner returns zero",
			ownerID: &ownerID,
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return nil, nil
				},
			},
			want: 0,
		},
		{
			name:    "returns owner discount rate",
			ownerID: &ownerID,
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: ownerID, DiscountRate: 15}, nil
				},
			},
			want: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &billingItemService{ownerRepo: tt.ownerRepo}
			got := svc.resolveOwnerDiscountRate(context.Background(), 1, tt.ownerID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- resolveAutoDiscount ----

func TestBillingItemService_ResolveAutoDiscount(t *testing.T) {
	ownerID := uint64(7)

	tests := []struct {
		name         string
		campaignRepo CampaignRepository // interface type: keeps the "nil campaignRepo" case a true nil interface, not a typed-nil *mockCampaignRepository
		billingRepo  *mockAccountingRepository
		ownerRepo    *mockOwnerRepository
		input        *CreateBillingItemInput
		want         int64
	}{
		{
			name:         "nil campaignRepo returns zero",
			campaignRepo: nil,
			input:        &CreateBillingItemInput{ClinicID: 1, BillingID: 10, UnitPrice: 1000, Quantity: 1},
			want:         0,
		},
		{
			name:         "billing not found returns zero",
			campaignRepo: &mockCampaignRepository{},
			billingRepo: &mockAccountingRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
					return nil, errors.New("not found")
				},
			},
			input: &CreateBillingItemInput{ClinicID: 1, BillingID: 10, UnitPrice: 1000, Quantity: 1},
			want:  0,
		},
		{
			name: "campaign lookup error falls back to owner-rate-only (best-effort)",
			campaignRepo: &mockCampaignRepository{
				findApplicableForItemFn: func(_ context.Context, _ uint64, _ time.Time, _ model.ItemCategory, _ *uint64) (*model.Campaign, error) {
					return nil, errors.New("lookup failed")
				},
			},
			billingRepo: &mockAccountingRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
					return &model.Billing{ID: 10, OwnerID: &ownerID, ScheduledDate: time.Now()}, nil
				},
			},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: ownerID, DiscountRate: 10}, nil
				},
			},
			input: &CreateBillingItemInput{ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryExamination), UnitPrice: 1000, Quantity: 1},
			want:  100, // 1000 * 10%
		},
		{
			name: "applies campaign discount when higher than owner rate",
			campaignRepo: &mockCampaignRepository{
				findApplicableForItemFn: func(_ context.Context, _ uint64, _ time.Time, _ model.ItemCategory, _ *uint64) (*model.Campaign, error) {
					return &model.Campaign{ID: 1, DiscountType: model.CampaignDiscountTypeRate, DiscountValue: 30}, nil
				},
			},
			billingRepo: &mockAccountingRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
					return &model.Billing{ID: 10, OwnerID: &ownerID, ScheduledDate: time.Now()}, nil
				},
			},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: ownerID, DiscountRate: 10}, nil
				},
			},
			input: &CreateBillingItemInput{ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryExamination), UnitPrice: 1000, Quantity: 1},
			want:  300, // 1000 * 30%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &billingItemService{campaignRepo: tt.campaignRepo, billingRepo: tt.billingRepo, ownerRepo: tt.ownerRepo}
			got := svc.resolveAutoDiscount(
				context.Background(),
				tt.input,
				model.ItemCategory(tt.input.Category),
				tt.input.UnitPrice,
				tt.input.Quantity,
			)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- NewBillingItemServiceWithCampaign ----

func TestNewBillingItemServiceWithCampaign_AppliesAutoDiscount(t *testing.T) {
	ownerID := uint64(9)
	repo := defaultMockBillingItemRepo()
	repo.createFn = func(_ context.Context, item *model.BillingItem) error {
		item.ID = 1
		return nil
	}
	billingRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return &model.Billing{ID: 10, OwnerID: &ownerID, ScheduledDate: time.Now()}, nil
		},
	}
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: ownerID, DiscountRate: 20}, nil
		},
	}
	campaignRepo := &mockCampaignRepository{}

	svc := NewBillingItemServiceWithCampaign(repo, billingRepo, defaultMockTreatmentRepo(), &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), campaignRepo, ownerRepo)

	result, err := svc.CreateItem(context.Background(), &CreateBillingItemInput{
		ClinicID:  1,
		BillingID: 10,
		Category:  string(model.ItemCategoryExamination),
		Name:      "診察料",
		UnitPrice: 1000,
		Quantity:  1,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(200), result.DiscountAmount) // 1000 * 20% (owner rate, no campaign configured)
}

// ---- GetDiscountSuggestions ----

func TestBillingItemService_GetDiscountSuggestions(t *testing.T) {
	ownerID := uint64(3)

	tests := []struct {
		name          string
		itemFindFn    func(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error)
		billingFindFn func(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
		campaignRepo  CampaignRepository // interface type: keeps the "no campaignRepo" case a true nil interface, not a typed-nil *mockCampaignRepository
		ownerRepo     *mockOwnerRepository
		wantErr       bool
		wantLen       int
	}{
		{
			name: "returns error when item not found",
			itemFindFn: func(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
				return nil, errors.New("not found")
			},
			wantErr: true,
		},
		{
			name: "returns error when billing not found",
			billingFindFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
				return nil, errors.New("not found")
			},
			wantErr: true,
		},
		{
			name: "no campaignRepo returns owner-only suggestion",
			billingFindFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
				return &model.Billing{ID: 10, OwnerID: &ownerID, ScheduledDate: time.Now()}, nil
			},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: ownerID, DiscountRate: 10}, nil
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "campaign lookup error is best-effort and still returns owner suggestion",
			billingFindFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
				return &model.Billing{ID: 10, OwnerID: &ownerID, ScheduledDate: time.Now()}, nil
			},
			campaignRepo: &mockCampaignRepository{
				findAllApplicableForItemFn: func(_ context.Context, _ uint64, _ time.Time, _ model.ItemCategory, _ *uint64) ([]*model.Campaign, error) {
					return nil, errors.New("lookup failed")
				},
			},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: ownerID, DiscountRate: 10}, nil
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "returns both campaign and owner suggestions",
			billingFindFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
				return &model.Billing{ID: 10, OwnerID: &ownerID, ScheduledDate: time.Now()}, nil
			},
			campaignRepo: &mockCampaignRepository{
				findAllApplicableForItemFn: func(_ context.Context, _ uint64, _ time.Time, _ model.ItemCategory, _ *uint64) ([]*model.Campaign, error) {
					return []*model.Campaign{{ID: 1, Name: "夏季キャンペーン", DiscountType: model.CampaignDiscountTypeRate, DiscountValue: 20}}, nil
				},
			},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: ownerID, DiscountRate: 10}, nil
				},
			},
			wantErr: false,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := defaultMockBillingItemRepo()
			repo.findByIDFn = tt.itemFindFn
			if repo.findByIDFn == nil {
				repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
					return &model.BillingItem{ID: 1, BillingID: 10, UnitPrice: 1000, Quantity: 1, Category: model.ItemCategoryExamination}, nil
				}
			}
			billingRepo := defaultMockBillingRepo()
			if tt.billingFindFn != nil {
				billingRepo.findByIDFn = tt.billingFindFn
			}
			svc := &billingItemService{repo: repo, billingRepo: billingRepo, campaignRepo: tt.campaignRepo, ownerRepo: tt.ownerRepo}

			suggestions, err := svc.GetDiscountSuggestions(context.Background(), 1, 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, suggestions, tt.wantLen)
			}
		})
	}
}

func TestBillingItemService_GetDiscountSuggestions_PassesMerchandiseItemID(t *testing.T) {
	merchandiseItemID := uint64(77)
	repo := defaultMockBillingItemRepo()
	repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
		return &model.BillingItem{
			ID:                1,
			BillingID:         10,
			UnitPrice:         1000,
			Quantity:          1,
			Category:          model.ItemCategoryGoods,
			MerchandiseItemID: &merchandiseItemID,
		}, nil
	}
	campaignRepo := &mockCampaignRepository{
		findAllApplicableForItemFn: func(_ context.Context, _ uint64, _ time.Time, _ model.ItemCategory, got *uint64) ([]*model.Campaign, error) {
			assert.Equal(t, &merchandiseItemID, got)
			return nil, nil
		},
	}
	svc := &billingItemService{
		repo:         repo,
		billingRepo:  defaultMockBillingRepo(),
		campaignRepo: campaignRepo,
	}

	_, err := svc.GetDiscountSuggestions(context.Background(), 1, 1)

	assert.NoError(t, err)
}

// ---- CreateItem: additional validation and transaction-error branches ----

func TestBillingItemService_CreateItem_QuantityValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateBillingItemInput
		wantErr bool
	}{
		{
			name: "returns error for zero quantity",
			input: &CreateBillingItemInput{
				ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryExamination), Name: "診察料", UnitPrice: 3000, Quantity: 0,
			},
			wantErr: true,
		},
		{
			name: "returns error for negative quantity",
			input: &CreateBillingItemInput{
				ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryExamination), Name: "診察料", UnitPrice: 3000, Quantity: -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := defaultMockBillingItemRepo()
			svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
			result, err := svc.CreateItem(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBillingItemService_CreateItem_TransactionErrors(t *testing.T) {
	validInput := &CreateBillingItemInput{
		ClinicID: 1, BillingID: 10, Category: string(model.ItemCategoryExamination), Name: "診察料", UnitPrice: 3000, Quantity: 1,
	}

	t.Run("propagates recalculateTotals error", func(t *testing.T) {
		repo := defaultMockBillingItemRepo()
		repo.createFn = func(_ context.Context, item *model.BillingItem) error {
			item.ID = 1
			return nil
		}
		repo.findByBillingIDFn = func(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
			return nil, errors.New("find failed")
		}
		svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		result, err := svc.CreateItem(context.Background(), validInput)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("propagates transactor error", func(t *testing.T) {
		repo := defaultMockBillingItemRepo()
		repo.createFn = func(_ context.Context, item *model.BillingItem) error {
			item.ID = 1
			return nil
		}
		svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{withTxErr: errors.New("tx failed")}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		result, err := svc.CreateItem(context.Background(), validInput)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// ---- UpdateItem: additional validation and transaction-error branches ----

func TestBillingItemService_UpdateItem_QuantityValidation(t *testing.T) {
	existingItem := &model.BillingItem{ID: 1, BillingID: 10, UnitPrice: 3000}
	zeroQty := 0.0
	negQty := -1.0

	tests := []struct {
		name  string
		input *UpdateBillingItemInput
	}{
		{name: "returns error for zero quantity", input: &UpdateBillingItemInput{Quantity: &zeroQty}},
		{name: "returns error for negative quantity", input: &UpdateBillingItemInput{Quantity: &negQty}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := defaultMockBillingItemRepo()
			repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
				return existingItem, nil
			}
			svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
			result, err := svc.UpdateItem(context.Background(), 1, 1, tt.input)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestBillingItemService_UpdateItem_TransactionErrors(t *testing.T) {
	existingItem := &model.BillingItem{ID: 1, BillingID: 10, UnitPrice: 3000}
	newPrice := int64(5000)

	t.Run("propagates recalculateTotals error", func(t *testing.T) {
		repo := defaultMockBillingItemRepo()
		repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
			return existingItem, nil
		}
		repo.updateFieldsFn = func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil }
		repo.findByBillingIDFn = func(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
			return nil, errors.New("find failed")
		}
		svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		result, err := svc.UpdateItem(context.Background(), 1, 1, &UpdateBillingItemInput{UnitPrice: &newPrice})
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("propagates FindByID-after-update error", func(t *testing.T) {
		callCount := 0
		repo := defaultMockBillingItemRepo()
		repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
			callCount++
			if callCount == 1 {
				return existingItem, nil
			}
			return nil, errors.New("find after update failed")
		}
		repo.updateFieldsFn = func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil }
		svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		result, err := svc.UpdateItem(context.Background(), 1, 1, &UpdateBillingItemInput{UnitPrice: &newPrice})
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("propagates transactor error", func(t *testing.T) {
		repo := defaultMockBillingItemRepo()
		repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
			return existingItem, nil
		}
		svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{withTxErr: errors.New("tx failed")}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		result, err := svc.UpdateItem(context.Background(), 1, 1, &UpdateBillingItemInput{UnitPrice: &newPrice})
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// ---- DeleteItem: transaction-error branches ----

func TestBillingItemService_DeleteItem_TransactionErrors(t *testing.T) {
	existingItem := &model.BillingItem{ID: 1, BillingID: 10}

	t.Run("propagates recalculateTotals error", func(t *testing.T) {
		repo := defaultMockBillingItemRepo()
		repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
			return existingItem, nil
		}
		repo.deleteFn = func(_ context.Context, _, _ uint64) error { return nil }
		repo.findByBillingIDFn = func(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
			return nil, errors.New("find failed")
		}
		svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		err := svc.DeleteItem(context.Background(), 1, 1, &DeleteBillingItemInput{})
		assert.Error(t, err)
	})

	t.Run("propagates transactor error", func(t *testing.T) {
		repo := defaultMockBillingItemRepo()
		repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
			return existingItem, nil
		}
		svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{withTxErr: errors.New("tx failed")}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		err := svc.DeleteItem(context.Background(), 1, 1, &DeleteBillingItemInput{})
		assert.Error(t, err)
	})
}

// ---- GetUnbilledItems: error branches ----

func TestBillingItemService_GetUnbilledItems_Errors(t *testing.T) {
	t.Run("propagates treatment repo error", func(t *testing.T) {
		treatmentRepo := &mockTreatmentRepositoryForBilling{
			findUnbilledByPetIDFn: func(_ context.Context, _, _ uint64) ([]model.Treatment, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewBillingItemServiceWithCampaign(defaultMockBillingItemRepo(), defaultMockBillingRepo(), treatmentRepo, &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		items, err := svc.GetUnbilledItems(context.Background(), 1, 20)
		assert.Error(t, err)
		assert.Nil(t, items)
	})

	t.Run("propagates trimming finder error", func(t *testing.T) {
		repo := &mockBillingItemRepositoryWithTrimming{
			mockBillingItemRepository: defaultMockBillingItemRepo(),
			findUnbilledTrimmingItemsByPetIDFn: func(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
				return nil, errors.New("trimming db error")
			},
		}
		svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), defaultMockTreatmentRepo(), &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		items, err := svc.GetUnbilledItems(context.Background(), 1, 20)
		assert.Error(t, err)
		assert.Nil(t, items)
	})
}

// ---- GetUngroupedSameDaySummary: error and trimming-counter branches ----

func TestBillingItemService_GetUngroupedSameDaySummary_Errors(t *testing.T) {
	t.Run("propagates medical record count error", func(t *testing.T) {
		treatmentRepo := &mockTreatmentRepositoryForBilling{
			countFinalizedFn: func(_ context.Context, _, _ uint64, _ time.Time) (int64, error) {
				return 0, errors.New("db error")
			},
		}
		svc := NewBillingItemServiceWithCampaign(defaultMockBillingItemRepo(), defaultMockBillingRepo(), treatmentRepo, &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		_, err := svc.GetUngroupedSameDaySummary(context.Background(), 1, 20, time.Now())
		assert.Error(t, err)
	})

	t.Run("propagates trimming count error", func(t *testing.T) {
		repo := &mockBillingItemRepositoryWithTrimmingCounter{
			mockBillingItemRepository: defaultMockBillingItemRepo(),
			countNonAccountingTrimmingByPetAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (int64, error) {
				return 0, errors.New("trimming count error")
			},
		}
		treatmentRepo := &mockTreatmentRepositoryForBilling{
			countFinalizedFn: func(_ context.Context, _, _ uint64, _ time.Time) (int64, error) { return 1, nil },
		}
		svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), treatmentRepo, &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		_, err := svc.GetUngroupedSameDaySummary(context.Background(), 1, 20, time.Now())
		assert.Error(t, err)
	})

	t.Run("returns combined counts when trimming counter is implemented", func(t *testing.T) {
		repo := &mockBillingItemRepositoryWithTrimmingCounter{
			mockBillingItemRepository: defaultMockBillingItemRepo(),
			countNonAccountingTrimmingByPetAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (int64, error) {
				return 3, nil
			},
		}
		treatmentRepo := &mockTreatmentRepositoryForBilling{
			countFinalizedFn: func(_ context.Context, _, _ uint64, _ time.Time) (int64, error) { return 1, nil },
		}
		svc := NewBillingItemServiceWithCampaign(repo, defaultMockBillingRepo(), treatmentRepo, &mockTransactor{}, okTrimmingCourseRepo(), okTrimmingOptionRepo(), nil, nil)
		summary, err := svc.GetUngroupedSameDaySummary(context.Background(), 1, 20, time.Now())
		assert.NoError(t, err)
		assert.Equal(t, int64(1), summary.MedicalRecordCount)
		assert.Equal(t, int64(3), summary.TrimmingCount)
	})
}

// ---- recalculateTotals (unexported method) ----

func TestBillingItemService_RecalculateTotals(t *testing.T) {
	t.Run("propagates FindByBillingID error", func(t *testing.T) {
		repo := defaultMockBillingItemRepo()
		repo.findByBillingIDFn = func(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
			return nil, errors.New("find failed")
		}
		svc := &billingItemService{repo: repo}
		err := svc.recalculateTotals(context.Background(), 1, 10)
		assert.Error(t, err)
	})

	t.Run("propagates UpdateBillingTotals error", func(t *testing.T) {
		repo := defaultMockBillingItemRepo()
		repo.updateBillingTotals = func(_ context.Context, _, _ uint64, _, _, _ int64) error {
			return errors.New("update failed")
		}
		svc := &billingItemService{repo: repo}
		err := svc.recalculateTotals(context.Background(), 1, 10)
		assert.Error(t, err)
	})

	t.Run("succeeds and updates totals", func(t *testing.T) {
		repo := defaultMockBillingItemRepo()
		repo.findByBillingIDFn = func(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
			return []model.BillingItem{
				{UnitPrice: 1000, Quantity: 1, TaxRate: 0.10, TaxType: model.TaxTypeExcluded},
			}, nil
		}
		svc := &billingItemService{repo: repo}
		err := svc.recalculateTotals(context.Background(), 1, 10)
		assert.NoError(t, err)
	})
}
