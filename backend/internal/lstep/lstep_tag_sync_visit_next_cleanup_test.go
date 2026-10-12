package lstep

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

type narrowTagCacheRepo struct{}

func (*narrowTagCacheRepo) UpsertTag(context.Context, uint64, uint64, string, string, string) error {
	return nil
}
func (*narrowTagCacheRepo) DeleteTag(context.Context, uint64, uint64, string) error { return nil }
func (*narrowTagCacheRepo) DeleteAllByOwner(context.Context, uint64, uint64) error  { return nil }
func (*narrowTagCacheRepo) FindByOwner(context.Context, uint64, uint64) ([]*model.LstepTagCache, error) {
	return nil, nil
}
func (*narrowTagCacheRepo) FindByOwners(context.Context, uint64, []uint64) (map[uint64][]*model.LstepTagCache, error) {
	return map[uint64][]*model.LstepTagCache{}, nil
}
func (*narrowTagCacheRepo) TagSummary(context.Context, uint64) ([]TagSummaryRow, int64, error) {
	return nil, 0, nil
}
func (*narrowTagCacheRepo) FindOwnersByTag(context.Context, uint64, string, string, int, int) ([]TagOwnerRow, int64, error) {
	return nil, 0, nil
}
func (*narrowTagCacheRepo) FindOwnerIDsByTag(context.Context, uint64, string) ([]uint64, error) {
	return nil, nil
}

type nextVisitCleanupSpy struct {
	mu             sync.Mutex
	getCalls       int
	addCalls       int
	finderCalls    int
	removed        []string
	cacheDeleted   []string
	ackRevisions   []uint64
	appointmentHit bool
}

func (s *nextVisitCleanupSpy) recordRemove(tag string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removed = append(s.removed, tag)
}
func (s *nextVisitCleanupSpy) recordCacheDelete(tag string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cacheDeleted = append(s.cacheDeleted, tag)
}
func (s *nextVisitCleanupSpy) recordAck(revision uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ackRevisions = append(s.ackRevisions, revision)
}

func deletedCleanupOwner(t *testing.T, clinicID, ownerID uint64, lineUserID string, optOut bool) *model.Owner {
	t.Helper()
	uid := lineUserID
	return &model.Owner{
		ID:          ownerID,
		ClinicID:    clinicID,
		LineUserID:  &uid,
		LstepOptOut: optOut,
		DeletedAt:   gorm.DeletedAt{Time: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), Valid: true},
	}
}

type cleanupSyncFixture struct {
	spy    *nextVisitCleanupSpy
	svc    *lstepTagSyncService
	client *mockLstepAPIClient
}

func newCleanupSyncFixture(
	t *testing.T,
	intent *NextVisitCleanupIntent,
	owner *model.Owner,
	remote []string,
	cached []*model.LstepTagCache,
) *cleanupSyncFixture {
	t.Helper()
	spy := &nextVisitCleanupSpy{}
	var remoteMu sync.Mutex
	remoteSet := append([]string(nil), remote...)
	cache := &mockLstepTagCacheRepository{
		findNextVisitCleanupIntentFn: func(_ context.Context, clinicID, ownerID uint64) (*NextVisitCleanupIntent, error) {
			spy.mu.Lock()
			defer spy.mu.Unlock()
			spy.finderCalls++
			return intent, nil
		},
		findOwnerForNextVisitCleanupFn: func(context.Context, uint64, uint64) (*model.Owner, error) {
			return owner, nil
		},
		withNextVisitOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
			return fn(ctx)
		},
		findByOwnerFn: func(context.Context, uint64, uint64) ([]*model.LstepTagCache, error) {
			return cached, nil
		},
		deleteTagFn: func(_ context.Context, _ uint64, _ uint64, tag string) error {
			spy.recordCacheDelete(tag)
			return nil
		},
		acknowledgeNextVisitSyncFn: func(_ context.Context, _, _ uint64, revision uint64) error {
			spy.recordAck(revision)
			return nil
		},
	}
	client := &mockLstepAPIClient{
		getUserTagsFn: func(context.Context, string) ([]string, error) {
			remoteMu.Lock()
			defer remoteMu.Unlock()
			spy.getCalls++
			return append([]string(nil), remoteSet...), nil
		},
		addTagFn: func(context.Context, string, string) error {
			spy.addCalls++
			t.Error("cleanup must never add tags")
			return nil
		},
		removeTagFn: func(_ context.Context, _ string, tag string) error {
			spy.recordRemove(tag)
			remoteMu.Lock()
			defer remoteMu.Unlock()
			for i, existing := range remoteSet {
				if existing == tag {
					remoteSet = append(remoteSet[:i], remoteSet[i+1:]...)
				}
			}
			return nil
		},
	}
	svc := &lstepTagSyncService{
		settingsSvc:  &mockLstepSettingsService{},
		ownerRepo:    &mockOwnerRepository{},
		tagCacheRepo: cache,
		nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
			spy.appointmentHit = true
			return nil, errors.New("cleanup must never read appointments")
		}},
		buildClientFn: func(context.Context, uint64) (lstepapi.Client, error) { return client, nil },
	}
	return &cleanupSyncFixture{spy: spy, svc: svc, client: client}
}

