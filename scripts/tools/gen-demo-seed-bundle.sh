#!/usr/bin/env bash
# Generate the 004_demo CSV seed bundle (backend/migrations/seeds/004_demo/)
# from the validated API demo seeder (scripts/tools/seed-demo-data.py).
#
# Isolation: everything runs inside throwaway docker containers on a private
# docker network (postgres:18-alpine + golang:1.26-alpine). The compose stack,
# local dev volumes, STG, and production are never touched. Containers and the
# network are removed on exit; only the committed CSVs remain.
#
# Usage:
#   bash scripts/tools/gen-demo-seed-bundle.sh
#
# Requires: docker, python3, curl.
#
# Pipeline:
#   1. start disposable postgres + create demo_seed_tmp
#   2. cmd/migrate (001_init + 002_master + 003_login) via a golang container
#   3. start cmd/api against the disposable DB
#   4. snapshot exact per-table row counts
#   5. seed-demo-data.py --apply (creates the 【デモ】 dataset for clinic 1)
#   6. diff the snapshots to find the demo-owned tables, topologically sort
#      them by FK edges, and COPY-dump each one (same format as
#      cmd/seed-export: PK-ordered CSV + header), then write manifest.json
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
NET="emr443-demoseed-net"
PG_CONTAINER="emr443-demoseed-pg"
API_CONTAINER="emr443-demoseed-api"
DB_NAME="demo_seed_tmp"
PG_USER="postgres"
PG_PASSWORD="demoseed-local-only"
DEMO_LOGIN_EMAIL="stg-staff-10000021@example.test"
DEMO_LOGIN_PASSWORD="demoseed-local-only"
API_HOST_PORT="18085"
OUT_DIR="$ROOT/backend/migrations/seeds/004_demo"
STATE_FILE="$ROOT/sensitive-local/demo-seed/EMR443-DEMO/state.json"
STATE_BACKUP="$STATE_FILE.pre-demogen"

die() { printf 'BLOCKED: %s\n' "$*" >&2; exit 1; }

command -v docker >/dev/null || die "docker is required"
command -v python3 >/dev/null || die "python3 is required"
command -v curl >/dev/null || die "curl is required"
[[ -f "$ROOT/backend/migrations/001_init.sql" ]] || die "001_init.sql not found"

cleanup() {
  docker rm -f "$PG_CONTAINER" "$API_CONTAINER" >/dev/null 2>&1 || true
  docker network rm "$NET" >/dev/null 2>&1 || true
  docker volume rm emr443-demoseed-gomod emr443-demoseed-gobuild >/dev/null 2>&1 || true
  if [[ -f "$STATE_BACKUP" ]]; then
    mv "$STATE_BACKUP" "$STATE_FILE"
  fi
}
trap cleanup EXIT

echo "=== 1/6 starting disposable postgres ($PG_CONTAINER) ==="
docker network create "$NET" >/dev/null
docker run -d --name "$PG_CONTAINER" --network "$NET" \
  -e POSTGRES_USER="$PG_USER" -e POSTGRES_PASSWORD="$PG_PASSWORD" -e POSTGRES_DB="$DB_NAME" \
  postgres:18-alpine >/dev/null

for _ in $(seq 1 60); do
  if docker exec "$PG_CONTAINER" pg_isready -U "$PG_USER" -d "$DB_NAME" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
docker exec "$PG_CONTAINER" pg_isready -U "$PG_USER" -d "$DB_NAME" >/dev/null

echo "=== 2/6 running cmd/migrate (001_init + 002_master + 003_login) ==="
# Note: the login seed runs with a throwaway password so the demo staffs
# exist as FK targets and the seeder can log in. None of these rows are
# dumped — accounts/staffs stay owned by the runtime login seed.
docker run --rm --network "$NET" \
  -v "$ROOT/backend:/app" -w /app \
  -v emr443-demoseed-gomod:/go/pkg/mod -v emr443-demoseed-gobuild:/root/.cache/go-build \
  -e DB_HOST="$PG_CONTAINER" -e DB_PORT=5432 \
  -e DB_USER="$PG_USER" -e DB_PASSWORD="$PG_PASSWORD" -e DB_NAME="$DB_NAME" \
  -e DB_SSL_MODE=disable \
  -e APP_ENV=staging -e SEEDLOGIN_DEMO_PASSWORD="$DEMO_LOGIN_PASSWORD" \
  golang:1.26-alpine sh -c "go run ./cmd/migrate 2>&1" | tail -4

