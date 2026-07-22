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