func TestNextVisitCleanup_RemoveOnlyConvergesAndAcknowledges(t *testing.T) {
	intent := &NextVisitCleanupIntent{ClinicID: 1, OwnerID: 2, Revision: 9, LineUserID: "Udeleted"}
	owner := deletedCleanupOwner(t, 1, 2, "Udeleted", false)
	fx := newCleanupSyncFixture(t, intent, owner,
		[]string{"next_visit_2026-11-01", "next_visit_2026-10-15", "friend_rank_a"},
		[]*model.LstepTagCache{
			{TagName: "next_visit_2026-10-15"},
			{TagName: "next_visit_2026-09-30"},
			{TagName: "friend_rank_a"},
		})

	require.NoError(t, fx.svc.SyncNextVisitTag(context.Background(), 1, 2))
	assert.False(t, fx.spy.appointmentHit, "cleanup must not use the appointment finder")
	assert.Zero(t, fx.spy.addCalls)
	assert.ElementsMatch(t,
		[]string{"next_visit_2026-10-15", "next_visit_2026-11-01"},
		fx.spy.removed,
		"only remote-present next_visit tags are removed from the provider")
	assert.ElementsMatch(t,
		[]string{"next_visit_2026-10-15", "next_visit_2026-11-01", "next_visit_2026-09-30"},
		fx.spy.cacheDeleted,
		"local stale next_visit rows are cleared even when already absent remotely")
	assert.Equal(t, []uint64{9}, fx.spy.ackRevisions)
}

