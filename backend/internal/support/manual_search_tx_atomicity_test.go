package support

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// reviewAmbientTransaction bounds the probe and rolls back at most once.
// LIFO cleanup rolls back before deleting the probe's rows.
func reviewAmbientTransaction(t *testing.T, db *gorm.DB) (context.Context, *gorm.DB, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	tx := db.WithContext(ctx).Begin()
	require.NoError(t, tx.Error)
	done := false
	rollback := func() {
		if done {
			return
		}
		done = true
		require.NoError(t, tx.Rollback().Error)
	}
	t.Cleanup(rollback)
	return ctx, tx, rollback
}

func testReviewAmbientCacheRollback(t *testing.T, db *gorm.DB, namespace string, ambient bool) {
	t.Helper()
	model := namespace + "-model"
	key := manualarticle.ChunkCacheKey(model, manualarticle.TaskTypeRetrievalDocument, namespace+"-cache")
	cleanupReviewEmbeddingRows(t, db, key)

	ctx, tx, rollback := reviewAmbientTransaction(t, db)
	actionCtx := ctx
	if ambient {
		actionCtx = persistence.WithTxValue(ctx, tx)
	}

	cache := NewManualEmbeddingCache(db, model)
	require.NoError(t, cache.StoreWithKey(
		actionCtx, key, model, manualarticle.TaskTypeRetrievalDocument,
		manualarticle.ChunkerVersion, manualarticle.EmbeddingDimensions, reviewVector(0),
	))
	stored, err := cache.Load(actionCtx, []string{key})
	require.NoError(t, err)
	require.Len(t, stored, 1)
	require.Len(t, stored[key], manualarticle.EmbeddingDimensions)
	require.InDelta(t, 1, stored[key][0], 1e-5)

	want := 1
	if ambient {
		want = 0
	}
	outside, err := cache.Load(ctx, []string{key})
	require.NoError(t, err)
	assert.Len(t, outside, want)

	rollback()
	after, err := cache.Load(ctx, []string{key})
	require.NoError(t, err)
	assert.Len(t, after, want)
}

func testReviewAmbientBudgetRollback(t *testing.T, db *gorm.DB, namespace string, ambient bool) {
	t.Helper()
	scope := namespace + "-scope"
	model := namespace + "-model"
	cleanupReviewBudgetRows(t, db, reviewBudgetKey{scope: scope, model: model})

	ctx, tx, rollback := reviewAmbientTransaction(t, db)
	actionCtx := ctx
	if ambient {
		actionCtx = persistence.WithTxValue(ctx, tx)
	}

	granter := NewLLMBudgetGranter(db)
	granted, err := granter.GrantAll(actionCtx, []BudgetGrant{{Scope: scope, Model: model, Interval: time.Hour}})
	require.NoError(t, err)
	require.True(t, granted)

	var inside int64
	require.NoError(t, tx.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM support_llm_budgets WHERE scope = ? AND model = ?", scope, model,
	).Scan(&inside).Error)
	assert.EqualValues(t, 1, inside)

	want := int64(1)
	if ambient {
		want = 0
	}
	var before int64
	require.NoError(t, db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM support_llm_budgets WHERE scope = ? AND model = ?", scope, model,
	).Scan(&before).Error)
	assert.Equal(t, want, before)

	rollback()
	var after int64
	require.NoError(t, db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM support_llm_budgets WHERE scope = ? AND model = ?", scope, model,
	).Scan(&after).Error)
	assert.Equal(t, want, after)
}

func TestRealDB_SupportManualSearchAmbientTx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real database test in short mode")
	}
	if os.Getenv("SUPPORT_MANUAL_SQL_INTEGRATION") != "1" {
		t.Skip("requires an isolated PostgreSQL instance and explicit human approval; set SUPPORT_MANUAL_SQL_INTEGRATION=1")
	}
	db := testdb.SetupTestDB(t)
	ensureSupportManualSearchTables(t, db)

	cases := []struct {
		name    string
		run     func(t *testing.T, db *gorm.DB, namespace string, ambient bool)
		ambient bool
	}{
		{name: "cache_ambient_rollback", run: testReviewAmbientCacheRollback, ambient: true},
		{name: "cache_without_ambient_control", run: testReviewAmbientCacheRollback, ambient: false},
		{name: "budget_ambient_rollback", run: testReviewAmbientBudgetRollback, ambient: true},
		{name: "budget_without_ambient_control", run: testReviewAmbientBudgetRollback, ambient: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.run(t, db, reviewManualSearchNamespace(t), tc.ambient)
		})
	}
}