echo "=== 3/6 starting disposable cmd/api ==="
docker run -d --name "$API_CONTAINER" --network "$NET" \
  -v "$ROOT/backend:/app" -w /app \
  -v emr443-demoseed-gomod:/go/pkg/mod -v emr443-demoseed-gobuild:/root/.cache/go-build \
  -e DB_HOST="$PG_CONTAINER" -e DB_PORT=5432 \
  -e DB_USER="$PG_USER" -e DB_PASSWORD="$PG_PASSWORD" -e DB_NAME="$DB_NAME" \
  -e DB_SSL_MODE=disable \
  -e APP_ENV=local -e JWT_SECRET="$PG_PASSWORD" \
  -p "127.0.0.1:${API_HOST_PORT}:8080" \
  golang:1.26-alpine sh -c "go run ./cmd/api" >/dev/null

api_ready=0
for _ in $(seq 1 120); do
  if curl -fsS "http://127.0.0.1:${API_HOST_PORT}/api/v1/health" >/dev/null 2>&1; then
    api_ready=1
    break
  fi
  sleep 2
done
[[ "$api_ready" == "1" ]] || { docker logs "$API_CONTAINER" 2>&1 | tail -20; die "API did not become healthy"; }
echo "API healthy: http://127.0.0.1:${API_HOST_PORT}"

echo "=== 4/6 seeding demo data and dumping CSVs ==="
# fresh state: the seeder must create everything from scratch in this DB
if [[ -f "$STATE_FILE" ]]; then
  mv "$STATE_FILE" "$STATE_BACKUP"
fi
PG_CONTAINER="$PG_CONTAINER" PG_USER="$PG_USER" DB_NAME="$DB_NAME" \
API_HOST_PORT="$API_HOST_PORT" OUT_DIR="$OUT_DIR" \
DEMO_LOGIN_EMAIL="$DEMO_LOGIN_EMAIL" DEMO_LOGIN_PASSWORD="$DEMO_LOGIN_PASSWORD" \
ROOT_DIR="$ROOT" \
python3 - <<'PY'
import json
import os
import subprocess
import sys
import time
from pathlib import Path

root = Path(os.environ["ROOT_DIR"])
out_dir = Path(os.environ["OUT_DIR"])
pg_container = os.environ["PG_CONTAINER"]
pg_user = os.environ["PG_USER"]
db_name = os.environ["DB_NAME"]

# tables already owned by 002_master / the login seed / migrate machinery
EXCLUDED = {
    # 002_master manifest ownership
    "companies", "animal_species", "lstep_auto_managed_prefixes",
    "lstep_condition_tag_mappings", "lstep_send_purpose_tag_prefixes",
    "clinics", "clinic_settings", "exam_types", "reservation_types",
    "payment_methods", "permission_groups", "permission_group_rules",
    # login seed (phase 3) ownership
    "accounts", "staffs", "staff_clinic_assignments", "staff_permission_groups",
    # migrate machinery
    "schema_migrations",
    # trigger-owned per-clinic defaults
    "trimming_course_types",
}


def run_psql(sql: str, stdout_to: str | None = None) -> str:
    handle = None
    if stdout_to is not None:
        handle = open(stdout_to, "wb")
    try:
        proc = subprocess.run(
            ["docker", "exec", pg_container, "psql", "-X", "-q", "-At",
             "-v", "ON_ERROR_STOP=1", "-U", pg_user, "-d", db_name, "-c", sql],
            check=True, stdout=handle or subprocess.PIPE, text=handle is None,
        )
        return proc.stdout or ""
    finally:
        if handle is not None:
            handle.close()