func TestNextVisitCleanup_FailuresRetainIntent(t *testing.T) {
	intent := &NextVisitCleanupIntent{ClinicID: 1, OwnerID: 2, Revision: 7, LineUserID: "Udeleted"}
	owner := deletedCleanupOwner(t, 1, 2, "Udeleted", false)
	remote := []string{"next_visit_2026-11-01", "next_visit_2026-10-15", "friend_rank_a"}

	for _, tc := range []struct {
		name    string
		mutate  func(fx *cleanupSyncFixture, failure error)
		wantAck bool
	}{
		{"provider GetUserTags failure", func(fx *cleanupSyncFixture, failure error) {
			fx.client = &mockLstepAPIClient{getUserTagsFn: func(context.Context, string) ([]string, error) {
				return nil, failure
			}}
			fx.svc.buildClientFn = func(context.Context, uint64) (lstepapi.Client, error) { return fx.client, nil }
		}, false},
		{"provider RemoveTag failure on one of two tags", func(fx *cleanupSyncFixture, failure error) {
			failed := false
			fx.client = &mockLstepAPIClient{
				getUserTagsFn: func(context.Context, string) ([]string, error) { return remote, nil },
				removeTagFn: func(_ context.Context, _ string, tag string) error {
					if tag == "next_visit_2026-10-15" && !failed {
						failed = true
						return failure
					}
					fx.spy.recordRemove(tag)
					return nil
				},
				addTagFn: func(context.Context, string, string) error {
					fx.spy.addCalls++
					t.Error("cleanup must never add tags")
					return nil
				},
			}
			fx.svc.buildClientFn = func(context.Context, uint64) (lstepapi.Client, error) { return fx.client, nil }
		}, false},
		{"cache delete failure", func(fx *cleanupSyncFixture, failure error) {
			fx.svc.tagCacheRepo.(*mockLstepTagCacheRepository).deleteTagFn = func(context.Context, uint64, uint64, string) error {
				return failure
			}
		}, false},
		{"acknowledge failure", func(fx *cleanupSyncFixture, failure error) {
			fx.svc.tagCacheRepo.(*mockLstepTagCacheRepository).acknowledgeNextVisitSyncFn = func(context.Context, uint64, uint64, uint64) error {
				return failure
			}
		}, false},
		{"lock transaction commit failure", func(fx *cleanupSyncFixture, failure error) {
			fx.svc.tagCacheRepo.(*mockLstepTagCacheRepository).withNextVisitOwnerLockFn = func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
				if err := fn(ctx); err != nil {
					return err
				}
				return failure
			}
		}, true},
		{"cancellation during cleanup", func(fx *cleanupSyncFixture, failure error) {
			fx.svc.tagCacheRepo.(*mockLstepTagCacheRepository).withNextVisitOwnerLockFn = func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
				return context.Canceled
			}
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := newCleanupSyncFixture(t, intent, owner, remote,
				[]*model.LstepTagCache{{TagName: "next_visit_2026-10-15"}})
			failure := errors.New(tc.name)
			tc.mutate(fx, failure)

			err := fx.svc.SyncNextVisitTag(context.Background(), 1, 2)
			require.Error(t, err, "cleanup failure must surface so the intent stays discoverable")
			assert.Zero(t, fx.spy.addCalls)
			if !tc.wantAck {
				assert.Empty(t, fx.spy.ackRevisions, "a failed cleanup must not acknowledge the intent")
			}
		})
	}
}

func TestNextVisitCleanup_RetryConvergesAfterPartialRemoval(t *testing.T) {
	intent := &NextVisitCleanupIntent{ClinicID: 1, OwnerID: 2, Revision: 5, LineUserID: "Udeleted"}
	owner := deletedCleanupOwner(t, 1, 2, "Udeleted", false)
	fx := newCleanupSyncFixture(t, intent, owner,
		[]string{"next_visit_2026-11-01", "next_visit_2026-10-15"},
		[]*model.LstepTagCache{
			{TagName: "next_visit_2026-10-15"},
			{TagName: "next_visit_2026-11-01"},
		})
	cache := fx.svc.tagCacheRepo.(*mockLstepTagCacheRepository)
	cache.deleteTagFn = func(context.Context, uint64, uint64, string) error {
		return errors.New("cache write lost")
	}

	require.Error(t, fx.svc.SyncNextVisitTag(context.Background(), 1, 2))
	assert.NotEmpty(t, fx.spy.removed, "external removals may partially complete before the local failure")
	assert.Empty(t, fx.spy.ackRevisions)

	cache.deleteTagFn = func(_ context.Context, _ uint64, _ uint64, tag string) error {
		fx.spy.recordCacheDelete(tag)
		return nil
	}
	require.NoError(t, fx.svc.SyncNextVisitTag(context.Background(), 1, 2))
	assert.Equal(t, []uint64{5}, fx.spy.ackRevisions, "a later run converges and acknowledges the same revision")
}

