package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// This file extends checkup_sync_service_test.go's CreateCheckupSync coverage with branches not
// exercised there: empty OwnerIDs, ownerRepo/petRepo repository errors, unresolved owner ids,
// and AddTag failure (unreachable via the real httpLstepClient, whose AddTag is a documented
// [DISABLED] no-op — reached here only via buildClientFn injection).

func TestCheckupSyncService_CreateCheckupSync_EmptyOwnerIDs(t *testing.T) {
	svc := NewCheckupSyncService(
		&mockCheckupSyncRepository{},
		&mockOwnerRepository{},
		&mockPetRepository{},
		&mockLstepTagCacheRepository{},
		&mockLstepSettingsService{
			getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
				return "test-key", "https://example.com", "", nil
			},
		},
		&mockAuditService{},
	)
	result, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{},
		TagName:     "campaign_2026",
	}, nil)
	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, result.SkippedCount)
		assert.Equal(t, 0, result.FailedCount)
		assert.Empty(t, result.FailedOwnerIDs)
	}
}

// TestCheckupSyncService_CreateCheckupSync_EmptyOwnerIDs_RecordsAudit は PERF-M3 の受入条件を検証する:
// OwnerIDs が空でも早期 return 前に監査ログが 1 回記録され、owner_count: 0 が metadata に残ること。
func TestCheckupSyncService_CreateCheckupSync_EmptyOwnerIDs_RecordsAudit(t *testing.T) {
	spy := &spyCheckupAuditService{}
	svc := NewCheckupSyncService(
		&mockCheckupSyncRepository{},
		&mockOwnerRepository{},
		&mockPetRepository{},
		&mockLstepTagCacheRepository{},
		&mockLstepSettingsService{
			getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
				return "test-key", "https://example.com", "", nil
			},
		},
		spy,
	)
	result, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{},
		TagName:     "campaign_2026",
	}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	if assert.True(t, spy.called, "empty OwnerIDs でも監査ログが記録されること") {
		assert.Equal(t, uint64(1), spy.capturedClinicID)
		metadata, ok := spy.capturedMetadata.(map[string]any)
		if assert.True(t, ok, "metadata は map[string]any であること") {
			assert.Equal(t, 0, metadata["owner_count"])
		}
	}
}

func TestCheckupSyncService_CreateCheckupSync_BuildClientCredentialsError(t *testing.T) {
	svc := NewCheckupSyncService(
		&mockCheckupSyncRepository{},
		&mockOwnerRepository{},
		&mockPetRepository{},
		&mockLstepTagCacheRepository{},
		&mockLstepSettingsService{
			getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
				return "", "", "", errors.New("db error")
			},
		},
		&mockAuditService{},
	)
	_, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1},
		TagName:     "campaign_2026",
	}, nil)
	assert.Error(t, err)
}

func TestCheckupSyncService_CreateCheckupSync_FindByIDsError(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, _ uint64, _ []uint64) ([]*model.Owner, error) {
			return nil, errors.New("db error")
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-key", "https://example.com", "", nil
		},
	}
	svc := NewCheckupSyncService(&mockCheckupSyncRepository{}, ownerRepo, &mockPetRepository{}, &mockLstepTagCacheRepository{}, settingsSvc, &mockAuditService{})
	_, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1},
		TagName:     "campaign_2026",
	}, nil)
	assert.Error(t, err)
}

func TestCheckupSyncService_CreateCheckupSync_UnresolvedOwnerIDsAreFailed(t *testing.T) {
	// owner 1 exists; owner 999 does not -> must be recorded in FailedOwnerIDs.
	ownerRepo := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, _ uint64, ids []uint64) ([]*model.Owner, error) {
			var out []*model.Owner
			for _, id := range ids {
				if id == 1 {
					out = append(out, &model.Owner{ID: 1, LstepOptOut: true}) // opt-out so it also skips cleanly
				}
			}
			return out, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-key", "https://example.com", "", nil
		},
	}
	svc := NewCheckupSyncService(&mockCheckupSyncRepository{}, ownerRepo, &mockPetRepository{}, &mockLstepTagCacheRepository{}, settingsSvc, &mockAuditService{})
	result, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1, 999},
		TagName:     "campaign_2026",
	}, nil)
	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, 1, result.FailedCount)
		assert.Contains(t, result.FailedOwnerIDs, uint64(999))
		assert.Equal(t, 1, result.SkippedCount, "owner 1 is opted out and skipped")
	}
}

