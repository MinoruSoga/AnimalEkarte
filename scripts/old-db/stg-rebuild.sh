#!/usr/bin/env bash
# Reset and rebuild the STG main schema through cmd/migrate.
#
# USER-RUN ONLY. This is destructive: it drops the STG public schema and
# reapplies 001_init.sql + 002_master + login seed. Run it only after the
# restore branch and preservation export are confirmed.
#
# Usage:
#   CONFIRM_STG_REBUILD=YES bash scripts/old-db/stg-rebuild.sh
#
# Optional:
#   ROLE_NAME=...                 short-lived PlanetScale role name
#   SEEDLOGIN_DEMO_PASSWORD=...   pass the STG demo-login secret to migrate
#
# The script creates a 4h role, runs the migrate, and always deletes the role
# with successor postgres. It never prints the password.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RUN_DIR="${RUN_DIR:-$ROOT/sensitive-local/stg-rebuild/$(date -u +%Y%m%dT%H%M%SZ)}"
ROLE_NAME="${ROLE_NAME:-emr443-rebuild-$(date -u +%Y%m%d%H%M%S)}"
ROLE_JSON="$RUN_DIR/pscale-role.json"
LOG="$RUN_DIR/migrate.log"

die() {
  printf 'BLOCKED: %s\n' "$*" >&2
  exit 1
}

[[ "${CONFIRM_STG_REBUILD:-}" == "YES" ]] || die "set CONFIRM_STG_REBUILD=YES after confirming the restore branch and export"
command -v pscale >/dev/null || die "pscale CLI is required"
command -v python3 >/dev/null || die "python3 is required"
[[ -f "$ROOT/scripts/old-db/stg-uat-old-db-handoff.local.env" ]] || die "STG env file is missing"

mkdir -p "$RUN_DIR"
chmod 700 "$RUN_DIR"

cleanup() {
  if [[ -s "$ROLE_JSON" ]]; then
    local role_id
    role_id="$(python3 - "$ROLE_JSON" <<'PY'
import json
import sys
d = json.load(open(sys.argv[1]))
print(d.get("id") or d.get("role_id") or "")
PY
)"
    if [[ -n "$role_id" ]]; then
      echo "Deleting short-lived role $role_id with successor postgres"
      pscale role delete animalekarte-stg main "$role_id" \
        --org noah-animalekarte --successor postgres --force || true
    fi
    rm -f "$ROLE_JSON"
  fi
}
trap cleanup EXIT

echo "Creating short-lived role $ROLE_NAME"
CREATE_ERR="$RUN_DIR/pscale-create.err"
if ! pscale role create animalekarte-stg main "$ROLE_NAME" \
  --org noah-animalekarte --inherited-roles postgres --ttl 4h --no-color -f json \
  2> "$CREATE_ERR" | sed -n '/^[[:space:]]*{/,$p' > "$ROLE_JSON"; then
  echo "BLOCKED: pscale role create failed" >&2
  [[ -s "$CREATE_ERR" ]] && sed -n '1,20p' "$CREATE_ERR" >&2
  exit 1
fi
chmod 600 "$ROLE_JSON"
if [[ ! -s "$ROLE_JSON" ]]; then
  echo "BLOCKED: pscale role create returned no JSON" >&2
  [[ -s "$CREATE_ERR" ]] && sed -n '1,20p' "$CREATE_ERR" >&2
  exit 1
fi

CONN="$(python3 - "$ROLE_JSON" <<'PY'
import json
import sys
import urllib.parse

try:
    d = json.load(open(sys.argv[1]))
    host = d["access_host_url"]
    if "://" in host:
        host = urllib.parse.urlparse(host).hostname or ""
    values = [host, d["username"], d["password"]]
    if not all(values):
        raise ValueError("empty connection field")
    print("\t".join(values))
except Exception as exc:
    print(f"BLOCKED: role JSON parse failed: {exc}", file=sys.stderr)
    raise SystemExit(1)
PY
)" || die "could not parse role JSON"
IFS=$'\t' read -r DB_HOST DB_USER DB_PASSWORD <<< "$CONN"
[[ -n "$DB_HOST" && -n "$DB_USER" && -n "$DB_PASSWORD" ]] || die "could not read role connection fields"

