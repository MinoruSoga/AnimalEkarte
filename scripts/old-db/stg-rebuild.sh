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
ROLE_NAME="${ROLE_NAME:-emr443-rebuild-$(date -u +%Y%m%dT%H%M%SZ)}"
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
pscale role create animalekarte-stg main "$ROLE_NAME" \
  --org noah-animalekarte --inherited-roles postgres --ttl 4h -f json > "$ROLE_JSON"
chmod 600 "$ROLE_JSON"

CONN="$(python3 - "$ROLE_JSON" <<'PY'
import json
import sys
import urllib.parse
d = json.load(open(sys.argv[1]))
host = d["access_host_url"]
if "://" in host:
    host = urllib.parse.urlparse(host).hostname or ""
print("\t".join([host, d["username"], d["password"]]))
PY
)"
IFS=$'\t' read -r DB_HOST DB_USER DB_PASSWORD <<< "$CONN"
[[ -n "$DB_HOST" && -n "$DB_USER" && -n "$DB_PASSWORD" ]] || die "could not read role connection fields"

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
  "${EXTRA_ENV[@]}" \
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