func TestCheckupSyncService_CreateCheckupSync_CountLivingByOwnerIDsError(t *testing.T) {
	lineID := "U_test"
	ownerRepo := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, _ uint64, ids []uint64) ([]*model.Owner, error) {
			return []*model.Owner{{ID: ids[0], LineUserID: &lineID, LstepOptOut: false}}, nil
		},
	}
	petRepo := &mockPetRepository{
		countLivingByOwnerIDsFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64]int64, error) {
			return nil, errors.New("db error")
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-key", "https://example.com", "", nil
		},
	}
	svc := NewCheckupSyncService(&mockCheckupSyncRepository{}, ownerRepo, petRepo, &mockLstepTagCacheRepository{}, settingsSvc, &mockAuditService{})
	_, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1},
		TagName:     "campaign_2026",
	}, nil)
	assert.Error(t, err)
}

func TestCheckupSyncService_CreateCheckupSync_AddTagFailureIsRecordedAsFailed(t *testing.T) {
	lineID := "U_test"
	ownerRepo := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, _ uint64, ids []uint64) ([]*model.Owner, error) {
			return []*model.Owner{{ID: ids[0], LineUserID: &lineID, LstepOptOut: false}}, nil
		},
	}
	petRepo := &mockPetRepository{
		countLivingByOwnerIDsFn: func(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64]int64, error) {
			m := make(map[uint64]int64, len(ownerIDs))
			for _, id := range ownerIDs {
				m[id] = 1
			}
			return m, nil
		},
	}
	client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
	svc := &checkupSyncService{
		repo:          &mockCheckupSyncRepository{},
		ownerRepo:     ownerRepo,
		petRepo:       petRepo,
		tagCacheRepo:  &mockLstepTagCacheRepository{},
		settingsSvc:   &mockLstepSettingsService{},
		auditSvc:      &mockAuditService{},
		buildClientFn: func(_ context.Context, _ uint64) (lstepapi.Client, error) { return client, nil },
	}
	result, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1},
		TagName:     "campaign_2026",
	}, nil)
	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, 1, result.FailedCount)
		assert.Contains(t, result.FailedOwnerIDs, uint64(1))
		assert.Equal(t, 0, result.SuccessCount)
	}
}

func TestCheckupSyncService_CreateCheckupSync_SuccessUpsertsCache(t *testing.T) {
	lineID := "U_test"
	ownerRepo := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, _ uint64, ids []uint64) ([]*model.Owner, error) {
			return []*model.Owner{{ID: ids[0], LineUserID: &lineID, LstepOptOut: false}}, nil
		},
	}
	petRepo := &mockPetRepository{
		countLivingByOwnerIDsFn: func(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64]int64, error) {
			m := make(map[uint64]int64, len(ownerIDs))
			for _, id := range ownerIDs {
				m[id] = 1
			}
			return m, nil
		},
	}
	var upsertedTag string
	var addedTagCalled bool
	client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { addedTagCalled = true; return nil }}
	svc := &checkupSyncService{
		repo:      &mockCheckupSyncRepository{},
		ownerRepo: ownerRepo,
		petRepo:   petRepo,
		tagCacheRepo: &mockLstepTagCacheRepository{
			upsertTagFn: func(_ context.Context, _, _ uint64, tagName, _, _ string) error {
				upsertedTag = tagName
				return nil
			},
		},
		settingsSvc:   &mockLstepSettingsService{},
		auditSvc:      &mockAuditService{},
		buildClientFn: func(_ context.Context, _ uint64) (lstepapi.Client, error) { return client, nil },
	}
	result, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1},
		TagName:     "campaign_2026",
	}, nil)
	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, 1, result.SuccessCount)
	}
	assert.True(t, addedTagCalled)
	assert.Equal(t, "campaign_2026", upsertedTag)
}

