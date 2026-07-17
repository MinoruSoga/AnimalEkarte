package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// TestSyncHealthPreventionTagsForClinic_SkipsWhenSyncDisabled covers the shouldSkipSync
// short-circuit branch (settingsSvc.IsSyncEnabled=false).
func TestSyncHealthPreventionTagsForClinic_SkipsWhenSyncDisabled(t *testing.T) {
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
	}
	count, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.Empty(t, errs)
}

func TestSyncHealthPreventionTagsForClinic_ShouldSkipSyncError(t *testing.T) {
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("db error") }},
	}
	count, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.NotEmpty(t, errs)
}

func TestSyncHealthPreventionTagsForClinic_FindOwnersError(t *testing.T) {
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{},
		ownerRepo: &mockOwnerRepository{
			findAllWithLineUserIDCursorFn: func(_ context.Context, _ uint64, _ uint64, _ int) ([]model.Owner, error) {
				return nil, errors.New("db error")
			},
		},
	}
	count, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.NotEmpty(t, errs)
}

// TestSyncHealthPreventionTagsForClinic_TagCodeMappingCacheErrorIsNonFatal verifies that a
// failure caching tag-code mappings up front falls back to per-owner fetch instead of aborting
// the whole batch (production comment: "fallback to per-owner fetch").
func TestSyncHealthPreventionTagsForClinic_TagCodeMappingCacheErrorIsNonFatal(t *testing.T) {
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{},
		ownerRepo: &mockOwnerRepository{
			findAllWithLineUserIDCursorFn: func(_ context.Context, _ uint64, _ uint64, _ int) ([]model.Owner, error) {
				return nil, nil // no owners, isolates this branch
			},
		},
		tagCodeRepo: &mockLstepTagCodeMappingRepository{
			findByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) ([]*model.LstepTagCodeMapping, error) {
				return nil, errors.New("cache fetch failed")
			},
		},
	}
	count, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.Empty(t, errs, "mapping cache failure alone must not surface as a batch error with zero owners")
}

func TestSyncHealthPreventionTagsForClinic_ThresholdsError(t *testing.T) {
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{
			getHealthPreventionThresholdsFn: func(_ context.Context, _ uint64) (model.HealthPreventionThresholds, error) {
				return model.HealthPreventionThresholds{}, errors.New("thresholds db error")
			},
		},
		ownerRepo: &mockOwnerRepository{
			findAllWithLineUserIDCursorFn: func(_ context.Context, _ uint64, _ uint64, _ int) ([]model.Owner, error) {
				return []model.Owner{{ID: 1}}, nil
			},
		},
		// tagCodeRepo must be non-nil: the batch-level mapping cache lookup runs unconditionally
		// once owners are found, before the thresholds fetch below it fails.
		tagCodeRepo: &mockLstepTagCodeMappingRepository{},
	}
	count, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.NotEmpty(t, errs)
}

// TestSyncHealthPreventionTagsForClinic_SuccessCountsOwnersWithNoFailures verifies that, with
// owners that have no LINE user id, every FEAT-379 sub-sync short-circuits cleanly right after
// checkOptOut and every owner counts as successfully processed. billingItemRepo is left nil so
// SyncFleaTickTag/SyncFoodPurchaseTag no-op via their own "tagCodeRepo/billingItemRepo nil"
// contract instead.
func TestSyncHealthPreventionTagsForClinic_SuccessCountsOwnersWithNoFailures(t *testing.T) {
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{},
		ownerRepo: &mockOwnerRepository{
			findAllWithLineUserIDCursorFn: func(_ context.Context, _ uint64, _ uint64, _ int) ([]model.Owner, error) {
				return []model.Owner{{ID: 1}, {ID: 2}}, nil
			},
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return &model.Owner{ID: id, LineUserID: nil}, nil
			},
		},
		vacRepo:     &mockVaccinationRepository{},
		tagCodeRepo: &mockLstepTagCodeMappingRepository{},
	}
	count, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), 1)
	assert.Equal(t, 2, count)
	assert.Empty(t, errs)
}

// TestSyncHealthPreventionTagsForClinic_OwnerFailureExcludesFromCount verifies that when a
// per-owner sub-sync fails, that owner is excluded from count and the error is aggregated
// (batch itself does not abort — "全体は失敗しない").
func TestSyncHealthPreventionTagsForClinic_OwnerFailureExcludesFromCount(t *testing.T) {
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{},
		ownerRepo: &mockOwnerRepository{
			findAllWithLineUserIDCursorFn: func(_ context.Context, _ uint64, _ uint64, _ int) ([]model.Owner, error) {
				return []model.Owner{{ID: 1}}, nil
			},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				// checkOptOut's FindByID fails for every sub-sync that reaches it.
				return nil, errors.New("owner lookup failed")
			},
		},
		tagCodeRepo: &mockLstepTagCodeMappingRepository{
			findByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) ([]*model.LstepTagCodeMapping, error) {
				return nil, nil
			},
		},
		// billingItemRepo intentionally nil: SyncFleaTickTag/SyncFoodPurchaseTag no-op before
		// reaching checkOptOut, so only the other 4 sub-syncs contribute failures.
	}
	count, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), 1)
	assert.Equal(t, 0, count, "owner with any failing sub-sync must not be counted as processed")
	assert.NotEmpty(t, errs)
}

// TestSyncHealthPreventionTagsForClinic_PaginatesAcrossMultiplePages verifies PERF-FOLLOWUP-02
// cursor pagination: when the first page is exactly full (pageSize), a second page fetch is
// issued using the last owner's ID as the cursor, and owners from both pages are processed.
func TestSyncHealthPreventionTagsForClinic_PaginatesAcrossMultiplePages(t *testing.T) {
	fetchCalls := make([]uint64, 0, 3)
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{},
		ownerRepo: &mockOwnerRepository{
			findAllWithLineUserIDCursorFn: func(_ context.Context, _ uint64, afterID uint64, limit int) ([]model.Owner, error) {
				fetchCalls = append(fetchCalls, afterID)
				assert.Equal(t, lstepBatchPageSize, limit)
				switch afterID {
				case 0:
					owners := make([]model.Owner, lstepBatchPageSize)
					for i := range owners {
						owners[i] = model.Owner{ID: uint64(i + 1)}
					}
					return owners, nil
				case uint64(lstepBatchPageSize):
					return []model.Owner{{ID: uint64(lstepBatchPageSize + 1)}}, nil
				default:
					return nil, nil
				}
			},
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return &model.Owner{ID: id, LineUserID: nil}, nil
			},
		},
		vacRepo:     &mockVaccinationRepository{},
		tagCodeRepo: &mockLstepTagCodeMappingRepository{},
	}
	count, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), 1)
	assert.Empty(t, errs)
	assert.Equal(t, lstepBatchPageSize+1, count, "owners from both pages must be processed")
	assert.Equal(t, []uint64{0, uint64(lstepBatchPageSize)}, fetchCalls, "cursor must advance using the last owner ID of the previous page, no duplicates/no skips")
}
