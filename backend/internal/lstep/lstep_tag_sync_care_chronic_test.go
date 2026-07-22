package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestSyncChronicConditionTags(t *testing.T) {
	lineUID := "U_chronic"

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
			},
		}
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
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
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
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
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
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
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
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
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
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
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when tagConfigRepo lookup fails", func(t *testing.T) {
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
			tagConfigRepo: &mockLstepTagConfigRepository{
				findAllConditionTagMappingsFn: func(_ context.Context) ([]*model.LstepConditionTagMapping, error) {
					return nil, errors.New("config db error")
				},
			},
		}
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
		assert.Error(t, err)
	})

	t.Run("uses empty condition map when tagConfigRepo is nil", func(t *testing.T) {
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
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			tagConfigRepo: nil,
		}
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
		assert.NoError(t, err)
		assert.Empty(t, addedTags, "unmapped condition codes yield no tags")
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
			tagConfigRepo: &mockLstepTagConfigRepository{
				findAllConditionTagMappingsFn: func(_ context.Context) ([]*model.LstepConditionTagMapping, error) {
					return []*model.LstepConditionTagMapping{{ConditionCode: "ckd", TagName: "chronic_ckd"}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("cache db error")
				},
			},
		}
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
		assert.Error(t, err)
	})

	t.Run("success: adds active tags, removes inactive tags, ignores non-chronic tags", func(t *testing.T) {
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
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return client, nil
			},
			tagConfigRepo: &mockLstepTagConfigRepository{
				findAllConditionTagMappingsFn: func(_ context.Context) ([]*model.LstepConditionTagMapping, error) {
					return []*model.LstepConditionTagMapping{
						{ConditionCode: "ckd", TagName: "chronic_ckd"},
						{ConditionCode: "heart", TagName: "chronic_heart"},
					}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{
						{TagName: "chronic_heart"}, // no longer active — should be removed
						{TagName: "cpm_core"},      // non-chronic tag — must be left untouched
						{TagName: "chronic_ckd"},   // still active — should stay (no add call)
					}, nil
				},
			},
		}
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
		assert.NoError(t, err)
		assert.Contains(t, removedTags, "chronic_heart")
		assert.NotContains(t, removedTags, "cpm_core")
		assert.NotContains(t, addedTags, "chronic_ckd", "already-cached active tag should not be re-added")
	})

	t.Run("adds a brand new active tag not present in cache", func(t *testing.T) {
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
			tagConfigRepo: &mockLstepTagConfigRepository{
				findAllConditionTagMappingsFn: func(_ context.Context) ([]*model.LstepConditionTagMapping, error) {
					return []*model.LstepConditionTagMapping{{ConditionCode: "ckd", TagName: "chronic_ckd"}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
		assert.NoError(t, err)
		assert.Contains(t, addedTags, "chronic_ckd")
	})

	t.Run("api failures on remove/add are recorded but do not fail the sync", func(t *testing.T) {
		incrementCalls := 0
		client := &mockLstepAPIClient{
			addTagFn:    func(_ context.Context, _, _ string) error { return errors.New("add failed") },
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
			tagConfigRepo: &mockLstepTagConfigRepository{
				findAllConditionTagMappingsFn: func(_ context.Context) ([]*model.LstepConditionTagMapping, error) {
					return []*model.LstepConditionTagMapping{{ConditionCode: "ckd", TagName: "chronic_ckd"}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "chronic_liver"}}, nil
				},
			},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
					incrementCalls++
					return 1, nil
				},
			},
		}
		err := svc.SyncChronicConditionTags(context.Background(), 1, 2, []string{"ckd"})
		assert.NoError(t, err)
		assert.Positive(t, incrementCalls)
	})
}
