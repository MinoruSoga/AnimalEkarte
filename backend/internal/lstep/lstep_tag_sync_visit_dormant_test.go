package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

func dormantTestThresholds() model.DormantThresholds {
	return model.DormantThresholds{Stage180: 180, Stage210: 210, Stage240: 240, Stage365: 365}
}

func buildDormantTestSvc(
	settingsSvc *mockLstepSettingsService,
	ownerRepo *mockOwnerRepository,
	tagCache *mockLstepTagCacheRepository,
	errorCounterRepo *mockErrorCounterRepo,
	client *mockLstepAPIClient,
) *lstepTagSyncService {
	// NOTE: a nil *mockErrorCounterRepo assigned directly to the
	// repository.LstepSyncErrorCounterRepository interface field would produce a
	// non-nil interface wrapping a nil pointer, defeating the `s.errorCounterRepo == nil`
	// nil-check in notifyAPIFailure/notifyAPISuccess and panicking on first use.
	// Substitute a real mock so callers can still pass nil for "don't care". The
	// substitute's FindByOwner must return a non-nil counter (FailureCount: 0) because
	// notifyAPISuccess dereferences counter.FailureCount without a nil guard.
	if errorCounterRepo == nil {
		errorCounterRepo = &mockErrorCounterRepo{
			findFn: func(_ context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
				return &model.LstepSyncErrorCounter{FailureCount: 0}, nil
			},
		}
	}
	return &lstepTagSyncService{
		settingsSvc:      settingsSvc,
		ownerRepo:        ownerRepo,
		tagCacheRepo:     tagCache,
		errorCounterRepo: errorCounterRepo,
		buildClientFn:    func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
	}
}

func enabledDormantSettings() *mockLstepSettingsService {
	return &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
	}
}

// ---- SyncDormantTagsWithThresholds ----
//
// B-3: 旧公開ラッパー SyncDormantTags（thresholds fetch → SyncDormantTagsWithThresholds 委譲）は
// 呼び出し元ゼロのため削除済み。ラッパー固有だった「thresholds 取得失敗」の分岐は
// SyncDormantTagsWithThresholds のシグネチャ（thresholds を値で受け取る）には存在しないため、
// 対応するテストは削除した。shouldSkipSync 失敗パスは下記
// TestSyncDormantTagsWithThresholds_ShouldSkipSyncError でカバーされる。

func TestSyncDormantTagsWithThresholds_ShouldSkipSyncError(t *testing.T) {
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("settings error") },
		},
	}
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 200, dormantTestThresholds())
	assert.Error(t, err)
}

func TestSyncDormantTagsWithThresholds_SkippedWhenSyncDisabled(t *testing.T) {
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
		},
	}
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 200, dormantTestThresholds())
	assert.NoError(t, err)
}

func TestSyncDormantTagsWithThresholds_CheckOptOutError(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(),
		&mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, errors.New("db error")
			},
		}, &mockLstepTagCacheRepository{}, nil, &mockLstepAPIClient{})
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 200, dormantTestThresholds())
	assert.Error(t, err)
}

func TestSyncDormantTagsWithThresholds_OptedOutIsNoop(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(),
		ownerRepoReturning(&model.Owner{ID: 2, LstepOptOut: true}), &mockLstepTagCacheRepository{}, nil, &mockLstepAPIClient{})
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 200, dormantTestThresholds())
	assert.NoError(t, err)
}

func TestSyncDormantTagsWithThresholds_NoLineUserIDIsNoop(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(),
		ownerRepoReturning(&model.Owner{ID: 2, LineUserID: nil}), &mockLstepTagCacheRepository{}, nil, &mockLstepAPIClient{})
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 200, dormantTestThresholds())
	assert.NoError(t, err)

	empty := ""
	svc2 := buildDormantTestSvc(enabledDormantSettings(),
		ownerRepoReturning(&model.Owner{ID: 2, LineUserID: &empty}), &mockLstepTagCacheRepository{}, nil, &mockLstepAPIClient{})
	err2 := svc2.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 200, dormantTestThresholds())
	assert.NoError(t, err2)
}

func TestSyncDormantTagsWithThresholds_BuildClientError(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(),
		ownerRepoReturning(defaultOwnerWithLineID()), &mockLstepTagCacheRepository{}, nil, &mockLstepAPIClient{})
	svc.buildClientFn = func(_ context.Context, _ uint64) (lstep.Client, error) {
		return nil, errors.New("client build failed")
	}
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 200, dormantTestThresholds())
	assert.Error(t, err)
}