func TestCheckupSyncService_CreateCheckupSync_PropagatesClinicScopeToBatchDependencies(t *testing.T) {
	const clinicID = uint64(17)
	lineID := "U_scope"
	settings := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, gotClinicID uint64) (bool, error) {
			assert.Equal(t, clinicID, gotClinicID)
			return true, nil
		},
		getRawCredentialsFn: func(_ context.Context, gotClinicID uint64) (string, string, string, error) {
			assert.Equal(t, clinicID, gotClinicID)
			return "test-key", "https://example.com", "", nil
		},
	}
	owners := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, gotClinicID uint64, ownerIDs []uint64) ([]*model.Owner, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, []uint64{1, 2}, ownerIDs)
			return []*model.Owner{
				{ID: 1, ClinicID: clinicID, LineUserID: &lineID},
				{ID: 2, ClinicID: clinicID, LstepOptOut: true},
			}, nil
		},
	}
	pets := &mockPetRepository{
		countLivingByOwnerIDsFn: func(_ context.Context, gotClinicID uint64, ownerIDs []uint64) (map[uint64]int64, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, []uint64{1}, ownerIDs, "opt-out owner must not enter the pet batch")
			return map[uint64]int64{1: 1}, nil
		},
	}
	cache := &mockLstepTagCacheRepository{
		upsertTagFn: func(_ context.Context, gotClinicID, ownerID uint64, _, _, _ string) error {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, uint64(1), ownerID)
			return nil
		},
	}
	audit := &mockAuditService{
		logLstepOperationWithMetadataFn: func(_ context.Context, gotClinicID uint64, _ *uint64, action, _ string, _ *uint64, _ any) error {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, "checkup_sync", action)
			return nil
		},
	}
	// Gate-independent: inject mock write client so default-off LSTEP_WRITE_API_ENABLED
	// (TASK-610) does not mask clinic-scope propagation assertions. Do not enable the
	// deploy gate; keep SuccessCount/SkippedCount as the clinic-scope oracle.
	svc := &checkupSyncService{
		repo:         &mockCheckupSyncRepository{},
		ownerRepo:    owners,
		petRepo:      pets,
		tagCacheRepo: cache,
		settingsSvc:  settings,
		auditSvc:     audit,
		buildClientFn: func(_ context.Context, gotClinicID uint64) (lstepapi.Client, error) {
			assert.Equal(t, clinicID, gotClinicID)
			return &mockLstepAPIClient{
				addTagFn: func(_ context.Context, _, _ string) error { return nil },
			}, nil
		},
	}

	result, err := svc.CreateCheckupSync(context.Background(), clinicID, CreateCheckupSyncInput{
		OwnerIDs: []uint64{1, 2},
		TagName:  "campaign_2026",
	}, nil)

	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 1, result.SkippedCount)
	}
}

func TestCheckupSyncService_CreateCheckupSync_AddTagFailureIsPerOwnerAndCacheFailureIsNonFatal(t *testing.T) {
	lineIDs := map[uint64]string{1: "U1", 2: "U2", 3: "U3"}
	owners := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, _ uint64, ownerIDs []uint64) ([]*model.Owner, error) {
			result := make([]*model.Owner, 0, len(ownerIDs))
			for _, id := range ownerIDs {
				lineID := lineIDs[id]
				result = append(result, &model.Owner{ID: id, LineUserID: &lineID})
			}
			return result, nil
		},
	}
	pets := &mockPetRepository{
		countLivingByOwnerIDsFn: func(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64]int64, error) {
			counts := make(map[uint64]int64, len(ownerIDs))
			for _, id := range ownerIDs {
				counts[id] = 1
			}
			return counts, nil
		},
	}
	client := &mockLstepAPIClient{
		addTagFn: func(_ context.Context, lineUserID, _ string) error {
			if lineUserID == "U1" {
				return errors.New("first owner failed")
			}
			return nil
		},
	}
	var upsertedOwnerIDs []uint64
	cache := &mockLstepTagCacheRepository{
		upsertTagFn: func(_ context.Context, _ uint64, ownerID uint64, _, _, _ string) error {
			upsertedOwnerIDs = append(upsertedOwnerIDs, ownerID)
			if ownerID == 2 {
				return errors.New("cache unavailable")
			}
			return nil
		},
	}
	svc := &checkupSyncService{
		ownerRepo:    owners,
		petRepo:      pets,
		tagCacheRepo: cache,
		auditSvc:     &mockAuditService{},
		buildClientFn: func(_ context.Context, _ uint64) (lstepapi.Client, error) {
			return client, nil
		},
	}

	result, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		OwnerIDs: []uint64{1, 2, 3},
		TagName:  "campaign_2026",
	}, nil)

	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, 2, result.SuccessCount, "cache failure remains non-fatal")
		assert.Equal(t, 1, result.FailedCount)
		assert.Equal(t, []uint64{1}, result.FailedOwnerIDs)
	}
	assert.Equal(t, []uint64{2, 3}, upsertedOwnerIDs, "processing continues after per-owner failures")
}
