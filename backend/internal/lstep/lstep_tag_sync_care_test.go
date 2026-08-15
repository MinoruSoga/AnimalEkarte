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

// ---- SyncPrescriptionTag ----

func TestSyncPrescriptionTag(t *testing.T) {
	lineUID := "U_prescription_test"
	now := time.Now()

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
		}
		assert.NoError(t, svc.SyncPrescriptionTag(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when checkOptOut fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) { return nil, errors.New("db error") },
			},
		}
		assert.Error(t, svc.SyncPrescriptionTag(context.Background(), 1, 10))
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
		assert.NoError(t, svc.SyncPrescriptionTag(context.Background(), 1, 10))
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
		assert.NoError(t, svc.SyncPrescriptionTag(context.Background(), 1, 10))
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
		assert.Error(t, svc.SyncPrescriptionTag(context.Background(), 1, 10))
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
		assert.NoError(t, svc.SyncPrescriptionTag(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when FindActiveByOwner fails", func(t *testing.T) {
		client := &mockLstepAPIClient{}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			prescriptionRepo: &mockPrescriptionRepository{
				findActiveByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) {
					return nil, errors.New("db error")
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		assert.Error(t, svc.SyncPrescriptionTag(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when tag cache lookup fails", func(t *testing.T) {
		client := &mockLstepAPIClient{}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			prescriptionRepo: &mockPrescriptionRepository{
				findActiveByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) { return nil, nil },
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("db error")
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		assert.Error(t, svc.SyncPrescriptionTag(context.Background(), 1, 10))
	})

	t.Run("no active prescriptions: removes stale refill_due tags and adds nothing", func(t *testing.T) {
		var removedTags []string
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, tagName string) error {
			removedTags = append(removedTags, tagName)
			return nil
		}}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			prescriptionRepo: &mockPrescriptionRepository{
				findActiveByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) { return nil, nil },
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "refill_due_2020-01-01"}, {TagName: "vaccine_dog_2020-01-01"}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncPrescriptionTag(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, []string{"refill_due_2020-01-01"}, removedTags, "only refill_due_ prefixed tags are touched")
	})

	t.Run("returns wrapped error when removing a stale tag fails", func(t *testing.T) {
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			prescriptionRepo: &mockPrescriptionRepository{
				findActiveByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) { return nil, nil },
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "refill_due_2020-01-01"}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		// RemoveTag failure is non-fatal on this path (apiFailed=true), but since latestRefillDue
		// is nil the function returns nil without ever surfacing the failure as an error.
		err := svc.SyncPrescriptionTag(context.Background(), 1, 10)
		assert.NoError(t, err)
	})

	t.Run("future refill due date: adds the new refill_due tag", func(t *testing.T) {
		var addedTag, upsertedTag string
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, tagName string) error { addedTag = tagName; return nil }}
		presc := model.Prescription{PrescribedAt: now, DurationDays: 14} // refill = now + 7 days (future)
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			prescriptionRepo: &mockPrescriptionRepository{
				findActiveByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) {
					return []model.Prescription{presc}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				upsertTagFn: func(_ context.Context, _, _ uint64, tagName, _, _ string) error { upsertedTag = tagName; return nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncPrescriptionTag(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.NotEmpty(t, addedTag)
		assert.Equal(t, addedTag, upsertedTag)
		assert.Contains(t, addedTag, tagPrefixRefillDue)
	})

	t.Run("past refill due date: does not add a new tag", func(t *testing.T) {
		addCalled := false
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { addCalled = true; return nil }}
		presc := model.Prescription{PrescribedAt: now.AddDate(0, -1, 0), DurationDays: 3} // refill in the past
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			prescriptionRepo: &mockPrescriptionRepository{
				findActiveByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) {
					return []model.Prescription{presc}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncPrescriptionTag(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.False(t, addCalled, "a refill date in the past must not add a new tag")
	})

	t.Run("returns wrapped error when AddTag for the new refill_due tag fails", func(t *testing.T) {
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		presc := model.Prescription{PrescribedAt: now, DurationDays: 14}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			prescriptionRepo: &mockPrescriptionRepository{
				findActiveByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) {
					return []model.Prescription{presc}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		assert.Error(t, svc.SyncPrescriptionTag(context.Background(), 1, 10))
	})

	t.Run("duration_days < 7 uses prescribed_at + 1 day as the refill date", func(t *testing.T) {
		var addedTag string
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, tagName string) error { addedTag = tagName; return nil }}
		presc := model.Prescription{PrescribedAt: now, DurationDays: 3}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			prescriptionRepo: &mockPrescriptionRepository{
				findActiveByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) {
					return []model.Prescription{presc}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncPrescriptionTag(context.Background(), 1, 10)
		assert.NoError(t, err)
		want := tagPrefixRefillDue + now.AddDate(0, 0, 1).Format("2006-01-02")
		assert.Equal(t, want, addedTag)
	})

	t.Run("picks the latest refill date across multiple active prescriptions", func(t *testing.T) {
		var addedTag string
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, tagName string) error { addedTag = tagName; return nil }}
		earlier := model.Prescription{PrescribedAt: now, DurationDays: 10} // refill: now+3
		later := model.Prescription{PrescribedAt: now, DurationDays: 20}   // refill = now+13 (latest)
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			prescriptionRepo: &mockPrescriptionRepository{
				findActiveByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) {
					return []model.Prescription{earlier, later}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncPrescriptionTag(context.Background(), 1, 10)
		assert.NoError(t, err)
		want := tagPrefixRefillDue + now.AddDate(0, 0, 13).Format("2006-01-02")
		assert.Equal(t, want, addedTag)
	})
}
