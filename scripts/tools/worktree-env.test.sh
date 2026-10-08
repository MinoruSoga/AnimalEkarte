#!/usr/bin/env bash
# worktree-env.test.sh — worktree-env.sh の退行テスト。
# 実 repo を触らないよう WORKTREE_ENV_ROOT/EXAMPLE/FILE を temp dir に向けて実行する。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="$SCRIPT_DIR/worktree-env.sh"

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

PASS_COUNT=0
FAIL_COUNT=0

ok() {
  PASS_COUNT=$((PASS_COUNT + 1))
  echo "PASS  $1"
}

ng() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  echo "FAIL  $1" >&2
}

assert_contains() { # file needle label
  if grep -qF "$2" "$1"; then ok "$3"; else ng "$3 (missing: $2 in $1)"; fi
}

assert_not_contains() { # file needle label
  if grep -qF "$2" "$1"; then ng "$3 (unexpected: $2 in $1)"; else ok "$3"; fi
}

write_example() {
  cat >"$1/.env.example" <<'EOF'
DB_HOST=db
DB_PORT=5432
DB_USER=ekarte_user
DB_PASSWORD=ekarte_local_dev_password
DB_NAME=ekarte_db
CORS_ALLOWED_ORIGIN=http://localhost:3003
JWT_SECRET=ekarte-local-dev-jwt-secret-not-for-production-use-32c
EOF
}

new_case() {
  local dir="$TMP_ROOT/$1"
  mkdir -p "$dir"
  write_example "$dir"
  printf '%s' "$dir"
}

# ── Case 1: fresh .env.local generation with offset ─────────────────────────
case1="$(new_case case1)"
WORKTREE_ENV_ROOT="$case1" "$TARGET" 300 >/dev/null 2>&1 \
  || { ng "case1: exit non-zero"; }
env_file="$case1/.env.local"
if [[ -f "$env_file" ]]; then
  assert_contains "$env_file" "COMPOSE_PROJECT_NAME=animalekarte-wt300" "case1: project name offset"
  assert_contains "$env_file" "DB_HOST_PORT=5734" "case1: db port = 5434+300"
  assert_contains "$env_file" "BACKEND_HOST_PORT=8380" "case1: backend port = 8080+300"
  assert_contains "$env_file" "FRONTEND_HOST_PORT=3303" "case1: frontend port = 3003+300"
  assert_contains "$env_file" "STORYBOOK_HOST_PORT=6306" "case1: storybook port"
  assert_contains "$env_file" "SWAGGER_HOST_PORT=8381" "case1: swagger port"
  assert_contains "$env_file" "REDOC_HOST_PORT=8382" "case1: redoc port"
  assert_contains "$env_file" "POSTGRES_DATA_VOLUME=ekarte-postgres-data-wt300" "case1: db volume isolated"
  assert_contains "$env_file" "FRONTEND_NODE_MODULES_VOLUME=ekarte-frontend-node-modules-wt300" "case1: node_modules isolated"
  assert_contains "$env_file" "CORS_ALLOWED_ORIGIN=http://localhost:3303" "case1: CORS follows frontend port"
  assert_contains "$env_file" "JWT_SECRET=ekarte-local-dev-jwt-secret-not-for-production-use-32c" "case1: template content kept"
else
  ng "case1: .env.local not created"
fi

# ── Case 2: idempotent re-run (single managed block) ────────────────────────
WORKTREE_ENV_ROOT="$case1" "$TARGET" 300 >/dev/null 2>&1 || true
WORKTREE_ENV_ROOT="$case1" "$TARGET" 300 >/dev/null 2>&1 || true
block_count="$(grep -c 'worktree-runtime managed' "$env_file" || true)"
if [[ "$block_count" == "2" ]]; then
  ok "case2: single managed block after re-run"
else
  ng "case2: managed block count = $block_count (want 2 marker lines)"
fi

# ── Case 3: existing .env.local preserved + CORS rewritten ──────────────────
case3="$(new_case case3)"
cat >"$case3/.env.local" <<'EOF'
DB_HOST=db
DB_USER=ekarte_user
DB_PASSWORD=my_custom_secret_keepme
CORS_ALLOWED_ORIGIN=http://localhost:3003
SUPPORT_LLM_API_KEY=keep-this-key
EOF
WORKTREE_ENV_ROOT="$case3" "$TARGET" 7 >/dev/null 2>&1 || true
env3="$case3/.env.local"
assert_contains "$env3" "DB_PASSWORD=my_custom_secret_keepme" "case3: custom value preserved"
assert_contains "$env3" "SUPPORT_LLM_API_KEY=keep-this-key" "case3: extra key preserved"
assert_contains "$env3" "CORS_ALLOWED_ORIGIN=http://localhost:3010" "case3: CORS rewritten to offset port"
if [[ "$(grep -c '^CORS_ALLOWED_ORIGIN=' "$env3")" == "1" ]]; then
  ok "case3: exactly one CORS line (no duplicate-key ambiguity)"
else
  ng "case3: CORS line count != 1"
fi

# ── Case 4: offset change rewrites managed values only ──────────────────────
WORKTREE_ENV_ROOT="$case3" "$TARGET" 42 >/dev/null 2>&1 || true
assert_contains "$env3" "DB_HOST_PORT=5476" "case4: db port re-offset (5434+42)"
assert_contains "$env3" "animalekarte-wt42" "case4: project re-slugged"
assert_contains "$env3" "DB_PASSWORD=my_custom_secret_keepme" "case4: custom value still preserved"

# ── Case 5: custom slug ─────────────────────────────────────────────────────
case5="$(new_case case5)"
WORKTREE_ENV_ROOT="$case5" "$TARGET" 10 knip-fix >/dev/null 2>&1 || true
assert_contains "$case5/.env.local" "COMPOSE_PROJECT_NAME=animalekarte-knip-fix" "case5: custom slug project"
assert_contains "$case5/.env.local" "ekarte-postgres-data-knip-fix" "case5: custom slug volume"

# ── Case 6: invalid inputs are rejected ─────────────────────────────────────
case6="$(new_case case6)"
for bad in "abc" "0" "57001" "-5"; do
  if WORKTREE_ENV_ROOT="$case6" "$TARGET" "$bad" >/dev/null 2>&1; then
    ng "case6: offset $bad accepted (want reject)"
  else
    ok "case6: offset $bad rejected"
  fi
done
if WORKTREE_ENV_ROOT="$case6" "$TARGET" 5 'BAD SLUG!' >/dev/null 2>&1; then
  ng "case6: invalid slug accepted"
else
  ok "case6: invalid slug rejected"
fi

# ── Summary ─────────────────────────────────────────────────────────────────
echo
echo "worktree-env.test.sh: $PASS_COUNT passed, $FAIL_COUNT failed"
[[ "$FAIL_COUNT" -eq 0 ]]
