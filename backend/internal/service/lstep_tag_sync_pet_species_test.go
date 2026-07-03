package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestSyncPetSpeciesTags(t *testing.T) {
	lineUID := "U_species"
	dog := &model.AnimalSpecies{Name: "犬"}
	cat := &model.AnimalSpecies{Name: "猫"}

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
			},
		}
		err := svc.SyncPetSpeciesTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when checkOptOut fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return nil, errors.New("owner db error")
				},
			},
		}
		err := svc.SyncPetSpeciesTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("skips when owner opted out", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LstepOptOut: true}, nil
				},
			},
		}
		err := svc.SyncPetSpeciesTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("skips when owner has no LINE user id", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2}, nil
				},
			},
		}
		err := svc.SyncPetSpeciesTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when pet lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
					return nil, errors.New("pet db error")
				},
			},
		}
		err := svc.SyncPetSpeciesTags(context.Background(), 1, 2)
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
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) { return nil, nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return nil, errors.New("credentials error")
			},
		}
		err := svc.SyncPetSpeciesTags(context.Background(), 1, 2)
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
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) { return nil, nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil },
		}
		err := svc.SyncPetSpeciesTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("has both dog and cat: adds both species tags", func(t *testing.T) {
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
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
					return []model.Pet{{AnimalSpecies: dog}, {AnimalSpecies: cat}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncPetSpeciesTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"PET_犬あり", "PET_猫あり"}, addedTags)
	})

	t.Run("no dog or cat: removes both species tags", func(t *testing.T) {
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
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) { return nil, nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncPetSpeciesTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"PET_犬あり", "PET_猫あり"}, removedTags)
	})

	t.Run("adding a species tag fails: recorded as api failure but does not propagate", func(t *testing.T) {
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error { return errors.New("add failed") },
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
					return []model.Pet{{AnimalSpecies: dog}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) { return 1, nil },
			},
		}
		err := svc.SyncPetSpeciesTags(context.Background(), 1, 2)
		// AddTag failure is recorded and the loop continues (apiFailed=true) rather
		// than returning early — verifying the sync completes without error.
		assert.NoError(t, err)
	})

	t.Run("removing a species tag fails: recorded as api failure but does not propagate", func(t *testing.T) {
		incrementCalls := 0
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("remove failed") },
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			petRepo: &mockPetRepository{
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) { return nil, nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
					incrementCalls++
					return 1, nil
				},
			},
		}
		err := svc.SyncPetSpeciesTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Positive(t, incrementCalls)
	})
}
