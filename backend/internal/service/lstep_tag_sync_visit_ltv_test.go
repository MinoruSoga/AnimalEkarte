package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// SyncLTVTopPercent starts at 0% coverage. Uses cpmAccountingRepositoryWrapper /
// newCPMAccountingRepository defined in lstep_tag_sync_visit_cpm_test.go (same package).

func TestSyncLTVTopPercent(t *testing.T) {
	lineUID1 := "U_ltv_top"
	lineUID2 := "U_ltv_other"

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.Empty(t, errs)
	})

	t.Run("returns wrapped error when shouldSkipSync check fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("db error") }},
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("returns wrapped error when FindOwnersByAnnualRevenue fails", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
			return nil, errors.New("db error")
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("returns wrapped error when FindAllWithLineUserID fails", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDFn: func(_ context.Context, _ uint64) ([]model.Owner, error) {
					return nil, errors.New("db error")
				},
			},
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("returns 0,nil when buildClient yields no client", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo:   &mockOwnerRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return nil, nil
			},
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.Empty(t, errs)
	})

	t.Run("returns wrapped error when buildClient itself fails", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo:   &mockOwnerRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return nil, errors.New("credentials error")
			},
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("skips owners without a line user id", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
			return []repository.OwnerAnnualRevenue{{OwnerID: 1}}, nil
		}
		client := &mockLstepAPIClient{}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDFn: func(_ context.Context, _ uint64) ([]model.Owner, error) {
					return []model.Owner{{ID: 1, LineUserID: nil}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.Empty(t, errs)
	})

	t.Run("top owner gets the tag added and cache upserted", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
			// 5 owners: top 20% = ceil(5*0.2) = 1 -> only revenues[0] (owner 1) is "top".
			return []repository.OwnerAnnualRevenue{{OwnerID: 1}, {OwnerID: 2}, {OwnerID: 3}, {OwnerID: 4}, {OwnerID: 5}}, nil
		}
		var addedTag string
		var upsertedTag string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error { addedTag = tagName; return nil },
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDFn: func(_ context.Context, _ uint64) ([]model.Owner, error) {
					return []model.Owner{{ID: 1, LineUserID: &lineUID1}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				upsertTagFn: func(_ context.Context, _, _ uint64, tagName, _, _ string) error {
					upsertedTag = tagName
					return nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Empty(t, errs)
		assert.Equal(t, 1, count)
		assert.Equal(t, ltvTop20Tag, addedTag)
		assert.Equal(t, ltvTop20Tag, upsertedTag)
	})

	t.Run("AddTag failure for a top owner is aggregated as an error and does not increment count", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
			return []repository.OwnerAnnualRevenue{{OwnerID: 1}}, nil
		}
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDFn: func(_ context.Context, _ uint64) ([]model.Owner, error) {
					return []model.Owner{{ID: 1, LineUserID: &lineUID1}}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("non-top owner with no cached tag counts as success without calling RemoveTag", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
			return nil, nil // topN=0, no owner is "top"
		}
		removeCalled := false
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, _ string) error { removeCalled = true; return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDFn: func(_ context.Context, _ uint64) ([]model.Owner, error) {
					return []model.Owner{{ID: 2, LineUserID: &lineUID2}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnersFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]*model.LstepTagCache, error) {
					return map[uint64][]*model.LstepTagCache{}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Empty(t, errs)
		assert.Equal(t, 1, count)
		assert.False(t, removeCalled, "owner without the cached tag must not trigger a RemoveTag call")
	})

	t.Run("non-top owner with cached tag has it removed", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
			return nil, nil
		}
		var removedTag string
		var deletedTag string
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, tagName string) error { removedTag = tagName; return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDFn: func(_ context.Context, _ uint64) ([]model.Owner, error) {
					return []model.Owner{{ID: 2, LineUserID: &lineUID2}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnersFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]*model.LstepTagCache, error) {
					return map[uint64][]*model.LstepTagCache{2: {{TagName: ltvTop20Tag}}}, nil
				},
				deleteTagFn: func(_ context.Context, _, _ uint64, tagName string) error { deletedTag = tagName; return nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Empty(t, errs)
		assert.Equal(t, 1, count)
		assert.Equal(t, ltvTop20Tag, removedTag)
		assert.Equal(t, ltvTop20Tag, deletedTag)
	})

	t.Run("tagCacheRepo.FindByOwners batch failure aborts the sync (fail-closed, not per-owner)", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
			return nil, nil
		}
		client := &mockLstepAPIClient{}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDFn: func(_ context.Context, _ uint64) ([]model.Owner, error) {
					return []model.Owner{{ID: 2, LineUserID: &lineUID2}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnersFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]*model.LstepTagCache, error) {
					return nil, errors.New("db error")
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("RemoveTag failure for a non-top owner is aggregated", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
			return nil, nil
		}
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDFn: func(_ context.Context, _ uint64) ([]model.Owner, error) {
					return []model.Owner{{ID: 2, LineUserID: &lineUID2}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnersFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]*model.LstepTagCache, error) {
					return map[uint64][]*model.LstepTagCache{2: {{TagName: ltvTop20Tag}}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})
}
