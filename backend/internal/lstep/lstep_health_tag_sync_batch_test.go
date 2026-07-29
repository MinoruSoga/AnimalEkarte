package lstep

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
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

// bulk-capable page loaders for G2F-02 query-count regression.
type bulkCheckupRepo struct {
	findByOwnerIDCalls  int64
	findByOwnerIDsCalls int64
	lastBulkSize        int
}

func (m *bulkCheckupRepo) FindByOwnerID(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
	atomic.AddInt64(&m.findByOwnerIDCalls, 1)
	return nil, nil
}

func (m *bulkCheckupRepo) FindByOwnerIDs(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64][]model.Checkup, error) {
	atomic.AddInt64(&m.findByOwnerIDsCalls, 1)
	m.lastBulkSize = len(ownerIDs)
	return map[uint64][]model.Checkup{}, nil
}

type bulkVaccinationRepo struct {
	findByOwnerCalls    int64
	findByOwnerIDsCalls int64
	lastBulkSize        int
}

func (m *bulkVaccinationRepo) FindByID(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
	return nil, nil
}

func (m *bulkVaccinationRepo) FindByOwner(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
	atomic.AddInt64(&m.findByOwnerCalls, 1)
	return nil, nil
}

func (m *bulkVaccinationRepo) FindByOwnerIDs(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64][]model.Vaccination, error) {
	atomic.AddInt64(&m.findByOwnerIDsCalls, 1)
	m.lastBulkSize = len(ownerIDs)
	return map[uint64][]model.Vaccination{}, nil
}

type bulkMedRecordRepo struct {
	findVisitSummaryCalls     int64
	findVisitSummariesCalls   int64
	lastBulkSize              int
}

func (m *bulkMedRecordRepo) FindOwnerVisitSummary(_ context.Context, _, _ uint64) (*medicalrecord.OwnerVisitSummary, error) {
	atomic.AddInt64(&m.findVisitSummaryCalls, 1)
	return &medicalrecord.OwnerVisitSummary{}, nil
}

func (m *bulkMedRecordRepo) FindLatestByOwner(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}

func (m *bulkMedRecordRepo) FindOwnerVisitSummariesByOwnerIDs(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64]*medicalrecord.OwnerVisitSummary, error) {
	atomic.AddInt64(&m.findVisitSummariesCalls, 1)
	m.lastBulkSize = len(ownerIDs)
	return map[uint64]*medicalrecord.OwnerVisitSummary{}, nil
}

// TestSyncHealthPreventionTagsForClinic_PageBulkChildQueriesArePageBounded proves G2F-02:
// over two owner pages, checkup/vaccination/visit-summary loads are 1 bulk call per page
// (page-bounded), never 1 call per owner (owner-linear).
func TestSyncHealthPreventionTagsForClinic_PageBulkChildQueriesArePageBounded(t *testing.T) {
	checkupRepo := &bulkCheckupRepo{}
	vacRepo := &bulkVaccinationRepo{}
	medRepo := &bulkMedRecordRepo{}

	lineID := "U_bulk_page"
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
			getHealthPreventionThresholdsFn: func(_ context.Context, _ uint64) (model.HealthPreventionThresholds, error) {
				return model.HealthPreventionThresholds{}.WithDefaults(), nil
			},
			// empty credentials → buildClient returns nil client after child inputs are consumed
		},
		ownerRepo: &mockOwnerRepository{
			findAllWithLineUserIDCursorFn: func(_ context.Context, _ uint64, afterID uint64, limit int) ([]model.Owner, error) {
				require.Equal(t, lstepBatchPageSize, limit)
				switch afterID {
				case 0:
					owners := make([]model.Owner, lstepBatchPageSize)
					for i := range owners {
						owners[i] = model.Owner{ID: uint64(i + 1), LineUserID: &lineID}
					}
					return owners, nil
				case uint64(lstepBatchPageSize):
					return []model.Owner{{ID: uint64(lstepBatchPageSize + 1), LineUserID: &lineID}}, nil
				default:
					return nil, nil
				}
			},
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				// LINE linked so sub-syncs reach child-input consumption (would hit per-owner
				// Find* if page bulk failed to preload).
				return &model.Owner{ID: id, LineUserID: &lineID}, nil
			},
		},
		checkupRepo:   checkupRepo,
		vacRepo:       vacRepo,
		medRecordRepo: medRepo,
		// Non-empty healthcheck mappings so healthcheck/annual4 consume preloaded checkups/visit.
		tagCodeRepo: &mockLstepTagCodeMappingRepository{
			findByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, tagName string) ([]*model.LstepTagCodeMapping, error) {
				if tagName == HlthHealthcheckDoneTag {
					return []*model.LstepTagCodeMapping{{
						TagName:  tagName,
						CodeType: model.CodeTypeCheckupType,
						Codes:    []string{"健診A"},
					}}, nil
				}
				return nil, nil
			},
		},
		tagCacheRepo: &mockLstepTagCacheRepository{},
	}

	_, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), 1)
	assert.Empty(t, errs)

	const pages = 2
	totalOwners := lstepBatchPageSize + 1

	assert.Equal(t, int64(pages), atomic.LoadInt64(&checkupRepo.findByOwnerIDsCalls),
		"checkup bulk must be 1 call per page (got %d)", atomic.LoadInt64(&checkupRepo.findByOwnerIDsCalls))
	assert.Equal(t, int64(0), atomic.LoadInt64(&checkupRepo.findByOwnerIDCalls),
		"per-owner checkup FindByOwnerID must not run when bulk loaded (got %d; owner-linear would be ~%d)",
		atomic.LoadInt64(&checkupRepo.findByOwnerIDCalls), totalOwners*2) // healthcheck + annual4

	assert.Equal(t, int64(pages), atomic.LoadInt64(&vacRepo.findByOwnerIDsCalls),
		"vaccination bulk must be 1 call per page")
	assert.Equal(t, int64(0), atomic.LoadInt64(&vacRepo.findByOwnerCalls),
		"per-owner vaccination FindByOwner must not run when bulk loaded")

	assert.Equal(t, int64(pages), atomic.LoadInt64(&medRepo.findVisitSummariesCalls),
		"visit-summary bulk must be 1 call per page")
	assert.Equal(t, int64(0), atomic.LoadInt64(&medRepo.findVisitSummaryCalls),
		"per-owner FindOwnerVisitSummary must not run when bulk loaded")

	// Second page has 1 owner; last bulk size should reflect that page.
	assert.Equal(t, 1, medRepo.lastBulkSize, "last page bulk size should equal last page owner count")
}
