#!/usr/bin/env bash
# scripts/check-test-worker-makefile.test.sh
#
# Makefile の test-worker ターゲット契約。ホスト pnpm 直叩きではなく
# Docker 経由で root package.json の vitest-pool-workers を走らせること。
# Docker 不要・Makefile の静的検査。
#
# Usage: bash scripts/check-test-worker-makefile.test.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MAKEFILE="$ROOT/Makefile"

if [[ ! -f "$MAKEFILE" ]]; then
  echo "FAIL  Makefile not found at $MAKEFILE"
  exit 1
fi

failures=0

assert_grep() {
  local name="$1"
  local pattern="$2"
  if grep -Eq "$pattern" "$MAKEFILE"; then
    echo "PASS  [$name]"
  else
    echo "FAIL  [$name] pattern not found: $pattern"
    failures=$((failures + 1))
  fi
}

assert_grep "phony-includes-test-worker" '^\.PHONY:.*[[:space:]]test-worker([[:space:]]|$)'
assert_grep "target-exists" '^test-worker:'
assert_grep "vitest-config" 'backend/worker/vitest\.config\.mts'
assert_grep "frozen-lockfile" 'pnpm install --frozen-lockfile'
assert_grep "help-mentions-test-worker" 'test-worker'

recipe="$(awk '/^test-worker:/{p=1;next} p && /^[^[:space:]#]/{exit} p' "$MAKEFILE")"
if [[ -z "$recipe" ]]; then
  echo "FAIL  [recipe-nonempty] test-worker recipe missing"
  failures=$((failures + 1))
elif printf '%s\n' "$recipe" | grep -Eq 'docker run'; then
  echo "PASS  [recipe-uses-docker-run]"
else
  echo "FAIL  [recipe-uses-docker-run] recipe does not call docker run"
  failures=$((failures + 1))
fi

if printf '%s\n' "$recipe" | grep -Eq '\$\(ARGS\)'; then
  echo "PASS  [args-forwarded]"
else
  echo "FAIL  [args-forwarded] test-worker recipe must forward \$(ARGS)"
  failures=$((failures + 1))
fi

if printf '%s\n' "$recipe" | grep -Eq '\$\(DC\) exec frontend'; then
  echo "FAIL  [not-frontend-exec] test-worker must not use the frontend container"
  failures=$((failures + 1))
else
  echo "PASS  [not-frontend-exec]"
fi

if [[ "$failures" -eq 0 ]]; then
  echo "OK  test-worker Makefile contract"
  exit 0
fi
echo "FAIL  $failures check(s)"
exit 1
