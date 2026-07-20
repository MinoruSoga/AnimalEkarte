package repository

// checkup_migration_ddl_helpers_test.go — audit_real_ddl_test.go 用の migration DDL 抽出ヘルパー。
//
// checkupMigration010Path / readCheckupMigration010 / extractCreateTableDDL は元々
// checkup_field_cascade_test.go に定義され、同 package の audit_real_ddl_test.go が
// 共有利用していた。BE9-2D roll-up で checkup_field_cascade_test.go を internal/medicalrecord へ
// 移動したため、残留する audit_real_ddl_test.go の呼び出しを無変更で通すべく、その 3 シンボルの
// 文書化付きローカルコピーをここへ残す（medicalrecord 側にも同名の複製がある。unexported かつ
// 別 package のため重複問題は無い）。名前は audit_real_ddl_test.go の既存呼び出しを保つため据え置く。

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const checkupMigration010Path = "../../migrations/001_init.sql"

// readCheckupMigration010 は 001_init.sql（旧 migration 010 統合先）の SQL テキストを返す。
func readCheckupMigration010(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(checkupMigration010Path)
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