func TestNextVisitCleanup_RelinkedRecipientDoesNoProviderWork(t *testing.T) {
	intent := &NextVisitCleanupIntent{ClinicID: 1, OwnerID: 2, Revision: 3, LineUserID: "Ureassigned"}
	owner := deletedCleanupOwner(t, 1, 2, "Ureassigned", false)
	fx := newCleanupSyncFixture(t, intent, owner, nil, nil)
	cache := fx.svc.tagCacheRepo.(*mockLstepTagCacheRepository)
	cache.countActiveOwnersWithLineUserFn = func(_ context.Context, clinicID uint64, lineUserID string) (int64, error) {
		assert.Equal(t, uint64(1), clinicID)
		assert.Equal(t, "Ureassigned", lineUserID)
		return 1, nil
	}
	providerCalled := false
	fx.svc.buildClientFn = func(context.Context, uint64) (lstepapi.Client, error) {
		return &mockLstepAPIClient{getUserTagsFn: func(context.Context, string) ([]string, error) {
			providerCalled = true
			return nil, nil
		}}, nil
	}

	err := fx.svc.SyncNextVisitTag(context.Background(), 1, 2)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "relinked recipient must conflict, got %v", err)
	assert.False(t, providerCalled, "a relinked recipient must never reach the provider")
	assert.Empty(t, fx.spy.ackRevisions)
}

func TestNextVisitCleanup_GuardsRetainIntent(t *testing.T) {
	intent := &NextVisitCleanupIntent{ClinicID: 1, OwnerID: 2, Revision: 4, LineUserID: "Udeleted"}
	fatalClient := func(providerCalled *bool) lstepapi.Client {
		return &mockLstepAPIClient{getUserTagsFn: func(context.Context, string) ([]string, error) {
			*providerCalled = true
			return nil, nil
		}}
	}

	t.Run("owner still active is a conflict without provider work", func(t *testing.T) {
		owner := deletedCleanupOwner(t, 1, 2, "Udeleted", false)
		owner.DeletedAt = gorm.DeletedAt{}
		fx := newCleanupSyncFixture(t, intent, owner, nil, nil)
		providerCalled := false
		fx.svc.buildClientFn = func(context.Context, uint64) (lstepapi.Client, error) {
			return fatalClient(&providerCalled), nil
		}
		err := fx.svc.SyncNextVisitTag(context.Background(), 1, 2)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.False(t, providerCalled)
	})

	t.Run("recipient mismatch with locked owner is a conflict", func(t *testing.T) {
		owner := deletedCleanupOwner(t, 1, 2, "Udifferent", false)
		fx := newCleanupSyncFixture(t, intent, owner, nil, nil)
		providerCalled := false
		fx.svc.buildClientFn = func(context.Context, uint64) (lstepapi.Client, error) {
			return fatalClient(&providerCalled), nil
		}
		err := fx.svc.SyncNextVisitTag(context.Background(), 1, 2)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.False(t, providerCalled)
	})

	t.Run("deleted owner who opted out is skipped with intent retained", func(t *testing.T) {
		owner := deletedCleanupOwner(t, 1, 2, "Udeleted", true)
		fx := newCleanupSyncFixture(t, intent, owner, nil, nil)
		providerCalled := false
		fx.svc.buildClientFn = func(context.Context, uint64) (lstepapi.Client, error) {
			return fatalClient(&providerCalled), nil
		}
		require.NoError(t, fx.svc.SyncNextVisitTag(context.Background(), 1, 2))
		assert.False(t, providerCalled)
		assert.Empty(t, fx.spy.ackRevisions)
	})

	t.Run("unavailable client retains the intent", func(t *testing.T) {
		owner := deletedCleanupOwner(t, 1, 2, "Udeleted", false)
		fx := newCleanupSyncFixture(t, intent, owner, nil, nil)
		fx.svc.buildClientFn = func(context.Context, uint64) (lstepapi.Client, error) { return nil, nil }
		require.NoError(t, fx.svc.SyncNextVisitTag(context.Background(), 1, 2))
		assert.Zero(t, fx.spy.getCalls)
		assert.Empty(t, fx.spy.ackRevisions)
	})

	t.Run("disabled sync retains the intent before repository discovery", func(t *testing.T) {
		owner := deletedCleanupOwner(t, 1, 2, "Udeleted", false)
		fx := newCleanupSyncFixture(t, intent, owner, nil, nil)
		fx.svc.settingsSvc = &mockLstepSettingsService{
			isSyncEnabledFn: func(context.Context, uint64) (bool, error) { return false, nil },
		}
		require.NoError(t, fx.svc.SyncNextVisitTag(context.Background(), 1, 2))
		assert.Zero(t, fx.spy.finderCalls, "disabled clinics skip before any repository access")
	})

	t.Run("nil intent inside the locked callback fails closed", func(t *testing.T) {
		owner := deletedCleanupOwner(t, 1, 2, "Udeleted", false)
		fx := newCleanupSyncFixture(t, intent, owner, nil, nil)
		calls := 0
		fx.svc.tagCacheRepo.(*mockLstepTagCacheRepository).findNextVisitCleanupIntentFn = func(context.Context, uint64, uint64) (*NextVisitCleanupIntent, error) {
			calls++
			if calls == 1 {
				return intent, nil
			}
			return nil, nil
		}
		providerCalled := false
		fx.svc.buildClientFn = func(context.Context, uint64) (lstepapi.Client, error) {
			return fatalClient(&providerCalled), nil
		}
		err := fx.svc.SyncNextVisitTag(context.Background(), 1, 2)
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, "INTERNAL", appErr.Code)
		assert.False(t, providerCalled)
	})

	t.Run("nil owner inside the locked callback fails closed", func(t *testing.T) {
		fx := newCleanupSyncFixture(t, intent, nil, nil, nil)
		providerCalled := false
		fx.svc.buildClientFn = func(context.Context, uint64) (lstepapi.Client, error) {
			return fatalClient(&providerCalled), nil
		}
		err := fx.svc.SyncNextVisitTag(context.Background(), 1, 2)
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, "INTERNAL", appErr.Code)
		assert.False(t, providerCalled)
	})
}

