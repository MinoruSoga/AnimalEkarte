#!/usr/bin/env bash
# Stage an old_db 21-table CSV bundle into the local PHI quarantine:
#   backend/migrations/seeds/_old_db_handoff/<clinic>/
# Staff CSV lives in 002_master/accounts/_old_db_handoff/<clinic>/.
#
# This does NOT register the bundle with cmd/migrate / make seed.
# Formal DB import remains make csv-import-* and requires TRUSTED_CANDIDATE.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

die() { echo "stage-old-db-handoff: $*" >&2; exit 1; }

CLINIC_CODE="${CLINIC_CODE:-}"
MIGRATION_RUN_ID="${MIGRATION_RUN_ID:-}"
SOURCE_DIR="${CSV_IMPORT_SOURCE_DIR:-${OLD_DB_CSV_SOURCE_DIR:-}}"

[[ -n "$CLINIC_CODE" ]] || die "CLINIC_CODE is required"
[[ -n "$MIGRATION_RUN_ID" ]] || die "MIGRATION_RUN_ID is required"
[[ -n "$SOURCE_DIR" ]] || die "CSV_IMPORT_SOURCE_DIR (or OLD_DB_CSV_SOURCE_DIR) is required"
[[ "$CLINIC_CODE" =~ ^[a-z][a-z0-9-]{0,31}$ ]] || die "CLINIC_CODE must be a lowercase slug"
[[ "$MIGRATION_RUN_ID" =~ ^[A-Za-z0-9._-]{1,64}$ ]] || die "MIGRATION_RUN_ID is unsafe"
[[ -d "$SOURCE_DIR" ]] || die "source dir does not exist: $SOURCE_DIR"
[[ -f "$SOURCE_DIR/manifest.json" ]] || die "manifest.json missing in source dir"

EXCLUDE_FILE="$(git rev-parse --git-path info/exclude)"
for EXCLUDE_LINE in backend/migrations/seeds/_old_db_handoff/ backend/migrations/seeds/002_master/accounts/_old_db_handoff/; do
if [[ ! -f "$EXCLUDE_FILE" ]] || ! grep -qxF "$EXCLUDE_LINE" "$EXCLUDE_FILE"; then
  mkdir -p "$(dirname "$EXCLUDE_FILE")"
  printf '\n# PHI-bearing old_db CSV quarantine (local only; never commit)\n%s\n' \
    "$EXCLUDE_LINE" >> "$EXCLUDE_FILE"
fi

git check-ignore -q --no-index "$EXCLUDE_LINE" \
  || die "git check-ignore failed for $EXCLUDE_LINE — refuse to stage PHI"
done

DEST="$ROOT/backend/migrations/seeds/_old_db_handoff/$CLINIC_CODE"
DEST_PARENT="$(dirname "$DEST")"
ACCOUNT_PARENT="$ROOT/backend/migrations/seeds/002_master/accounts/_old_db_handoff"
ACCOUNT_DEST="$ACCOUNT_PARENT/$CLINIC_CODE"
for directory in "$DEST_PARENT" "$DEST" "$ROOT/backend/migrations/seeds/002_master/accounts" "$ACCOUNT_PARENT" "$ACCOUNT_DEST"; do
  [[ ! -L "$directory" ]] || die "symlink destination is unsafe"
  [[ ! -e "$directory" || -d "$directory" ]] || die "destination must be a directory"
done
mkdir -p "$DEST_PARENT" "$ACCOUNT_PARENT"
chmod 700 "$DEST_PARENT" "$ACCOUNT_PARENT"
STAGE_ROOT="$(mktemp -d "$DEST_PARENT/.stage-old-db-handoff.XXXXXX")"
ACCOUNT_STAGE_ROOT="$(mktemp -d "$ACCOUNT_PARENT/.stage-old-db-handoff.XXXXXX")"
STAGED_DEST="$STAGE_ROOT/$CLINIC_CODE"
BACKUP_DEST="$STAGE_ROOT/previous"
STAGED_ACCOUNT="$ACCOUNT_STAGE_ROOT/$CLINIC_CODE"
BACKUP_ACCOUNT="$ACCOUNT_STAGE_ROOT/previous"
CLINICAL_PROMOTED=0
ACCOUNT_PROMOTED=0
COMMITTED=0
cleanup_stage() {
  local result=$?
  trap - EXIT HUP INT TERM
  if [[ "$COMMITTED" == 0 ]]; then
    if [[ "$CLINICAL_PROMOTED" == 1 && -d "$DEST" ]]; then mv "$DEST" "$STAGED_DEST" || return 1; fi
    if [[ "$ACCOUNT_PROMOTED" == 1 && -d "$ACCOUNT_DEST" ]]; then mv "$ACCOUNT_DEST" "$STAGED_ACCOUNT" || return 1; fi
    if [[ -d "$BACKUP_DEST" ]]; then mv "$BACKUP_DEST" "$DEST" || return 1; fi
    if [[ -d "$BACKUP_ACCOUNT" ]]; then mv "$BACKUP_ACCOUNT" "$ACCOUNT_DEST" || return 1; fi
  fi
  rm -rf "$STAGE_ROOT" "$ACCOUNT_STAGE_ROOT"
  return "$result"
}
trap cleanup_stage EXIT
trap 'exit 1' HUP INT TERM
mkdir -p "$STAGED_DEST"
rsync -a --delete "$SOURCE_DIR/" "$STAGED_DEST/"
find "$STAGED_DEST" -type d -exec chmod 700 {} +
find "$STAGED_DEST" -type f -exec chmod 600 {} +

