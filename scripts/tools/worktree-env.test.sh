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
  cat >"$1/fixture.template" <<'EOF'
DB_HOST=db
DB_PORT=5432
DB_USER=ekarte_user
CUSTOM_VALUE=template-kept
DB_NAME=ekarte_db
CORS_ALLOWED_ORIGIN=http://localhost:3003
EXTRA_VALUE=template-extra
EOF
}

new_case() {
  local dir="$TMP_ROOT/$1"
  mkdir -p "$dir"
  write_example "$dir"
  printf '%s' "$dir"
}

# ── Case 1: fresh fixture.runtime generation with offset ─────────────────────────
case1="$(new_case case1)"
WORKTREE_ENV_ROOT="$case1" WORKTREE_ENV_EXAMPLE="$case1/fixture.template" WORKTREE_ENV_FILE="$case1/fixture.runtime" "$TARGET" 300 >/dev/null 2>&1 \
  || { ng "case1: exit non-zero"; }
env_file="$case1/fixture.runtime"
if [[ -f "$env_file" ]]; then
  assert_contains "$env_file" "COMPOSE_PROJECT_NAME=animalekarte-wt300" "case1: project name offset"
  assert_contains "$env_file" "DB_HOST_PORT=11794" "case1: db port = block 300 start"
  assert_contains "$env_file" "BACKEND_HOST_PORT=11795" "case1: backend port = block 300 backend"
  assert_contains "$env_file" "FRONTEND_HOST_PORT=11796" "case1: frontend port = block 300 frontend"
  assert_contains "$env_file" "STORYBOOK_HOST_PORT=11797" "case1: storybook port"
  assert_contains "$env_file" "SWAGGER_HOST_PORT=11798" "case1: swagger port"
  assert_contains "$env_file" "REDOC_HOST_PORT=11799" "case1: redoc port"
  assert_contains "$env_file" "POSTGRES_DATA_VOLUME=ekarte-postgres-data-wt300" "case1: db volume isolated"
  assert_contains "$env_file" "FRONTEND_NODE_MODULES_VOLUME=ekarte-frontend-node-modules-wt300" "case1: node_modules isolated"
  assert_contains "$env_file" "CORS_ALLOWED_ORIGIN=http://localhost:11796" "case1: CORS follows frontend port"
  assert_contains "$env_file" "EXTRA_VALUE=template-extra" "case1: template content kept"
else
  ng "case1: fixture.runtime not created"
fi

# ── Case 2: idempotent re-run (single managed block) ────────────────────────
WORKTREE_ENV_ROOT="$case1" WORKTREE_ENV_EXAMPLE="$case1/fixture.template" WORKTREE_ENV_FILE="$case1/fixture.runtime" "$TARGET" 300 >/dev/null 2>&1 || true
WORKTREE_ENV_ROOT="$case1" WORKTREE_ENV_EXAMPLE="$case1/fixture.template" WORKTREE_ENV_FILE="$case1/fixture.runtime" "$TARGET" 300 >/dev/null 2>&1 || true
block_count="$(grep -c 'worktree-runtime managed' "$env_file" || true)"
if [[ "$block_count" == "2" ]]; then
  ok "case2: single managed block after re-run"
else
  ng "case2: managed block count = $block_count (want 2 marker lines)"
fi

# ── Case 3: existing fixture.runtime preserved + CORS rewritten ──────────────────
case3="$(new_case case3)"
cat >"$case3/fixture.runtime" <<'EOF'
DB_HOST=db
DB_USER=ekarte_user
CUSTOM_VALUE=custom-kept
CORS_ALLOWED_ORIGIN=http://localhost:3003
EXTRA_VALUE=extra-kept
EOF
WORKTREE_ENV_ROOT="$case3" WORKTREE_ENV_EXAMPLE="$case3/fixture.template" WORKTREE_ENV_FILE="$case3/fixture.runtime" "$TARGET" 7 >/dev/null 2>&1 || true
env3="$case3/fixture.runtime"
assert_contains "$env3" "CUSTOM_VALUE=custom-kept" "case3: custom value preserved"
assert_contains "$env3" "EXTRA_VALUE=extra-kept" "case3: extra key preserved"
assert_contains "$env3" "CORS_ALLOWED_ORIGIN=http://localhost:10038" "case3: CORS rewritten to offset port"
if [[ "$(grep -c '^CORS_ALLOWED_ORIGIN=' "$env3")" == "1" ]]; then
  ok "case3: exactly one CORS line (no duplicate-key ambiguity)"
