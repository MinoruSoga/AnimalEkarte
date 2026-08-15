#!/usr/bin/env bash
# scripts/check-workflow-remote-exec-policy.sh
#
# Fail if Dockerfiles or GitHub Actions workflows contain remote pipe-to-shell
# installers (docs/ops/ci-policy.md — supply-chain ban):
#   - curl|sh / curl|bash (and wget variants)
#   - bash <(curl ...) / sh <(curl ...) process substitution
#   - raw.githubusercontent.com/.../(master|main|HEAD)/... used in a pipe
#
# Usage: bash scripts/check-workflow-remote-exec-policy.sh [repo-root]
# Exit: 0 = PASS, 1 = FAIL
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="${1:-$SCRIPT_DIR/..}"
ROOT="$(cd "$ROOT" && pwd)"

if [[ ! -d "$ROOT" ]]; then
  echo "FAIL  repo root not found: $ROOT"
  exit 1
fi

# Collect targets: Dockerfile* under repo, and .github/workflows/*.{yml,yaml}
targets=()
while IFS= read -r -d '' f; do
  targets+=("$f")
done < <(find "$ROOT" \
  \( -path "$ROOT/.git" -o -path "$ROOT/node_modules" -o -path "$ROOT/**/node_modules" \) -prune -o \
  -type f \( -name 'Dockerfile' -o -name 'Dockerfile.*' -o -name '*.Dockerfile' \) -print0 2>/dev/null)

if [[ -d "$ROOT/.github/workflows" ]]; then
  shopt -s nullglob
  for f in "$ROOT/.github/workflows"/*.yml "$ROOT/.github/workflows"/*.yaml; do
    targets+=("$f")
  done
  shopt -u nullglob
fi

if [[ ${#targets[@]} -eq 0 ]]; then
  echo "FAIL  no Dockerfiles or workflow files found under $ROOT"
  exit 1
fi

# Collapse backslash-continued lines so multi-line RUN curl ... | sh is one unit.
# Drop full-line comments (# ...) so ban documentation does not false-positive.
normalize_file() {
  local file="$1"
  # strip full-line comments, then join "\\\n" continuations
  sed -E 's/^[[:space:]]*#.*$//' "$file" \
    | sed -e ':a' -e '/\\$/N; s/\\\n//; ta'
}

errors=0

# Patterns (ERE). Avoid newlines inside character classes (grep: brackets not balanced).
# 1) curl|sh / curl|bash / wget|sh (optional whitespace around |)
# 2) bash <(curl ...) / sh <(curl ...)
# 3) raw.githubusercontent.com/.../(master|main|HEAD)/... appearing with a pipe-to-shell
re_curl_pipe='(curl|wget)[^|]*\|[[:space:]]*(ba)?sh'
re_proc_sub='(ba)?sh[[:space:]]*<\([[:space:]]*curl'
re_raw_mutable='raw\.githubusercontent\.com/[^[:space:]"'\'']+/(master|main|HEAD)/'
re_pipe_shell='\|[[:space:]]*(ba)?sh'

for f in "${targets[@]}"; do
  # skip empty / unreadable
  [[ -r "$f" ]] || continue
  content="$(normalize_file "$f")"
  rel="${f#"$ROOT"/}"

  # Line-by-line on normalized content
  while IFS= read -r line || [[ -n "$line" ]]; do
    [[ -z "${line//[[:space:]]/}" ]] && continue

    if echo "$line" | grep -Eiq "$re_curl_pipe"; then
      echo "FAIL  remote pipe-to-shell (curl|sh) in $rel"
      echo "      $line"
      errors=$((errors + 1))
      continue
    fi

    if echo "$line" | grep -Eiq "$re_proc_sub"; then
      echo "FAIL  process-substitution remote exec (bash <(curl)) in $rel"
      echo "      $line"
      errors=$((errors + 1))
      continue
    fi

    if echo "$line" | grep -Eiq "$re_raw_mutable" && echo "$line" | grep -Eiq "$re_pipe_shell"; then
      echo "FAIL  raw.githubusercontent.com mutable-ref pipe install in $rel"
      echo "      $line"
      errors=$((errors + 1))
      continue
    fi

    # Also ban mutable raw.githubusercontent.com install.sh even without pipe on same
    # visual line after partial join failures (explicit install.sh from master/main).
    if echo "$line" | grep -Eiq "$re_raw_mutable"'[^[:space:]"'\'']*install\.sh'; then
      echo "FAIL  raw.githubusercontent.com mutable-ref install.sh in $rel"
      echo "      $line"
      errors=$((errors + 1))
      continue
    fi
  done <<< "$content"
done

if [[ "$errors" -gt 0 ]]; then
  echo "----"
  printf 'RESULT  %d remote-exec policy violation(s) — pin release artifacts with checksum or use a pinned image/action (docs/ops/ci-policy.md)\n' "$errors"
  exit 1
fi

echo "PASS  no remote pipe-to-shell patterns in Dockerfiles/workflows"
exit 0
