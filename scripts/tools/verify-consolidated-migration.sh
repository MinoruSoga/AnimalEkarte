#!/usr/bin/env bash
# Verify the consolidated 001_init.sql against a disposable local database.
#
# USER-RUN ONLY. This script creates and drops a scratch database, applies the
# consolidated baseline through cmd/migrate, and diffs its schema against the
# pre-consolidation capture in sensitive-local/stg-rebuild-emr-443/.
#
# It refuses to run unless:
#   - APP_ENV is unset/development/local
#   - CONFIRM_LOCAL_SCRATCH_DB=YES
#   - the db service is running in this Compose project
#
# Usage:
#   CONFIRM_LOCAL_SCRATCH_DB=YES bash scripts/tools/verify-consolidated-migration.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRATCH_DB="${SCRATCH_DB:-ekarte_baseline_check}"
BASELINE_DIR="${BASELINE_DIR:-$ROOT/sensitive-local/stg-rebuild-emr-443}"
BASELINE_SCHEMA="$BASELINE_DIR/db_a.normalized.sql"
scratch_created=0

die() {
  printf 'BLOCKED: %s\n' "$*" >&2
  exit 1
}

cleanup() {
  if [[ "$scratch_created" == "1" ]]; then
    docker compose exec -T db sh -c 'dropdb --if-exists -U "$POSTGRES_USER" "$1"' sh "$SCRATCH_DB" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

case "${APP_ENV:-development}" in
  development|local|"") ;;
  *) die "APP_ENV=${APP_ENV} is not local/development; refusing to create a scratch DB" ;;
esac
[[ "${CONFIRM_LOCAL_SCRATCH_DB:-}" == "YES" ]] || die "set CONFIRM_LOCAL_SCRATCH_DB=YES after reviewing the scratch DB name: $SCRATCH_DB"
[[ -f "$BASELINE_SCHEMA" ]] || die "missing pre-consolidation schema capture: $BASELINE_SCHEMA"

cd "$ROOT"
docker compose ps db >/dev/null 2>&1 || die "db service is not running in this Compose project"

printf 'Creating scratch DB %s\n' "$SCRATCH_DB"
docker compose exec -T db sh -c 'dropdb --if-exists -U "$POSTGRES_USER" "$1" && createdb -U "$POSTGRES_USER" "$1"' sh "$SCRATCH_DB" >/dev/null
scratch_created=1

printf 'Applying consolidated baseline through cmd/migrate\n'
docker compose exec -T -e DB_NAME="$SCRATCH_DB" -e DB_RESET=true backend go run ./cmd/migrate

printf 'Dumping scratch schema\n'
docker compose exec -T db sh -c 'pg_dump -U "$POSTGRES_USER" -d "$1" --schema-only --no-owner --exclude-table=public.schema_migrations' sh "$SCRATCH_DB" > "$BASELINE_DIR/db_b.raw.sql"

python3 - "$BASELINE_DIR/db_b.raw.sql" "$BASELINE_DIR/db_b.normalized.sql" <<'PY'
import re
import sys
from pathlib import Path

source = Path(sys.argv[1])
target = Path(sys.argv[2])
lines = [
    line
    for line in source.read_text().splitlines()
    if not line.startswith("\\")
    and not line.startswith("SET ")
    and not line.startswith("SELECT pg_catalog.set_config(")
]
text = "\n".join(lines) + "\n"
text = re.sub(
    r"CREATE EXTENSION IF NOT EXISTS (btree_gist|pg_trgm) WITH SCHEMA public;",
    r"CREATE EXTENSION IF NOT EXISTS \1;",
    text,
)
# `public` is created by initdb in the long-lived DB and by DB_RESET in the
# scratch DB. Their COMMENT/ACL defaults differ by environment; neither is an
# application schema object. Compare the application objects only.
text = re.sub(r"--\n-- Name: public; Type: SCHEMA;.*?\n\n\n", "", text, flags=re.S)
text = re.sub(r"--\n-- Name: SCHEMA public; Type: COMMENT;.*?\n\n\n", "", text, flags=re.S)
text = re.sub(r"--\n-- Name: SCHEMA public; Type: ACL;.*?\n\n\n", "", text, flags=re.S)
text = re.sub(r"COMMENT ON SCHEMA public IS '';\n", "", text)
text = re.sub(r"REVOKE USAGE ON SCHEMA public FROM PUBLIC;\n", "", text)
target.write_text(text)
PY

if ! diff -u "$BASELINE_SCHEMA" "$BASELINE_DIR/db_b.normalized.sql"; then
  die "schema diff is not empty; the consolidated baseline is not equivalent"
fi

printf 'PASS: consolidated schema matches the pre-consolidation capture\n'
