package medicalrecord

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExamTypeFieldDeleteIndexMigration_IsIndexOnly(t *testing.T) {
	// 003 は 2026-07-29 に 001_init.sql §9 へ原文アーカイブされた。
	// 独立ファイルは存在しないため、Source file マーカーで本文を取り出す
	// (先例: exam_reference_range_migration_test.go)。
	sql := strings.TrimSpace(readArchivedExamTypeFieldIndexMigration(t))
	assert.Equal(
		t,
		"CREATE INDEX idx_exam_results_exam_type_field_id ON exam_results (exam_type_field_id);",
		sql,
	)
	assert.Equal(t, 1, strings.Count(sql, ";"))
}

// readArchivedExamTypeFieldIndexMigration は 001_init.sql §9 に統合された
// 003_add_exam_results_exam_type_field_id_index.sql の SQL 本文を返す。
func readArchivedExamTypeFieldIndexMigration(t *testing.T) string {
	t.Helper()

	raw, err := os.ReadFile("../../migrations/001_init.sql")
	require.NoError(t, err)
	initial := string(raw)

	const sourceMarker = "-- Source file: 003_add_exam_results_exam_type_field_id_index.sql"
	const nextSourceMarker = "-- Source file: 004_add_inventory_quantity_check.sql"
	start := strings.Index(initial, sourceMarker)
	require.GreaterOrEqual(t, start, 0, "001_init.sql must contain the archived exam type field index migration")

	endOffset := strings.Index(initial[start:], "\n"+nextSourceMarker)
	require.Greater(t, endOffset, 0, "archived exam type field index migration must end at the 004 source marker")
	block := initial[start : start+endOffset]

	shaOffset := strings.Index(block, "-- Source SHA-256:")
	require.GreaterOrEqual(t, shaOffset, 0, "archived exam type field index migration must contain its SHA-256 header")
	bodyOffset := strings.Index(block[shaOffset:], "\n")
	require.GreaterOrEqual(t, bodyOffset, 0, "archived migration metadata must end before its SQL body")

	return block[shaOffset+bodyOffset+1:]
}
