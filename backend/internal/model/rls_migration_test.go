package model

import (
	"os"
	"strings"
	"testing"
)

func TestRLSMigrationCoversTenantTables(t *testing.T) {
	t.Parallel()

	sql := readMigrationFile(t, "../../migrations/010_enable_rls_tenant_policies.sql")

	requiredSnippets := []string{
		"CREATE SCHEMA IF NOT EXISTS app_private",
		"CREATE OR REPLACE FUNCTION app_private.current_clinic_ids()",
		"CREATE OR REPLACE FUNCTION app_private.has_clinic_access(row_clinic_id bigint)",
		"ALTER TABLE %s ENABLE ROW LEVEL SECURITY",
		"GRANT USAGE ON SCHEMA app_private TO PUBLIC",
		"REVOKE ALL ON FUNCTION app_private.apply_rls_policy(regclass, text, text, text) FROM PUBLIC",
		"a.attname = 'clinic_id'",
		"'tenant_clinic_id_isolation'",
		"'tenant_clinics_isolation'",
		"'tenant_accounts_isolation'",
		"app_private.has_clinic_access(clinic_id)",
		"app_private.has_clinic_access(id)",
		"s.account_id = accounts.id",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("RLS migration missing required snippet: %s", snippet)
		}
	}
}

func TestRLSMigrationCoversParentScopedChildTables(t *testing.T) {
	t.Parallel()

	sql := readMigrationFile(t, "../../migrations/010_enable_rls_tenant_policies.sql")

	childTables := []string{
		"exam_type_fields",
		"permission_group_rules",
		"staff_permission_groups",
		"appointment_trimming_options",
		"inquiries",
		"clinical_plans",
		"treatments",
		"medical_record_images",
		"billing_confirmations",
		"exam_results",
		"care_plan_items",
		"estimate_items",
		"staff_notes",
		"billing_items",
		"payments",
		"staff_reservation_exclusions",
		"shift_entry_breaks",
		"shift_template_breaks",
		"campaign_target_categories",
		"campaign_target_items",
	}
	for _, table := range childTables {
		if !strings.Contains(sql, "'"+table+"'") {
			t.Fatalf("RLS migration missing parent-scoped child table policy: %s", table)
		}
	}
}

func readMigrationFile(t *testing.T, path string) string {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration file: %v", err)
	}
	return string(b)
}
