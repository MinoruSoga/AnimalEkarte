#!/usr/bin/env bash
# scripts/local-db-reset-contract.sh
#
# Local-only single entry for recoverable, fail-closed DB rebuild.
# Invoked by `make reset`. USER-only destructive path — agents must not run execute mode.
#
# Contract (fail-closed):
#   1. Fixed expected project + volume names must match compose reality; refuse wrong env.
#   2. umask 077; owner-only gzipped pg_dumpall under .local-db-backups/<UTC>/,
#      plus SHA-256 and target volume + DDL/seed key manifest. Empty dump/digest/disk → stop.
#   3. Stop services without deleting volumes; remove ONLY ekarte-postgres-data;
#      keep cache volumes (frontend node_modules, go module/build cache).
#   4. Restart long-lived services; postflight: migration key coverage missing=0,
#      all root DDL keys, seed key 002_master, schema_migrations,
#      backend healthy + /health HTTP 200.
#   5. Snapshot failure must not proceed to volume delete.
#   6. Never wipe all compose volumes; never prune the volume store.
#
# Usage:
#   bash scripts/local-db-reset-contract.sh              # full execute (USER)
#   bash scripts/local-db-reset-contract.sh --contract-only
#
# Test overrides (fixture harness only):
#   LOCAL_DB_RESET_ROOT, LOCAL_DB_RESET_DOCKER, LOCAL_DB_RESET_BACKUP_ROOT,
#   LOCAL_DB_RESET_HEALTH_URL, LOCAL_DB_RESET_CURL, LOCAL_DB_RESET_COMPOSE_FILE,
#   LOCAL_DB_RESET_ENV_FILE, LOCAL_DB_RESET_EXPECTED_PROJECT,
#   LOCAL_DB_RESET_MIGRATIONS_DIR, LOCAL_DB_RESET_MAKEFILE
#
# Exit codes:
#   0  contract satisfied (and execute completed when not --contract-only)
#   1  contract violation or execute failure
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── Fixed expected values (local AnimalEkarte only) ─────────────────────────
EXPECTED_PROJECT_NAME="${LOCAL_DB_RESET_EXPECTED_PROJECT:-animalekarte}"
EXPECTED_DB_VOLUME="ekarte-postgres-data"
EXPECTED_CACHE_VOLUMES="ekarte-frontend-node-modules ekarte-go-mod-cache ekarte-go-build-cache"
EXPECTED_SEED_BUNDLES="002_master"
# Substrings that mark a project name as non-local (refuse).
FORBIDDEN_PROJECT_SUBSTRINGS="staging prod production stg a4 f8-g4 planetscale"

# ── Overridable paths / tools (tests inject fixtures and mocks) ─────────────
ROOT="${LOCAL_DB_RESET_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"
DOCKER_BIN="${LOCAL_DB_RESET_DOCKER:-docker}"
CURL_BIN="${LOCAL_DB_RESET_CURL:-curl}"
BACKUP_BASE="${LOCAL_DB_RESET_BACKUP_ROOT:-$ROOT/.local-db-backups}"
COMPOSE_FILE="${LOCAL_DB_RESET_COMPOSE_FILE:-$ROOT/docker-compose.yml}"
ENV_FILE="${LOCAL_DB_RESET_ENV_FILE:-$ROOT/.env.local}"
HEALTH_URL="${LOCAL_DB_RESET_HEALTH_URL:-http://127.0.0.1:8080/health}"
MIGRATIONS_DIR="${LOCAL_DB_RESET_MIGRATIONS_DIR:-$ROOT/backend/migrations}"
MAKEFILE_PATH="${LOCAL_DB_RESET_MAKEFILE:-$ROOT/Makefile}"

MODE="execute"
if [[ "${1:-}" == "--contract-only" ]]; then
  MODE="contract-only"
elif [[ -n "${1:-}" ]]; then
  echo "FAIL  unknown argument: $1 (use --contract-only or no args)"
  exit 1
fi