func TestNextVisitCleanup_MissingIntentFallsThroughToNormalSync(t *testing.T) {
	owner := &model.Owner{ID: 2, ClinicID: 1}
	uid := "Uactive"
	owner.LineUserID = &uid
	fx := newCleanupSyncFixture(t, nil, owner, nil, nil)
	fx.svc.ownerRepo = &mockOwnerRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Owner, error) {
		return owner, nil
	}}
	fx.svc.nextVisitAppointmentRepo = &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
		fx.spy.appointmentHit = true
		return nil, nil
	}}
	require.NoError(t, fx.svc.SyncNextVisitTag(context.Background(), 1, 2))
	assert.True(t, fx.spy.appointmentHit, "no cleanup intent means the normal appointment path runs")
}

func TestNextVisitCleanup_MisconfigurationFailsClosed(t *testing.T) {
	t.Run("cleanup finder not configured", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc:  &mockLstepSettingsService{},
			tagCacheRepo: &narrowTagCacheRepo{},
		}
		err := svc.SyncNextVisitTag(context.Background(), 1, 2)
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, "INTERNAL", appErr.Code)
	})

	t.Run("cleanup queue not configured", func(t *testing.T) {
		svc := &lstepTagSyncService{tagCacheRepo: &narrowTagCacheRepo{}}
		err := svc.EnqueueNextVisitCleanup(context.Background(), 1, 2, "Utarget")
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, "INTERNAL", appErr.Code)
	})

	t.Run("enqueue delegates to the configured queue", func(t *testing.T) {
		var got struct {
			clinicID, ownerID uint64
			lineUserID        string
		}
		svc := &lstepTagSyncService{tagCacheRepo: &mockLstepTagCacheRepository{
			enqueueNextVisitCleanupFn: func(_ context.Context, clinicID, ownerID uint64, lineUserID string) error {
				got.clinicID, got.ownerID, got.lineUserID = clinicID, ownerID, lineUserID
				return nil
			},
		}}
		require.NoError(t, svc.EnqueueNextVisitCleanup(context.Background(), 1, 2, "Utarget"))
		assert.Equal(t, uint64(1), got.clinicID)
		assert.Equal(t, uint64(2), got.ownerID)
		assert.Equal(t, "Utarget", got.lineUserID)
	})
}

