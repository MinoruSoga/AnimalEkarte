package model

import (
	"os"
	"strings"
	"testing"
)

func TestRLSMigrationCoversTenantTables(t *testing.T) {
	t.Parallel()

	sql := readMigrationFile(t, "../../migrations/001_init.sql")

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

	sql := readMigrationFile(t, "../../migrations/001_init.sql")

	childTables := []string{
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

func TestLabDevicesMigrationHasDirectClinicRLS(t *testing.T) {
	t.Parallel()

	sql := readMigrationFile(t, "../../migrations/001_init.sql")
	const policy = "SELECT app_private.apply_rls_policy(\n    'lab_devices'::regclass,\n    'tenant_clinic_id_isolation',\n    'app_private.has_clinic_access(clinic_id)',\n    'app_private.has_clinic_access(clinic_id)'\n);"
	if !strings.Contains(sql, policy) {
		t.Fatalf("lab_devices RLS policy must directly scope USING and WITH CHECK by clinic_id:\n%s", policy)
	}
}

func TestEstimatesMigrationBindsPetToClinic(t *testing.T) {
	t.Parallel()

	sql := readMigrationFile(t, "../../migrations/001_init.sql")
	requiredSnippets := []string{
		"ALTER TABLE pets\n    ADD CONSTRAINT uq_pets_id_clinic UNIQUE (id, clinic_id);",
		"ALTER TABLE estimates\n  ADD COLUMN IF NOT EXISTS pet_id bigint;",
		"ALTER TABLE estimates\n  ADD CONSTRAINT fk_estimates_pet_clinic\n  FOREIGN KEY (clinic_id, pet_id)\n  REFERENCES pets (clinic_id, id)\n  ON DELETE SET NULL;",
	}
	for _, required := range requiredSnippets {
		if !strings.Contains(sql, required) {
			t.Fatalf("estimates migration must bind pet_id to the estimate clinic:\n%s", required)
		}
	}
}

// TestExamResultsTenantBoundaryScopedViaParentExam — BE-refactor.md R3-7 (D13) exam_results 部分。
//
// exam_results は clinic_id カラムを持たないため、checkup_field_results 同型の
// (exam_type_field_id, clinic_id) 複合 FK は構造的に張れない（migration 012 は checkup 側のみ）。
// exam_type_fields は旧 migration 005 で clinic_id と複合 FK を獲得済み。
//
// exam_results の**実効的な**テナント境界はアプリ層が担う:
//   - examination_repository の全クエリが親 exams への JOIN + WHERE clinic_id 述語で scope する。
//   - examination_service.ReplaceItems の #124 検証（exam_type_field が caller clinic の exam_type に
//     属するか。TestExaminationService_ReplaceItems が RED/GREEN で固定）。
//
// ⚠️ DB レベル RLS は**定義済みだが現行構成では未 enforce**（001_init.sql:2896-2905 に明記の baseline）:
// ポリシーは `ENABLE ROW LEVEL SECURITY` のみで `FORCE` しておらず、アプリはテーブル owner ロールで
// 接続するため owner に RLS は適用されない（さらに Go 側は app.current_clinic_ids を SET していない）。
// よって RLS は runtime の越境遮断を**していない**。RLS を実効化するには非 owner ロール +
// FORCE ROW LEVEL SECURITY + 全 repository の SET LOCAL app.current_clinic_ids 配線が必要で、R3-7 の
// スコープ外の別タスク（プロジェクト全 clinic_id テーブル共通の既存ギャップ）。
//
// 本テストが固定するのは runtime の遮断ではなく migration の**構造前提**:
//  1. exam_results の RLS ポリシー定義が親 exams.clinic_id 経由で正しく書かれていること
//     （RLS を将来 FORCE 化した際の defense-in-depth の正しさ + ポリシー誤削除の検出）。
//  2. exam_results に clinic_id カラムが無いこと（追加された場合は複合 FK が可能になり
//     R3-7 の「app 層が実効境界」という設計前提が変わる → 見直しの合図）。
func TestExamResultsTenantBoundaryScopedViaParentExam(t *testing.T) {
	t.Parallel()
	sql := readMigrationFile(t, "../../migrations/001_init.sql")

	// (1) exam_results の RLS ポリシー定義が親 exams.clinic_id 経由でスコープされること
	// （runtime enforce の証明ではない。上記コメント参照）。
	const examParentScope = "EXISTS (SELECT 1 FROM exams e WHERE e.id = exam_results.exam_id AND app_private.has_clinic_access(e.clinic_id))"
	if !strings.Contains(sql, examParentScope) {
		t.Fatalf("exam_results RLS policy definition must scope via parent exams.clinic_id; missing expression:\n  %s\n"+
			"(RLS を将来 FORCE 化する際の defense-in-depth の正しさを固定。実効境界は app 層)", examParentScope)
	}
	if !strings.Contains(sql, "'tenant_exam_results_isolation'") {
		t.Fatal("exam_results RLS policy name 'tenant_exam_results_isolation' missing")
	}

	// (2) 構造前提の固定: exam_results の CREATE TABLE に clinic_id カラムが無いこと
	// （あれば checkup 同型の複合 FK が可能になり、本テストの設計前提が変わる → 見直しの合図）。
	ddl := extractCreateTable(t, sql, "exam_results")
	for line := range strings.SplitSeq(ddl, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "clinic_id ") {
			t.Fatalf("exam_results に clinic_id カラムが追加されている。複合 FK が可能になったので R3-7 の設計を見直すこと: %s", trimmed)
		}
	}
}

func TestRLSMigrationExamTypeFieldsUsesFinalDirectClinicBoundary(t *testing.T) {
	t.Parallel()
	sql := readMigrationFile(t, "../../migrations/001_init.sql")

	finalSchemaMutations := []struct {
		required  string
		forbidden []string
	}{
		{
			required: "ALTER TABLE exam_type_fields\n    ADD COLUMN clinic_id bigint;",
			forbidden: []string{
				"ALTER TABLE exam_type_fields\n    DROP COLUMN clinic_id",
				"ALTER TABLE exam_type_fields\n    DROP COLUMN IF EXISTS clinic_id",
			},
		},
		{
			required: "ALTER TABLE exam_type_fields\n    ALTER COLUMN clinic_id SET NOT NULL,",
			forbidden: []string{
				"ALTER TABLE exam_type_fields\n    ALTER COLUMN clinic_id DROP NOT NULL",
			},
		},
		{
			required: "ADD CONSTRAINT uq_exam_type_fields_id_clinic UNIQUE (id, clinic_id);",
			forbidden: []string{
				"ALTER TABLE exam_type_fields\n    DROP CONSTRAINT uq_exam_type_fields_id_clinic",
				"ALTER TABLE exam_type_fields\n    DROP CONSTRAINT IF EXISTS uq_exam_type_fields_id_clinic",
			},
		},
		{
			required: "ADD CONSTRAINT fk_exam_type_fields_type_clinic\n    FOREIGN KEY (exam_type_id, clinic_id)\n    REFERENCES exam_types (id, clinic_id)",
			forbidden: []string{
				"ALTER TABLE exam_type_fields\n    DROP CONSTRAINT fk_exam_type_fields_type_clinic",
				"ALTER TABLE exam_type_fields\n    DROP CONSTRAINT IF EXISTS fk_exam_type_fields_type_clinic",
			},
		},
	}
	for _, mutation := range finalSchemaMutations {
		requiredAt := strings.LastIndex(sql, mutation.required)
		if requiredAt < 0 {
			t.Fatalf("exam_type_fields final schema missing required clinic boundary:\n%s", mutation.required)
		}
		for _, forbidden := range mutation.forbidden {
			if forbiddenAt := strings.LastIndex(sql, forbidden); forbiddenAt > requiredAt {
				t.Fatalf("exam_type_fields final schema removes a required clinic boundary after it is established:\n%s", forbidden)
			}
		}
	}

	policy := extractFinalRLSPolicyApplication(t, sql, "exam_type_fields")
	const directClinicScope = "'app_private.has_clinic_access(clinic_id)'"
	if strings.Count(policy, directClinicScope) != 2 {
		t.Fatalf("final exam_type_fields RLS policy must directly scope USING and WITH CHECK by clinic_id:\n%s", policy)
	}
	if strings.Contains(policy, "EXISTS (") {
		t.Fatalf("final exam_type_fields RLS policy must not retain the historical parent-scoped expression:\n%s", policy)
	}
}

func extractFinalRLSPolicyApplication(t *testing.T, sql, table string) string {
	t.Helper()
	marker := "SELECT app_private.apply_rls_policy(\n    '" + table + "',"
	start := strings.LastIndex(sql, marker)
	if start < 0 {
		t.Fatalf("RLS policy application for %s not found", table)
	}
	policy, _, found := strings.Cut(sql[start:], "\n);")
	if !found {
		t.Fatalf("RLS policy application for %s has no terminator", table)
	}
	return policy
}

// extractCreateTable は 001_init.sql から `CREATE TABLE <table> ( ... );` ブロックを取り出す。
func extractCreateTable(t *testing.T, sql, table string) string {
	t.Helper()
	marker := "CREATE TABLE " + table + " ("
	start := strings.Index(sql, marker)
	if start < 0 {
		t.Fatalf("CREATE TABLE %s not found", table)
	}
	rest := sql[start:]
	body, _, found := strings.Cut(rest, "\n);")
	if !found {
		t.Fatalf("CREATE TABLE %s terminator not found", table)
	}
	return body
}

func readMigrationFile(t *testing.T, path string) string {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration file: %v", err)
	}
	return string(b)
}
