#!/bin/sh
# Warn about staged source files that exceed the 800-line guideline.
# 800 lines is a maintainability guideline, not a hard limit: this check never
# blocks the commit (documented exceptions such as frontend/src/lib/design-tokens.ts
# exist). Soft guidance remains 500 (see frontend/CLAUDE.md and .claude hooks).
# Skips generated, migrations, vendor, and test files.

set -e

GUIDELINE_LINES=800
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
    *_test.go|*.test.ts|*.test.tsx|*.test.js|*.test.jsx|*.test.mjs|*.test.cjs|*.spec.ts|*.spec.tsx|*.spec.mjs|*.spec.cjs) continue ;;
  esac
  [ -f "$f" ] || continue
  lines=$(wc -l < "$f" | tr -d ' ')
  if [ "$lines" -gt "$GUIDELINE_LINES" ]; then
    OVER="$OVER\n  $f ($lines lines)"
  fi
done

if [ -n "$OVER" ]; then
  echo "WARNING: staged source files exceed the ${GUIDELINE_LINES}-line guideline (not blocking):"
  printf "%b\n" "$OVER"
  echo "Consider splitting unless the file is a documented exception."
fi
exit 0
