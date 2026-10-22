package medicalrecord

// EMR-447 A-2/D2: examination type Update TX の前値読取ロック強度。
// 実 DB / trigger / 実行は NOT_RUN（Lead 指示）。DryRun で検査する。

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// --- repository: ambient-TX FindByID のロック強度（A-2/D2） ---

func TestExamTypeRepository_FindByIDLockStrength(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"),
		&gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
	require.NoError(t, err)
	conn, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	var queries []string
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("capture_find_lock",
		func(tx *gorm.DB) { queries = append(queries, tx.Statement.SQL.String()) }))
	repo := &examTypeRepository{db: db}
	txCtx := persistence.WithTxValue(t.Context(), db)
	// 参照用途（ambient tx 内の既定 read）: FOR SHARE。
	_, _ = repo.FindByID(txCtx, 2, 17)
	share := queries[0]
	// 更新 TX の前値読取: FOR UPDATE で先に排他ロックする。
	queries = nil
	_, _ = repo.FindByIDForUpdate(txCtx, 2, 17)
	upd := queries[0]
	// ambient tx なし: ロック句なし。
	queries = nil
	_, _ = repo.FindByID(t.Context(), 2, 17)
	require.Contains(t, share, "FOR SHARE")
	require.Contains(t, upd, "FOR UPDATE")
	for _, q := range queries {
		require.NotContains(t, q, "FOR ")
	}
}
