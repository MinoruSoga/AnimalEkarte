package billing

// billing_item_cross_tenant_master_fk_write_test.go — BE9-2C B③:
// service/cross_tenant_master_fk_write_test.go から billingItemService 節を同名移動。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestBillingItemService_CreateItem_RejectsCrossClinicTrimmingFK(t *testing.T) {
	const clinicID = uint64(1)
	const billingID = uint64(10)
	const ownedCourseID = uint64(300)
	const foreignCourseID = uint64(999)
	const ownedOptionID = uint64(400)
	const foreignOptionID = uint64(998)

	newSvc := func(created *bool, courseRepo trimmingCourseFinder, optionRepo trimmingOptionFinder) BillingItemService {
		repo := defaultMockBillingItemRepo()
		repo.createFn = func(_ context.Context, item *model.BillingItem) error { *created = true; item.ID = 1; return nil }
		billingRepo := &mockAccountingRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			return &model.Billing{ID: id}, nil
		}}
		return NewBillingItemServiceWithCampaign(repo, billingRepo, defaultMockTreatmentRepo(), &mockTransactor{}, courseRepo, optionRepo, nil, nil)
	}

	baseInput := func() *CreateBillingItemInput {
		return &CreateBillingItemInput{
			ClinicID:  clinicID,
			BillingID: billingID,
			Category:  string(model.ItemCategoryProcedure),
			Name:      "トリミング",
			UnitPrice: 5000,
			Quantity:  1,
		}
	}

	t.Run("rejects cross-clinic trimming_course_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID), okTrimmingOptionRepo())
		foreign := foreignCourseID
		input := baseInput()
		input.TrimmingCourseID = &foreign
		out, err := svc.CreateItem(context.Background(), input)
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "billing item must NOT be persisted referencing another clinic's trimming course")
	})

	t.Run("rejects cross-clinic trimming_option_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created, okTrimmingCourseRepo(), rejectTrimmingOptionRepo(ownedOptionID))
		foreign := foreignOptionID
		input := baseInput()
		input.TrimmingOptionID = &foreign
		out, err := svc.CreateItem(context.Background(), input)
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "billing item must NOT be persisted referencing another clinic's trimming option")
	})

	t.Run("accepts same-clinic trimming_course_id/trimming_option_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created, rejectTrimmingCourseRepo(ownedCourseID), rejectTrimmingOptionRepo(ownedOptionID))
		owned1, owned2 := ownedCourseID, ownedOptionID
		input := baseInput()
		input.TrimmingCourseID = &owned1
		input.TrimmingOptionID = &owned2
		out, err := svc.CreateItem(context.Background(), input)
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

// ── reservationValidators.ValidateAndCreate / liffService.CreateReservation
//    (X-14/U6a): ReservationTypeID / TrimmingCourseID / TrimmingOptionIDs ──
