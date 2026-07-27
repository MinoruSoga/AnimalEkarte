package medicalrecord

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExamTypeFieldDeleteIndexMigration_IsIndexOnly(t *testing.T) {
	const path = "../../migrations/003_add_exam_results_exam_type_field_id_index.sql"
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	sql := strings.TrimSpace(string(content))
	assert.Equal(
		t,
		"CREATE INDEX idx_exam_results_exam_type_field_id ON exam_results (exam_type_field_id);",
		sql,
	)
	assert.Equal(t, 1, strings.Count(sql, ";"))
}
