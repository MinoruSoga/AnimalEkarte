package billing

// campaign_cross_tenant_master_fk_write_test.go — BE9-2C B①:
// service/cross_tenant_master_fk_write_test.go から campaignService 節を同名移動。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestCampaignService_Create_RejectsCrossClinicTargetItemFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedItemID = uint64(10)
	const foreignItemID = uint64(999)

	newSvc := func(created *bool) CampaignService {
		repo := &mockCampaignRepository{
			createFn: func(_ context.Context, m *model.Campaign) (*model.Campaign, error) {
				*created = true
				m.ID = 1
				return m, nil
			},
		}
		return NewCampaignService(repo, rejectMerchandiseItemRepo(ownedItemID))
	}

	baseInput := func(itemID uint64) *CreateCampaignInput {
		return &CreateCampaignInput{
			Name:          "Autumn Sale",
			StartDate:     time.Now(),
			EndDate:       time.Now().Add(24 * time.Hour),
			DiscountType:  model.CampaignDiscountTypeRate,
			DiscountValue: 10.0,
			TargetItemIDs: []uint64{itemID},
		}
	}

	t.Run("rejects cross-clinic target_item_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, baseInput(foreignItemID))
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "campaign must NOT be persisted referencing another clinic's merchandise item")
	})

	t.Run("accepts same-clinic target_item_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, baseInput(ownedItemID))
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestCampaignService_Update_RejectsCrossClinicTargetItemFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedItemID = uint64(10)
	const foreignItemID = uint64(999)

	newSvc := func(replaced *bool) CampaignService {
		current := &model.Campaign{ID: 100, ClinicID: clinicID}
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
				return current, nil
			},
			replaceTargetsFn: func(_ context.Context, _ uint64, _ []model.ItemCategory, _ []uint64) error {
				*replaced = true
				return nil
			},
		}
		return NewCampaignService(repo, rejectMerchandiseItemRepo(ownedItemID))
	}

	t.Run("rejects cross-clinic target_item_id on update and does not persist", func(t *testing.T) {
		replaced := false
		svc := newSvc(&replaced)
		ids := []uint64{foreignItemID}
		out, err := svc.Update(context.Background(), clinicID, 100, &UpdateCampaignInput{TargetItemIDs: &ids})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, replaced, "campaign targets must NOT be replaced referencing another clinic's merchandise item")
	})

	t.Run("accepts same-clinic target_item_id (no false-reject)", func(t *testing.T) {
		replaced := false
		svc := newSvc(&replaced)
		ids := []uint64{ownedItemID}
		out, err := svc.Update(context.Background(), clinicID, 100, &UpdateCampaignInput{TargetItemIDs: &ids})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, replaced)
	})
}

// ── X-14 batch: self-ref ParentID ownership guard ──
// (checkupType/consultation/examType/procedure/vaccine)
//
// Each of these five master-data services carries a self-referencing ParentID (a
// sub-category pointing at its own parent row in the same table). Prior to this batch,
// Create/Update persisted a request-supplied ParentID without verifying it belongs to
// the caller's clinic. Each service has a single repo dependency (self-ref), so the
// mock below wires findByIDFn (used both for the parent-ownership guard and, on Update,
// the pre-existing self-entity existence check) alongside createFn/updateFieldsFn.

// mockMerchandiseItemFinder — merchandiseItemFinder view の最小モック（service側builderのview型版複製）。
type mockMerchandiseItemFinder struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
}

func (m *mockMerchandiseItemFinder) FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func rejectMerchandiseItemRepo(ownedID uint64) merchandiseItemFinder {
	return &mockMerchandiseItemFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.MerchandiseItem, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("merchandise_item", "foreign")
		}
		return &model.MerchandiseItem{ID: id}, nil
	}}
}
