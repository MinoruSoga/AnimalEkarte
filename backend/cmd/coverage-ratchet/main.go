// Command coverage-ratchet は backend のカバレッジが記録済みベースラインから低下していないかを判定する
// （BE-refactor.md R3-5 / bug.md H-7・coverage-policy.md Phase 1 ratchet）。
//
// 使い方:
//
//	go tool cover -func=coverage.out | go run ./cmd/coverage-ratchet -baseline .coverage-baseline
//
// もしくは -func-file で `go tool cover -func` の出力ファイルを渡す。
//
// ベースラインファイル（.coverage-baseline）は「backend 総計カバレッジ %」を1行の数値で持つ。
// 未記録（プレースホルダ 0 / ファイル欠落）の場合は warn のみで exit 0 とし、初回 CI 実測での
// ベースライン記録を促す（coverage-policy.md のベースライン記録手順）。ベースラインが記録済みなら
// 現在値がベースライン - tolerance を下回ったとき exit 非0（fail）でリファクタによる低下を検出する。
//
// 本ツール自体は cmd/ 配下のためカバレッジ計測対象外だが、判定ロジック（parseTotalCoverage /
// evaluateRatchet）は純粋関数として main_test.go で単体検証する。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	baselinePath := flag.String("baseline", ".coverage-baseline", "path to committed baseline file (single float percent)")
	funcFile := flag.String("func-file", "", "path to `go tool cover -func` output; if empty, read stdin")
	tolerance := flag.Float64("tolerance", 0.5, "allowed downward drift in percentage points before failing")
	warnOnly := flag.Bool("warn-only", false, "never exit non-zero; print WARN instead of failing (Phase 1)")
	flag.Parse()

	var coverageOut string
	if *funcFile != "" {
		b, err := os.ReadFile(*funcFile) //nolint:gosec // path is an operator-provided CLI flag (CI-controlled), not untrusted input
		if err != nil {
			fmt.Fprintf(os.Stderr, "coverage-ratchet: read func file: %v\n", err)
			os.Exit(2)
		}
		coverageOut = string(b)
	} else {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "coverage-ratchet: read stdin: %v\n", err)
			os.Exit(2)
		}
		coverageOut = string(b)
	}

	current, err := parseTotalCoverage(coverageOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coverage-ratchet: %v\n", err)
		os.Exit(2)
	}

	baseline, err := readBaseline(*baselinePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coverage-ratchet: baseline error: %v\n", err)
		os.Exit(2)
	}

	res := evaluateRatchet(current, baseline, *tolerance)
	fmt.Println(res.message)
	if res.fail && !*warnOnly {
		os.Exit(1)
	}
}

// parseTotalCoverage は `go tool cover -func` 出力の末尾 "total:" 行から総計カバレッジ % を取り出す。
// 例: "total:\t(statements)\t83.4%" → 83.4
func parseTotalCoverage(funcOutput string) (float64, error) {
	var lastTotal string
	for line := range strings.SplitSeq(funcOutput, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "total:") {
			lastTotal = line
		}
	}
	if lastTotal == "" {
		return 0, fmt.Errorf("no 'total:' line found in coverage output")
	}
	fields := strings.Fields(lastTotal)
	pctStr := fields[len(fields)-1]
	pctStr = strings.TrimSuffix(pctStr, "%")
	pct, err := strconv.ParseFloat(pctStr, 64)
	if err != nil {
		return 0, fmt.Errorf("parse total coverage %q: %w", pctStr, err)
	}
	return pct, nil
}

// readBaseline は baseline ファイルから記録済みカバレッジ % を読む。
// CMD-03: ファイル欠落は未記録 (0, nil)。読取/解析失敗は error で分離する。
func readBaseline(path string) (float64, error) {
	b, err := os.ReadFile(path) //nolint:gosec // path is an operator-provided CLI flag (CI-controlled), not untrusted input
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read baseline %s: %w", path, err)
	}
	sawContent := false
	for line := range strings.SplitSeq(string(b), "\n") {
		s := strings.TrimSpace(line)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		sawContent = true
		if v, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64); err == nil {
			return v, nil
		}
	}
	if sawContent {
		return 0, fmt.Errorf("baseline %s has no parseable coverage value", path)
	}
	return 0, nil
}

type ratchetResult struct {
	fail    bool
	message string
}

// evaluateRatchet は現在値とベースラインを比較する純粋関数。
//   - baseline <= 0（未記録）: warn のみ、fail=false。初回 CI 実測でのベースライン記録を促す。
//   - current >= baseline - tolerance: OK。
//   - current < baseline - tolerance: fail=true（リファクタによるカバレッジ低下）。
func evaluateRatchet(current, baseline, tolerance float64) ratchetResult {
	if baseline <= 0 {
		return ratchetResult{
			fail: false,
			message: fmt.Sprintf("coverage-ratchet: baseline not recorded (current=%.1f%%). "+
				"Record this value in .coverage-baseline / docs/ops/coverage-policy.md to arm the ratchet.", current),
		}
	}
	floor := baseline - tolerance
	if current < floor {
		return ratchetResult{
			fail: true,
			message: fmt.Sprintf("coverage-ratchet: FAIL — coverage %.1f%% dropped below baseline %.1f%% "+
				"(tolerance %.1fpp, floor %.1f%%). A refactor must not lower coverage; add tests or update "+
				"the baseline with justification.", current, baseline, tolerance, floor),
		}
	}
	return ratchetResult{
		fail: false,
		message: fmt.Sprintf("coverage-ratchet: OK — coverage %.1f%% ≥ floor %.1f%% (baseline %.1f%%, tolerance %.1fpp).",
			current, floor, baseline, tolerance),
	}
}
