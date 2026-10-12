package lstep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	ownerdomain "github.com/animal-ekarte/backend/internal/owner"
)

type cleanupDeleteService struct {
	ownerdomain.Service
	deleted bool
}

func (s *cleanupDeleteService) Delete(_ context.Context, clinicID, ownerID uint64) error {
	s.deleted = clinicID == 1 && ownerID == 2
	return nil
}

func TestOwnerDeletionHTTP_RecipientChangedWhileWaitingDoesNoProviderWork(t *testing.T) {
	var providerCalls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"tags":["next_visit_2026-11-01"]}`))
		assert.NoError(t, err)
	}))
	defer provider.Close()
	waiting, release := make(chan struct{}), make(chan struct{})
	uid := "Uold"
	locked := false
	reads := 0
	owners := &mockOwnerRepository{findByIDFn: func(_ context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
		assert.Equal(t, uint64(1), clinicID)
		assert.Equal(t, uint64(2), ownerID)
		reads++
		if reads > 1 {
			require.True(t, locked)
		}
		current := uid
		return &model.Owner{ID: ownerID, ClinicID: clinicID, LineUserID: &current}, nil
	}}
	cache := &mockLstepTagCacheRepository{withNextVisitOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
		close(waiting)
		<-release // A/U snapshot, then A/V and B/U committed before lock acquisition.
		uid = "Vnew"
		locked = true
		defer func() { locked = false }()
		return fn(ctx)
	}, deleteTagFn: func(context.Context, uint64, uint64, string) error {
		t.Error("stale recipient must not delete cache")
		return nil
	}}
	lifecycle := &lstepLifecycleService{ownerRepo: owners, tagCacheRepo: cache, auditSvc: &mockAuditService{}, settingsSvc: &mockLstepSettingsService{
		getRawCredentialsFn: func(context.Context, uint64) (string, string, string, error) {
			return "mock-only", provider.URL, "", nil
		},
	}}
	svc := &cleanupDeleteService{}
	handler := ownerdomain.NewHandler(svc, lifecycle, nil, nil)
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Set("clinic_id", "1")
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/v1/owners/2", nil)
	done := make(chan struct{})
	go func() { handler.DeleteOwner(c); c.Writer.WriteHeaderNow(); close(done) }()
	select {
	case <-waiting:
	case <-time.After(time.Second):
		t.Fatal("cleanup did not wait for lock")
	}
	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("delete did not finish")
	}
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.True(t, svc.deleted)
	assert.Equal(t, 2, reads)
	assert.Zero(t, providerCalls.Load(), "old recipient now belongs to B")
}

func TestNextVisitCleanup_SettingsChangedWhileWaitingRetainsIntent(t *testing.T) {
	for _, failure := range []bool{false, true} {
		name := "disabled"
		if failure {
			name = "settings read failed"
		}
		t.Run(name, func(t *testing.T) {
			intent := &NextVisitCleanupIntent{ClinicID: 1, OwnerID: 2, Revision: 9, LineUserID: "Udeleted"}
			fx := newCleanupSyncFixture(t, intent, deletedCleanupOwner(t, 1, 2, "Udeleted", false), []string{"next_visit_2026-11-01"}, nil)
			waiting, release := make(chan struct{}), make(chan struct{})
			enabled := true
			fx.svc.settingsSvc = &mockLstepSettingsService{isSyncEnabledFn: func(context.Context, uint64) (bool, error) {
				if !enabled && failure {
					return false, errors.New("settings unavailable")
				}
				return enabled, nil
			}}
			cache := fx.svc.tagCacheRepo.(*mockLstepTagCacheRepository)
			cache.withNextVisitOwnerLockFn = func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
				close(waiting)
				<-release
				enabled = false // OFF commit before the lock is released to this cleanup.
				return fn(ctx)
			}
			done := make(chan error, 1)
			go func() { done <- fx.svc.SyncNextVisitTag(context.Background(), 1, 2) }()
			select {
			case <-waiting:
			case <-time.After(time.Second):
				t.Fatal("cleanup did not wait for lock")
			}
			close(release)
			var err error
			select {
			case err = <-done:
			case <-time.After(time.Second):
				t.Fatal("cleanup did not finish")
			}
			if failure {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Zero(t, fx.spy.getCalls)
			assert.Empty(t, fx.spy.removed)
			assert.Empty(t, fx.spy.cacheDeleted)
			assert.Empty(t, fx.spy.ackRevisions)
			require.Same(t, intent, func() *NextVisitCleanupIntent {
				got, e := cache.FindNextVisitCleanupIntent(context.Background(), 1, 2)
				require.NoError(t, e)
				return got
			}())
			// Once resumed, the same retained revision is processed and ACKed.
			enabled = true
			cache.withNextVisitOwnerLockFn = func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error { return fn(ctx) }
			require.NoError(t, fx.svc.SyncNextVisitTag(context.Background(), 1, 2))
			assert.Equal(t, []uint64{9}, fx.spy.ackRevisions)
		})
	}
}

func TestOwnerDeletionHTTP_SyncDisabledWhileWaitingSkipsProviderAndCache(t *testing.T) {
	var providerCalls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		providerCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"tags":[]}`))
		assert.NoError(t, err)
	}))
	defer provider.Close()
	waiting, release := make(chan struct{}), make(chan struct{})
	enabled := true
	uid := "Uold"
	owners := &mockOwnerRepository{findByIDFn: func(_ context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
		assert.Equal(t, uint64(1), clinicID)
		assert.Equal(t, uint64(2), ownerID)
		return &model.Owner{ID: ownerID, ClinicID: clinicID, LineUserID: &uid}, nil
	}}
	cache := &mockLstepTagCacheRepository{withNextVisitOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
		close(waiting)
		<-release
		enabled = false
		return fn(ctx)
	}, findByOwnerFn: func(context.Context, uint64, uint64) ([]*model.LstepTagCache, error) {
		t.Error("sync disabled inside the lock must not read the cache")
		return nil, nil
	}, deleteTagFn: func(context.Context, uint64, uint64, string) error {
		t.Error("sync disabled inside the lock must not delete cache")
		return nil
	}}
	lifecycle := &lstepLifecycleService{ownerRepo: owners, tagCacheRepo: cache, auditSvc: &mockAuditService{}, settingsSvc: &mockLstepSettingsService{
		isSyncEnabledFn: func(context.Context, uint64) (bool, error) { return enabled, nil },
		getRawCredentialsFn: func(context.Context, uint64) (string, string, string, error) {
			return "mock-only", provider.URL, "", nil
		},
	}}
	svc := &cleanupDeleteService{}
	handler := ownerdomain.NewHandler(svc, lifecycle, nil, nil)
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Set("clinic_id", "1")
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/v1/owners/2", nil)
	started := make(chan struct{})
	done := make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() {
		select {
		case <-started:
			releaseOnce.Do(func() { close(release) })
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Error("delete goroutine did not exit")
			}
		default:
		}
	})
	go func() {
		close(started)
		handler.DeleteOwner(c)
		c.Writer.WriteHeaderNow()
		close(done)
	}()
	select {
	case <-waiting:
	case <-time.After(time.Second):
		t.Fatal("cleanup did not wait for lock")
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("delete did not finish")
	}
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.True(t, svc.deleted)
	assert.Zero(t, providerCalls.Load(), "disabled sync must not touch the provider")
}

