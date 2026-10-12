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

func TestSyncNextVisitTag(t *testing.T) {
	lineUID := "U_next_visit"

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
			},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
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
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
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
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
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
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
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
			tagCacheRepo: &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return nil, errors.New("credentials error")
			},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
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
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil },
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("returns error without touching tags when appointment repo is not configured", func(t *testing.T) {
		// EMR-308 fail-closed: 誤配線（cmd/lstep-migrate 等）で既存 next_visit_* が
		// 黙って剥奪されることを防ぐため、repo 未注入はエラーで返しタグに触れない。
		var addedTags, removedTags []string
		tagCacheLoaded := false
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
			buildClientFn:            func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			nextVisitAppointmentRepo: nil,
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					tagCacheLoaded = true
					return []*model.LstepTagCache{{TagName: "next_visit_2025-01-01"}}, nil
				},
			},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
		assert.Error(t, err)
		assert.False(t, tagCacheLoaded, "tag cache must not be read when the appointment repo is missing")
		assert.Empty(t, addedTags)
		assert.Empty(t, removedTags)
	})

	t.Run("returns wrapped error when next appointment lookup fails", func(t *testing.T) {
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
			nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{
				findNextAppointmentForOwnerFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return nil, errors.New("appointment db error")
				},
			},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
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
			nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("cache db error")
				},
			},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("no next appointment: removes stale next_visit tags and adds nothing", func(t *testing.T) {
		var addedTags, removedTags []string
		client := &mockLstepAPIClient{
			getUserTagsFn: func(context.Context, string) ([]string, error) { return []string{"next_visit_2025-01-01"}, nil },
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
			buildClientFn:            func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{
						{TagName: "next_visit_2025-01-01"},
						{TagName: "cpm_core"},
					}, nil
				},
			},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, removedTags, "next_visit_2025-01-01")
		assert.NotContains(t, removedTags, "cpm_core")
		assert.Empty(t, addedTags)
	})

	t.Run("tentative appointment without intended_date: removes stale tags, adds nothing", func(t *testing.T) {
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
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{
				findNextAppointmentForOwnerFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return &model.Reservation{ID: 5, Status: model.ReservationStatusTentative, IntendedDate: nil}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Empty(t, addedTags)
	})

	t.Run("success: adds new next_visit tag from confirmed appointment start_time", func(t *testing.T) {
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
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{
				findNextAppointmentForOwnerFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return &model.Reservation{
						ID:        5,
						Status:    model.ReservationStatusConfirmed,
						StartTime: time.Date(2026, 9, 15, 10, 30, 0, 0, time.UTC),
					}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, "next_visit_2026-09-15")
	})

	t.Run("success: adds new next_visit tag from tentative appointment intended_date", func(t *testing.T) {
		var addedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
		}
		intended := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{
				findNextAppointmentForOwnerFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return &model.Reservation{ID: 6, Status: model.ReservationStatusTentative, IntendedDate: &intended}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, "next_visit_2026-10-01")
	})

	t.Run("removing a stale next_visit tag fails: reported for batch retry", func(t *testing.T) {
		incrementCalls := 0
		client := &mockLstepAPIClient{
			getUserTagsFn: func(context.Context, string) ([]string, error) { return []string{"next_visit_2025-01-01"}, nil },
			removeTagFn:   func(_ context.Context, _, _ string) error { return errors.New("remove failed") },
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			buildClientFn:            func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "next_visit_2025-01-01"}}, nil
				},
			},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
					incrementCalls++
					return 1, nil
				},
			},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
		assert.Error(t, err)
		assert.Positive(t, incrementCalls)
	})

	t.Run("returns wrapped error when adding next_visit tag fails", func(t *testing.T) {
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
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{
				findNextAppointmentForOwnerFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return &model.Reservation{
						ID:        5,
						Status:    model.ReservationStatusConfirmed,
						StartTime: time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC),
					}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
		assert.Error(t, err)
	})
}
