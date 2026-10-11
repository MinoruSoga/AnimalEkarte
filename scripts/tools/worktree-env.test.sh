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

case9="$(new_case case9)"
WORKTREE_ENV_ROOT="$case9" WORKTREE_ENV_EXAMPLE="$case9/fixture.template" \
  WORKTREE_ENV_FILE="$case9/fixture.runtime" "$TARGET" 08 >/dev/null 2>&1 || true
assert_contains "$case9/fixture.runtime" "DB_HOST_PORT=10042" "case9: 08 parses as decimal offset 8"
assert_contains "$case9/fixture.runtime" "REDOC_HOST_PORT=10047" "case9: 08 block end"
assert_contains "$case9/fixture.runtime" "COMPOSE_PROJECT_NAME=animalekarte-wt08" "case9: raw 08 kept in default slug"
assert_contains "$case9/fixture.runtime" "ekarte-postgres-data-wt08" "case9: raw 08 kept in volume"

case9b="$(new_case case9b)"
WORKTREE_ENV_ROOT="$case9b" WORKTREE_ENV_EXAMPLE="$case9b/fixture.template" \
  WORKTREE_ENV_FILE="$case9b/fixture.runtime" "$TARGET" 0008 >/dev/null 2>&1 || true
assert_contains "$case9b/fixture.runtime" "DB_HOST_PORT=10042" "case9b: 0008 parses as decimal offset 8"
assert_contains "$case9b/fixture.runtime" "COMPOSE_PROJECT_NAME=animalekarte-wt0008" "case9b: raw 0008 kept in default slug"

case9c="$(new_case case9c)"
WORKTREE_ENV_ROOT="$case9c" WORKTREE_ENV_EXAMPLE="$case9c/fixture.template" \
  WORKTREE_ENV_FILE="$case9c/fixture.runtime" "$TARGET" 08 custom-slug >/dev/null 2>&1 || true
assert_contains "$case9c/fixture.runtime" "DB_HOST_PORT=10042" "case9c: 08 + custom slug keeps offset-8 ports"
assert_contains "$case9c/fixture.runtime" "ekarte-postgres-data-custom-slug" "case9c: explicit slug volume unchanged"

