#!/usr/bin/env bash
# pre-push runs Git LFS for GitHub and skips it for Cloudflare Artifacts.
set -euo pipefail

ROOT=$(git rev-parse --show-toplevel)
HOOK="$ROOT/.githooks/pre-push"
TMP=$(mktemp -d "${TMPDIR:-/tmp}/ae-pre-push-lfs.XXXXXXXX")
trap 'rm -rf "$TMP"' EXIT

cat > "$TMP/git-lfs" <<'EOF'
#!/bin/sh
printf '%s\n' "$*" >> "$LFS_LOG"
exit 1
EOF
chmod 755 "$TMP/git-lfs"

run_hook() {
  local url="$1"
  local log="$2"
  LFS_LOG="$log" PATH="$TMP:$PATH" "$HOOK" origin "$url" >/dev/null
}

artifacts_log="$TMP/artifacts.log"
: > "$artifacts_log"
if run_hook "https://example.artifacts.cloudflare.net/git/AnimalEkarte/AnimalEkarte.git" "$artifacts_log"; then
  echo "PASS  artifacts push skips git lfs"
else
  echo "FAIL  artifacts push should skip git lfs" >&2
  exit 1
fi
if [ -s "$artifacts_log" ]; then
  echo "FAIL  artifacts push invoked git lfs" >&2
  exit 1
fi

github_log="$TMP/github.log"
: > "$github_log"
if run_hook "https://github.com/MinoruSoga/AnimalEkarte.git" "$github_log"; then
  echo "FAIL  github push should still run git lfs" >&2
  exit 1
fi
if ! grep -q "pre-push" "$github_log"; then
  echo "FAIL  github push did not invoke git lfs pre-push" >&2
  exit 1
fi
echo "PASS  github push still runs git lfs"
