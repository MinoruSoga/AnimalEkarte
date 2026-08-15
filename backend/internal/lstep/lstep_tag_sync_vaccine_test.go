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

// ---- SyncVaccineTag ----

func TestSyncVaccineTag(t *testing.T) {
	lineUID := "U_vaccine_sync_test"
	dog := model.VaccineSpeciesDog
	vacDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
		}
		assert.NoError(t, svc.SyncVaccineTag(context.Background(), 1, 10, 100))
	})

	t.Run("returns wrapped error when checkOptOut fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) { return nil, errors.New("db error") },
			},
		}
		assert.Error(t, svc.SyncVaccineTag(context.Background(), 1, 10, 100))
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
		assert.NoError(t, svc.SyncVaccineTag(context.Background(), 1, 10, 100))
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
		assert.NoError(t, svc.SyncVaccineTag(context.Background(), 1, 10, 100))
	})

	t.Run("returns wrapped error when vacRepo.FindByID fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return nil, errors.New("db error")
				},
			},
		}
		assert.Error(t, svc.SyncVaccineTag(context.Background(), 1, 10, 100))
	})

	t.Run("noop when vaccine yields no tags (Vaccine relation is nil)", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return &model.Vaccination{Date: vacDate, Vaccine: nil}, nil
				},
			},
		}
		assert.NoError(t, svc.SyncVaccineTag(context.Background(), 1, 10, 100))
	})

	t.Run("returns nil when buildClient yields no client", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return &model.Vaccination{Date: vacDate, Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil },
		}
		assert.NoError(t, svc.SyncVaccineTag(context.Background(), 1, 10, 100))
	})

	t.Run("adds the tag and upserts the cache on success", func(t *testing.T) {
		var addedTag, upsertedTag string
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, tagName string) error { addedTag = tagName; return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return &model.Vaccination{Date: vacDate, Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				upsertTagFn: func(_ context.Context, _, _ uint64, tagName, _, _ string) error { upsertedTag = tagName; return nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncVaccineTag(context.Background(), 1, 10, 100)
		assert.NoError(t, err)
		assert.Equal(t, "vaccine_dog_2026-04-01", addedTag)
		assert.Equal(t, "vaccine_dog_2026-04-01", upsertedTag)
	})

	t.Run("returns wrapped error when AddTag fails", func(t *testing.T) {
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return &model.Vaccination{Date: vacDate, Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog}}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncVaccineTag(context.Background(), 1, 10, 100)
		assert.Error(t, err)
	})

	t.Run("removes stale same-category tags before adding the new one", func(t *testing.T) {
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
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return &model.Vaccination{Date: vacDate, Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "vaccine_dog_2025-01-01"}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncVaccineTag(context.Background(), 1, 10, 100)
		assert.NoError(t, err)
		assert.Contains(t, removedTags, "vaccine_dog_2025-01-01")
	})

	t.Run("cache load failure: zero AddTag and returns observable error", func(t *testing.T) {
		addCalls := 0
		removeCalls := 0
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error {
				addCalls++
				return nil
			},
			removeTagFn: func(_ context.Context, _, _ string) error {
				removeCalls++
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
			vacRepo: &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return &model.Vaccination{Date: vacDate, Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("cache db error")
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncVaccineTag(context.Background(), 1, 10, 100)
		assert.Error(t, err)
		assert.Zero(t, addCalls, "cache failure must fail closed: zero AddTag")
		assert.Zero(t, removeCalls, "cache failure must fail closed: zero RemoveTag")
	})

	t.Run("RemoveTag partial failure: desired AddTag continues and failure is accounted", func(t *testing.T) {
		var addedTags []string
		incrementCalls := 0
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
			removeTagFn: func(_ context.Context, _, _ string) error {
				return errors.New("remove failed")
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return &model.Vaccination{Date: vacDate, Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "vaccine_dog_2025-01-01"}}, nil
				},
			},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
					incrementCalls++
					return 1, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncVaccineTag(context.Background(), 1, 10, 100)
		assert.NoError(t, err, "RemoveTag partial failure must not abort desired adds")
		assert.Contains(t, addedTags, "vaccine_dog_2026-04-01")
		assert.Positive(t, incrementCalls, "durable failure accounting must remain")
	})
}

