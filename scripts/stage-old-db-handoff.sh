#!/usr/bin/env bash
# Stage an old_db 21-table CSV bundle into the local PHI quarantine:
#   backend/migrations/seeds/_old_db_handoff/<clinic>/
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

EXCLUDE_FILE="$ROOT/.git/info/exclude"
EXCLUDE_LINE="backend/migrations/seeds/_old_db_handoff/"
if [[ ! -f "$EXCLUDE_FILE" ]] || ! grep -qxF "$EXCLUDE_LINE" "$EXCLUDE_FILE"; then
  mkdir -p "$(dirname "$EXCLUDE_FILE")"
  printf '\n# PHI-bearing old_db CSV quarantine (local only; never commit)\n%s\n' \
    "$EXCLUDE_LINE" >> "$EXCLUDE_FILE"
fi

git check-ignore -q --no-index "$EXCLUDE_LINE" \
  || die "git check-ignore failed for $EXCLUDE_LINE — refuse to stage PHI"

DEST="$ROOT/backend/migrations/seeds/_old_db_handoff/$CLINIC_CODE"
mkdir -p "$DEST"
chmod 700 "$ROOT/backend/migrations/seeds/_old_db_handoff" "$DEST"
rsync -a --delete "$SOURCE_DIR/" "$DEST/"
chmod 700 "$DEST"
find "$DEST" -type f -exec chmod 600 {} +

git check-ignore -q --no-index "$DEST/manifest.json" \
  || die "staged manifest is not ignored — abort"

META_FILE="$(mktemp)"
python3 - "$DEST/manifest.json" >"$META_FILE" <<'PY'
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

SHA="$(shasum -a 256 "$DEST/manifest.json" | awk '{print $1}')"
echo "stage-old-db-handoff: staged $DEST"
echo "  manifest status=$MANIFEST_STATUS handoffEligibility=$HANDOFF"
echo "  manifestSha256=$SHA"
echo "  note: make seed / cmd/migrate will NOT load this directory"
if [[ "$HANDOFF" != "TRUSTED_CANDIDATE" || "$MANIFEST_STATUS" != "PASS" ]]; then
  echo "  note: formal csv-import-preflight will REJECT this bundle until TRUSTED_CANDIDATE/PASS"
fi
