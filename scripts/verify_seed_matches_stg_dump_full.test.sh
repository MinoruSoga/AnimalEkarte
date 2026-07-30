#!/usr/bin/env bash
# scripts/verify_seed_matches_stg_dump_full.test.sh
#
# verify_seed_matches_stg_dump_full.sh のセキュリティ契約を静的に検証する。
# Docker は起動しない。スクリプト本文のみを検査する。
#
# 契約 (SEC-CS-F03):
#   1. 全インタフェース公開の `-p "${PORT}:5432"` が無い
#      （ホスト公開するなら `127.0.0.1:${PORT}:5432` のみ。公開しないのも可）
#   2. 固定パスワード `POSTGRES_PASSWORD=verify` が無い
#   3. ループバック限定の公開、またはポート公開自体が無い
#   4. trap / cleanup が残っている（一時コンテナと作業 dir の後始末）
#
# Usage: bash scripts/verify_seed_matches_stg_dump_full.test.sh
# Exit codes:
#   0  全アサーション PASS
#   1  いずれかが FAIL
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="$SCRIPT_DIR/verify_seed_matches_stg_dump_full.sh"

if [[ ! -f "$TARGET" ]]; then
  echo "FAIL  target script not found at $TARGET"
  exit 1
fi

failures=0

assert_true() {
  local name="$1"
  shift
  if "$@"; then
    echo "PASS  [$name]"
  else
    echo "FAIL  [$name]"
    failures=$((failures + 1))
  fi
}

assert_false() {
  local name="$1"
  shift
  if "$@"; then
    echo "FAIL  [$name] (expected not to match)"
    failures=$((failures + 1))
  else
    echo "PASS  [$name]"
  fi
}

# --- 1. No all-interfaces host port publish: -p "${PORT}:5432" without 127.0.0.1 ---
# Matches docker -p forms that bind 0.0.0.0 (all interfaces), not loopback.
has_all_interfaces_port_publish() {
  # Literal pattern used by the vulnerable script: -p "${PORT}:5432"
  # Also catch unquoted / alternate expansions that still bind all interfaces.
  grep -nE -- \
    '-p[[:space:]]+"\$\{PORT\}:5432"|-p[[:space:]]+\$\{PORT\}:5432|-p[[:space:]]+"\$PORT:5432"|-p[[:space:]]+\$PORT:5432' \
    "$TARGET" >/dev/null 2>&1
}

assert_false "no all-interfaces -p \${PORT}:5432" has_all_interfaces_port_publish

# --- 2. No fixed POSTGRES_PASSWORD=verify ---
has_fixed_password_verify() {
  grep -nE -- 'POSTGRES_PASSWORD=verify([[:space:]]|$|"|'\'')' "$TARGET" >/dev/null 2>&1
}

assert_false "no fixed POSTGRES_PASSWORD=verify" has_fixed_password_verify

# --- 3. Loopback publication OR no host port publish at all ---
# Accept either:
#   a) -p "127.0.0.1:${PORT}:5432" (or equivalent loopback bind)
#   b) no -p host port publish for postgres 5432
has_loopback_port_publish() {
  grep -nE -- \
    '-p[[:space:]]+"127\.0\.0\.1:\$\{PORT\}:5432"|-p[[:space:]]+127\.0\.0\.1:\$\{PORT\}:5432|-p[[:space:]]+"127\.0\.0\.1:\$PORT:5432"|-p[[:space:]]+127\.0\.0\.1:\$PORT:5432' \
    "$TARGET" >/dev/null 2>&1
}

has_any_host_port_5432_publish() {
  # Any -p ...:5432 form (loopback or not). Used only to allow "no publish".
  grep -nE -- '-p[[:space:]]+[^[:space:]]*5432' "$TARGET" >/dev/null 2>&1
}

loopback_or_no_publish() {
  if has_loopback_port_publish; then
    return 0
  fi
  if ! has_any_host_port_5432_publish; then
    return 0
  fi
  return 1
}

assert_true "loopback publish or no host port publish" loopback_or_no_publish

# --- 4. trap / cleanup still present ---
has_cleanup_function() {
  grep -nE -- '^[[:space:]]*cleanup_all\(\)[[:space:]]*\{' "$TARGET" >/dev/null 2>&1
}

has_trap_cleanup() {
  # trap cleanup_all EXIT (or trap with cleanup_all and EXIT)
  grep -nE -- 'trap[[:space:]]+[^;]*cleanup_all[^;]*EXIT|trap[[:space:]]+cleanup_all[[:space:]]+EXIT' \
    "$TARGET" >/dev/null 2>&1
}

has_docker_rm_in_cleanup() {
  # cleanup must still remove the temporary container
  grep -nE -- 'docker[[:space:]]+rm[[:space:]]+-f' "$TARGET" >/dev/null 2>&1
}

assert_true "cleanup_all function present" has_cleanup_function
assert_true "trap cleanup_all EXIT present" has_trap_cleanup
assert_true "docker rm -f still in cleanup path" has_docker_rm_in_cleanup

# --- Summary ---
echo ""
if [[ "$failures" -eq 0 ]]; then
  echo "ALL PASS ($TARGET security contract)"
  exit 0
else
  echo "FAILED: $failures assertion(s)"
  exit 1
fi
