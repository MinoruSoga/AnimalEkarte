package lintscan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestMigrationNumberUniqueness は backend/migrations/ 直下の NNN_*.sql で
// 数字 prefix が重複していたら FAIL する。
//
// この gate が存在する理由: schema_migrations は filename を主キーに記録するため、
// 同一番号のファイルは適用順を番号で決められず、ファイル名順への暗黙依存になる。
// 加えて新規環境でどちらか一方だけが適用された場合、もう一方が同名テーブルを
// 作りに来て失敗し得る。実際に 2026-09-26、並行ブランチ由来で 011 が衝突した
// （011_care_plan_items_manual_other.sql / 011_support_bug_reports.sql）。
//
// 適用済みファイルのリネームは filename 主キーに反するため解消不能であり、
// 既知の 011 ペアのみを例外として許容する。新しい重複 prefix、または 011 への
// 第3ファイルの追加は FAIL とする。
func TestMigrationNumberUniqueness(t *testing.T) {
	moduleRoot := mustFindSeedCSVModuleRoot(t)
	migrationsDir := filepath.Join(moduleRoot, "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations directory %s: %v", migrationsDir, err)
	}

	byNumber := make(map[int][]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := seedCSVMigrationFilenamePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		number, err := strconv.Atoi(match[1])
		if err != nil {
			t.Fatalf("parse migration number from %s: %v", entry.Name(), err)
		}
		byNumber[number] = append(byNumber[number], entry.Name())
	}
	if len(byNumber) == 0 {
		t.Fatalf("no NNN_*.sql migrations found in %s", migrationsDir)
	}

	// 許容済みの既知重複: 011 の2ファイルのみ。リネームすると適用済み環境で
	// 新名が未適用扱いとなり DDL が重複実行されるため、共存を維持する。
	knownCollision := map[string]bool{
		"011_care_plan_items_manual_other.sql": true,
		"011_support_bug_reports.sql":          true,
	}

	var violations []string
	for number, names := range byNumber {
		if len(names) == 1 {
			continue
		}
		sort.Strings(names)
		if number == 11 {
			knownOnly := len(names) == len(knownCollision)
			for _, name := range names {
				if !knownCollision[name] {
					knownOnly = false
				}
			}
			if knownOnly {
				continue
			}
		}
		violations = append(violations, fmt.Sprintf("%03d: %s", number, strings.Join(names, ", ")))
	}
	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf(
			"duplicate migration number prefixes detected (schema_migrations は filename 主キーのため適用済みファイルのリネーム不可 — 新規ファイルは未使用の連番にすること):\n%s",
			strings.Join(violations, "\n"),
		)
	}
}
