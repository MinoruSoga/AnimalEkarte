#!/usr/bin/env bash
# Link old_db imported staffs via the authoritative cross-clinic identity map
# (old_db sensitive-local/staff-identity-map artifact).
#
#   scripts/staff-identity-map-link.sh <local|stg> [--dry-run]
#
# Env:
#   OLD_DB_STAFF_IDENTITY_MAP_CSV   path to staff-identity-map.csv (required;
#                                   owner-only file, verified against
#                                   SHA256SUMS in the same directory)
# local: docker compose -p animalekarte exec -T db psql (dev/reset DB)
# stg:   sources scripts/stg-uat-old-db-handoff.local.env (0600, non-symlink),
#        psql direct with verify-full + app.bypass_rls=on
# --dry-run runs the full statement set inside BEGIN...ROLLBACK and only
# prints the summary counts.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MODE="${1:-}"
DRY_RUN=0
[ "${2:-}" = "--dry-run" ] && DRY_RUN=1

if [[ "$MODE" != "local" && "$MODE" != "stg" ]]; then
  echo "usage: $0 <local|stg> [--dry-run]" >&2
  exit 2
fi

MAP_CSV="${OLD_DB_STAFF_IDENTITY_MAP_CSV:-}"
if [[ -z "$MAP_CSV" || ! -f "$MAP_CSV" || -L "$MAP_CSV" ]]; then
  echo "ERROR: OLD_DB_STAFF_IDENTITY_MAP_CSV must point to a regular map CSV file" >&2
  exit 1
fi
perm="$(stat -f '%Lp' "$MAP_CSV" 2>/dev/null || stat -c '%a' "$MAP_CSV")"
if [[ "$perm" != "600" && "$perm" != "400" ]]; then
  echo "ERROR: map file must be owner-only (0600/0400), got $perm" >&2
  exit 1
fi

ARTIFACT_DIR="$(dirname "$MAP_CSV")"
if [[ ! -f "$ARTIFACT_DIR/SHA256SUMS" ]]; then
  echo "ERROR: SHA256SUMS missing next to the map CSV" >&2
  exit 1
fi
expected="$(grep -E ' staff-identity-map\.csv$' "$ARTIFACT_DIR/SHA256SUMS" | awk '{print $1}' | head -1)"
actual="$(shasum -a 256 "$MAP_CSV" | awk '{print $1}')"
if [[ -z "$expected" || "$expected" != "$actual" ]]; then
  echo "ERROR: map CSV sha256 does not match SHA256SUMS" >&2
  exit 1
fi
echo "INFO  map CSV sha256 verified against SHA256SUMS"

if [[ -f "$ARTIFACT_DIR/manifest.json" ]]; then
  python3 - "$ARTIFACT_DIR/manifest.json" <<'PY'
import json, sys
m = json.load(open(sys.argv[1]))
print(f"INFO  manifest status={m.get('status')} productionEligible={m.get('productionEligible')} "
      f"schemaVersion={m.get('schemaVersion')}")
if m.get("schemaVersion") != "staff-identity-map-manifest-v1":
    sys.exit("ERROR: unexpected manifest schemaVersion")
PY
fi

build_stream() {
  printf 'BEGIN;\n'
  printf 'CREATE TEMP TABLE tmp_staff_identity_map (\n'
  printf '  person_group_id text NOT NULL, classification text NOT NULL,\n'
  printf '  basis text NOT NULL, clinic_code text NOT NULL,\n'
  printf '  legacy_staff_no text NOT NULL, csv_staff_id bigint NOT NULL\n'
  printf ') ON COMMIT DROP;\n'
  printf 'COPY tmp_staff_identity_map FROM STDIN WITH (FORMAT csv, HEADER true);\n'
  cat "$MAP_CSV"
  printf '\\.\n'
  cat "$ROOT/scripts/sql/link-old-db-staff-identity-map.sql"
  if [[ "$DRY_RUN" = "1" ]]; then
    printf 'ROLLBACK;\n'
  else
    printf 'COMMIT;\n'
  fi
}

case "$MODE" in
  local)
    DB_NAME_VAL="$(grep -E '^DB_NAME=' "$ROOT/.env.local" 2>/dev/null | head -1 | cut -d= -f2- | tr -d '"' || true)"
    DB_NAME_VAL="${DB_NAME_VAL:-${DB_NAME:-ekarte_db}}"
    echo "INFO  applying identity map link (local db: $DB_NAME_VAL, dry_run=$DRY_RUN)"
    build_stream | docker compose -p animalekarte -f "$ROOT/docker-compose.yml" \
      exec -T db psql -U ekarte_user -d "$DB_NAME_VAL" -v ON_ERROR_STOP=1
    ;;
  stg)
    ENV_FILE="${STG_UAT_HANDOFF_ENV_FILE:-$ROOT/scripts/stg-uat-old-db-handoff.local.env}"
    if [[ ! -f "$ENV_FILE" || -L "$ENV_FILE" ]]; then
      echo "ERROR: STG env file missing or symlink: $ENV_FILE" >&2
      exit 1
    fi
    env_perm="$(stat -f '%Lp' "$ENV_FILE" 2>/dev/null || stat -c '%a' "$ENV_FILE")"
    if [[ "$env_perm" != "600" ]]; then
      echo "ERROR: STG env file must be 0600" >&2
      exit 1
    fi
    # shellcheck disable=SC1090
    source "$ENV_FILE"
    : "${DB_HOST:?}" "${DB_USER:?}" "${DB_PASSWORD:?}"
    export PGHOST="$DB_HOST" PGUSER="$DB_USER" PGPASSWORD="$DB_PASSWORD"
    export PGDATABASE="${TARGET_DB_NAME:-postgres}" PGSSLMODE=verify-full PGSSLROOTCERT=system
    export PGOPTIONS="-c app.bypass_rls=on"
    echo "INFO  applying identity map link (stg db: $PGDATABASE, dry_run=$DRY_RUN)"
    build_stream | psql -v ON_ERROR_STOP=1
    ;;
esac

if [[ "$DRY_RUN" = "1" ]]; then
  echo "INFO  dry-run rolled back; no data changed"
else
  echo "INFO  identity map link committed"
fi
