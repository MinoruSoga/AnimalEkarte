#!/usr/bin/env bash
# scripts/check-csv-import-account-source.test.sh
#
# Makefile の CSV_IMPORT_ACCOUNT_SOURCE_DIR 解決の回帰テスト。
# Docker 不要。ローカル handoff は staffs.csv を
# seeds/002_master/accounts/_old_db_handoff/<clinic>/ に分離しており、
# --account-source-dir が落ちると preflight が /migration-input/staffs.csv を探す。
#
# 検証すること:
#   1. SOURCE が _old_db_handoff/<clinic> のとき ACCOUNT が accounts 側を指す
#   2. 親 make が空の CSV_IMPORT_ACCOUNT_SOURCE_DIR を export しても再計算する
#   3. SOURCE 側の seeds ツリーから解決する（CURDIR との大小文字差に依存しない）
#   4. 明示の非空 ACCOUNT は優先する
#   5. handoff 以外の SOURCE では空のまま
#
# Usage: bash scripts/check-csv-import-account-source.test.sh
# Exit: 0=PASS, 1=FAIL
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

failures=0

PROBE_MK="$TMP_ROOT/probe.mk"
cat >"$PROBE_MK" <<'EOF'
include Makefile
.DEFAULT_GOAL := print-csv-import-account-source
.PHONY: print-csv-import-account-source
print-csv-import-account-source:
	@printf '%s\n' '$(CSV_IMPORT_ACCOUNT_SOURCE_DIR)'
EOF

SEEDS="$TMP_ROOT/seeds"
CLINIC_DIR="$SEEDS/_old_db_handoff/probeclinic"
ACCOUNT_DIR="$SEEDS/002_master/accounts/_old_db_handoff/probeclinic"
mkdir -p "$CLINIC_DIR" "$ACCOUNT_DIR"
printf 'id\n' >"$ACCOUNT_DIR/staffs.csv"
chmod 700 "$CLINIC_DIR" "$ACCOUNT_DIR" \
  "$SEEDS/_old_db_handoff" "$SEEDS/002_master/accounts/_old_db_handoff"
chmod 600 "$ACCOUNT_DIR/staffs.csv"

OTHER_DIR="$TMP_ROOT/not-handoff/bundle"
mkdir -p "$OTHER_DIR"

resolve_account() {
  local source_dir="$1"
  shift
  env "$@" make --no-print-directory -f "$PROBE_MK" \
    CSV_IMPORT_SOURCE_DIR="$source_dir" 2>/dev/null
}

expect_eq() {
  local name="$1" expected="$2" actual="$3"
  if [[ "$actual" == "$expected" ]]; then
    echo "PASS  [$name]"
  else
    echo "FAIL  [$name]"
    echo "  expected: [$expected]"
    echo "  actual:   [$actual]"
    failures=$((failures + 1))
  fi
}

got="$(resolve_account "$CLINIC_DIR")"
expect_eq "handoff-resolves-account-dir" "$ACCOUNT_DIR" "$got"

got="$(resolve_account "$CLINIC_DIR" CSV_IMPORT_ACCOUNT_SOURCE_DIR=)"
expect_eq "empty-env-recomputes" "$ACCOUNT_DIR" "$got"

# Symlink seeds under another path prefix so SOURCE does not share CURDIR's
# string prefix (macOS Dev/Case vs dev/case regression).
LOWER_LINK="$TMP_ROOT/casealt-seeds"
ln -s "$SEEDS" "$LOWER_LINK"
CASE_SOURCE="$LOWER_LINK/_old_db_handoff/probeclinic"
got="$(resolve_account "$CASE_SOURCE")"
expect_eq "source-tree-resolves-without-curdir-prefix" \
  "$LOWER_LINK/002_master/accounts/_old_db_handoff/probeclinic" "$got"

EXPLICIT="$TMP_ROOT/explicit-account"
mkdir -p "$EXPLICIT"
printf 'id\n' >"$EXPLICIT/staffs.csv"
got="$(env CSV_IMPORT_ACCOUNT_SOURCE_DIR="$EXPLICIT" make --no-print-directory -f "$PROBE_MK" \
  CSV_IMPORT_SOURCE_DIR="$CLINIC_DIR" 2>/dev/null)"
expect_eq "explicit-account-kept" "$EXPLICIT" "$got"

got="$(resolve_account "$OTHER_DIR")"
expect_eq "non-handoff-empty" "" "$got"

if [[ "$failures" -ne 0 ]]; then
  echo "FAIL  $failures case(s)"
  exit 1
fi
echo "PASS  csv-import account source resolution"
exit 0
