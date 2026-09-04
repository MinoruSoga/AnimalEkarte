package lintscan

// migration_cascade_lint_test.go — BE-refactor.md R3-1 (D9・P0)
//
// ─── Purpose ────────────────────────────────────────────────────────────────────────
//
// migrations/CLAUDE.md は「CASCADE DELETE 禁止」を原則とするが、001_init.sql（初期スキーマ。
// 2026-07-04 と 2026-07-27 に旧増分マイグレーションを統合済み）には既に 50+ 件の
// "ON DELETE CASCADE" が存在する — 大半は junction テーブルや line item 等「純粋従属の
// 子テーブル」の正当な用法で、#211 の CASCADE 安全監査（checkup_field_cascade_test.go）でも
// 個別に安全性が確認されている（旧 010 由来の 2 件を含む）。
//
// これら既存サイトを1行ずつ allowlist に列挙するのは churn が大きく安全価値が薄い。本ゲートは
// 「新規 migration が未レビューの CASCADE を持ち込むことを検出する」（R3-1 の実際の目的）に
// 絞り、ファイル単位で CASCADE 出現数をピンする。migrations は追記のみ（migrations/CLAUDE.md:
// 適用済みファイルの編集は checksum mismatch）ため、確定済みファイルの出現数は正当な理由なく
// 変化しない。出現数の増加（既存ファイル編集）または新規ファイルでの非ゼロ出現は、
// 未レビューの CASCADE 混入を意味し fail する。
//
// #211 の懸念（臨床結果テーブルへの CASCADE 混入でマスタ削除時に患者結果値が silent 消去される）
// は、このファイル単位ゲートでは「新規 migration が exam_results/checkup_field_results 等に
// CASCADE を持ち込む」ケースを検出できる（新規ファイルの非ゼロ出現として即座に fail）。
// 既存の安全設計（SET NULL）そのものの回帰は checkup_field_cascade_test.go が個別に守る。
//
// ─── Technique ──────────────────────────────────────────────────────────────────────
//
// migrations/*.sql は backend/internal/lintscan/ から2階層上の backend/migrations/ にあり、
// ".." を使えない go:embed の相対パス制約に抵触する。internal/lintscan は internal/<package> と
// 同じ深さなので、既存の checkup_field_cascade_test.go / rls_migration_test.go と同じ
// os.ReadFile("../../migrations/...") 相対パス方式を維持し、directory 読み取りが壊れた場合は
// minDDLFiles の floor が vacuous pass を防ぐ。

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const migrationsDir = "../../migrations"

// migrationCascadeAllowlist はレビュー済みの CASCADE 出現数をファイル単位でピンする。
// 出現数が変化した場合、または新規ファイルで非ゼロが検出された場合はゲートが fail する。
//
//	001_init.sql: 初期スキーマの junction/line-item 等の純粋従属子50件（#211 監査時点でレビュー済み）
//	  + 旧 010_add_checkup_packages.sql 由来の checkup_type_fields.checkup_type_id と
//	  checkup_field_results.checkup_id の2件（純粋従属子・checkup_field_cascade_test.go で安全性実証済み。
//	  旧 010 は 2026-07-04 に 001_init.sql へ統合）
//	  + 旧 002_checkup_field_clinic_composite_fk.sql（#211 A6・2026-07-17 に 001 末尾へ統合）由来の1件
//	  （checkup_type_fields.checkup_type_id → checkup_types の単一列 CASCADE FK を複合FK
//	  （clinic_id 込み）へ置換するもの。挙動は既存 CASCADE から変更しない behavior-preserving）。
//	  + 2026-07-27 に統合した exam_type_fields の clinic 複合FK置換由来の1件
//	  （削除伝播の意味は変えず、テナント整合性を強化。新設 exam_reference_ranges は
//	  RESTRICT のみ）。
//	  + 2026-09-02: trimming details/options の (appointment_id, clinic_id) 複合 FK CASCADE 2 件。
//	  合計 56。
var migrationCascadeAllowlist = map[string]int{
	"001_init.sql": 56,
}

