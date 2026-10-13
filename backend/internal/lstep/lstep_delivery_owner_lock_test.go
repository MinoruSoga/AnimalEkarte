package lstep

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// unlockeableTagCacheRepo satisfies deliveryTagCacheRepository but not
// deliveryOwnerLocker — processSingleOwner must fail closed without it.
type unlockeableTagCacheRepo struct {
	tagsByOwner map[uint64][]*model.LstepTagCache
}

func (r *unlockeableTagCacheRepo) FindByOwner(_ context.Context, _, ownerID uint64) ([]*model.LstepTagCache, error) {
	return r.tagsByOwner[ownerID], nil
}
func (r *unlockeableTagCacheRepo) FindOwnerIDsByTag(_ context.Context, _ uint64, _ string) ([]uint64, error) {
	return nil, nil
}

func newOwnerLockUnitService(
	logRepo LstepDeliveryTriggerLogRepository,
	tagRepo deliveryTagCacheRepository,
	ownerRepo deliveryOwnerRepository,
	settingsSvc *mockLstepSettingsService,
) *lstepDeliveryTriggerService {
	return &lstepDeliveryTriggerService{
		ownerRepo:      ownerRepo,
		tagCacheRepo:   tagRepo,
		triggerLogRepo: logRepo,
		settingsSvc:    settingsSvc,
	}
}

// Missing owner locker fails closed: no provider call, and the claim gets the
// retryable known-unsent marker rather than staying scheduled forever.
func TestProcessSingleOwner_FailClosedWithoutOwnerLocker(t *testing.T) {
	var markCalls, addCalls int
	logRepo := &mockDeliveryTriggerLogRepository{
		markKnownUnsentFn: func(_ context.Context, _, _ uint64, reason string) error {
			markCalls++
			assert.Equal(t, preSendRevalidationFailure, reason)
			return nil
		},
	}
	svc := newOwnerLockUnitService(logRepo, &unlockeableTagCacheRepo{}, validOwnerRepoForDelivery(), &mockLstepSettingsService{})
	client := &mockLstepClientForDelivery{
		addTagFn: func(context.Context, string, string) error { addCalls++; return nil },
	}

	fired, err := svc.processSingleOwner(context.Background(), client, 1, 10, "trigger_x", "tag_x", time.Now(), nil, nil)
	require.Error(t, err)
	assert.False(t, fired)
	assert.Equal(t, 0, addCalls, "provider must never be invoked without the owner lock")
	assert.Equal(t, 1, markCalls, "the claimed slot must be marked known-unsent")
}

// Clinic sync OFF re-checked inside the lock marks the claim excluded — the
// provider is never invoked and the day slot stays closed.
func TestProcessSingleOwner_LockSettingsOffMarksExcluded(t *testing.T) {
	var addCalls int
	var gotStatus, gotReason string
	logRepo := &mockDeliveryTriggerLogRepository{
		updateStatusFn: func(_ context.Context, _ uint64, status string, _ *time.Time, reason *string) error {
			gotStatus = status
			if reason != nil {
				gotReason = *reason
			}
			return nil
		},
	}
	tagRepo := &mockTagCacheRepoForDelivery{}
	svc := newOwnerLockUnitService(logRepo, tagRepo, validOwnerRepoForDelivery(), &mockLstepSettingsService{
		isSyncEnabledFn: func(context.Context, uint64) (bool, error) { return false, nil },
	})
	client := &mockLstepClientForDelivery{
		addTagFn: func(context.Context, string, string) error { addCalls++; return nil },
	}

	fired, err := svc.processSingleOwner(context.Background(), client, 1, 10, "trigger_x", "tag_x", time.Now(), nil, nil)
	require.NoError(t, err)
	assert.False(t, fired)
	assert.Equal(t, 0, addCalls)
	assert.Equal(t, model.TriggerStatusExcluded, gotStatus)
	assert.Equal(t, "lstep_sync_disabled", gotReason)
}