def all_tables() -> list[str]:
    out = run_psql(
        "SELECT c.relname FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace "
        "WHERE n.nspname = 'public' AND c.relkind = 'r' AND c.relname <> 'schema_migrations' "
        "ORDER BY c.relname"
    )
    return [line.strip() for line in out.splitlines() if line.strip()]


def snapshot_counts() -> dict[str, int]:
    tables = all_tables()
    if not tables:
        raise SystemExit("BLOCKED: no tables found in the disposable DB")
    union = " UNION ALL ".join(
        f"SELECT '{t}' AS name, count(*) AS n FROM \"{t}\"" for t in tables
    )
    out = run_psql(f"{union} ORDER BY name")
    counts: dict[str, int] = {}
    for line in out.splitlines():
        if not line.strip():
            continue
        name, _, value = line.partition("|")
        counts[name] = int(value)
    return counts


before = snapshot_counts()

env = dict(os.environ)
env["DEMO_SEED_PASSWORD"] = env["DEMO_LOGIN_PASSWORD"]
print("--- running seed-demo-data.py --apply ---")
subprocess.run(
    [
        "python3", str(root / "scripts/tools/seed-demo-data.py"),
        "--base-url", f"http://127.0.0.1:{os.environ['API_HOST_PORT']}",
        "--clinic-id", "1",
        "--email", env["DEMO_LOGIN_EMAIL"],
        "--apply",
    ],
    check=True, env=env,
)

after = snapshot_counts()

grown = sorted(t for t, n in after.items() if n > before.get(t, 0) and t not in EXCLUDED)
if not grown:
    raise SystemExit("BLOCKED: no table grew after the seeder run")

in_set = ", ".join(f"'{t}'" for t in grown)
edges_sql = (
    "SELECT conrelid::regclass::text || '|' || confrelid::regclass::text "
    "FROM pg_constraint WHERE contype = 'f' "
    f"AND conrelid::regclass::text IN ({in_set}) "
    f"AND confrelid::regclass::text IN ({in_set})"
)
edges = [line for line in run_psql(edges_sql).splitlines() if line]

indegree = {t: 0 for t in grown}
children: dict[str, list[str]] = {t: [] for t in grown}
for edge in edges:
    child, _, parent = edge.partition("|")
    if child == parent:
        continue
    indegree[child] += 1
    children[parent].append(child)

order: list[str] = []
ready = sorted(t for t, d in indegree.items() if d == 0)
while ready:
    table = ready.pop(0)
    order.append(table)
    for child in sorted(children[table]):
        indegree[child] -= 1
        if indegree[child] == 0:
            ready.append(child)
if len(order) != len(grown):
    raise SystemExit(f"BLOCKED: FK cycle among demo tables: {sorted(set(grown) - set(order))}")

out_dir.mkdir(parents=True, exist_ok=True)
manifest_tables = []
for table in order:
    pk = run_psql(
        "SELECT string_agg(a.attname, ',' ORDER BY array_position(i.indkey, a.attnum)) "
        "FROM pg_index i JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey) "
        f"WHERE i.indrelid = '{table}'::regclass AND i.indisprimary"
    )
    if pk:
        ordered = ", ".join(f'"{c}"' for c in pk.split(","))
        copy_sql = f'COPY (SELECT * FROM "{table}" ORDER BY {ordered}) TO STDOUT WITH (FORMAT csv, HEADER true)'
    else:
        copy_sql = f'COPY "{table}" TO STDOUT WITH (FORMAT csv, HEADER true)'
    csv_path = out_dir / f"{table}.csv"
    run_psql(copy_sql, stdout_to=str(csv_path))
    print(f"dumped {table}: {after[table]} rows")
    manifest_tables.append({"table": table, "csvFile": f"{table}.csv"})

manifest = {"bundle": "004_demo", "tables": manifest_tables}
(out_dir / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
print(f"manifest written: {out_dir / 'manifest.json'} ({len(manifest_tables)} tables)")
PY

docker stop "$API_CONTAINER" >/dev/null 2>&1 || true

echo "=== 6/6 cleanup (containers removed by trap) ==="
echo "DONE: 004_demo generated at backend/migrations/seeds/004_demo/"
