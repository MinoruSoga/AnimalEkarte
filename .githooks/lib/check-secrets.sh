#!/bin/sh
# Secret scan for staged changes. Prefers gitleaks (same config as CI).
# Falls back to a pattern scan when gitleaks is unavailable.

set -e

ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"

if command -v gitleaks >/dev/null 2>&1; then
  echo "  [secrets] gitleaks protect --staged"
  gitleaks protect --staged --config .gitleaks.toml --verbose
  exit 0
fi

if command -v docker >/dev/null 2>&1; then
  echo "  [secrets] gitleaks via docker (local gitleaks missing)"
  docker run --rm -v "$ROOT:/repo" -w /repo zricethezav/gitleaks:latest \
    protect --staged --config .gitleaks.toml --verbose
  exit 0
fi

echo "  [secrets] gitleaks unavailable — fallback pattern scan on staged files"
STAGED=$(git diff --cached --name-only --diff-filter=ACM || true)
[ -z "$STAGED" ] && exit 0

# Patterns aligned with readiness/setup check-secrets.js examples.
PATTERN='sk-or-|sk-ant-|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{20,}|-----BEGIN ([A-Z0-9]+ )?PRIVATE KEY-----'
FOUND=0
for f in $STAGED; do
  [ -f "$f" ] || continue
  case "$f" in
    *.png|*.jpg|*.jpeg|*.gif|*.webp|*.pdf|*.zip|*.woff*|*.ttf) continue ;;
  esac
  if git grep -E -n -I --cached -e "$PATTERN" -- "$f" >/dev/null 2>&1; then
    echo "ERROR: possible secret pattern in staged file: $f"
    git grep -E -n -I --cached -e "$PATTERN" -- "$f" || true
    FOUND=1
  fi
done

if [ "$FOUND" -ne 0 ]; then
  echo "Install gitleaks for full CI-parity scanning: https://github.com/gitleaks/gitleaks"
  exit 1
fi