// A provider failure after send started is a blocking outcome — never the
// retryable pre_send marker.
func TestProcessSingleOwner_SendFailureIsBlockingNotRetryable(t *testing.T) {
	var gotStatus, gotReason string
	var markCalls int
	logRepo := &mockDeliveryTriggerLogRepository{
		updateStatusFn: func(_ context.Context, _ uint64, status string, _ *time.Time, reason *string) error {
			gotStatus = status
			if reason != nil {
				gotReason = *reason
			}
			return nil
		},
		markKnownUnsentFn: func(context.Context, uint64, uint64, string) error {
			markCalls++
			return nil
		},
	}
	svc := newOwnerLockUnitService(logRepo, &mockTagCacheRepoForDelivery{}, validOwnerRepoForDelivery(), &mockLstepSettingsService{})
	client := &mockLstepClientForDelivery{
		addTagFn: func(context.Context, string, string) error { return errors.New("provider exploded") },
	}

	fired, err := svc.processSingleOwner(context.Background(), client, 1, 10, "trigger_x", "tag_x", time.Now(), nil, nil)
	require.Error(t, err)
	assert.False(t, fired)
	assert.Equal(t, model.TriggerStatusFailed, gotStatus)
	assert.True(t, strings.HasPrefix(gotReason, "lstep_add_tag_failed"), "blocking reason, got %q", gotReason)
	assert.Equal(t, 0, markCalls, "a started send must never get the retryable marker")
}

// Provider success followed by a serialization failure (e.g. commit error)
// still persists the fired outcome — the slot stays claimed and is never
// resent — and the failure stays observable.
func TestProcessSingleOwner_CommitFailureAfterSendPersistsFired(t *testing.T) {
	var gotStatus string
	var firedAtSet bool
	logRepo := &mockDeliveryTriggerLogRepository{
		updateStatusFn: func(_ context.Context, _ uint64, status string, firedAt *time.Time, _ *string) error {
			gotStatus = status
			firedAtSet = firedAt != nil
			return nil
		},
	}
	tagRepo := &mockTagCacheRepoForDelivery{
		withDeliveryOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
			if err := fn(ctx); err != nil {
				return err
			}
			return errors.New("serialization commit failed")
		},
	}
	svc := newOwnerLockUnitService(logRepo, tagRepo, validOwnerRepoForDelivery(), &mockLstepSettingsService{})
	client := &mockLstepClientForDelivery{
		addTagFn: func(context.Context, string, string) error { return nil },
	}

	fired, err := svc.processSingleOwner(context.Background(), client, 1, 10, "trigger_x", "tag_x", time.Now(), nil, nil)
	require.Error(t, err, "serialization failure must surface")
	assert.False(t, fired)
	assert.Equal(t, model.TriggerStatusFired, gotStatus, "provider success must still persist fired")
	assert.True(t, firedAtSet)
}

// Provider success with a failed outcome persistence must surface the error —
// never silently claim a clean delivery.
func TestProcessSingleOwner_FiredPersistFailureIsObservable(t *testing.T) {
	logRepo := &mockDeliveryTriggerLogRepository{
		updateStatusFn: func(context.Context, uint64, string, *time.Time, *string) error {
			return errors.New("db write lost")
		},
	}
	svc := newOwnerLockUnitService(logRepo, &mockTagCacheRepoForDelivery{}, validOwnerRepoForDelivery(), &mockLstepSettingsService{})
	client := &mockLstepClientForDelivery{
		addTagFn: func(context.Context, string, string) error { return nil },
	}

	fired, err := svc.processSingleOwner(context.Background(), client, 1, 10, "trigger_x", "tag_x", time.Now(), nil, nil)
	require.Error(t, err)
	assert.False(t, fired, "a send whose outcome is not durable must not count as delivered")
}

// A hook reporting valid without invoking send is a contract violation — the
// claim becomes known-unsent rather than a phantom delivery.
func TestProcessSingleOwner_HookValidWithoutSendFailsClosed(t *testing.T) {
	var markCalls, addCalls int
	logRepo := &mockDeliveryTriggerLogRepository{
		markKnownUnsentFn: func(context.Context, uint64, uint64, string) error {
			markCalls++
			return nil
		},
	}
	svc := newOwnerLockUnitService(logRepo, &mockTagCacheRepoForDelivery{}, validOwnerRepoForDelivery(), &mockLstepSettingsService{})
	client := &mockLstepClientForDelivery{
		addTagFn: func(context.Context, string, string) error { addCalls++; return nil },
	}
	revalidate := func(context.Context, uint64, func(context.Context) error) (bool, string, error) {
		return true, "", nil
	}

	fired, err := svc.processSingleOwner(context.Background(), client, 1, 10, "trigger_x", "tag_x", time.Now(), revalidate, nil)
	require.Error(t, err)
	assert.False(t, fired)
	assert.Equal(t, 0, addCalls)
	assert.Equal(t, 1, markCalls)
}
