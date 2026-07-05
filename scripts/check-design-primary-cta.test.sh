#!/usr/bin/env bash
# scripts/check-design-primary-cta.test.sh
#
# check-design-primary-cta.mjs の回帰テスト。
# 一時 fixture で
#   - クリーン tree      → exit 0
#   - 意図的 violation   → exit 1
# を確認する。Docker / pnpm 不要。
#
# Usage: bash scripts/check-design-primary-cta.test.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHECK="$SCRIPT_DIR/check-design-primary-cta.mjs"

if [[ ! -f "$CHECK" ]]; then
  echo "FAIL  check script not found at $CHECK"
  exit 1
fi

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

failures=0

run_expect() {
  local name="$1" expected_exit="$2" root="$3"
  local out actual_exit
  out="$(node "$CHECK" --root "$root" 2>&1)" && actual_exit=0 || actual_exit=$?
  if [[ "$actual_exit" -eq "$expected_exit" ]]; then
    echo "PASS  [$name] exit=$actual_exit (expected $expected_exit)"
  else
    echo "FAIL  [$name] exit=$actual_exit (expected $expected_exit)"
    echo "----- output -----"
    printf '%s\n' "$out"
    echo "------------------"
    failures=$((failures + 1))
  fi
}

# 1. 本番 tree は clean
run_expect "production-tree" 0 "$SCRIPT_DIR/.."

# 2. 意図的 violation fixture
FIXTURE="$TMP_ROOT/violation"
mkdir -p "$FIXTURE/frontend/src/features/demo"
cat > "$FIXTURE/frontend/src/features/demo/BadCta.tsx" <<'EOF'
export function BadCta() {
  return (
    <SubmitButton className={STYLE.confirmPrimary}>
      保存
    </SubmitButton>
  );
}
EOF

run_expect "violation-fixture" 1 "$FIXTURE"

if [[ "$failures" -gt 0 ]]; then
  echo "FAIL  $failures test case(s) failed"
  exit 1
fi

echo "PASS  all check-design-primary-cta tests"