// onDeleteCascadeRE matches PostgreSQL-equivalent ON DELETE CASCADE spellings:
// case folding, flexible whitespace (incl. newlines), without requiring exact
// "ON DELETE CASCADE" literal. MDL-05: exact Count was false-negative for variants.
var onDeleteCascadeRE = regexp.MustCompile(`(?is)\bon\s+delete\s+cascade\b`)

// countCascadeOccurrences は SQL テキスト中の ON DELETE CASCADE 出現数を数える純粋関数。
func countCascadeOccurrences(sql string) int {
	return len(onDeleteCascadeRE.FindAllStringIndex(sql, -1))
}

// reconcileMigrationCascade は found（実測 ファイル名→出現数）と allowlist を突合する純粋関数。
// 違反があれば人間可読な文字列を返す（空 = クリーン）。
func reconcileMigrationCascade(found, allowlist map[string]int) []string {
	var violations []string

	for file, count := range found {
		allowed, ok := allowlist[file]
		switch {
		case !ok && count > 0:
			violations = append(violations,
				"migration "+file+" has "+strconv.Itoa(count)+" ON DELETE CASCADE occurrence(s) "+
					"not on the allowlist. This migration is not applied yet (append-only rule), so "+
					"either remove the CASCADE (prefer SET NULL/RESTRICT for tables that may hold "+
					"patient/financial result values) or add a reviewed allowlist entry with a "+
					"documented reason (pure dependent child table only).")
		case ok && count != allowed:
			violations = append(violations,
				"migration "+file+" CASCADE count changed: found "+strconv.Itoa(count)+
					", allowlist pins "+strconv.Itoa(allowed)+". Applied migrations must not be edited "+
					"(checksum mismatch); if this is a legitimate new migration file reusing an "+
					"already-allowlisted name, that itself is a bug. Investigate before updating the pin.")
		}
	}
	for file, allowed := range allowlist {
		if _, ok := found[file]; !ok {
			violations = append(violations,
				"allowlist entry for "+file+" no longer matches an existing migration file "+
					"(renamed/removed?) — remove the stale entry.")
			continue
		}
		if allowed == 0 {
			violations = append(violations,
				"allowlist entry for "+file+" pins 0 — remove the entry instead of pinning zero.")
		}
	}
	return violations
}

// walkMigrationsForCascade reads every *.sql file under migrationsDir and counts CASCADE occurrences.
func walkMigrationsForCascade(t *testing.T) map[string]int {
	t.Helper()
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}
	found := make(map[string]int)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(migrationsDir, e.Name()))
		if err != nil {
			t.Fatalf("read migration %s: %v", e.Name(), err)
		}
		found[e.Name()] = countCascadeOccurrences(string(raw))
	}
	return found
}

// TestMigrationCascadeInventory_NoUnreviewedCascade is the gate: every migration file's
// "ON DELETE CASCADE" occurrence count must match the reviewed allowlist. A floor guards
// against a vacuous pass if the migrations directory glob silently breaks.
//
// Floor rationale (2026-07-27): DDL は統合スキーマ 001_init.sql の単一ファイル。
// The floor tracks the current known minimum .sql file count, not an arbitrary buffer;
// it must be revisited whenever migrations are consolidated/split or the directory's .sql
// file count otherwise legitimately changes.
func TestMigrationCascadeInventory_NoUnreviewedCascade(t *testing.T) {
	found := walkMigrationsForCascade(t)

	const minDDLFiles = 1
	if len(found) < minDDLFiles {
		t.Fatalf("expected at least %d migration .sql file(s) under %s, found %d "+
			"(directory read broken or files missing).", minDDLFiles, migrationsDir, len(found))
	}

	for _, v := range reconcileMigrationCascade(found, migrationCascadeAllowlist) {
		t.Error(v)
	}
}