func TestSyncDormantTagsWithThresholds_ClientNilIsNoop(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(),
		ownerRepoReturning(defaultOwnerWithLineID()), &mockLstepTagCacheRepository{}, nil, nil)
	svc.buildClientFn = func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil }
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 200, dormantTestThresholds())
	assert.NoError(t, err)
}

func TestSyncDormantTagsWithThresholds_TagCacheFindError(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{
			findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
				return nil, errors.New("db error")
			},
		}, nil, &mockLstepAPIClient{})
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 200, dormantTestThresholds())
	assert.Error(t, err)
}

func TestSyncDormantTagsWithThresholds_TargetTagSelection(t *testing.T) {
	tests := []struct {
		name     string
		days     int
		wantTags []string
	}{
		// days >= Stage240 also adds the CPM dormant stage tag ("cpm_dormant") in a
		// second AddTag call (see lstep_tag_sync_visit_dormant.go: needsCpmDormant).
		{"negative days (no visit) -> 365d", -1, []string{"dormant_365d", "cpm_dormant"}},
		{"exactly 365 -> 365d", 365, []string{"dormant_365d", "cpm_dormant"}},
		{"between 240 and 365 -> 240d", 250, []string{"dormant_240d", "cpm_dormant"}},
		{"between 210 and 240 -> 210d", 220, []string{"dormant_210d"}},
		{"between 180 and 210 -> 180d", 190, []string{"dormant_180d"}},
		{"below 180 -> no tag", 100, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addedTags []string
			client := &mockLstepAPIClient{
				addTagFn: func(_ context.Context, _, tag string) error { addedTags = append(addedTags, tag); return nil },
			}
			svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
				&mockLstepTagCacheRepository{}, nil, client)
			err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, tt.days, dormantTestThresholds())
			assert.NoError(t, err)
			assert.Equal(t, tt.wantTags, addedTags)
		})
	}
}

func TestSyncDormantTagsWithThresholds_RemovesStaleDormantTag(t *testing.T) {
	var removedTag string
	var deletedCacheTag string
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: "dormant_180d"}}, nil
		},
		deleteTagFn: func(_ context.Context, _, _ uint64, tag string) error {
			deletedCacheTag = tag
			return nil
		},
	}
	client := &mockLstepAPIClient{
		addTagFn:    func(_ context.Context, _, _ string) error { return nil },
		removeTagFn: func(_ context.Context, _, tag string) error { removedTag = tag; return nil },
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()), tagCache, nil, client)
	// days=365 targets dormant_365d, so the cached dormant_180d is stale and should be removed.
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 365, dormantTestThresholds())
	assert.NoError(t, err)
	assert.Equal(t, "dormant_180d", removedTag)
	assert.Equal(t, "dormant_180d", deletedCacheTag)
}

func TestSyncDormantTagsWithThresholds_RemoveStaleDormantTagFailureSetsAPIFailed(t *testing.T) {
	incrementCalled := false
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: "dormant_180d"}}, nil
		},
	}
	client := &mockLstepAPIClient{
		addTagFn:    func(_ context.Context, _, _ string) error { return nil },
		removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("remove failed") },
	}
	errCounter := &mockErrorCounterRepo{
		incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
			incrementCalled = true
			return 1, nil
		},
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()), tagCache, errCounter, client)
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 365, dormantTestThresholds())
	assert.NoError(t, err) // apiFailed does not propagate as an error
	assert.True(t, incrementCalled, "RemoveTag 失敗時は notifyAPIFailure が呼ばれる")
}

func TestSyncDormantTagsWithThresholds_RemovesCpmDormantWhenNotNeeded(t *testing.T) {
	var removedTags []string
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: "cpm_dormant"}}, nil
		},
	}
	client := &mockLstepAPIClient{
		addTagFn:    func(_ context.Context, _, _ string) error { return nil },
		removeTagFn: func(_ context.Context, _, tag string) error { removedTags = append(removedTags, tag); return nil },
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()), tagCache, nil, client)
	// days=190 -> dormant_180d (needsCpmDormant=false) -> cached cpm_dormant should be removed
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 190, dormantTestThresholds())
	assert.NoError(t, err)
	assert.Contains(t, removedTags, "cpm_dormant")
}

