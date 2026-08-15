package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestSyncVisitCompletionTags(t *testing.T) {
	lineUID := "U_visit"

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
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
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
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
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
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
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when visit summary lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{
				findOwnerVisitSummaryFn: func(_ context.Context, _, _ uint64) (*medicalrecord.OwnerVisitSummary, error) {
					return nil, errors.New("summary db error")
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when LTV sum lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo: &mockAccountingRepository{
				sumPaidByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return 0, errors.New("ltv db error")
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
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
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo:   &mockAccountingRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil },
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("success: adds fresh visit tags and clears dormant/legacy tags", func(t *testing.T) {
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
		first := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		last := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{
				findOwnerVisitSummaryFn: func(_ context.Context, _, _ uint64) (*medicalrecord.OwnerVisitSummary, error) {
					return &medicalrecord.OwnerVisitSummary{FirstVisitAt: &first, LastVisitAt: &last, AnnualCount: 6}, nil
				},
			},
			accountRepo: &mockAccountingRepository{
				sumPaidByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) { return 60_000, nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{
						{TagName: "dormant_180d"},
						{TagName: "VISIT_120日超"},
						{TagName: "cpm_dormant"},
					}, nil
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, "first_visit_2024-01-01")
		assert.Contains(t, addedTags, "last_visit_2026-04-01")
		assert.Contains(t, addedTags, "ltv_amount_5")
		assert.Contains(t, addedTags, "visit_count_annual_5")
		assert.Contains(t, removedTags, "dormant_180d")
		assert.Contains(t, removedTags, "VISIT_120日超")
		assert.Contains(t, removedTags, "cpm_dormant")
		assert.Contains(t, removedTags, "dormant")
		assert.Contains(t, removedTags, "noshow")
		assert.Contains(t, removedTags, "reserved")
	})

	t.Run("returns wrapped error when adding a visit tag fails", func(t *testing.T) {
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
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo:   &mockAccountingRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when loading tag cache for dormant cleanup fails", func(t *testing.T) {
		// First FindByOwner (stale cleanup) succeeds; second (dormant cleanup) fails after desired adds.
		findCalls := 0
		client := &mockLstepAPIClient{}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo:   &mockAccountingRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					findCalls++
					if findCalls == 1 {
						return nil, nil // stale cleanup: empty cache
					}
					return nil, errors.New("cache db error")
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("stale-tag cache load failure: zero AddTag and returns observable error", func(t *testing.T) {
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
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo:   &mockAccountingRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("cache db error")
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.Error(t, err)
		assert.Zero(t, addCalls, "cache failure must fail closed: zero AddTag")
		assert.Zero(t, removeCalls, "cache failure must fail closed: zero RemoveTag")
	})

	t.Run("stale RemoveTag partial failure: desired AddTag continues and failure is accounted", func(t *testing.T) {
		var addedTags []string
		incrementCalls := 0
		findCalls := 0
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
			removeTagFn: func(_ context.Context, _, tagName string) error {
				if tagName == "ltv_amount_0" {
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
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo: &mockAccountingRepository{
				sumPaidByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) { return 60_000, nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					findCalls++
					if findCalls == 1 {
						// stale cleanup: old LTV bracket to remove
						return []*model.LstepTagCache{{TagName: "ltv_amount_0"}}, nil
					}
					// dormant cleanup: none
					return nil, nil
				},
			},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
					incrementCalls++
					return 1, nil
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.NoError(t, err, "RemoveTag partial failure must not abort desired adds")
		assert.Contains(t, addedTags, "ltv_amount_5")
		assert.Positive(t, incrementCalls, "durable failure accounting must remain")
	})

	t.Run("removing a dormant tag fails: recorded as api failure but does not propagate", func(t *testing.T) {
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
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo:   &mockAccountingRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "dormant_180d"}}, nil
				},
			},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
					incrementCalls++
					return 1, nil
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Positive(t, incrementCalls)
	})
}
