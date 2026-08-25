package lintscan

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestDocsMigrationInventoryDrift fails when a document states the migration inventory as a
// fixed value instead of pointing at the command that produces it.
//
// Why this gate exists: `backend/migrations/` 直下の DDL は増分の追加と `001_init.sql` への統合で
// 増減する。ところが 2026-07-29 時点で 7 個のドキュメントがその在庫（「DDL は 1 本」「schema_migrations
// は 4 行」）を各自の散文で複製していた。複製は独立に腐り、USER は誤った期待値で再構築の成否を照合する
// ことになる（4 行を期待して 6 行を見て「失敗した」と誤判断する）。同じ欠陥を 5 ラウンドかけて手で
// 追いかけたが、検出の仕組みが無い限り再発する。
//
// 規則: **ドキュメントに書いてよいのは「意図」と「コマンド」であり「値」ではない。** 在庫を知りたい
// 読者には `ls backend/migrations/*.sql` を案内する。行数を知りたい読者には直下 DDL の本数 + seed
// バンドル数という導出規則を案内する。固定値を書いてはならない。
//
// 過去形の履歴記述（「2026-07-27 に統合して単一ファイルへ戻した」）は真であり保持してよい。ただし
// 散文から「過去の事実」と「現在の断定」を機械的に弁別することはできない（erd.md の方針転換記録が
// その実例である）ため、本 gate は同 package の dbOrTx / audit-tx inventory lint と同じ idiom を採る:
// 許容する出現を理由付きで allowlist へ登録し、未登録の出現は全て FAIL とする。
//
// allowlist の key は **行番号ではなく本文の一部** である。行番号は上流の編集で必ずずれ、ずれた瞬間に
// 「doc が嘘をついた」ように見える偽 FAIL を出すためである。
func TestDocsMigrationInventoryDrift(t *testing.T) {
	repoRoot := repositoryRootForDocsGate(t)
	documents := discoverInventoryGateDocuments(t, repoRoot)

	if len(documents) == 0 {
		t.Fatalf("docs inventory gate: 走査対象が 0 件。探索経路が壊れている（違反ゼロと区別できないため fail-closed）")
	}

	var violations []string
	for _, relativePath := range sortedDocumentPaths(documents) {
		for lineNumber, line := range strings.Split(string(documents[relativePath]), "\n") {
			rule := matchFixedInventoryAssertion(line)
			if rule == "" {
				continue
			}
			if justification := allowedFixedInventoryAssertion(line); justification != "" {
				continue
			}
			violations = append(violations, fmt.Sprintf(
				"%s:%d [%s] %s", relativePath, lineNumber+1, rule, strings.TrimSpace(line),
			))
		}
	}

	if len(violations) > 0 {
		t.Fatalf(
			"docs が migration 在庫を固定値で断定している（%d 件）。値ではなくコマンド（`ls backend/migrations/*.sql`）"+
				"または導出規則（直下 DDL 本数 + seed バンドル数）を書くこと。過去形の履歴記述であれば "+
				"docsMigrationInventoryAllowlist へ理由付きで登録する:\n  %s",
			len(violations), strings.Join(violations, "\n  "),
		)
	}
}

// docsMigrationInventoryAllowlist は「固定値に見えるが許容してよい出現」を本文の一部で登録する。
// key は本文中の distinctive な部分文字列、value はなぜ許容できるかの理由。
var docsMigrationInventoryAllowlist = map[string]string{
	"同日中に「DDL は `001_init.sql` 単一ファイル」へ方針転換し": "2026-07-17 の方針転換を記録した過去形の履歴（erd.md の追記ブロック）",
	"直下 DDL を単一ファイルへ戻した":                      "2026-07-27 時点の統合作業を述べた過去形の履歴（backend/migrations/CLAUDE.md）",
	"2026-07-27に当時のincremental 002〜009は":      "ADR-003 が決定時点の状況を述べた過去形の記録。ADR は決定の記録であり現況の宣言ではない",
	"2026-07-27時点の現行構成を以下の通り整理しました":           "erd.md の静的照合が実施日を明示した調査記録。日付で時点が固定されている",
	"2026-08-20 統合第8回":                        "2026-08-20 の統合第8回を記録した過去形の履歴記述であり、当時の物理テーブル総数128が不変だった事実を残している。将来の在庫を規定する現在形の断定ではない",
	"2026-08-25 統合第9回":                        "2026-08-25 の統合第9回を記録した過去形の履歴記述であり、当時の物理テーブル総数128が不変だった事実を残している。将来の在庫を規定する現在形の断定ではない",
}