case10="$(new_case case10)"
for spec in "1:10000:10005" "2:10006:10011" "7:10036:10041" "8:10042:10047" \
            "42:10246:10251" "300:11794:11799" "1000:15994:15999" \
            "2000:21994:21999" "3000:27994:27999" "9256:65530:65535"; do
  vec_off=${spec%%:*}; vec_rest=${spec#*:}
  vec_first=${vec_rest%%:*}; vec_last=${vec_rest##*:}
  WORKTREE_ENV_ROOT="$case10" WORKTREE_ENV_EXAMPLE="$case10/fixture.template" \
    WORKTREE_ENV_FILE="$case10/fixture.runtime" "$TARGET" "$vec_off" >/dev/null 2>&1 || true
  assert_contains "$case10/fixture.runtime" "DB_HOST_PORT=$vec_first" "case10: offset $vec_off block start $vec_first"
  assert_contains "$case10/fixture.runtime" "REDOC_HOST_PORT=$vec_last" "case10: offset $vec_off block end $vec_last"
done

case11="$(new_case case11)"
printf 'PRESERVED=pre-existing\nCORS_ALLOWED_ORIGIN=http://localhost:9999\n' > "$case11/fixture.runtime"
cp "$case11/fixture.runtime" "$case11/before.bytes"
for bad in "0" "0000" "9257" "-1" "abc" "14" "014" "181" "0181" \
           "999999999999999999999999999999999999999999"; do
  if WORKTREE_ENV_ROOT="$case11" WORKTREE_ENV_EXAMPLE="$case11/fixture.template" \
      WORKTREE_ENV_FILE="$case11/fixture.runtime" "$TARGET" "$bad" >/dev/null 2>&1; then
    ng "case11: offset $bad accepted (want reject)"
  elif cmp -s "$case11/before.bytes" "$case11/fixture.runtime"; then
    ok "case11: offset $bad rejected, file bytes unchanged"
  else
    ng "case11: offset $bad rejected but file bytes changed"
  fi
done

case12="$(new_case case12)"
WORKTREE_ENV_ROOT="$case12" WORKTREE_ENV_EXAMPLE="$case12/fixture.template" \
  WORKTREE_ENV_FILE="$case12/fixture.runtime" "$TARGET" 300 >/dev/null 2>&1 || true
cp "$case12/fixture.runtime" "$case12/first.bytes"
WORKTREE_ENV_ROOT="$case12" WORKTREE_ENV_EXAMPLE="$case12/fixture.template" \
  WORKTREE_ENV_FILE="$case12/fixture.runtime" "$TARGET" 300 >/dev/null 2>&1 || true
if cmp -s "$case12/first.bytes" "$case12/fixture.runtime"; then
  ok "case12: repeated regeneration byte-identical"
else
  ng "case12: regenerated file differs (blank-line accumulation?)"
fi
blank_runs="$(grep -c '^$' "$case12/fixture.runtime" || true)"
if [[ "$blank_runs" == "1" ]]; then
  ok "case12: exactly one separator blank line"
else
  ng "case12: separator blank line count = $blank_runs (want 1)"
fi

actual_ports=""
for spec in "1:10000:10005" "2:10006:10011" "7:10036:10041" "8:10042:10047" \
            "42:10246:10251" "300:11794:11799" "1000:15994:15999" \
            "2000:21994:21999" "3000:27994:27999" "9256:65530:65535"; do
  vec_off=${spec%%:*}; vec_rest=${spec#*:}
  vec_first=${vec_rest%%:*}; vec_last=${vec_rest##*:}
  vec_dir="$(new_case "case13-$vec_off")"
  WORKTREE_ENV_ROOT="$vec_dir" WORKTREE_ENV_EXAMPLE="$vec_dir/fixture.template" \
    WORKTREE_ENV_FILE="$vec_dir/fixture.runtime" "$TARGET" "$vec_off" >/dev/null 2>&1 \
    || { ng "case13: offset $vec_off generation failed"; continue; }
  vec_ports="$(sed -n 's/^[A-Z_]*HOST_PORT=//p' "$vec_dir/fixture.runtime")"
  vec_count="$(printf '%s\n' "$vec_ports" | wc -l | tr -d ' ')"
  vec_uniq="$(printf '%s\n' "$vec_ports" | sort -un | wc -l | tr -d ' ')"
  if [[ "$vec_count" == "6" && "$vec_uniq" == "6" ]]; then
    ok "case13: offset $vec_off emits six unique host ports"
  else
    ng "case13: offset $vec_off host ports = $vec_count lines / $vec_uniq unique (want 6/6)"
  fi
  vec_min="$(printf '%s\n' "$vec_ports" | sort -n | head -1)"
  vec_max="$(printf '%s\n' "$vec_ports" | sort -n | tail -1)"
  if [[ "$vec_min" == "$vec_first" && "$vec_max" == "$vec_last" ]]; then
    ok "case13: offset $vec_off block bounds $vec_first..$vec_last"
  else
    ng "case13: offset $vec_off actual bounds $vec_min..$vec_max (want $vec_first..$vec_last)"
  fi
  actual_ports="$(printf '%s\n%s\n' "$actual_ports" "$vec_ports")"
done
main_defaults="$(printf '%s\n' 5434 8080 3003 6006 8081 8082)"
known_legacy="$(printf '%s\n' 6434 9080 4003 7006 9081 9082 \
  7434 10080 5003 8006 10081 10082 8434 11080 6003 9006 11081 11082)"
all_known="$(printf '%s\n%s\n%s\n' "$actual_ports" "$main_defaults" "$known_legacy" | sed '/^$/d')"
if [[ "$(printf '%s\n' "$all_known" | sort -u | wc -l | tr -d ' ')" == \
      "$(printf '%s\n' "$all_known" | wc -l | tr -d ' ')" ]]; then
  ok "case13: generated blocks disjoint from main defaults and all known legacy ports"
else
  ng "case13: port collision among generated/main/known-legacy sets"
fi
legacy_in_band="$(printf '%s\n' "$known_legacy" | awk '$1 >= 10000 { print }' | sort -nu)"
band_blocks="$(printf '%s\n' "$legacy_in_band" | awk '{ print 1 + int(($1 - 10000) / 6) }' | sort -nu | tr '\n' ' ' | sed 's/ $//')"
if [[ "$band_blocks" == "14 181" ]]; then
  ok "case13: reserved offsets cover exactly the legacy band blocks (14 181)"
else
  ng "case13: legacy band blocks = '$band_blocks' (want '14 181')"
fi

# ── Summary ─────────────────────────────────────────────────────────────────
echo
echo "worktree-env.test.sh: $PASS_COUNT passed, $FAIL_COUNT failed"
[[ "$FAIL_COUNT" -eq 0 ]]