# Select exactly one staff source; clinical files and manifest retain their bytes.
ACCOUNT_SOURCE="${CSV_IMPORT_ACCOUNT_SOURCE_DIR:-}"
if [[ -z "$ACCOUNT_SOURCE" && ! -e "$SOURCE_DIR/staffs.csv" && ! -e "$SOURCE_DIR/accounts" ]]; then
  if [[ "$(cd "$SOURCE_DIR" && pwd -P)" == "$(cd "$DEST_PARENT" && pwd -P)/$CLINIC_CODE" ]]; then
    ACCOUNT_SOURCE="$ACCOUNT_DEST"
  fi
fi
mkdir -m 700 "$STAGED_ACCOUNT"
python3 - "$STAGED_DEST" "$ACCOUNT_SOURCE" "$STAGED_ACCOUNT" <<'PY'
import pathlib, shutil, sys
clinical, external, target = sys.argv[1:]
root = pathlib.Path(clinical)
candidates = [root / "staffs.csv", root / "accounts/staffs.csv"]
if external:
    candidates.append(pathlib.Path(external) / "staffs.csv")
existing = [p for p in candidates if p.exists() or p.is_symlink()]
if len(existing) != 1 or existing[0].is_symlink() or not existing[0].is_file():
    raise SystemExit("expected exactly one regular staff CSV source")
source = existing[0]
if source.parent.is_symlink():
    raise SystemExit("unsafe account source directory")
shutil.copyfile(source, pathlib.Path(target) / "staffs.csv")
(pathlib.Path(target) / "staffs.csv").chmod(0o600)
if source in candidates[:2]:
    source.unlink()
if (root / "accounts").exists():
    (root / "accounts").rmdir()
if len(list(root.glob("*.csv"))) != 20:
    raise SystemExit("expected 20 clinical CSV files")
PY

git check-ignore -q --no-index "$STAGED_DEST/manifest.json" \
  || die "staged manifest is not ignored — abort"

META_FILE="$(mktemp)"
python3 - "$STAGED_DEST/manifest.json" >"$META_FILE" <<'PY'
import json, sys
m = json.load(open(sys.argv[1], encoding="utf-8"))
print(m.get("status") or "")
print(m.get("handoffEligibility") or "")
print(m.get("clinicCode") or "")
print(m.get("sourceRunId") or m.get("migrationRunId") or "")
print(len(m.get("tables") or []))
PY
MANIFEST_STATUS="$(sed -n '1p' "$META_FILE")"
HANDOFF="$(sed -n '2p' "$META_FILE")"
MANIFEST_CLINIC="$(sed -n '3p' "$META_FILE")"
MANIFEST_RUN="$(sed -n '4p' "$META_FILE")"
TABLE_COUNT="$(sed -n '5p' "$META_FILE")"
rm -f "$META_FILE"

[[ "$MANIFEST_CLINIC" == "$CLINIC_CODE" ]] || die "manifest clinicCode=$MANIFEST_CLINIC != CLINIC_CODE=$CLINIC_CODE"
[[ "$MANIFEST_RUN" == "$MIGRATION_RUN_ID" ]] || die "manifest run=$MANIFEST_RUN != MIGRATION_RUN_ID=$MIGRATION_RUN_ID"
[[ "$TABLE_COUNT" == "21" ]] || die "expected 21 tables, got $TABLE_COUNT"

if [[ -e "$DEST" ]]; then
  mv "$DEST" "$BACKUP_DEST"
fi
if [[ -e "$ACCOUNT_DEST" ]]; then
  mv "$ACCOUNT_DEST" "$BACKUP_ACCOUNT"
fi
CLINICAL_PROMOTED=1
mv "$STAGED_DEST" "$DEST"
ACCOUNT_PROMOTED=1
mv "$STAGED_ACCOUNT" "$ACCOUNT_DEST"
COMMITTED=1
SHA="$(shasum -a 256 "$DEST/manifest.json" | awk '{print $1}')"
echo "stage-old-db-handoff: staged $DEST"
echo "  staff CSV: $ACCOUNT_DEST/staffs.csv"
echo "  manifest status=$MANIFEST_STATUS handoffEligibility=$HANDOFF"
echo "  manifestSha256=$SHA"
echo "  note: make seed / cmd/migrate will NOT load this directory"
if [[ "$HANDOFF" != "TRUSTED_CANDIDATE" || "$MANIFEST_STATUS" != "PASS" ]]; then
  echo "  note: formal csv-import-preflight will REJECT this bundle until TRUSTED_CANDIDATE/PASS"
fi
