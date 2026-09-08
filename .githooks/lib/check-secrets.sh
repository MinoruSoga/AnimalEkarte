#!/bin/sh
# Scan staged contents, redact findings, and never pull an image automatically.
set -eu
ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"
if command -v gitleaks >/dev/null 2>&1; then
  gitleaks protect --staged --config .gitleaks.toml --redact
  exit 0
fi
# The fallback only detects common patterns. It cannot establish scanner parity.
PATTERN='sk-or-|sk-ant-|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{20,}|-----BEGIN ([A-Z0-9]+ )?PRIVATE KEY-----'
if git diff --cached --no-ext-diff --no-textconv --unified=0 | grep -E '^\+' | grep -E -q -- "$PATTERN"; then
  echo 'FAIL: a possible secret was detected in staged additions (content redacted).' >&2
  exit 1
fi
echo 'BLOCKED: gitleaks unavailable; fallback found no common pattern but is not a full scan. Install the approved scanner manually.' >&2
exit 2
