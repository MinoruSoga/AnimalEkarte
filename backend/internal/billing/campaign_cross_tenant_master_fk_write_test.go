package billing

// campaign_cross_tenant_master_fk_write_test.go — BE9-2C B①:
// service/cross_tenant_master_fk_write_test.go から campaignService 節を同名移動。
// Also hosts BE-ACT-MERCHANDISE-ATOMIC-DELETE interleaving proofs against campaign attach.

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/inventory"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
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
		return NewCampaignService(repo, rejectMerchandiseItemRepo(ownedItemID), noopTransactor{})
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
		return NewCampaignService(repo, rejectMerchandiseItemRepo(ownedItemID), noopTransactor{})
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

	// Ambient-tx path: validation must run inside WithTx so a foreign FK aborts before ReplaceTargets.
	t.Run("rejects cross-clinic target inside ambient transaction before replace", func(t *testing.T) {
		replaced := false
		inTx := false
		validatedInTx := false
		tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			inTx = true
			defer func() { inTx = false }()
			return fn(ctx)
		}}
		current := &model.Campaign{ID: 100, ClinicID: clinicID}
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
				return current, nil
			},
			replaceTargetsFn: func(_ context.Context, _ uint64, _ []model.ItemCategory, _ []uint64) error {
				replaced = true
				return nil
			},
		}
		merch := &mockMerchandiseItemFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.MerchandiseItem, error) {
			if !inTx {
				t.Error("cross-tenant merchandise check must run inside WithTx")
			}
			validatedInTx = inTx
			if id != ownedItemID {
				return nil, apperrors.WrapNotFound("merchandise_item", "foreign")
			}
			return &model.MerchandiseItem{ID: id, IsActive: true}, nil
		}}
		svc := NewCampaignService(repo, merch, tx)
		ids := []uint64{foreignItemID}
		out, err := svc.Update(context.Background(), clinicID, 100, &UpdateCampaignInput{TargetItemIDs: &ids})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.True(t, validatedInTx)
		assert.False(t, replaced)
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

// TestCampaignService_ConcurrentMerchandiseItemAtomicDelete proves both interleavings
// with the atomic merchandise Delete service (BE-ACT-MERCHANDISE-ATOMIC-DELETE):
//  1. attach-first → merchandise Delete serializes then Conflicts
//  2. delete-first → later campaign Create rejects inactive merchandise (NotFound)
func TestCampaignService_ConcurrentMerchandiseItemAtomicDelete(t *testing.T) {
	db := setupCampaignMerchandiseTargetTestDB(t)
	campaignRepo := NewCampaignRepository(db)
	merchRepo := inventory.NewMerchandiseItemRepository(db)
	merchSvc := inventory.NewMerchandiseItemService(merchRepo, persistence.NewTransactor(db))
	campaignSvc := NewCampaignService(campaignRepo, merchRepo, testNewTransactor(db))
	ctx := context.Background()
	const clinicID = uint64(1)

	t.Run("attach-first yields Conflict on merchandise atomic delete", func(t *testing.T) {
		item := &model.MerchandiseItem{
			ClinicID: clinicID, Name: "attach-first concurrent merch", Category: model.ItemCategoryGoods,
			UnitPrice: 1000, IsActive: true,
		}
		require.NoError(t, merchRepo.Create(ctx, item))

		// Campaign Create opens WithTx, share-locks merchandise, then writes targets.
		// Hold that validation path open while merchandise Delete runs.
		locked := make(chan struct{})
		release := make(chan struct{})
		holderDone := make(chan error, 1)
		go func() {
			holderDone <- testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
				// Same ambient FindByID FOR SHARE path Create uses for target validation.
				if _, err := merchRepo.FindByID(txCtx, clinicID, item.ID); err != nil {
					return err
				}
				// Persist campaign + target under the same ambient tx so post-serialization
				// CountUsage sees the reference after the share lock is released.
				camp := &model.Campaign{
					ClinicID: clinicID, Name: "attach-first holder",
					StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour),
					DiscountType: model.CampaignDiscountTypeRate, DiscountValue: 5, IsActive: true,
					TargetItems: []model.CampaignTargetItem{{MerchandiseItemID: item.ID}},
				}
				if _, err := campaignRepo.Create(txCtx, camp); err != nil {
					return err
				}
				close(locked)
				<-release
				return nil
			})
		}()
		<-locked

		deleteDone := make(chan error, 1)
		go func() {
			deleteDone <- merchSvc.Delete(ctx, clinicID, item.ID)
		}()

		select {
		case err := <-deleteDone:
			close(release)
			require.Failf(t, "merchandise atomic delete was not serialized behind campaign attach", "err=%v", err)
		case <-time.After(100 * time.Millisecond):
			// still waiting for exclusive soft-delete lock — expected
		}

		close(release)
		require.NoError(t, <-holderDone)

		deleteErr := <-deleteDone
		require.Error(t, deleteErr)
		assert.True(t, apperrors.IsConflict(deleteErr), "attach-first must Conflict, got %v", deleteErr)

		// Row must still exist (soft-delete rolled back).
		got, err := merchRepo.FindByID(ctx, clinicID, item.ID)
		require.NoError(t, err)
		assert.Equal(t, item.ID, got.ID)
	})

	t.Run("delete-first rejects subsequent campaign target attach", func(t *testing.T) {
		item := &model.MerchandiseItem{
			ClinicID: clinicID, Name: "delete-first concurrent merch", Category: model.ItemCategoryGoods,
			UnitPrice: 1000, IsActive: true,
		}
		require.NoError(t, merchRepo.Create(ctx, item))
		require.NoError(t, merchSvc.Delete(ctx, clinicID, item.ID))

		out, err := campaignSvc.Create(ctx, clinicID, &CreateCampaignInput{
			Name:          "attach after atomic delete",
			StartDate:     time.Now(),
			EndDate:       time.Now().Add(24 * time.Hour),
			DiscountType:  model.CampaignDiscountTypeRate,
			DiscountValue: 10,
			TargetItemIDs: []uint64{item.ID},
		})
		require.Error(t, err)
		assert.Nil(t, out)
		assert.True(t, apperrors.IsNotFound(err), "delete-first must make later attach reject inactive row")
	})
}