else
  ng "case3: CORS line count != 1"
fi

# ── Case 4: offset change rewrites managed values only ──────────────────────
WORKTREE_ENV_ROOT="$case3" WORKTREE_ENV_EXAMPLE="$case3/fixture.template" WORKTREE_ENV_FILE="$case3/fixture.runtime" "$TARGET" 42 >/dev/null 2>&1 || true
assert_contains "$env3" "DB_HOST_PORT=10246" "case4: db port re-offset (block 42 start)"
assert_contains "$env3" "animalekarte-wt42" "case4: project re-slugged"
assert_contains "$env3" "CUSTOM_VALUE=custom-kept" "case4: custom value still preserved"

# ── Case 5: custom slug ─────────────────────────────────────────────────────
case5="$(new_case case5)"
WORKTREE_ENV_ROOT="$case5" WORKTREE_ENV_EXAMPLE="$case5/fixture.template" WORKTREE_ENV_FILE="$case5/fixture.runtime" "$TARGET" 10 knip-fix >/dev/null 2>&1 || true
assert_contains "$case5/fixture.runtime" "COMPOSE_PROJECT_NAME=animalekarte-knip-fix" "case5: custom slug project"
assert_contains "$case5/fixture.runtime" "ekarte-postgres-data-knip-fix" "case5: custom slug volume"

# ── Case 6: invalid inputs are rejected ─────────────────────────────────────
case6="$(new_case case6)"
for bad in "abc" "0" "9257" "-5"; do
  if WORKTREE_ENV_ROOT="$case6" WORKTREE_ENV_EXAMPLE="$case6/fixture.template" WORKTREE_ENV_FILE="$case6/fixture.runtime" "$TARGET" "$bad" >/dev/null 2>&1; then
    ng "case6: offset $bad accepted (want reject)"
  else
    ok "case6: offset $bad rejected"
  fi
done
if WORKTREE_ENV_ROOT="$case6" WORKTREE_ENV_EXAMPLE="$case6/fixture.template" WORKTREE_ENV_FILE="$case6/fixture.runtime" "$TARGET" 5 'BAD SLUG!' >/dev/null 2>&1; then
  ng "case6: invalid slug accepted"
else
  ok "case6: invalid slug rejected"
fi

# ── Case 7: every service in adjacent blocks and main stays disjoint ─────────
case7="$(new_case case7)"
case8="$(new_case case8)"
for offset in 1 2; do
  target_dir="$case7"
  [[ "$offset" == "2" ]] && target_dir="$case8"
  WORKTREE_ENV_ROOT="$target_dir" WORKTREE_ENV_EXAMPLE="$target_dir/fixture.template" \
    WORKTREE_ENV_FILE="$target_dir/fixture.runtime" "$TARGET" "$offset" >/dev/null 2>&1
done
ports7="$(sed -n 's/^[A-Z_]*HOST_PORT=//p' "$case7/fixture.runtime")"
ports8="$(sed -n 's/^[A-Z_]*HOST_PORT=//p' "$case8/fixture.runtime")"
all_ports="$(printf '%s\n%s\n%s\n' "$ports7" "$ports8" $'5434\n8080\n3003\n6006\n8081\n8082')"
if [[ "$(printf '%s\n' "$all_ports" | sort -u | wc -l | tr -d ' ')" == "18" ]]; then
  ok "case7: adjacent worktrees and main have eighteen distinct service ports"
else
  ng "case7: service ports overlap"
fi

# ── Case 8: largest supported block stays within TCP port range ─────────────
WORKTREE_ENV_ROOT="$case8" WORKTREE_ENV_EXAMPLE="$case8/fixture.template" \
  WORKTREE_ENV_FILE="$case8/fixture.runtime" "$TARGET" 9256 >/dev/null 2>&1
assert_contains "$case8/fixture.runtime" "REDOC_HOST_PORT=65535" "case8: last supported port is 65535"

# ── Summary ─────────────────────────────────────────────────────────────────
echo
echo "worktree-env.test.sh: $PASS_COUNT passed, $FAIL_COUNT failed"
[[ "$FAIL_COUNT" -eq 0 ]]
