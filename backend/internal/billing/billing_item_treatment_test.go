package billing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestBillingItemService_CreateItem_DuplicateTreatmentProvenanceConflicts(t *testing.T) {
	treatmentID := uint64(100)
	repo := defaultMockBillingItemRepo()
	repo.validateCreateReferencesFn = func(_ context.Context, _, _ uint64, _, _, _, _, _ *uint64) (model.ItemCategory, error) {
		return model.ItemCategoryProcedure, nil
	}
	repo.createFn = func(_ context.Context, _ *model.BillingItem) error {
		return apperrors.WrapAlreadyExists("billing_item", "")
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
		ClinicID:    1,
		BillingID:   10,
		Category:    string(model.ItemCategoryProcedure),
		Name:        "処置",
		UnitPrice:   1000,
		Quantity:    1,
		TreatmentID: &treatmentID,
	})
	require.Error(t, err)
	assert.Nil(t, item)
	assert.True(t, apperrors.IsConflict(err), "duplicate treatment claim must conflict: %v", err)
}

func TestBillingItemTreatmentProvenance_SecondClaimConflicts(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	require.NoError(t, f.db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS uq_billing_items_treatment_lifetime
    ON billing_items (treatment_id)
    WHERE treatment_id IS NOT NULL
`).Error)

	svc := newBillingItemReferenceService(f, f.repo)
	input := billingItemReferenceCreateInput(f)
	input.Category = string(model.ItemCategoryProcedure)
	input.TreatmentID = &f.treatment.ID

	first, err := svc.CreateItem(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, first)
	require.NotNil(t, first.TreatmentID)
	assert.Equal(t, f.treatment.ID, *first.TreatmentID)

	second, err := svc.CreateItem(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, second)
	assert.True(t, apperrors.IsConflict(err), "duplicate treatment claim must conflict: %v", err)
}

// TestBillingItemService_CreateItem_LateItemOnCompletedConflicts pins UAT-R2-EXCLUSIVE-LOCK
// "late line item after complete": finalized billing status guard → Conflict, no Create.
func TestBillingItemService_CreateItem_LateItemOnCompletedConflicts(t *testing.T) {
	createCalled := false
	repo := defaultMockBillingItemRepo()
	repo.validateCreateReferencesFn = func(_ context.Context, _, _ uint64, _, _, _, _, _ *uint64) (model.ItemCategory, error) {
		return model.ItemCategoryExamination, nil
	}
	repo.createFn = func(_ context.Context, _ *model.BillingItem) error {
		createCalled = true
		return nil
	}
	billingRepo := defaultMockBillingRepo()
	billingRepo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.Billing, error) {
		return &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusCompleted}, nil
	}
	svc := NewBillingItemServiceWithCampaign(
		repo,
		billingRepo,
		defaultMockTreatmentRepo(),
		&mockTransactor{},
		okTrimmingCourseRepo(),
		okTrimmingOptionRepo(),
		nil,
		nil,
	)

	item, err := svc.CreateItem(context.Background(), &CreateBillingItemInput{
		ClinicID:  1,
		BillingID: 10,
		Category:  string(model.ItemCategoryExamination),
		Name:      "診察料",
		UnitPrice: 3000,
		Quantity:  1,
	})
	require.Error(t, err)
	assert.Nil(t, item)
	assert.True(t, apperrors.IsConflict(err), "late item on completed billing must conflict: %v", err)
	assert.False(t, createCalled, "create must not run after finalized status guard")
}

// TestBillingItemService_CreateItem_DuplicateExamProvenanceConflicts pins exam lifetime UNIQUE
// AlreadyExists → Conflict (same mapping as treatment/vaccination provenance).
func TestBillingItemService_CreateItem_DuplicateExamProvenanceConflicts(t *testing.T) {
	examID := uint64(77)
	repo := defaultMockBillingItemRepo()
	repo.validateCreateReferencesFn = func(_ context.Context, _, _ uint64, _, _, _, _, _ *uint64) (model.ItemCategory, error) {
		return model.ItemCategoryTest, nil
	}
	repo.createFn = func(_ context.Context, _ *model.BillingItem) error {
		return apperrors.WrapAlreadyExists("billing_item", "")
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
		ClinicID:  1,
		BillingID: 10,
		Category:  string(model.ItemCategoryTest),
		Name:      "血液検査",
		UnitPrice: 4200,
		Quantity:  1,
		ExamID:    &examID,
	})
	require.Error(t, err)
	assert.Nil(t, item)
	assert.True(t, apperrors.IsConflict(err), "duplicate exam claim must conflict: %v", err)
}