func TestOwnerDeletionCleanup_SettingsRecheckInsideLockGatesProviderAndCache(t *testing.T) {
	settingsErr := errors.New("settings store unavailable")
	for _, tt := range []struct {
		name    string
		enabled bool
		err     error
		wantErr error
	}{
		{name: "disabled inside lock"},
		{name: "settings error inside lock", err: settingsErr, wantErr: settingsErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var getCalls, removeCalls, deleteCalls atomic.Int32
			client := &mockLstepAPIClient{
				getUserTagsFn: func(context.Context, string) ([]string, error) {
					getCalls.Add(1)
					return nil, nil
				},
				removeTagFn: func(context.Context, string, string) error {
					removeCalls.Add(1)
					return nil
				},
			}
			uid := "Uold"
			lifecycle := &lstepLifecycleService{
				ownerRepo: &mockOwnerRepository{findByIDFn: func(_ context.Context, _, ownerID uint64) (*model.Owner, error) {
					return &model.Owner{ID: ownerID, ClinicID: 1, LineUserID: &uid}, nil
				}},
				tagCacheRepo: &mockLstepTagCacheRepository{
					findByOwnerFn: func(context.Context, uint64, uint64) ([]*model.LstepTagCache, error) {
						t.Error("settings gate must not read the cache")
						return nil, nil
					},
					deleteTagFn: func(context.Context, uint64, uint64, string) error {
						deleteCalls.Add(1)
						return nil
					},
				},
				auditSvc: &mockAuditService{},
				settingsSvc: &mockLstepSettingsService{
					isSyncEnabledFn: func(context.Context, uint64) (bool, error) { return tt.enabled, tt.err },
				},
			}

			err := lifecycle.clearTagsWithOwnerLock(context.Background(), client, 1, 2, "Uold")

			require.ErrorIs(t, err, tt.wantErr)
			assert.Zero(t, getCalls.Load(), "GetUserTags")
			assert.Zero(t, removeCalls.Load(), "RemoveTag")
			assert.Zero(t, deleteCalls.Load(), "DeleteTag")
		})
	}
}

func TestOwnerDeletionCleanup_SyncEnabledStillCleans(t *testing.T) {
	var providerCalls atomic.Int32
	client := &mockLstepAPIClient{
		getUserTagsFn: func(context.Context, string) ([]string, error) {
			providerCalls.Add(1)
			return []string{"next_visit_2026-11-01"}, nil
		},
		removeTagFn: func(context.Context, string, string) error {
			providerCalls.Add(1)
			return nil
		},
	}
	uid := "Uold"
	var deleted []string
	lifecycle := &lstepLifecycleService{
		ownerRepo: &mockOwnerRepository{findByIDFn: func(_ context.Context, _, ownerID uint64) (*model.Owner, error) {
			return &model.Owner{ID: ownerID, ClinicID: 1, LineUserID: &uid}, nil
		}},
		tagCacheRepo: &mockLstepTagCacheRepository{
			withNextVisitOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
				return fn(ctx)
			},
			findByOwnerFn: func(context.Context, uint64, uint64) ([]*model.LstepTagCache, error) {
				return []*model.LstepTagCache{{TagName: "next_visit_2026-11-01"}}, nil
			},
			deleteTagFn: func(_ context.Context, _, _ uint64, tag string) error {
				deleted = append(deleted, tag)
				return nil
			},
		},
		auditSvc: &mockAuditService{},
		settingsSvc: &mockLstepSettingsService{
			isSyncEnabledFn: func(context.Context, uint64) (bool, error) { return true, nil },
		},
	}

	require.NoError(t, lifecycle.clearTagsWithOwnerLock(context.Background(), client, 1, 2, "Uold"))
	assert.Equal(t, int32(2), providerCalls.Load(), "tag discovery + removal must reach the provider")
	assert.Equal(t, []string{"next_visit_2026-11-01"}, deleted)
}