func allowedFixedInventoryAssertion(line string) string {
	for fragment, justification := range docsMigrationInventoryAllowlist {
		if strings.Contains(line, fragment) {
			return justification
		}
	}
	return ""
}

var fixedInventoryRules = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{
		// 「001_init.sql が唯一の DDL である」型の断定。
		name:    "single-ddl-assertion",
		pattern: regexp.MustCompile("001_init\\.sql[^\n]{0,40}?(1 ?本|1 ?件|単一|のみ)|(単一|1 ?本|1 ?件)[^\n]{0,40}?001_init\\.sql"),
	},
	{
		// 「schema_migrations は N 行」型の固定期待値。
		name:    "fixed-schema-migrations-row-count",
		pattern: regexp.MustCompile("schema_migrations[^\n]{0,40}?[0-9]+ ?行|[0-9]+ ?行[^\n]{0,40}?schema_migrations"),
	},
	{
		// 「DDL 1 + seed 3」型の固定内訳。
		name:    "fixed-ddl-seed-breakdown",
		pattern: regexp.MustCompile("(統合)?DDL ?1 ?(\\+|件|本)[^\n]{0,20}?seed"),
	},
}

// sourceCitationPattern は `001_init.sql:2735` のような **行番号付きの出典引用** に一致する。
// 引用は在庫の主張ではないため、規則の照合前に取り除く。これを取り除かないと
// 「payment 系トリガーは …（`001_init.sql:2735`）のみで」のような、`のみ` が DDL 在庫ではなく
// 別の対象を修飾している行を誤検出する（本 gate の初回逆テストで実測した偽陽性）。
var sourceCitationPattern = regexp.MustCompile(`001_init\.sql:[0-9]+`)

func matchFixedInventoryAssertion(line string) string {
	trimmed := strings.TrimSpace(sourceCitationPattern.ReplaceAllString(line, "«citation»"))
	if trimmed == "" {
		return ""
	}
	for _, rule := range fixedInventoryRules {
		if rule.pattern.MatchString(trimmed) {
			return rule.name
		}
	}
	return ""
}

// discoverInventoryGateDocuments は本 gate の走査対象を集める。lintscan.WalkInternalTree は
// internal/**/*.go 専用の discovery なのでここでは使えない（同 package の docstring がその境界を
// 明示している）。erd_table_count_drift_test.go と同じく module root からの相対 path で辿る。
func discoverInventoryGateDocuments(t *testing.T, repoRoot string) map[string][]byte {
	t.Helper()

	documents := make(map[string][]byte)

	addFile := func(relativePath string) {
		content, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
		if err != nil {
			if os.IsNotExist(err) {
				return
			}
			t.Fatalf("docs inventory gate: read %s: %v", relativePath, err)
		}
		documents[relativePath] = content
	}

	// docs ツリー配下の Markdown を全数。
	docsRoot := filepath.Join(repoRoot, "docs")
	if err := filepath.WalkDir(docsRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}
		relativePath, relErr := filepath.Rel(repoRoot, path)
		if relErr != nil {
			return relErr
		}
		addFile(relativePath)
		return nil
	}); err != nil {
		t.Fatalf("docs inventory gate: walk docs/: %v", err)
	}

	// migration 手順の正本と、repo 直下の計画・台帳ドキュメント。
	addFile(filepath.Join("backend", "migrations", "CLAUDE.md"))
	for _, rootDocument := range []string{"3-session-agent.html", "q&a.html", "phase2.html", "BE-pending.md", "BE-refactor.md", "README.md"} {
		addFile(rootDocument)
	}

	return documents
}

func repositoryRootForDocsGate(t *testing.T) string {
	t.Helper()

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("docs inventory gate: os.Getwd: %v", err)
	}
	moduleRoot, err := FindModuleRoot(workingDirectory)
	if err != nil {
		t.Fatalf("docs inventory gate: FindModuleRoot: %v", err)
	}
	return filepath.Join(moduleRoot, "..")
}

func sortedDocumentPaths(documents map[string][]byte) []string {
	keys := make([]string, 0, len(documents))
	for key := range documents {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
