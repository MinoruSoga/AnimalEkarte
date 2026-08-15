package audit

// repository_test_helpers_test.go — audit の real-DDL test 用 migration DDL 抽出ヘルパー。
//
// initMigrationPath / readInitMigrationSQL / extractCreateTableDDL は、repository_real_ddl_test.go が
// 001_init.sql の実 DDL から audit_logs を再作成するために使う package-local helper である。

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testdb.CloseSharedTestDB()
	os.Exit(code)
}

const initMigrationPath = "../../migrations/001_init.sql"

// readInitMigrationSQL は 001_init.sql（旧 migration 010 統合先）の SQL テキストを返す。
func readInitMigrationSQL(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(initMigrationPath)
	require.NoError(t, err, "migration 010 を読めること")
	return string(raw)
}

// extractCreateTableDDL は migration SQL から `CREATE TABLE <table> ( ... );` ブロックを取り出す
// （終端の ");" を含むため、そのまま Exec できる）。
func extractCreateTableDDL(t *testing.T, sql, table string) string {
	t.Helper()
	marker := "CREATE TABLE " + table + " ("
	start := strings.Index(sql, marker)
	require.GreaterOrEqual(t, start, 0, "CREATE TABLE %s がマイグレーションに見つからない", table)
	rest := sql[start:]
	// CREATE TABLE 本体は最初の行頭 ");" で閉じる（10桁整形の本マイグレーションで安定）。
	end := strings.Index(rest, "\n);")
	require.GreaterOrEqual(t, end, 0, "CREATE TABLE %s の終端 ');' が見つからない", table)
	return rest[:end+len("\n);")]
}