// TestReconcileMigrationCascade_Analyzer pins detection on inline fixtures: known CASCADE
// occurrences must be counted correctly, and the reconciler must flag each failure mode.
func TestReconcileMigrationCascade_Analyzer(t *testing.T) {
	t.Run("counts multiple occurrences in one file", func(t *testing.T) {
		sql := "a REFERENCES x(id) ON DELETE CASCADE,\nb REFERENCES y(id) ON DELETE CASCADE,\nc REFERENCES z(id) ON DELETE SET NULL"
		if got := countCascadeOccurrences(sql); got != 2 {
			t.Fatalf("got %d, want 2", got)
		}
	})

	t.Run("detects case and whitespace variants (MDL-05)", func(t *testing.T) {
		variants := []string{
			"on delete cascade",
			"ON DELETE  CASCADE",
			"ON DELETE\n    CASCADE",
			"On Delete Cascade",
		}
		for _, sql := range variants {
			if got := countCascadeOccurrences(sql); got != 1 {
				t.Fatalf("sql %q: got %d, want 1", sql, got)
			}
		}
	})

	t.Run("zero occurrences", func(t *testing.T) {
		sql := "a REFERENCES x(id) ON DELETE SET NULL,\nb REFERENCES y(id) ON DELETE RESTRICT"
		if got := countCascadeOccurrences(sql); got != 0 {
			t.Fatalf("got %d, want 0", got)
		}
	})

	baseAllowlist := map[string]int{"001_init.sql": 50, "010_add_checkup_packages.sql": 2}
	baseFound := map[string]int{"001_init.sql": 50, "010_add_checkup_packages.sql": 2}

	t.Run("clean baseline reports no violation", func(t *testing.T) {
		got := reconcileMigrationCascade(baseFound, baseAllowlist)
		if len(got) != 0 {
			t.Fatalf("expected 0 violations, got %v", got)
		}
	})

	t.Run("new migration with unreviewed CASCADE fails", func(t *testing.T) {
		found := map[string]int{"001_init.sql": 50, "010_add_checkup_packages.sql": 2, "012_new.sql": 1}
		got := reconcileMigrationCascade(found, baseAllowlist)
		if len(got) != 1 || !strings.Contains(got[0], "012_new.sql") || !strings.Contains(got[0], "not on the allowlist") {
			t.Fatalf("expected unreviewed CASCADE to be flagged, got %v", got)
		}
	})

	t.Run("new migration with no CASCADE does not fail", func(t *testing.T) {
		found := map[string]int{"001_init.sql": 50, "010_add_checkup_packages.sql": 2, "012_new.sql": 0}
		got := reconcileMigrationCascade(found, baseAllowlist)
		if len(got) != 0 {
			t.Fatalf("expected 0 violations for CASCADE-free migration, got %v", got)
		}
	})

	t.Run("increased count on existing file fails (edited applied migration)", func(t *testing.T) {
		found := map[string]int{"001_init.sql": 51, "010_add_checkup_packages.sql": 2}
		got := reconcileMigrationCascade(found, baseAllowlist)
		if len(got) != 1 || !strings.Contains(got[0], "001_init.sql") || !strings.Contains(got[0], "count changed") {
			t.Fatalf("expected count-change violation, got %v", got)
		}
	})

	t.Run("stale allowlist entry fails", func(t *testing.T) {
		found := map[string]int{"010_add_checkup_packages.sql": 2}
		got := reconcileMigrationCascade(found, baseAllowlist)
		if len(got) != 1 || !strings.Contains(got[0], "001_init.sql") || !strings.Contains(got[0], "stale") {
			t.Fatalf("expected stale entry violation, got %v", got)
		}
	})

	t.Run("zero-pinned allowlist entry fails", func(t *testing.T) {
		found := map[string]int{"001_init.sql": 0}
		got := reconcileMigrationCascade(found, map[string]int{"001_init.sql": 0})
		if len(got) != 1 || !strings.Contains(got[0], "pins 0") {
			t.Fatalf("expected zero-pin violation, got %v", got)
		}
	})
}
