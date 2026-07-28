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

// ---- SyncPetBasicInfoTags ----

func TestSyncPetBasicInfoTags(t *testing.T) {
	lineUID := "U_pet_basic_test"

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
		}
		assert.NoError(t, svc.SyncPetBasicInfoTags(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when checkOptOut fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) { return nil, errors.New("db error") },
			},
		}
		assert.Error(t, svc.SyncPetBasicInfoTags(context.Background(), 1, 10))
	})

	t.Run("noop when owner opted out", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LstepOptOut: true}, nil
				},
			},
		}
		assert.NoError(t, svc.SyncPetBasicInfoTags(context.Background(), 1, 10))
	})

	t.Run("noop when owner has no line user id", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: nil}, nil
				},
			},
		}
		assert.NoError(t, svc.SyncPetBasicInfoTags(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when FindLivingByOwner fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
					return nil, errors.New("db error")
				},
			},
		}
		assert.Error(t, svc.SyncPetBasicInfoTags(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when tag cache lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) { return nil, nil },
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("db error")
				},
			},
		}
		assert.Error(t, svc.SyncPetBasicInfoTags(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when auto-managed prefix lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) { return nil, nil },
			},
			tagCacheRepo: &mockLstepTagCacheRepository{},
			tagConfigRepo: &mockLstepTagConfigRepository{
				findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
					return nil, errors.New("db error")
				},
			},
		}
		assert.Error(t, svc.SyncPetBasicInfoTags(context.Background(), 1, 10))
	})

	t.Run("returns nil when buildClient yields no client", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) { return nil, nil },
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil },
		}
		assert.NoError(t, svc.SyncPetBasicInfoTags(context.Background(), 1, 10))
	})

	t.Run("adds new tags and removes stale tags, updating the cache", func(t *testing.T) {
		bd := time.Date(2020, 4, 20, 0, 0, 0, 0, time.UTC)
		dog := model.AnimalSpecies{Name: "犬"}
		pets := []model.Pet{{Breed: "柴犬", Gender: model.PetGenderMale, BirthDate: &bd, AnimalSpecies: &dog}}

		var addedTags, removedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error { addedTags = append(addedTags, tagName); return nil },
			removeTagFn: func(_ context.Context, _, tagName string) error {
				removedTags = append(removedTags, tagName)
				return nil
			},
		}

		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) { return pets, nil },
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "sex_female"}}, nil // stale: pet is now male
				},
			},
			tagConfigRepo: &mockLstepTagConfigRepository{
				findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
					return []*model.LstepAutoManagedPrefix{{Prefix: "sex_", Category: "C1"}, {Prefix: "breed_", Category: "C1"}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}

		err := svc.SyncPetBasicInfoTags(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.Contains(t, removedTags, "sex_female")
		assert.Contains(t, addedTags, "sex_male")
	})

	t.Run("returns wrapped error when RemoveTag AddTag for a new tag fails", func(t *testing.T) {
		pets := []model.Pet{{Breed: "", Gender: model.PetGenderUnknown}}
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) { return pets, nil },
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncPetBasicInfoTags(context.Background(), 1, 10)
		assert.Error(t, err)
	})

	t.Run("nil tagConfigRepo skips prefix loading (empty c1Prefixes)", func(t *testing.T) {
		client := &mockLstepAPIClient{}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) { return nil, nil },
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "sex_female"}}, nil
				},
			},
			// tagConfigRepo intentionally nil
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncPetBasicInfoTags(context.Background(), 1, 10)
		assert.NoError(t, err, "nil tagConfigRepo must not remove any tag since c1Prefixes stays empty")
	})
}
