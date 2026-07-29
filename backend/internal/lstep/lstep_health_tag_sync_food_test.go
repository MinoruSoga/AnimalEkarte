package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- local minimal BillingItemRepository mock scoped to this file ----

type foodMockBillingItemRepository struct {
	hasFoodPurchaseByOwnerSinceFn func(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
}

func (m *foodMockBillingItemRepository) FindByID(_ context.Context, _, _ uint64) (*model.BillingItem, error) {
	return nil, nil
}
func (m *foodMockBillingItemRepository) FindByBillingID(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
	return nil, nil
}
func (m *foodMockBillingItemRepository) Create(_ context.Context, _ *model.BillingItem) error {
	return nil
}
func (m *foodMockBillingItemRepository) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *foodMockBillingItemRepository) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *foodMockBillingItemRepository) UpdateBillingTotals(_ context.Context, _, _ uint64, _, _, _ int64) error {
	return nil
}
func (m *foodMockBillingItemRepository) HasItemByOwnerSince(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) {
	return false, nil
}
func (m *foodMockBillingItemRepository) HasFoodPurchaseByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	if m.hasFoodPurchaseByOwnerSinceFn != nil {
		return m.hasFoodPurchaseByOwnerSinceFn(ctx, clinicID, ownerID, since, names)
	}
	return false, nil
}
func (m *foodMockBillingItemRepository) FindUnbilledTrimmingItemsByPetID(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
	return nil, nil
}
func (m *foodMockBillingItemRepository) CountNonAccountingTrimmingByPetAndDate(_ context.Context, _, _ uint64, _ time.Time) (int64, error) {
	return 0, nil
}

func TestSyncFoodPurchaseTag_ExtraBranches(t *testing.T) {
	lineUID := "U_food_test"

	t.Run("noop when tagCodeRepo is nil", func(t *testing.T) {
		svc := &lstepTagSyncService{}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.NoError(t, err)
	})

	t.Run("noop when billingItemRepo is nil", func(t *testing.T) {
		svc := &lstepTagSyncService{tagCodeRepo: &mockLstepTagCodeMappingRepository{}}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.NoError(t, err)
	})

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			tagCodeRepo:     &mockLstepTagCodeMappingRepository{},
			billingItemRepo: &foodMockBillingItemRepository{},
			settingsSvc:     &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when tag code mapping lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			tagCodeRepo: &mockLstepTagCodeMappingRepository{
				findByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) ([]*model.LstepTagCodeMapping, error) {
					return nil, errors.New("db error")
				},
			},
			billingItemRepo: &foodMockBillingItemRepository{},
			settingsSvc:     &mockLstepSettingsService{},
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when checkOptOut fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			tagCodeRepo:     &mockLstepTagCodeMappingRepository{},
			billingItemRepo: &foodMockBillingItemRepository{},
			settingsSvc:     &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return nil, errors.New("db error")
				},
			},
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.Error(t, err)
	})

	t.Run("noop when owner opted out", func(t *testing.T) {
		svc := &lstepTagSyncService{
			tagCodeRepo:     &mockLstepTagCodeMappingRepository{},
			billingItemRepo: &foodMockBillingItemRepository{},
			settingsSvc:     &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LstepOptOut: true}, nil
				},
			},
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.NoError(t, err)
	})

	t.Run("noop when owner has no line user id", func(t *testing.T) {
		svc := &lstepTagSyncService{
			tagCodeRepo:     &mockLstepTagCodeMappingRepository{},
			billingItemRepo: &foodMockBillingItemRepository{},
			settingsSvc:     &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: nil}, nil
				},
			},
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when thresholds lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			tagCodeRepo:     &mockLstepTagCodeMappingRepository{},
			billingItemRepo: &foodMockBillingItemRepository{},
			settingsSvc: &mockLstepSettingsService{
				getHealthPreventionThresholdsFn: func(_ context.Context, _ uint64) (model.HealthPreventionThresholds, error) {
					return model.HealthPreventionThresholds{}, errors.New("db error")
				},
			},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when purchase check fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			tagCodeRepo: &mockLstepTagCodeMappingRepository{},
			billingItemRepo: &foodMockBillingItemRepository{hasFoodPurchaseByOwnerSinceFn: func(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) {
				return false, errors.New("db error")
			}},
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.Error(t, err)
	})

	t.Run("returns nil when buildClient yields no client (lstep not configured)", func(t *testing.T) {
		svc := &lstepTagSyncService{
			tagCodeRepo:     &mockLstepTagCodeMappingRepository{},
			billingItemRepo: &foodMockBillingItemRepository{},
			settingsSvc:     &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil },
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.NoError(t, err)
	})

	t.Run("adds the tag when a purchase is found", func(t *testing.T) {
		var addedTag string
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, tagName string) error { addedTag = tagName; return nil }}
		svc := &lstepTagSyncService{
			tagCodeRepo:     &mockLstepTagCodeMappingRepository{},
			billingItemRepo: &foodMockBillingItemRepository{hasFoodPurchaseByOwnerSinceFn: func(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) { return true, nil }},
			settingsSvc:     &mockLstepSettingsService{},
			tagCacheRepo:    &mockLstepTagCacheRepository{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, LtvFoodPurchaseTag, addedTag)
	})

	t.Run("returns wrapped error when AddTag fails", func(t *testing.T) {
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			tagCodeRepo:      &mockLstepTagCodeMappingRepository{},
			billingItemRepo:  &foodMockBillingItemRepository{hasFoodPurchaseByOwnerSinceFn: func(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) { return true, nil }},
			settingsSvc:      &mockLstepSettingsService{},
			tagCacheRepo:     &mockLstepTagCacheRepository{},
			errorCounterRepo: nil,
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.Error(t, err)
	})

	t.Run("removes the tag when no purchase is found", func(t *testing.T) {
		var removedTag string
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, tagName string) error { removedTag = tagName; return nil }}
		svc := &lstepTagSyncService{
			tagCodeRepo:     &mockLstepTagCodeMappingRepository{},
			billingItemRepo: &foodMockBillingItemRepository{hasFoodPurchaseByOwnerSinceFn: func(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) { return false, nil }},
			settingsSvc:     &mockLstepSettingsService{},
			tagCacheRepo:    &mockLstepTagCacheRepository{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, LtvFoodPurchaseTag, removedTag)
	})

	t.Run("RemoveTag failure propagates error (G2B-01)", func(t *testing.T) {
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			tagCodeRepo:     &mockLstepTagCodeMappingRepository{},
			billingItemRepo: &foodMockBillingItemRepository{hasFoodPurchaseByOwnerSinceFn: func(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) { return false, nil }},
			settingsSvc:     &mockLstepSettingsService{},
			tagCacheRepo:    &mockLstepTagCacheRepository{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 10, nil, nil)
		assert.Error(t, err)
	})
}
