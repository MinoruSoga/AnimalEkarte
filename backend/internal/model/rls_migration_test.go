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

// TestExamResultsTenantBoundaryScopedViaParentExam — BE-refactor.md R3-7 (D13) exam_results 部分。
//
// exam_results / exam_type_fields は clinic_id カラムを持たないため、checkup_field_results 同型の
// (id, clinic_id) 複合 FK は構造的に張れない（migration 012 は checkup 側のみ）。
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
//  2. exam_results / exam_type_fields に clinic_id カラムが無いこと（あれば複合 FK が可能になり
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

	// (2) 構造前提の固定: exam_results / exam_type_fields の CREATE TABLE に clinic_id カラムが無いこと
	// （あれば checkup 同型の複合 FK が可能になり、本テストの設計前提が変わる → 見直しの合図）。
	for _, table := range []string{"exam_results", "exam_type_fields"} {
		ddl := extractCreateTable(t, sql, table)
		for line := range strings.SplitSeq(ddl, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "clinic_id ") {
				t.Fatalf("%s に clinic_id カラムが追加されている。複合 FK が可能になったので R3-7 の設計を見直すこと: %s", table, trimmed)
			}
		}
	}
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
