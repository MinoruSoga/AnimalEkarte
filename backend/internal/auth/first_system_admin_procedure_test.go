package auth

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFirstSystemAdminProcedureMatchesInitSchema(t *testing.T) {
	schema, err := os.ReadFile("../../migrations/001_init.sql")
	require.NoError(t, err)
	sqlBytes, err := os.ReadFile("testdata/first_system_admin.sql")
	require.NoError(t, err)
	sql := strings.TrimSpace(string(sqlBytes))
	required := []string{
		"LOCK TABLE public.accounts, public.clinics, public.staffs,",
		"public.staff_clinic_assignments IN SHARE ROW EXCLUSIVE MODE",
		"INSERT INTO public.accounts (email, password_hash, is_active, is_system_admin)",
		"UPDATE public.staffs SET account_id = new_account_id, updated_at = now()",
		"INSERT INTO public.audit_logs (",
		"'system'",
		"account.bootstrap.create",
		"IF EXISTS (SELECT 1 FROM public.accounts WHERE is_system_admin)",
		"s.account_id IS NULL",
		"a.is_main AND a.deleted_at IS NULL",
	}
	for _, fragment := range required {
		require.Contains(t, sql, fragment)
	}
	require.NotContains(t, sql, "WHERE is_system_admin AND deleted_at IS NULL")

	procedure, err := os.ReadFile("../../../docs/ops/deploy/FIRST_SYSTEM_ADMIN.md")
	require.NoError(t, err, "the procedure document must be mounted for this contract")
	require.Equal(t, sql, extractMarkdownSQL(t, string(procedure)))
	require.Contains(t, string(procedure), "通信切断や COMMIT 応答不明の場合は**再実行せず**")
	require.Contains(t, string(procedure), "末尾の receipt SELECT")

	for table, columns := range map[string][]string{
		"accounts":                 {"id", "email", "password_hash", "is_active", "is_system_admin", "deleted_at"},
		"staffs":                   {"id", "clinic_id", "account_id", "is_active", "deleted_at", "updated_at"},
		"clinics":                  {"id", "is_active"},
		"staff_clinic_assignments": {"staff_id", "clinic_id", "is_main", "deleted_at"},
		"audit_logs":               {"clinic_id", "actor_id", "actor_type", "action", "resource", "resource_id", "old_value", "new_value", "metadata"},
	} {
		create := createTableSQL(t, string(schema), table)
		for _, column := range columns {
			require.Regexp(t, regexp.MustCompile(`(?m)^\s+`+column+`\s`), create, "table %s missing %s", table, column)
		}
	}
	require.Contains(t, createTableSQL(t, string(schema), "audit_logs"), "actor_type = 'system' AND actor_id IS NULL")
}

func extractMarkdownSQL(t *testing.T, markdown string) string {
	t.Helper()
	start := strings.Index(markdown, "```sql")
	require.Greater(t, start, -1)
	rest := markdown[start+len("```sql"):]
	end := strings.Index(rest, "```")
	require.Greater(t, end, -1)
	return strings.TrimSpace(rest[:end])
}

func createTableSQL(t *testing.T, schema, table string) string {
	t.Helper()
	marker := "CREATE TABLE " + table + " ("
	start := strings.Index(schema, marker)
	require.Greater(t, start, -1, table)
	rest := schema[start:]
	end := strings.Index(rest, "\n);")
	require.Greater(t, end, -1, table)
	return rest[:end]
}
