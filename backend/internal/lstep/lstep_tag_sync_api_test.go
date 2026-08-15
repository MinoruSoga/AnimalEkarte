package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- removeStaleTagsByPrefixes ----

func TestRemoveStaleTagsByPrefixes_FindByOwnerError(t *testing.T) {
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return nil, errors.New("db error")
		},
	}
	svc := &lstepTagSyncService{tagCacheRepo: tagCache}
	client := &mockLstepAPIClient{
		removeTagFn: func(_ context.Context, _, _ string) error {
			t.Fatal("RemoveTag should not be called when FindByOwner fails")
			return nil
		},
		addTagFn: func(_ context.Context, _, _ string) error {
			t.Fatal("AddTag should not be called when FindByOwner fails")
			return nil
		},
	}
	apiFailed, err := svc.removeStaleTagsByPrefixes(context.Background(), client, 1, 2, "u1", []string{"vaccine_dog_"}, map[string]struct{}{})
	// LSA-11: cache-read failure is observable (err) and fails closed (apiFailed), distinct from RemoveTag partial failure.
	assert.True(t, apiFailed)
	assert.Error(t, err)
}

func TestRemoveStaleTagsByPrefixes_SkipsTagsInSkipSet(t *testing.T) {
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: "vaccine_dog_2026-05-01"}}, nil
		},
	}
	svc := &lstepTagSyncService{tagCacheRepo: tagCache}
	removeCalled := false
	client := &mockLstepAPIClient{
		removeTagFn: func(_ context.Context, _, _ string) error {
			removeCalled = true
			return nil
		},
	}
	apiFailed, err := svc.removeStaleTagsByPrefixes(context.Background(), client, 1, 2, "u1",
		[]string{"vaccine_dog_"}, map[string]struct{}{"vaccine_dog_2026-05-01": {}})
	assert.NoError(t, err)
	assert.False(t, apiFailed)
	assert.False(t, removeCalled, "skipTags に含まれるタグは RemoveTag されない")
}

func TestRemoveStaleTagsByPrefixes_RemovesMatchingPrefixSuccess(t *testing.T) {
	var removedTag, deletedCacheTag string
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: "vaccine_dog_2026-05-01"}}, nil
		},
		deleteTagFn: func(_ context.Context, _, _ uint64, tag string) error {
			deletedCacheTag = tag
			return nil
		},
	}
	svc := &lstepTagSyncService{tagCacheRepo: tagCache}
	client := &mockLstepAPIClient{
		removeTagFn: func(_ context.Context, _, tag string) error {
			removedTag = tag
			return nil
		},
	}
	apiFailed, err := svc.removeStaleTagsByPrefixes(context.Background(), client, 1, 2, "u1", []string{"vaccine_dog_"}, map[string]struct{}{})
	assert.NoError(t, err)
	assert.False(t, apiFailed)
	assert.Equal(t, "vaccine_dog_2026-05-01", removedTag)
	assert.Equal(t, "vaccine_dog_2026-05-01", deletedCacheTag)
}

func TestRemoveStaleTagsByPrefixes_NoPrefixMatchLeavesTagAlone(t *testing.T) {
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: "cpm_core"}}, nil
		},
	}
	svc := &lstepTagSyncService{tagCacheRepo: tagCache}
	removeCalled := false
	client := &mockLstepAPIClient{
		removeTagFn: func(_ context.Context, _, _ string) error {
			removeCalled = true
			return nil
		},
	}
	apiFailed, err := svc.removeStaleTagsByPrefixes(context.Background(), client, 1, 2, "u1", []string{"vaccine_dog_"}, map[string]struct{}{})
	assert.NoError(t, err)
	assert.False(t, apiFailed)
	assert.False(t, removeCalled, "プレフィックス不一致のタグは触られない")
}

func TestRemoveStaleTagsByPrefixes_RemoveTagPartialFailureIsNotCacheError(t *testing.T) {
	// RemoveTag partial failure: durable failure accounting + apiFailed=true, but err=nil
	// so callers may still apply desired AddTag (distinct from cache-read fail-closed).
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: "vaccine_dog_2026-05-01"}}, nil
		},
	}
	incrementCalled := false
	repo := &mockErrorCounterRepo{
		incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
			incrementCalled = true
			return lstepSyncErrorThreshold - 1, nil
		},
	}
	svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: tagCache}
	client := &mockLstepAPIClient{
		removeTagFn: func(_ context.Context, _, _ string) error {
			return errors.New("lstep remove failed")
		},
	}
	apiFailed, err := svc.removeStaleTagsByPrefixes(context.Background(), client, 1, 2, "u1",
		[]string{"vaccine_dog_"}, map[string]struct{}{})
	assert.NoError(t, err, "RemoveTag partial failure must not surface as cache-read error")
	assert.True(t, apiFailed)
	assert.True(t, incrementCalled, "durable failure accounting must remain")
}