# DROP SCHEMA public CASCADE in one transaction exceeds the shared-memory lock
# budget on the STG branch (138 tables + RLS/functions). Drop objects statement
# by statement first; DB_RESET=true then only recreates the empty schema.
echo "Batched pre-drop of public objects"
psql_major="$(psql --version | grep -oE '[0-9]+' | head -1)"
ssl_root="${STG_SSL_ROOT_CERT:-system}"
if [[ "$ssl_root" == "system" && "${psql_major:-0}" -lt 17 ]]; then
  # libpq <17 lacks sslrootcert=system; use the OS CA bundle instead
  ssl_root="/etc/ssl/cert.pem"
fi
PGHOST="$DB_HOST" PGPORT=5432 PGUSER="$DB_USER" PGPASSWORD="$DB_PASSWORD" \
PGDATABASE=postgres PGSSLMODE=verify-full PGSSLROOTCERT="$ssl_root" \
PGOPTIONS='-c app.bypass_rls=on' \
psql -X -q -v ON_ERROR_STOP=1 -f - <<'SQL'
SELECT format('DROP TABLE IF EXISTS public.%I CASCADE', tablename)
  FROM pg_tables WHERE schemaname = 'public' \gexec
SELECT format('DROP MATERIALIZED VIEW IF EXISTS public.%I CASCADE', matviewname)
  FROM pg_matviews WHERE schemaname = 'public' \gexec
SELECT format('DROP VIEW IF EXISTS public.%I CASCADE', viewname)
  FROM pg_views WHERE schemaname = 'public' \gexec
SELECT format('DROP SEQUENCE IF EXISTS public.%I CASCADE', sequencename)
  FROM pg_sequences WHERE schemaname = 'public' \gexec
SELECT format('DROP EXTENSION IF EXISTS %I CASCADE', extname)
  FROM pg_extension WHERE extname IN ('pg_trgm', 'btree_gist') \gexec
SELECT format('DROP TYPE IF EXISTS public.%I CASCADE', typname)
  FROM pg_type WHERE typtype = 'e' AND typnamespace = 'public'::regnamespace \gexec
SQL

EXTRA_ENV=()
if [[ -n "${SEEDLOGIN_DEMO_PASSWORD:-}" ]]; then
  EXTRA_ENV+=(-e "SEEDLOGIN_DEMO_PASSWORD=$SEEDLOGIN_DEMO_PASSWORD")
fi

echo "Running DB_RESET=true migrate against ${DB_HOST}/${DB_USER}"
docker compose exec -T \
  -e DB_HOST="$DB_HOST" -e DB_PORT=5432 \
  -e DB_USER="$DB_USER" -e DB_PASSWORD="$DB_PASSWORD" \
  -e DB_NAME=postgres -e DB_SSL_MODE=verify-full -e DB_SSL_ROOT_CERT=system \
  -e DB_RESET=true -e APP_ENV=staging \
  "${EXTRA_ENV[@]+"${EXTRA_ENV[@]}"}" \
  backend go run ./cmd/migrate 2>&1 | tee "$LOG"

missing=()
for marker in "Schema reset completed" "Seed bundle loaded bundle=002_master" "Migration key coverage"; do
  if ! grep -Fq "$marker" "$LOG"; then
    missing+=("$marker")
  fi
done
if grep -Fq "Migration key coverage" "$LOG" && ! grep -Eq "Migration key coverage.*missing=0" "$LOG"; then
  missing+=("Migration key coverage missing=0")
fi
if ! grep -Fq "Login seed applied" "$LOG"; then
  if [[ -z "${SEEDLOGIN_DEMO_PASSWORD:-}" ]]; then
    echo "WARN: Login seed applied was not observed and SEEDLOGIN_DEMO_PASSWORD was not passed; demo login may be locked." >&2
  else
    missing+=("Login seed applied")
  fi
fi
if ((${#missing[@]} > 0)); then
  printf 'FAIL: migrate did not show required markers: %s\n' "${missing[*]}" >&2
  printf 'Log: %s\n' "$LOG" >&2
  exit 1
fi

echo "PASS: STG schema reset and migrate completed"
echo "Log: $LOG"
