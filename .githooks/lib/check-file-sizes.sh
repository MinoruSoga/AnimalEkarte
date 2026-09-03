#!/bin/sh
# Block staged source files that exceed the project hard line limit (800).
# Soft guidance remains 500 (see frontend/CLAUDE.md and .claude hooks).
# Skips generated, migrations, vendor, and test files.

set -e

MAX_LINES=800
ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"

STAGED=$(git diff --cached --name-only --diff-filter=ACM || true)
[ -z "$STAGED" ] && exit 0

OVER=""
for f in $STAGED; do
  case "$f" in
    *.go|*.ts|*.tsx|*.js|*.jsx|*.mjs|*.cjs) ;;
    *) continue ;;
  esac
  case "$f" in
    *generated*|*migrations*|*vendor*|*node_modules*) continue ;;
    *_test.go|*.test.ts|*.test.tsx|*.test.js|*.test.jsx|*.spec.ts|*.spec.tsx) continue ;;
  esac
  [ -f "$f" ] || continue
  lines=$(wc -l < "$f" | tr -d ' ')
  if [ "$lines" -gt "$MAX_LINES" ]; then
    OVER="$OVER\n  $f ($lines lines)"
  fi
done

if [ -n "$OVER" ]; then
  echo "ERROR: staged source files exceed ${MAX_LINES}-line hard limit:"
  printf "%b\n" "$OVER"
  echo "Split into smaller modules before committing."
  exit 1
fi