type exclNotifyFixture struct {
	svc        *lstepTagSyncService
	counter    *mockErrorCounterRepo
	ownerRepo  *mockOwnerRepository
	cache      *mockLstepTagCacheRepository
	client     *mockLstepAPIClient
	addCalls   int
	addedTags  []string
	removeTags []string
	upserts    []string
	cacheDels  []string
	lockCalls  int
}

func newExclNotifyFixture(t *testing.T) *exclNotifyFixture {
	t.Helper()
	fx := &exclNotifyFixture{}
	uid := "Uoriginal"
	fx.ownerRepo = &mockOwnerRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Owner, error) {
		return &model.Owner{ID: 2, ClinicID: 1, LineUserID: &uid}, nil
	}}
	fx.cache = &mockLstepTagCacheRepository{
		withNextVisitOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
			fx.lockCalls++
			return fn(ctx)
		},
		upsertTagFn: func(_ context.Context, _, _ uint64, tagName, _, _ string) error {
			fx.upserts = append(fx.upserts, tagName)
			return nil
		},
		deleteTagFn: func(_ context.Context, _, _ uint64, tagName string) error {
			fx.cacheDels = append(fx.cacheDels, tagName)
			return nil
		},
	}
	fx.client = &mockLstepAPIClient{
		addTagFn: func(_ context.Context, _ string, tag string) error {
			fx.addCalls++
			fx.addedTags = append(fx.addedTags, tag)
			return nil
		},
		removeTagFn: func(_ context.Context, _ string, tag string) error {
			fx.removeTags = append(fx.removeTags, tag)
			return nil
		},
	}
	fx.svc = &lstepTagSyncService{
		settingsSvc:   &mockLstepSettingsService{},
		ownerRepo:     fx.ownerRepo,
		tagCacheRepo:  fx.cache,
		buildClientFn: func(context.Context, uint64) (lstepapi.Client, error) { return fx.client, nil },
	}
	return fx
}

func TestNextVisitNotify_APIFailureRevalidatesTargetInsideLock(t *testing.T) {
	other := "Udifferent"

	for _, tc := range []struct {
		name        string
		incrementFn func(*exclNotifyFixture) (int, error)
		wantAdd     int
		wantLocks   int
	}{
		{"threshold met and recipient unchanged adds EXCL inside the lock", func(*exclNotifyFixture) (int, error) {
			return lstepSyncErrorThreshold, nil
		}, 1, 1},
		{"owner deleted between the failure and the locked callback adds nothing", func(fx *exclNotifyFixture) (int, error) {
			fx.ownerRepo.findByIDFn = func(context.Context, uint64, uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "2")
			}
			return lstepSyncErrorThreshold, nil
		}, 0, 1},
		{"relinked recipient between the failure and the locked callback adds nothing", func(fx *exclNotifyFixture) (int, error) {
			fx.ownerRepo.findByIDFn = func(context.Context, uint64, uint64) (*model.Owner, error) {
				return &model.Owner{ID: 2, ClinicID: 1, LineUserID: &other}, nil
			}
			return lstepSyncErrorThreshold, nil
		}, 0, 1},
		{"below threshold never reacquires the lock", func(*exclNotifyFixture) (int, error) {
			return lstepSyncErrorThreshold - 1, nil
		}, 0, 0},
		{"increment failure skips the protected mutation", func(*exclNotifyFixture) (int, error) {
			return 0, errors.New("counter write failed")
		}, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := newExclNotifyFixture(t)
			fx.counter = &mockErrorCounterRepo{incrementFn: func(context.Context, uint64, uint64) (int, error) {
				return tc.incrementFn(fx)
			}}
			fx.svc.errorCounterRepo = fx.counter

			fx.svc.notifyNextVisitAPIFailure(context.Background(), fx.client, 1, 2, "Uoriginal")

			assert.Equal(t, tc.wantLocks, fx.lockCalls)
			assert.Equal(t, tc.wantAdd, fx.addCalls)
			if tc.wantAdd == 0 {
				assert.Empty(t, fx.upserts, "no EXCL cache write without a provider add")
			} else {
				assert.Equal(t, []string{lstepErrorTag}, fx.addedTags)
				assert.Equal(t, []string{lstepErrorTag}, fx.upserts)
			}
		})
	}
}

