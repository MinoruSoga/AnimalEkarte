package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestSyncSeniorTag(t *testing.T) {
	lineUID := "U_senior"
	dog := &model.AnimalSpecies{Name: "犬"}
	seniorBirth := time.Now().AddDate(-8, 0, 0)
	youngBirth := time.Now().AddDate(-2, 0, 0)

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
			},
		}
		err := svc.SyncSeniorTag(context.Background(), 1, 2)
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
		err := svc.SyncSeniorTag(context.Background(), 1, 2)
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
		err := svc.SyncSeniorTag(context.Background(), 1, 2)
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
		err := svc.SyncSeniorTag(context.Background(), 1, 2)
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
		err := svc.SyncSeniorTag(context.Background(), 1, 2)
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
		err := svc.SyncSeniorTag(context.Background(), 1, 2)
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
		err := svc.SyncSeniorTag(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("has a senior pet: adds PET_シニア対象 tag", func(t *testing.T) {
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
					return []model.Pet{{AnimalSpecies: dog, BirthDate: &seniorBirth}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncSeniorTag(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, "PET_シニア対象")
	})

	t.Run("no senior pet: removes PET_シニア対象 tag", func(t *testing.T) {
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
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
					return []model.Pet{{AnimalSpecies: dog, BirthDate: &youngBirth}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncSeniorTag(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, removedTags, "PET_シニア対象")
	})

	t.Run("returns wrapped error when adding senior tag fails", func(t *testing.T) {
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
					return []model.Pet{{AnimalSpecies: dog, BirthDate: &seniorBirth}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncSeniorTag(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("removing senior tag fails: recorded as api failure but does not propagate", func(t *testing.T) {
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
				findLivingByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Pet, error) {
					return []model.Pet{{AnimalSpecies: dog, BirthDate: &youngBirth}}, nil
				},
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
		err := svc.SyncSeniorTag(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Positive(t, incrementCalls)
	})
}
