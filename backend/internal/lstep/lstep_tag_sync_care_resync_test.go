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

func TestResyncOwnerCheckupTags(t *testing.T) {
	lineUID := "U_resync"

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
			},
		}
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
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
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
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
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
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
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when buildClient fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return nil, errors.New("credentials error")
			},
		}
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
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
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return nil, nil
			},
		}
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when checkup lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return &mockLstepAPIClient{}, nil
			},
			checkupRepo: &mockCheckupRepository{
				findByOwnerIDFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
					return nil, errors.New("checkup db error")
				},
			},
		}
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when tag cache lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return &mockLstepAPIClient{}, nil
			},
			checkupRepo: &mockCheckupRepository{
				findByOwnerIDFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
					return nil, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("cache db error")
				},
			},
		}
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("success: rebuilds tag set, removing stale tags and adding fresh ones", func(t *testing.T) {
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
		checkupDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return client, nil
			},
			checkupRepo: &mockCheckupRepository{
				findByOwnerIDFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
					return []model.Checkup{{CheckupTypeID: 3, Date: checkupDate}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{
						{TagName: "checkup_done_3_2025-12"},  // stale — must be removed
						{TagName: "next_checkup_2025-06-01"}, // stale — must be removed
						{TagName: "cpm_core"},                // untouched
					}, nil
				},
			},
		}
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, removedTags, "checkup_done_3_2025-12")
		assert.Contains(t, removedTags, "next_checkup_2025-06-01")
		assert.NotContains(t, removedTags, "cpm_core")
		assert.Contains(t, addedTags, "checkup_done_3_2026-04")
	})

	t.Run("keeps a tag that already matches the rebuilt set without re-adding", func(t *testing.T) {
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
		checkupDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return client, nil
			},
			checkupRepo: &mockCheckupRepository{
				findByOwnerIDFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
					return []model.Checkup{{CheckupTypeID: 3, Date: checkupDate}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "checkup_done_3_2026-04"}}, nil
				},
			},
		}
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Empty(t, removedTags, "tag already matching the fresh set should not be removed")
		assert.Contains(t, addedTags, "checkup_done_3_2026-04", "AddTag is always called for every entry in the fresh set")
	})

	t.Run("removing a stale tag fails: recorded as api failure but continues", func(t *testing.T) {
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
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return client, nil
			},
			checkupRepo: &mockCheckupRepository{
				findByOwnerIDFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
					return nil, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "checkup_done_3_2025-12"}}, nil
				},
			},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
					incrementCalls++
					return 1, nil
				},
			},
		}
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Positive(t, incrementCalls)
	})

	t.Run("returns wrapped error when adding a fresh tag fails", func(t *testing.T) {
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error { return errors.New("add failed") },
		}
		checkupDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return client, nil
			},
			checkupRepo: &mockCheckupRepository{
				findByOwnerIDFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
					return []model.Checkup{{CheckupTypeID: 3, Date: checkupDate}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.ResyncOwnerCheckupTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})
}
