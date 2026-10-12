#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

if [[ ! -f "$REPO_ROOT/scripts/ci/ci-python.sh" ]]; then
  echo "FAIL  helper not found at $REPO_ROOT/scripts/ci/ci-python.sh"
  exit 1
fi
. "$REPO_ROOT/scripts/ci/ci-python.sh"

if ! declare -F ci_py >/dev/null 2>&1; then
  echo "FAIL  ci_py not defined after sourcing ci-python.sh"
  exit 1
fi

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

MOCK_LOG="$TMP_ROOT/exec.log"
OUT="$TMP_ROOT/stdout"
ERR="$TMP_ROOT/stderr"
failures=0

make_mock() {
  local path="$1" probe_rc="$2" exec_rc="$3"
  mkdir -p "$(dirname "$path")"
  cat > "$path" <<EOF
#!/bin/bash
if [[ "\$1" == "-c" ]]; then
  printf 'PROBE %s\n' "\${0##*/}" >> "$MOCK_LOG"
  exit $probe_rc
fi
{
  printf 'EXEC %s\n' "\${0##*/}"
  for _a in "\$@"; do printf 'ARG %s\n' "\$_a"; done
} >> "$MOCK_LOG"
exit $exec_rc
EOF
  chmod +x "$path"
}

invoke() {
  local mock_path="$1" ci_python="$2"
  shift 2
  local rc=0
  ( PATH="$mock_path" CI_PYTHON="$ci_python" ci_py "$@" ) >"$OUT" 2>"$ERR" || rc=$?
  return "$rc"
}

pass() { echo "PASS  [$1] $2"; }
fail() {
  echo "FAIL  [$1] $2"
  [[ -f "$MOCK_LOG" ]] && sed 's/^/      log: /' "$MOCK_LOG"
  [[ -s "$ERR" ]] && sed 's/^/      err: /' "$ERR"
  failures=$((failures + 1))
}

d="$TMP_ROOT/case1"; mkdir -p "$d"
make_mock "$d/python3" 1 0
make_mock "$d/python3.14" 0 0
: > "$MOCK_LOG"
rc=0; invoke "$d" "" script.py arg1 || rc=$?
if [[ "$rc" -eq 0 ]] \
  && grep -qxF "PROBE python3" "$MOCK_LOG" \
  && grep -qxF "PROBE python3.14" "$MOCK_LOG" \
  && grep -qxF "EXEC python3.14" "$MOCK_LOG" \
  && grep -qxF "ARG script.py" "$MOCK_LOG" \
  && ! grep -qxF "EXEC python3" "$MOCK_LOG"; then
  pass "old-python3-skipped" "3.9 相当の python3 を飛ばして python3.14 を選択 (rc=0)"
else
  fail "old-python3-skipped" "rc=$rc"
fi

d="$TMP_ROOT/case2"; mkdir -p "$d"
make_mock "$d/python3" 0 0
make_mock "$d/python3.14" 0 0
: > "$MOCK_LOG"
rc=0; invoke "$d" "" script.py || rc=$?
if [[ "$rc" -eq 0 ]] \
  && grep -qxF "EXEC python3" "$MOCK_LOG" \
  && ! grep -q "python3\.14" "$MOCK_LOG"; then
  pass "bare-python3-preferred" "先頭候補の python3 で即決し後続候補は probe しない"
else
  fail "bare-python3-preferred" "rc=$rc"
fi

d="$TMP_ROOT/case3"; mkdir -p "$d"
make_mock "$d/python3" 1 0
: > "$MOCK_LOG"
rc=0; invoke "$d" "" script.py || rc=$?
if [[ "$rc" -ne 0 ]] \
  && grep -q "3\.11" "$ERR" \
  && grep -qi "CI_PYTHON" "$ERR" \
  && ! grep -q "EXEC" "$MOCK_LOG"; then
  pass "no-valid-candidate-fails" "rc=$rc, stderr に 3.11+ 要求と CI_PYTHON 回避策"
else
  fail "no-valid-candidate-fails" "rc=$rc"
fi

d="$TMP_ROOT/case4"; mkdir -p "$d" "$TMP_ROOT/py dir"
make_mock "$TMP_ROOT/py dir/my python" 0 0
: > "$MOCK_LOG"
rc=0; invoke "$d" "$TMP_ROOT/py dir/my python" -B "arg with space.py" plain || rc=$?
if [[ "$rc" -eq 0 ]] \
  && grep -qxF "PROBE my python" "$MOCK_LOG" \
  && grep -qxF "EXEC my python" "$MOCK_LOG" \
  && grep -qxF "ARG -B" "$MOCK_LOG" \
  && grep -qxF "ARG arg with space.py" "$MOCK_LOG" \
  && grep -qxF "ARG plain" "$MOCK_LOG"; then
  pass "explicit-ci-python-spaced-path" "スペース入り CI_PYTHON を検証して実行・引数を完全転送"
else
  fail "explicit-ci-python-spaced-path" "rc=$rc"
fi

d="$TMP_ROOT/case5"; mkdir -p "$d"
make_mock "$d/python3" 0 0
: > "$MOCK_LOG"
rc=0; invoke "$d" "$TMP_ROOT/no-such-dir/no python" script.py || rc=$?
if [[ "$rc" -ne 0 ]] \
  && grep -q "CI_PYTHON" "$ERR" \
  && ! grep -q "PROBE\|EXEC" "$MOCK_LOG"; then
  pass "explicit-missing-fails" "fallback の python3 を probe せず即 FAIL (rc=$rc)"
else
  fail "explicit-missing-fails" "rc=$rc"
fi

make_mock "$d/py-old" 1 0
: > "$MOCK_LOG"
rc=0; invoke "$d" "$d/py-old" script.py || rc=$?
if [[ "$rc" -ne 0 ]] \
  && grep -q "CI_PYTHON" "$ERR" \
  && grep -qxF "PROBE py-old" "$MOCK_LOG" \
  && ! grep -q "python3" "$MOCK_LOG"; then
  pass "explicit-old-fails" "旧版の明示指定は fallback を試さず FAIL (rc=$rc)"
else
  fail "explicit-old-fails" "rc=$rc"
fi

d="$TMP_ROOT/case6"; mkdir -p "$d"
make_mock "$d/python3" 0 23
make_mock "$d/python3.14" 0 0
: > "$MOCK_LOG"
rc=0; invoke "$d" "" script.py || rc=$?
if [[ "$rc" -eq 23 ]] \
  && grep -qxF "EXEC python3" "$MOCK_LOG" \
  && ! grep -q "python3\.14" "$MOCK_LOG"; then
  pass "script-exit-propagated" "exit 23 をそのまま返し python3.14 は probe もしない"
else
  fail "script-exit-propagated" "rc=$rc (expected 23)"
fi

echo "----"
if [[ "$failures" -gt 0 ]]; then
  echo "RESULT  $failures case(s) failed"
  exit 1
fi
echo "RESULT  all cases passed"
exit 0
