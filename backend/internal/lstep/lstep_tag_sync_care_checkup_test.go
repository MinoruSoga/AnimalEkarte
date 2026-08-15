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

func TestSyncCheckupTag(t *testing.T) {
	lineUID := "U_checkup"
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	next := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
			},
		}
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, nil)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when shouldSkipSync fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("db down") },
			},
		}
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, nil)
		assert.Error(t, err)
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
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, nil)
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
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, nil)
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
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, nil)
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
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, nil)
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
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, nil)
		assert.NoError(t, err)
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
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("cache db error")
				},
			},
		}
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, nil)
		assert.Error(t, err)
	})

	t.Run("success: removes stale checkup/next tags and adds new ones", func(t *testing.T) {
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
		findCalled := false
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
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{
						{TagName: "checkup_done_3_2025-01"},
						{TagName: "next_checkup_2025-06-01"},
						{TagName: "checkup_done_9_2025-01"}, // different type — must NOT be removed
					}, nil
				},
			},
			errorCounterRepo: &mockErrorCounterRepo{
				findFn: func(_ context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
					findCalled = true
					return &model.LstepSyncErrorCounter{FailureCount: 0}, nil
				},
			},
		}
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, &next)
		assert.NoError(t, err)
		assert.Contains(t, removedTags, "checkup_done_3_2025-01")
		assert.Contains(t, removedTags, "next_checkup_2025-06-01")
		assert.NotContains(t, removedTags, "checkup_done_9_2025-01")
		assert.Contains(t, addedTags, "checkup_done_3_2026-05")
		assert.Contains(t, addedTags, "next_checkup_2026-08-01")
		assert.True(t, findCalled, "notifyAPISuccess should be invoked when no api failures occurred")
	})

	t.Run("next tag is not added when nextDate is nil", func(t *testing.T) {
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
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return client, nil
			},
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, nil)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, "checkup_done_3_2026-05")
		for _, tag := range addedTags {
			assert.NotContains(t, tag, "next_checkup_")
		}
	})

	t.Run("removing a stale tag fails: api failure recorded but sync continues", func(t *testing.T) {
		incrementCalled := false
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, _ string) error {
				return errors.New("lstep api down")
			},
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
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "checkup_done_3_2025-01"}}, nil
				},
			},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
					incrementCalled = true
					return 1, nil
				},
			},
		}
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, nil)
		assert.NoError(t, err, "stale removal failure is best-effort and does not fail the sync")
		assert.True(t, incrementCalled)
	})

	t.Run("returns wrapped error when adding checkup tag fails", func(t *testing.T) {
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
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return client, nil
			},
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, &next)
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when adding next_checkup tag fails", func(t *testing.T) {
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				if tagName == "next_checkup_2026-08-01" {
					return errors.New("add next failed")
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
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return client, nil
			},
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.SyncCheckupTag(context.Background(), 1, 2, 3, now, &next)
		assert.Error(t, err)
	})
}