func TestSyncDormantTagsWithThresholds_NoTargetTagCallsNotifyAPISuccess(t *testing.T) {
	findCalled := false
	errCounter := &mockErrorCounterRepo{
		findFn: func(_ context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
			findCalled = true
			return &model.LstepSyncErrorCounter{FailureCount: 0}, nil
		},
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{}, errCounter, &mockLstepAPIClient{})
	// days below all thresholds -> targetTag == "" -> notifyAPISuccess should still be invoked
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 50, dormantTestThresholds())
	assert.NoError(t, err)
	assert.True(t, findCalled, "targetTag が空でも apiFailed=false なら notifyAPISuccess が呼ばれる")
}

func TestSyncDormantTagsWithThresholds_AddTagErrorPropagates(t *testing.T) {
	client := &mockLstepAPIClient{
		addTagFn: func(_ context.Context, _, _ string) error { return errors.New("add tag failed") },
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{}, nil, client)
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 365, dormantTestThresholds())
	assert.Error(t, err)
}

func TestSyncDormantTagsWithThresholds_AddTagSuccessCacheUpsertErrorIsNonFatal(t *testing.T) {
	tagCache := &mockLstepTagCacheRepository{
		upsertTagFn: func(_ context.Context, _, _ uint64, _, _, _ string) error {
			return errors.New("cache upsert failed")
		},
	}
	client := &mockLstepAPIClient{
		addTagFn: func(_ context.Context, _, _ string) error { return nil },
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()), tagCache, nil, client)
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 365, dormantTestThresholds())
	assert.NoError(t, err)
}

func TestSyncDormantTagsWithThresholds_NeedsCpmDormantAddsSuccessfully(t *testing.T) {
	var addedTags []string
	client := &mockLstepAPIClient{
		addTagFn: func(_ context.Context, _, tag string) error { addedTags = append(addedTags, tag); return nil },
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{}, nil, client)
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 365, dormantTestThresholds())
	assert.NoError(t, err)
	assert.Contains(t, addedTags, "dormant_365d")
	assert.Contains(t, addedTags, "cpm_dormant")
}

func TestSyncDormantTagsWithThresholds_NeedsCpmDormantAddErrorPropagates(t *testing.T) {
	client := &mockLstepAPIClient{
		addTagFn: func(_ context.Context, _, tag string) error {
			if tag == "cpm_dormant" {
				return errors.New("cpm_dormant add failed")
			}
			return nil
		},
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{}, nil, client)
	err := svc.SyncDormantTagsWithThresholds(context.Background(), 1, 2, 365, dormantTestThresholds())
	assert.Error(t, err)
}

// ---- SyncVisitDormantTags ----

func TestSyncVisitDormantTags_ShouldSkipSyncError(t *testing.T) {
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("settings error") },
		},
	}
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 130)
	assert.Error(t, err)
}

func TestSyncVisitDormantTags_SkippedWhenSyncDisabled(t *testing.T) {
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
		},
	}
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 130)
	assert.NoError(t, err)
}

func TestSyncVisitDormantTags_CheckOptOutError(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(),
		&mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, errors.New("db error")
			},
		}, &mockLstepTagCacheRepository{}, nil, &mockLstepAPIClient{})
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 130)
	assert.Error(t, err)
}

func TestSyncVisitDormantTags_OptedOutIsNoop(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(),
		ownerRepoReturning(&model.Owner{ID: 2, LstepOptOut: true}), &mockLstepTagCacheRepository{}, nil, &mockLstepAPIClient{})
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 130)
	assert.NoError(t, err)
}

func TestSyncVisitDormantTags_NoLineUserIDIsNoop(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(),
		ownerRepoReturning(&model.Owner{ID: 2, LineUserID: nil}), &mockLstepTagCacheRepository{}, nil, &mockLstepAPIClient{})
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 130)
	assert.NoError(t, err)
}

func TestSyncVisitDormantTags_BuildClientError(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{}, nil, &mockLstepAPIClient{})
	svc.buildClientFn = func(_ context.Context, _ uint64) (lstep.Client, error) {
		return nil, errors.New("client build failed")
	}
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 130)
	assert.Error(t, err)
}

func TestSyncVisitDormantTags_ClientNilIsNoop(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{}, nil, nil)
	svc.buildClientFn = func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil }
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 130)
	assert.NoError(t, err)
}

