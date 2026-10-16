package support

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type reviewBudgetKey struct {
	scope string
	model string
}

func reviewManualSearchNamespace(t *testing.T) string {
	t.Helper()
	var raw [16]byte
	_, err := rand.Read(raw[:])
	require.NoError(t, err)
	return "emr410-review-" + hex.EncodeToString(raw[:])
}

func reviewVector(index int) []float64 {
	vec := make([]float64, manualarticle.EmbeddingDimensions)
	vec[index] = 1
	return vec
}

func reviewArrayLiteral(length, dimensions, lowerBound int) string {
	values := make([]string, length)
	for i := range values {
		values[i] = "0"
	}
	if len(values) > 0 {
		values[0] = "1"
	}
	inner := strings.Join(values, ",")
	switch {
	case dimensions > 1:
		return "{" + "{" + inner + "}" + "}"
	case lowerBound != 1:
		return fmt.Sprintf("[%d:%d]={%s}", lowerBound, lowerBound+length-1, inner)
	default:
		return "{" + inner + "}"
	}
}

func requireReviewPostgresCode(t *testing.T, err error, code string) {
	t.Helper()
	require.Error(t, err)
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	assert.Equal(t, code, pgErr.Code)
}

func cleanupReviewBudgetRows(t *testing.T, db *gorm.DB, keys ...reviewBudgetKey) {
	t.Helper()
	t.Cleanup(func() {
		for _, key := range keys {
			require.NoError(t, db.Exec(
				"DELETE FROM support_llm_budgets WHERE scope = ? AND model = ?",
				key.scope,
				key.model,
			).Error)
		}
	})
}

func cleanupReviewEmbeddingRows(t *testing.T, db *gorm.DB, keys ...string) {
	t.Helper()
	t.Cleanup(func() {
		for _, key := range keys {
			require.NoError(t, db.Exec(
				"DELETE FROM support_manual_embeddings WHERE cache_key = ?",
				key,
			).Error)
		}
	})
}

func TestRealDB_SupportManualSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("requires an isolated PostgreSQL instance and explicit human approval; not run in -short mode")
	}
	if os.Getenv("SUPPORT_MANUAL_SQL_INTEGRATION") != "1" {
		t.Skip("requires an isolated PostgreSQL instance and explicit human approval; set SUPPORT_MANUAL_SQL_INTEGRATION=1")
	}

	db := testdb.SetupTestDB(t)
	migration, err := os.ReadFile("../../migrations/037_support_manual_search.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(migration)).Error)
	namespace := reviewManualSearchNamespace(t)

	cases := []struct {
		name string
		run  func(*testing.T, *gorm.DB, string)
	}{
		{name: "concurrent budget grants", run: testReviewConcurrentBudgetGrants},
		{name: "all-or-none budget rollback", run: testReviewBudgetRollback},
		{name: "canceled budget grant", run: testReviewCanceledBudgetGrant},
		{name: "embedding cache persistence", run: testReviewEmbeddingCache},
		{name: "database constraints", run: testReviewManualSearchConstraints},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.run(t, db, namespace+"-"+strings.ReplaceAll(tc.name, " ", "-"))
		})
	}
}

