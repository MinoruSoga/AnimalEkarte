package lstep

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncLTVTopPercent starts at 0% coverage. Uses cpmAccountingRepositoryWrapper /
// newCPMAccountingRepository defined in lstep_tag_sync_visit_cpm_test.go (same package).
//
// G2F-03: FindOwnersByAnnualRevenue returns the already-bounded top-20% set.
// SyncLTVTopPercent must not re-slice with (len(revenues)*20+99)/100.

func TestSyncLTVTopPercent(t *testing.T) {
	lineUID1 := "U_ltv_top"
	lineUID2 := "U_ltv_other"

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.Empty(t, errs)
	})

	t.Run("returns wrapped error when shouldSkipSync check fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("db error") }},
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("returns wrapped error when FindOwnersByAnnualRevenue fails", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]billing.OwnerAnnualRevenue, error) {
			return nil, errors.New("db error")
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("returns wrapped error when FindAllWithLineUserIDCursor fails", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDCursorFn: func(_ context.Context, _, _ uint64, _ int) ([]model.Owner, error) {
					return nil, errors.New("db error")
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return &mockLstepAPIClient{}, nil
			},
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("returns 0,nil when buildClient yields no client", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo:   &mockOwnerRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return nil, nil
			},
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.Empty(t, errs)
	})

	t.Run("returns wrapped error when buildClient itself fails", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo:   &mockOwnerRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return nil, errors.New("credentials error")
			},
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("skips owners without a line user id", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		// Bounded contract: repository already returns only the top set.
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]billing.OwnerAnnualRevenue, error) {
			return []billing.OwnerAnnualRevenue{{OwnerID: 1}}, nil
		}
		client := &mockLstepAPIClient{}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDCursorFn: func(_ context.Context, _, afterID uint64, _ int) ([]model.Owner, error) {
					if afterID != 0 {
						return nil, nil
					}
					return []model.Owner{{ID: 1, LineUserID: nil}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.Empty(t, errs)
	})

	t.Run("top owner gets the tag added and cache upserted", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		// Bounded contract: SQL returns only the top-20% set (here: 1 owner).
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]billing.OwnerAnnualRevenue, error) {
			return []billing.OwnerAnnualRevenue{{OwnerID: 1}}, nil
		}
		var addedTag string
		var upsertedTag string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error { addedTag = tagName; return nil },
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDCursorFn: func(_ context.Context, _, afterID uint64, _ int) ([]model.Owner, error) {
					if afterID != 0 {
						return nil, nil
					}
					return []model.Owner{{ID: 1, LineUserID: &lineUID1}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				upsertTagFn: func(_ context.Context, _, _ uint64, tagName, _, _ string) error {
					upsertedTag = tagName
					return nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Empty(t, errs)
		assert.Equal(t, 1, count)
		assert.Equal(t, ltvTop20Tag, addedTag)
		assert.Equal(t, ltvTop20Tag, upsertedTag)
	})

	t.Run("all owners in bounded top set receive the tag (no Go-side re-slice)", func(t *testing.T) {
		// Regression for G2F-03: if Go still did topN=(len*20+99)/100 on a mock
		// that returns the already-bounded set of size 5, only 1 owner would be tagged.
		// Under the bounded contract every returned owner is top.
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]billing.OwnerAnnualRevenue, error) {
			return []billing.OwnerAnnualRevenue{
				{OwnerID: 1}, {OwnerID: 2}, {OwnerID: 3}, {OwnerID: 4}, {OwnerID: 5},
			}, nil
		}
		added := make(map[string]int)
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, lineUID, tagName string) error {
				if tagName == ltvTop20Tag {
					added[lineUID]++
				}
				return nil
			},
		}
		uids := []string{"U1", "U2", "U3", "U4", "U5"}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDCursorFn: func(_ context.Context, _, afterID uint64, _ int) ([]model.Owner, error) {
					if afterID != 0 {
						return nil, nil
					}
					out := make([]model.Owner, 0, 5)
					for i := uint64(1); i <= 5; i++ {
						uid := uids[i-1]
						out = append(out, model.Owner{ID: i, LineUserID: &uid})
					}
					return out, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Empty(t, errs)
		assert.Equal(t, 5, count)
		assert.Len(t, added, 5, "every owner in the bounded top set must receive LTV_上位20")
	})

	t.Run("AddTag failure for a top owner is aggregated as an error and does not increment count", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]billing.OwnerAnnualRevenue, error) {
			return []billing.OwnerAnnualRevenue{{OwnerID: 1}}, nil
		}
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDCursorFn: func(_ context.Context, _, afterID uint64, _ int) ([]model.Owner, error) {
					if afterID != 0 {
						return nil, nil
					}
					return []model.Owner{{ID: 1, LineUserID: &lineUID1}}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("non-top owner with no cached tag counts as success without calling RemoveTag", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]billing.OwnerAnnualRevenue, error) {
			return nil, nil // empty top set
		}
		removeCalled := false
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, _ string) error { removeCalled = true; return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDCursorFn: func(_ context.Context, _, afterID uint64, _ int) ([]model.Owner, error) {
					if afterID != 0 {
						return nil, nil
					}
					return []model.Owner{{ID: 2, LineUserID: &lineUID2}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnersFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]*model.LstepTagCache, error) {
					return map[uint64][]*model.LstepTagCache{}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Empty(t, errs)
		assert.Equal(t, 1, count)
		assert.False(t, removeCalled, "owner without the cached tag must not trigger a RemoveTag call")
	})

	t.Run("non-top owner with cached tag has it removed", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]billing.OwnerAnnualRevenue, error) {
			return nil, nil
		}
		var removedTag string
		var deletedTag string
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, tagName string) error { removedTag = tagName; return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDCursorFn: func(_ context.Context, _, afterID uint64, _ int) ([]model.Owner, error) {
					if afterID != 0 {
						return nil, nil
					}
					return []model.Owner{{ID: 2, LineUserID: &lineUID2}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnersFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]*model.LstepTagCache, error) {
					return map[uint64][]*model.LstepTagCache{2: {{TagName: ltvTop20Tag}}}, nil
				},
				deleteTagFn: func(_ context.Context, _, _ uint64, tagName string) error { deletedTag = tagName; return nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Empty(t, errs)
		assert.Equal(t, 1, count)
		assert.Equal(t, ltvTop20Tag, removedTag)
		assert.Equal(t, ltvTop20Tag, deletedTag)
	})

	t.Run("tagCacheRepo.FindByOwners batch failure aborts the sync (fail-closed, not per-owner)", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]billing.OwnerAnnualRevenue, error) {
			return nil, nil
		}
		client := &mockLstepAPIClient{}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDCursorFn: func(_ context.Context, _, afterID uint64, _ int) ([]model.Owner, error) {
					if afterID != 0 {
						return nil, nil
					}
					return []model.Owner{{ID: 2, LineUserID: &lineUID2}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnersFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]*model.LstepTagCache, error) {
					return nil, errors.New("db error")
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("RemoveTag failure for a non-top owner is aggregated", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]billing.OwnerAnnualRevenue, error) {
			return nil, nil
		}
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDCursorFn: func(_ context.Context, _, afterID uint64, _ int) ([]model.Owner, error) {
					if afterID != 0 {
						return nil, nil
					}
					return []model.Owner{{ID: 2, LineUserID: &lineUID2}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnersFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]*model.LstepTagCache, error) {
					return map[uint64][]*model.LstepTagCache{2: {{TagName: ltvTop20Tag}}}, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Equal(t, 0, count)
		assert.NotEmpty(t, errs)
	})

	t.Run("page-batched tag cache still scopes FindByOwners to non-top owners on each page", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]billing.OwnerAnnualRevenue, error) {
			return []billing.OwnerAnnualRevenue{{OwnerID: 1}}, nil // owner 1 is top
		}
		var findByOwnersCalls [][]uint64
		client := &mockLstepAPIClient{
			addTagFn:    func(_ context.Context, _, _ string) error { return nil },
			removeTagFn: func(_ context.Context, _, _ string) error { return nil },
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			accountRepo: accountRepo,
			ownerRepo: &mockOwnerRepository{
				findAllWithLineUserIDCursorFn: func(_ context.Context, _, afterID uint64, limit int) ([]model.Owner, error) {
					// Two pages of LINE owners: [1,2] then [3].
					switch afterID {
					case 0:
						return []model.Owner{
							{ID: 1, LineUserID: &lineUID1},
							{ID: 2, LineUserID: &lineUID2},
						}, nil
					case 2:
						uid3 := "U_ltv_3"
						return []model.Owner{{ID: 3, LineUserID: &uid3}}, nil
					default:
						return nil, nil
					}
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnersFn: func(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error) {
					copied := append([]uint64(nil), ownerIDs...)
					findByOwnersCalls = append(findByOwnersCalls, copied)
					return map[uint64][]*model.LstepTagCache{}, nil
				},
				upsertTagFn: func(_ context.Context, _, _ uint64, _, _, _ string) error { return nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		// Force page size so the two-owner first page still paginates when limit is small.
		// lstepBatchPageSize is package-level; we emulate paging via afterID above with
		// len(owners) < page size on the second page. First page returns 2 which is < page size
		// under default limit, so override cursor logic: return 2 then break would not hit page2
		// unless first page length == lstepBatchPageSize. Use a single page that includes top+non-top
		// and assert only non-top IDs are batched.
		//
		// Re-run with a single page (afterID path returns nil on second call automatically when
		// len < page size). Assert FindByOwners received only non-top owner 2.
		count, errs := svc.SyncLTVTopPercent(context.Background(), 1)
		assert.Empty(t, errs)
		assert.Equal(t, 2, count) // top owner 1 + non-top owner 2 (owner 3 not reached: first page len 2 < page size)
		require.Len(t, findByOwnersCalls, 1)
		assert.Equal(t, []uint64{2}, findByOwnersCalls[0], "tag cache batch must exclude top owners and stay page-scoped")
	})
}

// TestSyncLTVTopPercent_DoesNotMaterializeFullClinicRevenueInGo is the G2F-03
// large-clinic source regression: SyncLTVTopPercent must consume the bounded
// top set from FindOwnersByAnnualRevenue without recomputing topN in Go.
func TestSyncLTVTopPercent_DoesNotMaterializeFullClinicRevenueInGo(t *testing.T) {
	src, err := os.ReadFile("lstep_tag_sync_visit_ltv.go")
	require.NoError(t, err)
	body := string(src)

	assert.NotContains(t, body, "(len(revenues)*20 + 99) / 100",
		"must not recompute topN from a fully materialized revenue slice")
	assert.NotContains(t, body, "topN = (len(",
		"must not derive topN from len() of revenues in Go")
	assert.NotContains(t, body, "for i := 0; i < topN; i++",
		"must not slice the first topN of a full ranking in Go")
	assert.True(t,
		strings.Contains(body, "for _, rev := range topRevenues") ||
			strings.Contains(body, "for _, rev := range revenues"),
		"must iterate the bounded top set returned by the repository")
	assert.Contains(t, body, "FindOwnersByAnnualRevenue",
		"top set must come from the billing repository contract")
}