// ---- ResyncOwnerVaccineTags ----

func TestResyncOwnerVaccineTags(t *testing.T) {
	lineUID := "U_resync_test"
	dog := model.VaccineSpeciesDog
	vacDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
		}
		assert.NoError(t, svc.ResyncOwnerVaccineTags(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when checkOptOut fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) { return nil, errors.New("db error") },
			},
		}
		assert.Error(t, svc.ResyncOwnerVaccineTags(context.Background(), 1, 10))
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
		assert.NoError(t, svc.ResyncOwnerVaccineTags(context.Background(), 1, 10))
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
		assert.NoError(t, svc.ResyncOwnerVaccineTags(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when buildClient fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, errors.New("credentials error") },
		}
		assert.Error(t, svc.ResyncOwnerVaccineTags(context.Background(), 1, 10))
	})

	t.Run("returns nil when buildClient yields no client", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil },
		}
		assert.NoError(t, svc.ResyncOwnerVaccineTags(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when vacRepo.FindByOwner fails", func(t *testing.T) {
		client := &mockLstepAPIClient{}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return nil, errors.New("db error")
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		assert.Error(t, svc.ResyncOwnerVaccineTags(context.Background(), 1, 10))
	})

	t.Run("rebuilds vaccine_* tags from the latest per-species records", func(t *testing.T) {
		var addedTags []string
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, tagName string) error { addedTags = append(addedTags, tagName); return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return []model.Vaccination{{Date: vacDate, Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog}}}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.ResyncOwnerVaccineTags(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, "vaccine_dog_2026-04-01")
	})

	t.Run("no vaccinations: all stale tags removed, no new tags added", func(t *testing.T) {
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
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) { return nil, nil },
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "vaccine_dog_2025-01-01"}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.ResyncOwnerVaccineTags(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.Contains(t, removedTags, "vaccine_dog_2025-01-01")
	})

	t.Run("returns wrapped error when AddTag fails during resync", func(t *testing.T) {
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return []model.Vaccination{{Date: vacDate, Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog}}}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.ResyncOwnerVaccineTags(context.Background(), 1, 10)
		assert.Error(t, err)
	})

	t.Run("cache load failure: zero AddTag and returns observable error", func(t *testing.T) {
		addCalls := 0
		removeCalls := 0
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error {
				addCalls++
				return nil
			},
			removeTagFn: func(_ context.Context, _, _ string) error {
				removeCalls++
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
			vacRepo: &mockVaccinationRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return []model.Vaccination{{Date: vacDate, Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog}}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("cache db error")
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.ResyncOwnerVaccineTags(context.Background(), 1, 10)
		assert.Error(t, err)
		assert.Zero(t, addCalls, "cache failure must fail closed: zero AddTag")
		assert.Zero(t, removeCalls, "cache failure must fail closed: zero RemoveTag")
	})

	t.Run("RemoveTag partial failure: desired AddTag continues and failure is accounted", func(t *testing.T) {
		var addedTags []string
		incrementCalls := 0
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
			removeTagFn: func(_ context.Context, _, _ string) error {
				return errors.New("remove failed")
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return []model.Vaccination{{Date: vacDate, Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog}}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "vaccine_dog_2025-01-01"}}, nil
				},
			},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
					incrementCalls++
					return 1, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.ResyncOwnerVaccineTags(context.Background(), 1, 10)
		assert.NoError(t, err, "RemoveTag partial failure must not abort desired adds")
		assert.Contains(t, addedTags, "vaccine_dog_2026-04-01")
		assert.Positive(t, incrementCalls, "durable failure accounting must remain")
	})
}