func testReviewConcurrentBudgetGrants(t *testing.T, db *gorm.DB, namespace string) {
	t.Helper()
	scope := namespace + "-scope"
	model := namespace + "-model"
	otherModel := namespace + "-other-model"
	cleanupReviewBudgetRows(t, db,
		reviewBudgetKey{scope: scope, model: model},
		reviewBudgetKey{scope: scope, model: otherModel},
	)

	const workers = 12
	type result struct {
		granted bool
		err     error
	}
	start := make(chan struct{})
	results := make(chan result, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			granted, err := NewLLMBudgetGranter(db).GrantAll(context.Background(), []BudgetGrant{{
				Scope: scope, Model: model, Interval: time.Hour,
			}})
			results <- result{granted: granted, err: err}
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	grants := 0
	for result := range results {
		require.NoError(t, result.err)
		if result.granted {
			grants++
		}
	}
	require.Equal(t, 1, grants)

	granted, err := NewLLMBudgetGranter(db).GrantAll(context.Background(), []BudgetGrant{{
		Scope: scope, Model: model, Interval: time.Hour,
	}})
	require.NoError(t, err)
	assert.False(t, granted)

	granted, err = NewLLMBudgetGranter(db).GrantAll(context.Background(), []BudgetGrant{{
		Scope: scope, Model: otherModel, Interval: time.Hour,
	}})
	require.NoError(t, err)
	assert.True(t, granted)
}

func testReviewBudgetRollback(t *testing.T, db *gorm.DB, namespace string) {
	t.Helper()
	scopeA := namespace + "-a"
	scopeB := namespace + "-b"
	model := namespace + "-model"
	cleanupReviewBudgetRows(t, db,
		reviewBudgetKey{scope: scopeA, model: model},
		reviewBudgetKey{scope: scopeB, model: model},
	)
	require.NoError(t, db.Exec(
		"INSERT INTO support_llm_budgets (scope, model, next_allowed_at) VALUES (?, ?, now() + interval '1 hour')",
		scopeB,
		model,
	).Error)
	var before time.Time
	require.NoError(t, db.Raw(
		"SELECT next_allowed_at FROM support_llm_budgets WHERE scope = ? AND model = ?",
		scopeB,
		model,
	).Scan(&before).Error)

	granted, err := NewLLMBudgetGranter(db).GrantAll(context.Background(), []BudgetGrant{
		{Scope: scopeB, Model: model, Interval: time.Hour},
		{Scope: scopeA, Model: model, Interval: time.Hour},
	})
	require.NoError(t, err)
	assert.False(t, granted)

	var count int64
	require.NoError(t, db.Raw(
		"SELECT COUNT(*) FROM support_llm_budgets WHERE scope = ? AND model = ?",
		scopeA,
		model,
	).Scan(&count).Error)
	assert.Zero(t, count)
	var after time.Time
	require.NoError(t, db.Raw(
		"SELECT next_allowed_at FROM support_llm_budgets WHERE scope = ? AND model = ?",
		scopeB,
		model,
	).Scan(&after).Error)
	assert.True(t, before.Equal(after))
}

func testReviewCanceledBudgetGrant(t *testing.T, db *gorm.DB, namespace string) {
	t.Helper()
	scope := namespace + "-scope"
	model := namespace + "-model"
	cleanupReviewBudgetRows(t, db, reviewBudgetKey{scope: scope, model: model})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewLLMBudgetGranter(db).GrantAll(ctx, []BudgetGrant{{
		Scope: scope, Model: model, Interval: time.Hour,
	}})
	require.ErrorIs(t, err, context.Canceled)

	var count int64
	require.NoError(t, db.Raw(
		"SELECT COUNT(*) FROM support_llm_budgets WHERE scope = ? AND model = ?",
		scope,
		model,
	).Scan(&count).Error)
	assert.Zero(t, count)
}

func testReviewEmbeddingCache(t *testing.T, db *gorm.DB, namespace string) {
	t.Helper()
	model := namespace + "-model"
	key := manualarticle.ChunkCacheKey(model, manualarticle.TaskTypeRetrievalDocument, namespace+"-cache")
	cancelKey := manualarticle.ChunkCacheKey(model, manualarticle.TaskTypeRetrievalDocument, namespace+"-cancel")
	missingKey := manualarticle.ChunkCacheKey(model, manualarticle.TaskTypeRetrievalDocument, namespace+"-missing")
	cleanupReviewEmbeddingRows(t, db, key, cancelKey)
	cache := NewManualEmbeddingCache(db, model)
	first := reviewVector(0)
	first[0], first[1] = 3, 4
	require.NoError(t, cache.StoreWithKey(
		context.Background(), key, model, manualarticle.TaskTypeRetrievalDocument,
		manualarticle.ChunkerVersion, manualarticle.EmbeddingDimensions, first,
	))
	require.NoError(t, cache.StoreWithKey(
		context.Background(), key, model, manualarticle.TaskTypeRetrievalDocument,
		manualarticle.ChunkerVersion, manualarticle.EmbeddingDimensions, reviewVector(1),
	))

	loaded, err := cache.Load(context.Background(), []string{key, missingKey})
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	assert.InDelta(t, 0.6, loaded[key][0], 1e-5)
	assert.InDelta(t, 0.8, loaded[key][1], 1e-5)
	assert.Len(t, loaded[key], manualarticle.EmbeddingDimensions)

	wrongModel, err := NewManualEmbeddingCache(db, namespace+"-wrong-model").Load(context.Background(), []string{key})
	require.NoError(t, err)
	assert.Empty(t, wrongModel)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = cache.Load(canceled, []string{key})
	require.ErrorIs(t, err, context.Canceled)
	err = cache.StoreWithKey(
		canceled, cancelKey, model, manualarticle.TaskTypeRetrievalDocument,
		manualarticle.ChunkerVersion, manualarticle.EmbeddingDimensions, reviewVector(0),
	)
	require.ErrorIs(t, err, context.Canceled)
	var count int64
	require.NoError(t, db.Raw(
		"SELECT COUNT(*) FROM support_manual_embeddings WHERE cache_key = ?",
		cancelKey,
	).Scan(&count).Error)
	assert.Zero(t, count)
}

func testReviewManualSearchConstraints(t *testing.T, db *gorm.DB, namespace string) {
	t.Helper()
	model := namespace + "-model"
	validVector := embeddingArray(make([]float32, manualarticle.EmbeddingDimensions))
	validVector[0] = 1
	invalidArray := reviewArrayLiteral(manualarticle.EmbeddingDimensions, 1, 0)
	non1DArray := reviewArrayLiteral(manualarticle.EmbeddingDimensions, 2, 1)
	shortVector := embeddingArray(make([]float32, manualarticle.EmbeddingDimensions-1))
	shortVector[0] = 1
	cases := []struct {
		name    string
		task    string
		version int
		dims    int
		vector  any
	}{
		{name: "task type", task: "RETRIEVAL_QUERY", version: manualarticle.ChunkerVersion, dims: manualarticle.EmbeddingDimensions, vector: validVector},
		{name: "chunker version", task: manualarticle.TaskTypeRetrievalDocument, version: manualarticle.ChunkerVersion + 1, dims: manualarticle.EmbeddingDimensions, vector: validVector},
		{name: "dimensions metadata", task: manualarticle.TaskTypeRetrievalDocument, version: manualarticle.ChunkerVersion, dims: manualarticle.EmbeddingDimensions - 1, vector: validVector},
		{name: "array length", task: manualarticle.TaskTypeRetrievalDocument, version: manualarticle.ChunkerVersion, dims: manualarticle.EmbeddingDimensions, vector: shortVector},
		{name: "multi-dimensional array", task: manualarticle.TaskTypeRetrievalDocument, version: manualarticle.ChunkerVersion, dims: manualarticle.EmbeddingDimensions, vector: non1DArray},
		{name: "non-one lower bound", task: manualarticle.TaskTypeRetrievalDocument, version: manualarticle.ChunkerVersion, dims: manualarticle.EmbeddingDimensions, vector: invalidArray},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := manualarticle.ChunkCacheKey(model, manualarticle.TaskTypeRetrievalDocument, namespace+"-"+tc.name)
			cleanupReviewEmbeddingRows(t, db, key)
			err := db.Exec(
				`INSERT INTO support_manual_embeddings
					(cache_key, model, task_type, chunker_version, dimensions, embedding)
					VALUES (?, ?, ?, ?, ?, ?::real[])`,
				key,
				model,
				tc.task,
				tc.version,
				tc.dims,
				tc.vector,
			).Error
			requireReviewPostgresCode(t, err, "23514")
		})
	}

	scope := namespace + "-duplicate"
	cleanupReviewBudgetRows(t, db, reviewBudgetKey{scope: scope, model: model})
	require.NoError(t, db.Exec(
		"INSERT INTO support_llm_budgets (scope, model, next_allowed_at) VALUES (?, ?, now())",
		scope,
		model,
	).Error)
	err := db.Exec(
		"INSERT INTO support_llm_budgets (scope, model, next_allowed_at) VALUES (?, ?, now())",
		scope,
		model,
	).Error
	requireReviewPostgresCode(t, err, "23505")
}
