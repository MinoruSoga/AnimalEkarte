#!/usr/bin/env bash
# Verify a staged old_db 21-table CSV bundle in the local PHI quarantine:
#   backend/migrations/seeds/_old_db_handoff/<clinic>/
#
# Layout only: git-ignore, 21 CSV files, manifest clinicCode/run/table count.
# Does not import, preflight payment graphs, or talk to PlanetScale / shared STG.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

die() { echo "old-db-handoff-check: $*" >&2; exit 1; }

CLINIC_CODE="${CLINIC_CODE:-}"
MIGRATION_RUN_ID="${MIGRATION_RUN_ID:-}"
[[ -n "$CLINIC_CODE" ]] || die "CLINIC_CODE is required"
[[ -n "$MIGRATION_RUN_ID" ]] || die "MIGRATION_RUN_ID is required"
[[ "$CLINIC_CODE" =~ ^[a-z][a-z0-9-]{0,31}$ ]] || die "CLINIC_CODE must be a lowercase slug"
[[ "$MIGRATION_RUN_ID" =~ ^[A-Za-z0-9._-]{1,64}$ ]] || die "MIGRATION_RUN_ID is unsafe"

DEST="$ROOT/backend/migrations/seeds/_old_db_handoff/$CLINIC_CODE"
ACCOUNT_PARENT="$ROOT/backend/migrations/seeds/002_master/accounts/_old_db_handoff"
ACCOUNT_DEST="$ACCOUNT_PARENT/$CLINIC_CODE"

git check-ignore -q --no-index backend/migrations/seeds/_old_db_handoff/ \
  || die "backend/migrations/seeds/_old_db_handoff/ is not git-ignored"
[[ -f "$DEST/manifest.json" ]] || die "missing staged manifest for $CLINIC_CODE"
git check-ignore -q --no-index "$ACCOUNT_PARENT/" || die "central accounts are not git-ignored"

csv_count="$(find "$DEST" -maxdepth 1 -type f -name '*.csv' | wc -l | tr -d '[:space:]')"
[[ "$csv_count" == "20" ]] || die "expected 20 clinical CSV files for $CLINIC_CODE, got $csv_count"
[[ ! -e "$DEST/accounts" && ! -L "$DEST/accounts" && ! -e "$DEST/staffs.csv" && ! -L "$DEST/staffs.csv" ]] || die "duplicate local account layout"

python3 - "$DEST/manifest.json" "$CLINIC_CODE" "$MIGRATION_RUN_ID" "$ACCOUNT_DEST" <<'PY'
import json, os, stat, sys
path, clinic, run_id = sys.argv[1], sys.argv[2], sys.argv[3]
account_dir = sys.argv[4]
for account_path, is_dir in [(os.path.dirname(account_dir), True), (account_dir, True),
                             (os.path.join(account_dir, "staffs.csv"), False)]:
    metadata = os.lstat(account_path)
    kind_ok = stat.S_ISDIR(metadata.st_mode) if is_dir else stat.S_ISREG(metadata.st_mode)
    required = 0o500 if is_dir else 0o400
    mode = stat.S_IMODE(metadata.st_mode)
    if not kind_ok or mode & 0o077 or mode & required != required or metadata.st_uid != os.getuid():
        raise SystemExit("account directory/file must be regular, owner-only and readable")
if os.listdir(account_dir) != ["staffs.csv"]:
    raise SystemExit("unexpected account directory entry")
with open(path, encoding="utf-8") as fh:
    m = json.load(fh)
got_clinic = m.get("clinicCode") or ""
got_run = m.get("sourceRunId") or m.get("migrationRunId") or ""
tables = m.get("tables") or []
if got_clinic != clinic:
    raise SystemExit(f"manifest clinicCode={got_clinic} != CLINIC_CODE={clinic}")
if got_run != run_id:
    raise SystemExit(f"manifest run={got_run} != MIGRATION_RUN_ID={run_id}")
if len(tables) != 21:
    raise SystemExit(f"expected 21 tables, got {len(tables)}")
PY

echo "old-db-handoff-check: PASS ($CLINIC_CODE 21 CSV + manifest, ignored)"