func TestSyncVisitDormantTags_TagCacheFindError(t *testing.T) {
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{
			findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
				return nil, errors.New("db error")
			},
		}, nil, &mockLstepAPIClient{})
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 130)
	assert.Error(t, err)
}

func TestSyncVisitDormantTags_AddsAllQualifyingThresholds(t *testing.T) {
	var addedTags []string
	client := &mockLstepAPIClient{
		addTagFn: func(_ context.Context, _, tag string) error { addedTags = append(addedTags, tag); return nil },
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{}, nil, client)
	// 241 days > all four thresholds (120/180/220/240) -> all four VISIT tags should be added
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 241)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{visitTag120, visitTag180, visitTag220, visitTag240}, addedTags)
}

func TestSyncVisitDormantTags_AddTagFailureSetsAPIFailedAndContinues(t *testing.T) {
	var addedTags []string
	incrementCalls := 0
	client := &mockLstepAPIClient{
		addTagFn: func(_ context.Context, _, tag string) error {
			if tag == visitTag120 {
				return errors.New("add failed")
			}
			addedTags = append(addedTags, tag)
			return nil
		},
	}
	errCounter := &mockErrorCounterRepo{
		incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
			incrementCalls++
			return 1, nil
		},
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{}, errCounter, client)
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 241)
	assert.NoError(t, err) // apiFailed does not propagate as an error
	assert.Equal(t, 1, incrementCalls, "1つの失敗でnotifyAPIFailureが1回呼ばれる")
	// remaining thresholds are still attempted despite the first failure (continue, not return)
	assert.Contains(t, addedTags, visitTag180)
	assert.Contains(t, addedTags, visitTag220)
	assert.Contains(t, addedTags, visitTag240)
}

func TestSyncVisitDormantTags_RemovesTagsNoLongerQualifying(t *testing.T) {
	var removedTags []string
	var deletedCacheTags []string
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: visitTag120}, {TagName: visitTag180}}, nil
		},
		deleteTagFn: func(_ context.Context, _, _ uint64, tag string) error {
			deletedCacheTags = append(deletedCacheTags, tag)
			return nil
		},
	}
	client := &mockLstepAPIClient{
		removeTagFn: func(_ context.Context, _, tag string) error { removedTags = append(removedTags, tag); return nil },
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()), tagCache, nil, client)
	// days=50 is below all thresholds -> both cached VISIT tags should be removed
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 50)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{visitTag120, visitTag180}, removedTags)
	assert.ElementsMatch(t, []string{visitTag120, visitTag180}, deletedCacheTags)
}

func TestSyncVisitDormantTags_RemoveTagFailureSetsAPIFailedAndContinues(t *testing.T) {
	incrementCalls := 0
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: visitTag120}}, nil
		},
	}
	client := &mockLstepAPIClient{
		removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("remove failed") },
	}
	errCounter := &mockErrorCounterRepo{
		incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
			incrementCalls++
			return 1, nil
		},
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()), tagCache, errCounter, client)
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 50)
	assert.NoError(t, err)
	assert.Equal(t, 1, incrementCalls)
}

func TestSyncVisitDormantTags_NoQualifyingNoCachedIsNoopPerThreshold(t *testing.T) {
	addCalled := false
	removeCalled := false
	client := &mockLstepAPIClient{
		addTagFn:    func(_ context.Context, _, _ string) error { addCalled = true; return nil },
		removeTagFn: func(_ context.Context, _, _ string) error { removeCalled = true; return nil },
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{}, nil, client)
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 50)
	assert.NoError(t, err)
	assert.False(t, addCalled)
	assert.False(t, removeCalled)
}

func TestSyncVisitDormantTags_AllSuccessCallsNotifyAPISuccess(t *testing.T) {
	findCalled := false
	errCounter := &mockErrorCounterRepo{
		findFn: func(_ context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
			findCalled = true
			return &model.LstepSyncErrorCounter{FailureCount: 0}, nil
		},
	}
	svc := buildDormantTestSvc(enabledDormantSettings(), ownerRepoReturning(defaultOwnerWithLineID()),
		&mockLstepTagCacheRepository{}, errCounter, &mockLstepAPIClient{})
	err := svc.SyncVisitDormantTags(context.Background(), 1, 2, 50)
	assert.NoError(t, err)
	assert.True(t, findCalled, "全件成功時は notifyAPISuccess が呼ばれる")
}