# Compose invocation: fixed project name, optional env-file when present.
compose() {
  # bash 3.2: build argv without namerefs
  if [[ -f "$ENV_FILE" ]]; then
    "$DOCKER_BIN" compose -p "$EXPECTED_PROJECT_NAME" -f "$COMPOSE_FILE" --env-file "$ENV_FILE" "$@"
  else
    "$DOCKER_BIN" compose -p "$EXPECTED_PROJECT_NAME" -f "$COMPOSE_FILE" "$@"
  fi
}

die() {
  echo "FAIL  $*" >&2
  exit 1
}

info() {
  echo "INFO  $*" >&2
}

# ── helpers ─────────────────────────────────────────────────────────────────
# True if space-separated list $1 contains exact token $2.
list_has() {
  local list="$1" token="$2" item
  for item in $list; do
    if [[ "$item" == "$token" ]]; then
      return 0
    fi
  done
  return 1
}

# Reads top-level compose `volumes:` mapping `name:` values (portable awk).
extract_compose_volume_names() {
  local file="$1"
  awk '
    /^volumes:[[:space:]]*$/ { inv=1; next }
    inv && /^[^[:space:]#]/ { exit }
    inv {
      # match "    name: ekarte-postgres-data" or quoted form
      if ($1 == "name:") {
        gsub(/["\047]/, "", $2)
        print $2
      }
    }
  ' "$file"
}

# Portable SHA-256 of a file → print hex digest only.
file_sha256() {
  local path="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$path" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$path" | awk '{print $1}'
  else
    die "neither sha256sum nor shasum is available"
  fi
}

list_root_ddl_basenames() {
  local dir="$1"
  local f
  if [[ ! -d "$dir" ]]; then
    return 0
  fi
  for f in "$dir"/*.sql; do
    [[ -e "$f" ]] || continue
    basename "$f"
  done | sort
}

# ── Static contract ─────────────────────────────────────────────────────────
validate_static_contract() {
  local errors=0
  local name
  local found_list=""
  local expected_list
  local v
  local reset_recipe

  expected_list="$EXPECTED_DB_VOLUME $EXPECTED_CACHE_VOLUMES"

  [[ -f "$COMPOSE_FILE" ]] || die "compose file not found: $COMPOSE_FILE"

  while IFS= read -r name; do
    [[ -z "$name" ]] && continue
    found_list="$found_list $name"
  done < <(extract_compose_volume_names "$COMPOSE_FILE")
  found_list="${found_list# }"

  if [[ -z "$found_list" ]]; then
    echo "FAIL  compose declares no named volumes with explicit name:"
    errors=$((errors + 1))
  fi

  for v in $expected_list; do
    if ! list_has "$found_list" "$v"; then
      echo "FAIL  compose missing expected volume name: $v"
      errors=$((errors + 1))
    fi
  done

  for v in $found_list; do
    if ! list_has "$expected_list" "$v"; then
      echo "FAIL  compose has unexpected volume name: $v"
      errors=$((errors + 1))
    fi
  done

  if [[ -f "$MAKEFILE_PATH" ]]; then
    reset_recipe="$(awk '
      /^reset:/ { inrecipe=1; next }
      inrecipe && /^[A-Za-z0-9_.-]+:/ { exit }
      inrecipe { print }
    ' "$MAKEFILE_PATH")"
    if printf '%s\n' "$reset_recipe" | grep -E 'down[[:space:]]+-v|[[:space:]]--volumes|volume[[:space:]]+prune' >/dev/null 2>&1; then
      echo "FAIL  Makefile reset recipe must not wipe all compose volumes or prune the volume store"
      errors=$((errors + 1))
    fi
    if ! printf '%s\n' "$reset_recipe" | grep -F 'local-db-reset-contract.sh' >/dev/null 2>&1; then
      echo "FAIL  Makefile reset recipe must invoke local-db-reset-contract.sh"
      errors=$((errors + 1))
    fi
  fi

  if [[ "$errors" -gt 0 ]]; then
    die "static contract failed ($errors error(s))"
  fi
  info "static contract OK (project=$EXPECTED_PROJECT_NAME db_volume=$EXPECTED_DB_VOLUME caches=$EXPECTED_CACHE_VOLUMES)"
}

# ── Live project / env guard ────────────────────────────────────────────────
validate_live_project_and_env() {
  local sub
  local app_env=""
  local found_list=""
  local expected_list
  local name
  local v
  local expected_count=0
  local found_count=0

  expected_list="$EXPECTED_DB_VOLUME $EXPECTED_CACHE_VOLUMES"
  for v in $expected_list; do
    expected_count=$((expected_count + 1))
  done

  for sub in $FORBIDDEN_PROJECT_SUBSTRINGS; do
    case "$EXPECTED_PROJECT_NAME" in
      *"$sub"*)
        die "expected project name '$EXPECTED_PROJECT_NAME' looks non-local (*$sub*)"
        ;;
    esac
  done

  if [[ -n "${COMPOSE_PROJECT_NAME:-}" && "${COMPOSE_PROJECT_NAME}" != "$EXPECTED_PROJECT_NAME" ]]; then
    die "COMPOSE_PROJECT_NAME='${COMPOSE_PROJECT_NAME}' != fixed expected '$EXPECTED_PROJECT_NAME' (refuse wrong env)"
  fi

  if [[ -f "$ENV_FILE" ]]; then
    app_env="$(grep -E '^[[:space:]]*APP_ENV=' "$ENV_FILE" | tail -1 | cut -d= -f2- | tr -d '[:space:]"'"'" || true)"
    case "$app_env" in
      ""|development|local|dev|test) ;;
      *)
        die "APP_ENV='$app_env' is not a local development value; refuse reset"
        ;;
    esac
  fi

  while IFS= read -r name; do
    [[ -z "$name" ]] && continue
    found_list="$found_list $name"
    found_count=$((found_count + 1))
  done < <(extract_compose_volume_names "$COMPOSE_FILE")
  found_list="${found_list# }"

  if [[ "$found_count" -ne "$expected_count" ]]; then
    die "compose volume name count $found_count != expected $expected_count (found: $found_list)"
  fi

  for v in $expected_list; do
    if ! list_has "$found_list" "$v"; then
      die "compose reality missing expected volume: $v"
    fi
  done

  info "live project/env guard OK"
}

# ── Snapshot (must succeed before any volume delete) ────────────────────────
create_recovery_snapshot() {
  local ts dir dump_gz digest_file manifest_file
  local ddl_list seed_list
  local digest
  local i
  local vol_exists=0

  umask 077

  if ! "$DOCKER_BIN" volume inspect "$EXPECTED_DB_VOLUME" >/dev/null 2>&1; then
    info "volume $EXPECTED_DB_VOLUME absent — skipping snapshot (nothing to recover)"
    # Print empty marker so caller can capture; still success.
    printf '%s\n' ""
    return 0
  fi
  vol_exists=1

  ts="$(date -u +%Y%m%dT%H%M%SZ)"
  dir="${BACKUP_BASE}/${ts}"
  mkdir -p "$dir" || die "cannot create backup dir $dir (disk?)"
  chmod 700 "$dir" || die "chmod 700 failed on $dir"

  dump_gz="${dir}/pg_dumpall.sql.gz"
  digest_file="${dir}/pg_dumpall.sql.gz.sha256"
  manifest_file="${dir}/manifest.txt"

  info "ensuring db is up for pg_dumpall..."
  compose up -d db >/dev/null
  i=0
  while [[ "$i" -lt 10 ]]; do
    if compose exec -T db sh -c 'pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"' >/dev/null 2>&1; then
      break
    fi
    i=$((i + 1))
    sleep 1
  done

  info "writing gzipped pg_dumpall → $dump_gz"
  if ! compose exec -T db sh -c 'pg_dumpall -U "$POSTGRES_USER"' 2>/dev/null | gzip -c >"$dump_gz"; then
    rm -f "$dump_gz"
    die "pg_dumpall | gzip failed — refusing volume delete"
  fi

  if [[ ! -s "$dump_gz" ]]; then
    die "snapshot dump is empty — refusing volume delete"
  fi

  # gzip(1) of a zero-byte stream still writes a small header; treat uncompressed
  # empty payload as failure so "empty dump" cannot slip through.
  local uncomp_bytes
  uncomp_bytes="$(gzip -dc "$dump_gz" 2>/dev/null | wc -c | tr -d ' ')"
  if [[ -z "$uncomp_bytes" || "$uncomp_bytes" -eq 0 ]]; then
    die "snapshot dump uncompressed payload is empty — refusing volume delete"
  fi

  digest="$(file_sha256 "$dump_gz")"
  if [[ -z "$digest" ]]; then
    die "empty SHA-256 digest — refusing volume delete"
  fi
  printf '%s  %s\n' "$digest" "pg_dumpall.sql.gz" >"$digest_file" || die "cannot write digest file"
  [[ -s "$digest_file" ]] || die "digest file empty — refusing volume delete"

  ddl_list="$(list_root_ddl_basenames "$MIGRATIONS_DIR" | tr '\n' ' ')"
  seed_list="$EXPECTED_SEED_BUNDLES"

  cat >"$manifest_file" <<EOF
# local DB reset recovery manifest (owner-only)
created_utc=${ts}
expected_project=${EXPECTED_PROJECT_NAME}
target_volume=${EXPECTED_DB_VOLUME}
kept_cache_volumes=${EXPECTED_CACHE_VOLUMES}
root_ddl_keys=${ddl_list}
seed_bundle_keys=${seed_list}
schema_migrations_seed_keys=seeds/002_master
dump_file=pg_dumpall.sql.gz
dump_sha256=${digest}
volume_existed_before_reset=${vol_exists}
EOF
  [[ -s "$manifest_file" ]] || die "manifest empty — refusing volume delete"
  chmod 600 "$dump_gz" "$digest_file" "$manifest_file" 2>/dev/null || true

  info "snapshot OK under $dir (sha256=${digest})"
  printf '%s\n' "$dir"
}

# ── Destructive phase: stop + DB volume only ────────────────────────────────
stop_services_keep_volumes() {
  info "stopping services (volumes retained)..."
  # Explicitly keep named volumes (do not pass compose volume-removal flags).
  compose down --remove-orphans
}

delete_db_volume_only() {
  local v
  info "deleting ONLY volume: $EXPECTED_DB_VOLUME"
  if "$DOCKER_BIN" volume inspect "$EXPECTED_DB_VOLUME" >/dev/null 2>&1; then
    "$DOCKER_BIN" volume rm "$EXPECTED_DB_VOLUME" || die "volume rm $EXPECTED_DB_VOLUME failed"
  else
    info "volume $EXPECTED_DB_VOLUME already absent"
  fi

  for v in $EXPECTED_CACHE_VOLUMES; do
    if "$DOCKER_BIN" volume inspect "$v" >/dev/null 2>&1; then
      info "kept cache volume: $v"
    else
      info "cache volume absent (ok if never created): $v"
    fi
  done
}

# ── Restart + postflight ────────────────────────────────────────────────────
restart_and_wait() {
  info "restarting db backend frontend with --wait..."
  compose up -d --build --wait --wait-timeout 1200 db backend frontend
}

# Run a SQL scalar query inside the db container; print trimmed stdout.
psql_scalar() {
  local sql="$1"
  compose exec -T db sh -c "psql -U \"\$POSTGRES_USER\" -d \"\$POSTGRES_DB\" -tAc $(printf '%q' "$sql")" 2>/dev/null | tr -d '[:space:]'
}

postflight_checks() {
  local errors=0
  local logs
  local ddl
  local key
  local missing=0
  local count_expected=0
  local count_recorded=""
  local health_code
  local ps_out
  local got

  info "postflight: backend logs for migration key coverage..."
  logs="$(compose logs backend 2>&1 || true)"
  if ! printf '%s\n' "$logs" | grep -E 'Migration key coverage' | tail -1 | grep -E 'missing=0' >/dev/null 2>&1; then
    echo "FAIL  postflight: Migration key coverage missing=0 not found in backend logs"
    errors=$((errors + 1))
  else
    info "postflight: migration key coverage missing=0"
  fi

  info "postflight: schema_migrations keys..."
  while IFS= read -r ddl; do
    [[ -z "$ddl" ]] && continue
    count_expected=$((count_expected + 1))
    got="$(psql_scalar "SELECT 1 FROM schema_migrations WHERE filename = '${ddl}' LIMIT 1" || true)"
    if [[ "$got" != "1" ]]; then
      echo "FAIL  postflight: schema_migrations missing DDL key: $ddl"
      missing=$((missing + 1))
      errors=$((errors + 1))
    fi
  done < <(list_root_ddl_basenames "$MIGRATIONS_DIR")

  for key in $EXPECTED_SEED_BUNDLES; do
    count_expected=$((count_expected + 1))
    got="$(psql_scalar "SELECT 1 FROM schema_migrations WHERE filename = 'seeds/${key}' LIMIT 1" || true)"
    if [[ "$got" != "1" ]]; then
      echo "FAIL  postflight: schema_migrations missing seed key: seeds/$key"
      missing=$((missing + 1))
      errors=$((errors + 1))
    fi
  done

  count_recorded="$(psql_scalar "SELECT COUNT(*) FROM schema_migrations" || true)"
  if [[ -z "$count_recorded" ]]; then
    echo "FAIL  postflight: could not read schema_migrations count"
    errors=$((errors + 1))
  else
    info "postflight: schema_migrations recorded=$count_recorded expected_min=$count_expected missing_keys=$missing"
  fi

  ps_out="$(compose ps backend 2>&1 || true)"
  if ! printf '%s\n' "$ps_out" | grep -Ei 'healthy|running' >/dev/null 2>&1; then
    echo "FAIL  postflight: backend not healthy/running"
    printf '%s\n' "$ps_out"
    errors=$((errors + 1))
  else
    info "postflight: backend healthy/running"
  fi

  health_code="$("$CURL_BIN" -s -o /dev/null -w '%{http_code}' "$HEALTH_URL" 2>/dev/null || echo "000")"
  if [[ "$health_code" != "200" ]]; then
    echo "FAIL  postflight: $HEALTH_URL returned HTTP $health_code (want 200)"
    errors=$((errors + 1))
  else
    info "postflight: /health HTTP 200"
  fi

  if [[ "$errors" -gt 0 ]]; then
    die "postflight failed ($errors error(s)) — DB was rebuilt; inspect logs and .local-db-backups/"
  fi
  info "postflight OK"
}

# ── Main ────────────────────────────────────────────────────────────────────
main() {
  info "local DB reset contract (mode=$MODE root=$ROOT)"
  validate_static_contract

  if [[ "$MODE" == "contract-only" ]]; then
    info "contract-only mode: static checks passed"
    exit 0
  fi

  validate_live_project_and_env

  # Snapshot first — failure must not reach volume delete.
  # Capture only the last line (snapshot dir); stream logs to stderr via info.
  local snap_dir=""
  local snap_out
  if ! snap_out="$(create_recovery_snapshot)"; then
    die "snapshot phase failed — volume NOT deleted"
  fi
  snap_dir="$(printf '%s\n' "$snap_out" | tail -1)"

  stop_services_keep_volumes
  delete_db_volume_only
  restart_and_wait
  postflight_checks

  # Local-only: if hospital CSV bundles are staged under
  # backend/migrations/seeds/_old_db_handoff/<clinic>/, import them now.
  if [[ -x "$SCRIPT_DIR/import-old-db-handoffs-on-reset.sh" ]]; then
    info "importing staged old_db handoffs (local rehearsal allowed)..."
    bash "$SCRIPT_DIR/import-old-db-handoffs-on-reset.sh" \
      || die "old_db handoff import failed after reset"
  fi

  info "reset complete (snapshot=${snap_dir:-none})"
  echo "✓ Local DB reset complete — only ${EXPECTED_DB_VOLUME} was removed; caches kept; postflight OK"
}

main "$@"