// ---- notifyAPIFailure additional branches ----

func TestNotifyAPIFailure_AddTagErrorAtThreshold(t *testing.T) {
	client := &mockLstepAPIClient{
		addTagFn: func(_ context.Context, _, _ string) error { return errors.New("add tag failed") },
	}
	repo := &mockErrorCounterRepo{
		incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
			return lstepSyncErrorThreshold, nil
		},
	}
	upsertCalled := false
	tagCache := &mockLstepTagCacheRepository{
		upsertTagFn: func(_ context.Context, _, _ uint64, _, _, _ string) error {
			upsertCalled = true
			return nil
		},
	}
	svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: tagCache}
	// AddTag 失敗時はキャッシュ更新まで到達しない
	svc.notifyAPIFailure(context.Background(), client, 1, 2, "u1")
	assert.False(t, upsertCalled, "AddTag に失敗した場合、タグキャッシュの更新は行われない")
}

func TestNotifyAPIFailure_UpsertTagCacheErrorAtThresholdIsNonFatal(t *testing.T) {
	client := &mockLstepAPIClient{
		addTagFn: func(_ context.Context, _, _ string) error { return nil },
	}
	repo := &mockErrorCounterRepo{
		incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
			return lstepSyncErrorThreshold, nil
		},
	}
	tagCache := &mockLstepTagCacheRepository{
		upsertTagFn: func(_ context.Context, _, _ uint64, _, _, _ string) error {
			return errors.New("cache upsert failed")
		},
	}
	svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: tagCache}
	// must not panic; best-effort failure — no observable return value.
	svc.notifyAPIFailure(context.Background(), client, 1, 2, "u1")
}

// ---- notifyAPISuccess additional branches ----

func TestNotifyAPISuccess_GenericFindByOwnerErrorIsNonFatal(t *testing.T) {
	removeTagCalled := false
	client := &mockLstepAPIClient{
		removeTagFn: func(_ context.Context, _, _ string) error {
			removeTagCalled = true
			return nil
		},
	}
	repo := &mockErrorCounterRepo{
		findFn: func(_ context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
			return nil, errors.New("db error")
		},
	}
	svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: &mockLstepTagCacheRepository{}}
	svc.notifyAPISuccess(context.Background(), client, 1, 2, "u1")
	assert.False(t, removeTagCalled)
}

func TestNotifyAPISuccess_DeleteTagCacheErrorIsNonFatal(t *testing.T) {
	client := &mockLstepAPIClient{
		removeTagFn: func(_ context.Context, _, _ string) error { return nil },
	}
	resetCalled := false
	tagCache := &mockLstepTagCacheRepository{
		deleteTagFn: func(_ context.Context, _, _ uint64, _ string) error {
			return errors.New("cache delete failed")
		},
	}
	repo := &mockErrorCounterRepo{
		findFn: func(_ context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
			return &model.LstepSyncErrorCounter{FailureCount: lstepSyncErrorThreshold}, nil
		},
		resetFn: func(_ context.Context, _, _ uint64) error {
			resetCalled = true
			return nil
		},
	}
	svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: tagCache}
	// Must not panic despite the cache delete error; reset should still be attempted.
	svc.notifyAPISuccess(context.Background(), client, 1, 2, "u1")
	assert.True(t, resetCalled, "キャッシュ削除に失敗してもカウンターリセットは継続する")
}

func TestNotifyAPISuccess_ResetFailureErrorIsNonFatal(t *testing.T) {
	client := &mockLstepAPIClient{
		removeTagFn: func(_ context.Context, _, _ string) error { return nil },
	}
	repo := &mockErrorCounterRepo{
		findFn: func(_ context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
			return &model.LstepSyncErrorCounter{FailureCount: lstepSyncErrorThreshold}, nil
		},
		resetFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("reset failed")
		},
	}
	svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: &mockLstepTagCacheRepository{}}
	// Must not panic; best-effort — no return value to observe beyond absence of panic.
	svc.notifyAPISuccess(context.Background(), client, 1, 2, "u1")
}
