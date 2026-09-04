package testdb

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// coreTruncateSQL is one statement so CASCADE lock acquisition cannot
// interleave with a sibling TRUNCATE on another pool connection.
const coreTruncateSQL = "TRUNCATE TABLE billing_refunds, payments, billings, medical_records, owners CASCADE"

const testDBTruncateLockKey = "testdb.shared-truncate"

var testDBTruncateMu sync.Mutex

func isDedicatedTestDatabaseName(name string) bool {
	return strings.HasSuffix(name, "_test")
}

func isDeadlockError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "40P01"
}

func isSafeTruncateTableName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}

func truncateCoreTestTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	execSharedTruncate(t, db, coreTruncateSQL)
}

// Truncate は SetupTestDB と同じ単一接続・advisory lock・deadlock retry で
// 指定テーブルを CASCADE TRUNCATE する。リーク tx 切断は retry 時のみ行う。
func Truncate(t *testing.T, db *gorm.DB, tables ...string) {
	t.Helper()
	if len(tables) == 0 {
		return
	}
	quoted := make([]string, 0, len(tables))
	for _, table := range tables {
		if !isSafeTruncateTableName(table) {
			t.Fatalf("invalid truncate table name %q", table)
		}
		quoted = append(quoted, quotePostgresIdentifier(table))
	}
	execSharedTruncate(t, db, "TRUNCATE TABLE "+strings.Join(quoted, ", ")+" CASCADE")
}

func execSharedTruncate(t *testing.T, db *gorm.DB, truncateSQL string) {
	t.Helper()
	testDBTruncateMu.Lock()
	defer testDBTruncateMu.Unlock()

	sqlDB, err := db.DB()
	require.NoError(t, err, "shared test db pool")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := sqlDB.Conn(ctx)
	require.NoError(t, err, "checkout truncate connection")
	defer func() { _ = conn.Close() }()

	var dbName string
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT current_database()").Scan(&dbName))
	if !isDedicatedTestDatabaseName(dbName) {
		t.Fatalf("refusing to truncate or terminate backends on non-test database %q", dbName)
	}

	_, err = conn.ExecContext(ctx, "SELECT pg_advisory_lock(hashtext($1))", testDBTruncateLockKey)
	require.NoError(t, err, "acquire testdb truncate advisory lock")
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock(hashtext($1))", testDBTruncateLockKey)
	}()

	_, err = conn.ExecContext(ctx, "SET lock_timeout = '500ms'")
	require.NoError(t, err, "set testdb truncate lock_timeout")
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "RESET lock_timeout")
	}()

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			if err := terminateIdleInTransactionBackends(ctx, conn); err != nil {
				t.Fatalf("failed to clear leftover testdb transactions: %v", err)
			}
			evictSharedPoolConns(sqlDB)
		}
		if _, err := conn.ExecContext(ctx, truncateSQL); err == nil {
			return
		} else {
			lastErr = err
			if !isDeadlockError(err) && !isLockTimeoutError(err) {
				t.Fatalf("failed to truncate test tables: %v", err)
			}
		}
	}
	t.Fatalf("failed to truncate test tables after deadlock retries: %v", lastErr)
}

func isLockTimeoutError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "55P03"
}

func evictSharedPoolConns(sqlDB *sql.DB) {
	sqlDB.SetMaxIdleConns(0)
	sqlDB.SetMaxIdleConns(2)
}

func terminateIdleInTransactionBackends(ctx context.Context, conn *sql.Conn) error {
	_, err := conn.ExecContext(ctx, `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = current_database()
		  AND pid <> pg_backend_pid()
		  AND state IN ('idle in transaction', 'idle in transaction (aborted)')
	`)
	return err
}