func TestNextVisitNotify_APISuccessRevalidatesTargetInsideLock(t *testing.T) {
	other := "Udifferent"
	counterWithFailures := &model.LstepSyncErrorCounter{FailureCount: 3}

	for _, tc := range []struct {
		name       string
		mutate     func(*exclNotifyFixture)
		wantLocks  int
		wantRemove int
		wantReset  bool
	}{
		{"unchanged recipient removes EXCL and resets the counter", nil, 1, 1, true},
		{"owner deleted before the locked callback removes nothing", func(fx *exclNotifyFixture) {
			fx.ownerRepo.findByIDFn = func(context.Context, uint64, uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "2")
			}
		}, 1, 0, false},
		{"relinked recipient before the locked callback removes nothing", func(fx *exclNotifyFixture) {
			fx.ownerRepo.findByIDFn = func(context.Context, uint64, uint64) (*model.Owner, error) {
				return &model.Owner{ID: 2, ClinicID: 1, LineUserID: &other}, nil
			}
		}, 1, 0, false},
		{"provider remove failure keeps the counter", func(fx *exclNotifyFixture) {
			fx.client.removeTagFn = func(context.Context, string, string) error {
				return errors.New("remove failed")
			}
		}, 1, 0, false},
		{"zero counter does no work", func(fx *exclNotifyFixture) {
			fx.counter.findFn = func(context.Context, uint64, uint64) (*model.LstepSyncErrorCounter, error) {
				return &model.LstepSyncErrorCounter{FailureCount: 0}, nil
			}
		}, 0, 0, false},
		{"missing counter does no work", func(fx *exclNotifyFixture) {
			fx.counter.findFn = func(context.Context, uint64, uint64) (*model.LstepSyncErrorCounter, error) {
				return nil, apperrors.WrapNotFound("lstep sync error counter", "none")
			}
		}, 0, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := newExclNotifyFixture(t)
			resetCalls := 0
			fx.counter = &mockErrorCounterRepo{
				findFn: func(context.Context, uint64, uint64) (*model.LstepSyncErrorCounter, error) {
					return counterWithFailures, nil
				},
				resetFn: func(context.Context, uint64, uint64) error {
					resetCalls++
					return nil
				},
			}
			fx.svc.errorCounterRepo = fx.counter
			if tc.mutate != nil {
				tc.mutate(fx)
			}

			fx.svc.notifyNextVisitAPISuccess(context.Background(), fx.client, 1, 2, "Uoriginal")

			assert.Equal(t, tc.wantLocks, fx.lockCalls)
			if tc.wantRemove == 0 {
				assert.Empty(t, fx.removeTags)
				assert.Empty(t, fx.cacheDels)
			} else {
				assert.Equal(t, []string{lstepErrorTag}, fx.removeTags)
				assert.Equal(t, []string{lstepErrorTag}, fx.cacheDels)
			}
			if tc.wantReset {
				assert.Equal(t, 1, resetCalls)
			} else {
				assert.Zero(t, resetCalls, "the counter only resets after a successful protected removal")
			}
		})
	}
}
