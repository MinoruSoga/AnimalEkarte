package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestSyncExclusionTags(t *testing.T) {
	lineUID := "U_exclusion"

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
			},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when owner lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return nil, errors.New("owner db error")
				},
			},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("skips when owner has no LINE user id (does not check opt-out)", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LstepOptOut: true}, nil
				},
			},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when total pet count fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				countByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return 0, errors.New("count error")
				},
			},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when living pet count fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				countByOwnerFn:       func(_ context.Context, _, _ uint64) (int64, error) { return 1, nil },
				countLivingByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) { return 0, errors.New("count error") },
			},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("returns raw error when buildClient fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				countByOwnerFn:       func(_ context.Context, _, _ uint64) (int64, error) { return 1, nil },
				countLivingByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) { return 1, nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return nil, errors.New("credentials error")
			},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("noop when buildClient returns nil client", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				countByOwnerFn:       func(_ context.Context, _, _ uint64) (int64, error) { return 1, nil },
				countLivingByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) { return 1, nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil },
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	buildPetRepo := func(total, living int64) *mockPetRepository {
		return &mockPetRepository{
			countByOwnerFn:       func(_ context.Context, _, _ uint64) (int64, error) { return total, nil },
			countLivingByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) { return living, nil },
		}
	}

	t.Run("all pets dead: adds EXCL_配信停止 tag", func(t *testing.T) {
		var addedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			petRepo:       buildPetRepo(2, 0),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, exclTagDeliveryStop)
	})

	t.Run("no pets at all is not treated as all-dead", func(t *testing.T) {
		var addedTags, removedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
			removeTagFn: func(_ context.Context, _, tagName string) error {
				removedTags = append(removedTags, tagName)
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			petRepo:       buildPetRepo(0, 0),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.NotContains(t, addedTags, exclTagDeliveryStop)
		assert.Contains(t, removedTags, exclTagDeliveryStop)
	})

	t.Run("LstepOptOut true: adds EXCL_配信停止 tag even with living pets", func(t *testing.T) {
		var addedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID, LstepOptOut: true}, nil
				},
			},
			petRepo:       buildPetRepo(1, 1),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, exclTagDeliveryStop)
	})

	t.Run("DeliveryExcluded true: adds EXCL_配信停止 tag", func(t *testing.T) {
		var addedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID, DeliveryExcluded: true}, nil
				},
			},
			petRepo:       buildPetRepo(1, 1),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, exclTagDeliveryStop)
	})

	t.Run("IsTransferred true: adds EXCL_配信停止 tag", func(t *testing.T) {
		var addedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID, IsTransferred: true}, nil
				},
			},
			petRepo:       buildPetRepo(1, 1),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, exclTagDeliveryStop)
	})

	t.Run("MembershipType deceased: adds EXCL_配信停止 tag", func(t *testing.T) {
		var addedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID, MembershipType: model.MembershipTypeDeceased}, nil
				},
			},
			petRepo:       buildPetRepo(1, 1),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, exclTagDeliveryStop)
	})

	t.Run("MembershipType transferred: adds EXCL_配信停止 tag", func(t *testing.T) {
		var addedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID, MembershipType: model.MembershipTypeTransferred}, nil
				},
			},
			petRepo:       buildPetRepo(1, 1),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, exclTagDeliveryStop)
	})

	t.Run("no exclusion conditions: removes EXCL_配信停止 tag", func(t *testing.T) {
		var removedTags []string
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, tagName string) error {
				removedTags = append(removedTags, tagName)
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			petRepo:       buildPetRepo(1, 1),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, removedTags, exclTagDeliveryStop)
	})

	t.Run("returns wrapped error when adding EXCL_配信停止 tag fails", func(t *testing.T) {
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				if tagName == exclTagDeliveryStop {
					return errors.New("add failed")
				}
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID, LstepOptOut: true}, nil
				},
			},
			petRepo:       buildPetRepo(1, 1),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("removing EXCL_配信停止 tag fails: recorded as api failure but does not propagate", func(t *testing.T) {
		incrementCalls := 0
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, tagName string) error {
				if tagName == exclTagDeliveryStop {
					return errors.New("remove failed")
				}
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			petRepo:       buildPetRepo(1, 1),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
					incrementCalls++
					return 1, nil
				},
			},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Positive(t, incrementCalls)
	})

	t.Run("adding EXCL_配信注意 tag fails: recorded as api failure but does not propagate", func(t *testing.T) {
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				if tagName == exclTagDeliveryCaution {
					return errors.New("add caution failed")
				}
				return nil
			},
			removeTagFn: func(_ context.Context, _, _ string) error { return nil },
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID, DeliveryCaution: true}, nil
				},
			},
			petRepo:       buildPetRepo(1, 1),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 2)
		assert.NoError(t, err, "caution tag failure is best-effort and must not fail the sync")
	})
}
