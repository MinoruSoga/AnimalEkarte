package reservation

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func nextVisitRetryCandidatesDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open("host=disabled.invalid user=test dbname=test sslmode=disable"), &gorm.Config{
		DisableAutomaticPing:   true,
		DryRun:                 true,
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)
	return db
}

// This empty relation fixture supports query-only database tests whose setup
// uses AutoMigrate. It does not apply migration 033 or validate its RLS policy.
// Call it only after setupReservationRepoTestDB has passed the -short gate.
func setupEmptyNextVisitRetryQueueFixture(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS lstep_next_visit_sync_retries (
		clinic_id BIGINT NOT NULL,
		owner_id BIGINT NOT NULL,
		revision BIGSERIAL NOT NULL,
		cleanup_line_user_id VARCHAR(64),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (clinic_id, owner_id))`).Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE lstep_next_visit_sync_retries").Error)
}

func TestNextVisitTagRepairCandidates_QueryContract(t *testing.T) {
	repo := &reservationRepository{db: nextVisitRetryCandidatesDryRunDB(t)}
	for _, tc := range []struct {
		name      string
		asOf      time.Time
		date      string
		limit     int
		wantLimit int
	}{
		{"before JST midnight", time.Date(2026, 10, 11, 14, 59, 59, 0, time.UTC), "2026-10-11", 1, 1},
		{"at JST midnight", time.Date(2026, 10, 11, 15, 0, 0, 0, time.UTC), "2026-10-12", 200, 200},
		{"zero limit", time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC), "2026-10-12", 0, 200},
		{"negative limit", time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC), "2026-10-12", -1, 200},
		{"oversized limit", time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC), "2026-10-12", 201, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var ids []uint64
			query := repo.nextVisitTagRepairOwnersQuery(context.Background(), 7, tc.asOf, 42, tc.limit).Find(&ids)
			require.NoError(t, query.Error)
			sql := query.Statement.SQL.String()

			assert.Contains(t, sql, "SELECT DISTINCT expiry_tag.owner_id AS owner_id")
			assert.Contains(t, sql, "SELECT retry.owner_id AS owner_id")
			assert.Contains(t, sql, "lstep_next_visit_sync_retries AS retry")
			assert.Contains(t, sql, "expiry_tag.clinic_id =")
			assert.Contains(t, sql, "expiry_tag.owner_id >")
			assert.Contains(t, sql, "owner.clinic_id = expiry_tag.clinic_id AND owner.deleted_at IS NULL")
			assert.Contains(t, sql, "retry.clinic_id =")
			assert.Contains(t, sql, "retry.owner_id >")
			assert.Contains(t, sql, "owner.clinic_id = retry.clinic_id AND (owner.deleted_at IS NULL OR retry.cleanup_line_user_id IS NOT NULL)")
			assert.Regexp(t, `\) UNION \(`, sql, "overlapping cache and retry owners must be deduplicated")
			assert.NotContains(t, sql, "UNION ALL")
			assert.Equal(t, 3, strings.Count(sql, "ORDER BY owner_id ASC"), "both source pages and the merged page use the same keyset order")
			assert.Equal(t, 3, strings.Count(sql, "LIMIT "), "each source and the merged page must be bounded")
			assert.NotContains(t, sql, "OFFSET")
			assert.NotContains(t, sql, "appointments")
			assert.Equal(t, []any{
				uint64(7), uint64(42), tc.date, tc.wantLimit,
				uint64(7), uint64(42), tc.wantLimit, tc.wantLimit,
			}, query.Statement.Vars, "clinic, cursor and the normalized limit must bind to both independent sources")
		})
	}
}

func TestNextVisitTagRepairCandidates_RetryDoesNotRequireCacheOrExpiredDate(t *testing.T) {
	repo := &reservationRepository{db: nextVisitRetryCandidatesDryRunDB(t)}
	var ids []uint64
	query := repo.nextVisitTagRepairOwnersQuery(context.Background(), 7,
		time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC), 0, 2).Find(&ids)
	require.NoError(t, query.Error)
	sql := query.Statement.SQL.String()
	start := strings.Index(sql, "SELECT retry.owner_id AS owner_id")
	require.NotEqual(t, -1, start, "a separate pending-work source must discover owners when no cache row exists")
	pending := sql[start:]
	assert.Contains(t, pending, "JOIN owners owner ON owner.id = retry.owner_id")
	assert.NotContains(t, pending, "lstep_tag_cache")
	assert.NotContains(t, pending, "tag_name")
	assert.NotContains(t, pending, "substring")
	assert.NotContains(t, pending, "appointments")
	assert.NotContains(t, pending, "pets")
}

func TestNextVisitTagRepairCandidates_UsesAmbientQuerySession(t *testing.T) {
	db := nextVisitRetryCandidatesDryRunDB(t)
	repo := &reservationRepository{db: db}
	ambient := db.Session(&gorm.Session{NewDB: true})
	ambient.Statement.Settings.Store("repair-candidates-ambient-proof", true)
	ctx := persistence.WithTxValue(context.Background(), ambient)
	query := repo.nextVisitTagRepairOwnersQuery(ctx, 7, time.Now(), 0, 1)
	value, ok := query.Statement.Settings.Load("repair-candidates-ambient-proof")
	assert.True(t, ok)
	assert.Equal(t, true, value)
	assert.Same(t, ctx, query.Statement.Context)
}

// This test documents query semantics only, not migration 033/RLS enforcement.
// It never opens a database during the authorized offline (-short) verification.
func TestNextVisitTagRepairCandidates_DBQueueOnlyDedupAndKeyset(t *testing.T) {
	if testing.Short() {
		t.Skip("database semantics require an explicitly authorized isolated PostgreSQL test environment")
	}
	db := setupReservationRepoTestDB(t)
	queueOnly := testdb.MakeTestOwner(t, db, 1, "repair queue only")
	shared := testdb.MakeTestOwner(t, db, 1, "repair shared cache and queue")
	cacheOnly := testdb.MakeTestOwner(t, db, 1, "repair cache only")
	deleted := testdb.MakeTestOwner(t, db, 1, "repair deleted")
	deletedCleanup := testdb.MakeTestOwner(t, db, 1, "repair deleted cleanup")
	foreign := testdb.MakeTestOwner(t, db, 2, "repair foreign")
	require.NoError(t, db.Delete(deleted).Error)
	require.NoError(t, db.Delete(deletedCleanup).Error)
	repo := NewReservationRepository(db)
	asOf := time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC)
	err := db.Transaction(func(tx *gorm.DB) error {
		// Temporary tables keep this query-only test independent of the retry
		// migration and isolate it from other tests' cache/retry fixtures.
		if err := tx.Exec(`CREATE TEMP TABLE lstep_next_visit_sync_retries (
			clinic_id BIGINT NOT NULL, owner_id BIGINT NOT NULL,
			cleanup_line_user_id VARCHAR(64),
			PRIMARY KEY (clinic_id, owner_id)) ON COMMIT DROP`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE TEMP TABLE lstep_tag_cache (
			clinic_id BIGINT NOT NULL, owner_id BIGINT NOT NULL,
			tag_name TEXT NOT NULL) ON COMMIT DROP`).Error; err != nil {
			return err
		}
		for _, row := range []struct {
			clinicID, ownerID uint64
			cleanup           *string
		}{
			{1, queueOnly.ID, nil}, {1, shared.ID, nil}, {1, deleted.ID, nil},
			{2, foreign.ID, nil}, {1, foreign.ID, nil},
			{1, deletedCleanup.ID, ptrString("Udeleted-cleanup")},
		} {
			if err := tx.Exec("INSERT INTO lstep_next_visit_sync_retries (clinic_id, owner_id, cleanup_line_user_id) VALUES (?, ?, ?)", row.clinicID, row.ownerID, row.cleanup).Error; err != nil {
				return err
			}
		}
		for _, ownerID := range []uint64{shared.ID, cacheOnly.ID} {
			if err := tx.Exec("INSERT INTO lstep_tag_cache (clinic_id, owner_id, tag_name) VALUES (?, ?, ?)", 1, ownerID, "next_visit_2024-02-29").Error; err != nil {
				return err
			}
		}
		var queueCacheCount int64
		if err := tx.Table("lstep_tag_cache").Where("owner_id = ?", queueOnly.ID).Count(&queueCacheCount).Error; err != nil {
			return err
		}
		require.Zero(t, queueCacheCount, "the failed initial sync owner has no local cache evidence")

		ctx := persistence.WithTxValue(context.Background(), tx)
		var got []uint64
		var cursor uint64
		for page := 0; page < 5; page++ {
			ids, err := repo.FindOwnerIDsWithExpiredTentativeTagCursor(ctx, 1, asOf, cursor, 1)
			if err != nil {
				return err
			}
			if len(ids) == 0 {
				break
			}
			require.Len(t, ids, 1)
			require.Greater(t, ids[0], cursor, "a merged page must advance the shared owner cursor")
			got = append(got, ids[0])
			cursor = ids[0]
		}
		want := []uint64{queueOnly.ID, shared.ID, cacheOnly.ID, deletedCleanup.ID}
		sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
		assert.Equal(t, want, got, "discover queue-only owners, deduplicate shared owners, include deleted cleanup intents, and exclude plain deleted/foreign owners")
		return nil
	})
	require.NoError(t, err)
}
