package lintscan

// staff_capability_backfill_lint_test.go — EMR-285
//
// ─── Purpose ────────────────────────────────────────────────────────────────────────
//
// 021_backfill_staff_reservation_capabilities.sql は TASK-021 Stage B 切替時に欠落した
// データ backfill を補う migration であり、runtime の fail-closed 契約そのものは変更しない。
// その代わりに「誰へ何を grant するか」の SELECT 条件が唯一の安全境界になるため、
// 本ゲートは backfill の不変条件を字義 pin する:
//
//   1. 未設定スタッフのみ対象 — 同一 (clinic_id, staff_id) の capability 行が既に
//      存在するスタッフ（明示設定済み）を決して上書きしない。
//   2. legacy 除外の尊重 — 旧 staff_reservation_exclusions（staff 単位）の指定を
//      grant 対象から除く（capable = universe \ excluded の facade と同型）。
//   3. clinic スコープ — staff_clinic_assignments の未削除行で所属医院へ限定し、
//      universe はその医院の is_active かつ未削除 reservation_types のみ。
//   4. 冪等 — ON CONFLICT DO NOTHING で再適用しても重複行を作らない。
//   5. INSERT のみ — UPDATE / DELETE / DDL を持ち込まない（読み取り専用 SELECT 由来の挿入）。
//
// これらは backfill が「旧既定（未設定=全区分対応可）を未設定スタッフへ復元する」
// 最小変更であることの構造的証跡であり、後続の編集で対象範囲が広がる回帰を機械検出する。

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const capabilityBackfillMigration = "021_backfill_staff_reservation_capabilities.sql"

func readCapabilityBackfillSQL(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(migrationsDir, capabilityBackfillMigration))
	if err != nil {
		t.Fatalf("read migration %s: %v", capabilityBackfillMigration, err)
	}
	// コメント除去は seed_csv_schema_drift_test.go の共有実装を使い、
	// 構造検査がコメント本文へ誤マッチしないようにする。
	stripped, err := stripSQLComments(string(raw))
	if err != nil {
		t.Fatalf("strip SQL comments in %s: %v", capabilityBackfillMigration, err)
	}
	return stripped
}

func TestCapabilityBackfillMigration_Contract(t *testing.T) {
	sql := readCapabilityBackfillSQL(t)
	upper := strings.ToUpper(sql)

	required := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		{
			name:    "INSERT INTO staff_reservation_capabilities のみ",
			pattern: regexp.MustCompile(`(?is)\bINSERT\s+INTO\s+staff_reservation_capabilities\b`),
		},
		{
			name: "所属医院スコープ（未削除 assignment のみ）",
			pattern: regexp.MustCompile(
				`(?is)FROM\s+staff_clinic_assignments\s+sca[\s\S]*?sca\.deleted_at\s+IS\s+NULL`),
		},
		{
			name: "スタッフ未削除（staffs.deleted_at IS NULL join）",
			pattern: regexp.MustCompile(
				`(?is)JOIN\s+staffs\s+s\b[\s\S]*?s\.deleted_at\s+IS\s+NULL`),
		},
		{
			name: "universe は医院の is_active + 未削除 reservation_types",
			pattern: regexp.MustCompile(
				`(?is)JOIN\s+reservation_types\s+rt\b[\s\S]*?rt\.is_active\s*=\s*TRUE[\s\S]*?rt\.deleted_at\s+IS\s+NULL`),
		},
		{
			name:    "clinic 所属一致（rt.clinic_id = sca.clinic_id）",
			pattern: regexp.MustCompile(`(?is)rt\.clinic_id\s*=\s*sca\.clinic_id`),
		},
		{
			name: "未設定スタッフのみ（capability NOT EXISTS ガード・同一 clinic+staff）",
			pattern: regexp.MustCompile(
				`(?is)NOT\s+EXISTS\s*\(\s*SELECT\s+1\s+FROM\s+staff_reservation_capabilities\s+cap\b[\s\S]*?cap\.clinic_id\s*=\s*sca\.clinic_id[\s\S]*?cap\.staff_id\s*=\s*sca\.staff_id`),
		},
		{
			name: "legacy 除外の尊重（exclusions NOT EXISTS・staff+type 一致）",
			pattern: regexp.MustCompile(
				`(?is)NOT\s+EXISTS\s*\(\s*SELECT\s+1\s+FROM\s+staff_reservation_exclusions\s+ex\b[\s\S]*?ex\.staff_id\s*=\s*sca\.staff_id[\s\S]*?ex\.reservation_type_id\s*=\s*rt\.id`),
		},
		{
			name:    "冪等（uq 制約への ON CONFLICT DO NOTHING）",
			pattern: regexp.MustCompile(`(?is)ON\s+CONFLICT\s+ON\s+CONSTRAINT\s+uq_staff_reservation_capability\s+DO\s+NOTHING`),
		},
	}
	for _, req := range required {
		if !req.pattern.MatchString(sql) {
			t.Errorf("backfill contract violation: %s — required clause not found in %s",
				req.name, capabilityBackfillMigration)
		}
	}

	// 変更系は INSERT 1 文のみ — UPDATE/DELETE/DDL の混入は fail。
	forbidden := []string{
		`\bUPDATE\b`, `\bDELETE\b`, `\bTRUNCATE\b`, `\bALTER\b`, `\bDROP\b`,
		`\bCREATE\b`, `\bMERGE\b`,
	}
	for _, kw := range forbidden {
		re := regexp.MustCompile(kw)
		if loc := re.FindStringIndex(upper); loc != nil {
			t.Errorf("backfill migration must be INSERT-only; found %s at byte %d",
				strings.Trim(kw, `\b`), loc[0])
		}
	}
	if strings.Count(upper, "INSERT INTO STAFF_RESERVATION_CAPABILITIES") != 1 {
		t.Errorf("expected exactly one INSERT INTO staff_reservation_capabilities statement")
	}
}
