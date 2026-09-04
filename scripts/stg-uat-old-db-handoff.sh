#!/usr/bin/env bash
# Bind every staged _old_db_handoff clinic onto shared STG via make stg-uat-import.
# Always processes jouto, shikishima, hakobuneco (missing dirs are skipped).
# Clinic selection arguments are rejected.
#
# Does NOT call cmd/migrate, make csv-import, or --allow-local-rehearsal.
# Does NOT print passwords / CSV cell values.
#
# Usage:
#   make stg-uat-handoff
#   bash scripts/stg-uat-old-db-handoff.sh import|preflight|verify
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

die() { echo "stg-uat-old-db-handoff: $*" >&2; exit 1; }

incoming_app_env="$(printf '%s' "${APP_ENV:-}" | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]')"
if [[ "$incoming_app_env" == "production" ]]; then
  die "refuses APP_ENV=production"
fi

# --- STG connection (written here so make に export は不要) ---
# パスワードは gitignored の local.env にだけ置く。この tracked ファイルへ書かない。
export APP_ENV=staging
export STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL=YES_I_UNDERSTAND
export DB_PORT=5432
export DB_NAME=postgres
export TARGET_DB_NAME=postgres
export DB_SSL_MODE=verify-full
export DB_SSL_ROOT_CERT=system

load_local_db_exports() {
  if [[ "${STG_UAT_HANDOFF_SKIP_LOCAL_ENV:-}" == "1" ]]; then
    return 0
  fi
  local env_file="${STG_UAT_HANDOFF_ENV_FILE:-$ROOT/scripts/stg-uat-old-db-handoff.local.env}"
  if [[ ! -f "$env_file" ]]; then
    die "missing $env_file (copy scripts/stg-uat-old-db-handoff.local.env.example, chmod 600, fill DB_HOST/DB_USER/DB_PASSWORD)"
  fi
  local mode
  mode="$(stat -f '%Lp' "$env_file" 2>/dev/null || stat -c '%a' "$env_file")"
  if [[ "$mode" != "600" && "$mode" != "0600" ]]; then
    die "$env_file must be mode 0600 (got $mode)"
  fi
  set -a
  # shellcheck disable=SC1090
  source "$env_file"
  set +a
}

load_local_db_exports

export TARGET_DB_NAME="${TARGET_DB_NAME:-${DB_NAME:-postgres}}"
export DB_NAME="$TARGET_DB_NAME"
export STG_UAT_CSV_IMPORT_CONFIRM_HOST="${STG_UAT_CSV_IMPORT_CONFIRM_HOST:-${DB_HOST:-}}"

COMMAND="${1:-}"
shift || true

HANDOFF_ROOT="${STG_UAT_HANDOFF_ROOT:-$ROOT/backend/migrations/seeds/_old_db_handoff}"
MASTER_DIR="${STG_UAT_HANDOFF_MASTER_DIR:-$ROOT/backend/migrations/seeds/002_master}"
REPORT_DIR="${STG_UAT_HANDOFF_REPORT_DIR:-$ROOT/sensitive-local/csv-import-reports}"
ALL_CLINICS=(jouto shikishima hakobuneco)

usage() {
  die "usage: $0 import|preflight|verify

Always imports jouto, shikishima, hakobuneco. Do not pass a clinic code.
STG_UAT_HANDOFF_DRY_RUN=1 binds only and does not invoke make."
}

case "$COMMAND" in
  import|preflight|verify) ;;
  *) usage ;;
esac

if [[ "$#" -gt 0 ]]; then
  die "clinic selection is not supported; this command always imports ${ALL_CLINICS[*]}"
fi

make_target=""
case "$COMMAND" in
  import) make_target="stg-uat-import" ;;
  preflight) make_target="stg-uat-csv-import-preflight" ;;
  verify) make_target="stg-uat-csv-import-verify" ;;
esac

clinics=("${ALL_CLINICS[@]}")

is_local_db_host() {
  local host="$1"
  case "$(printf '%s' "$host" | tr '[:upper:]' '[:lower:]')" in
    db|localhost|127.0.0.1|::1|0.0.0.0) return 0 ;;
    *) return 1 ;;
  esac
}

require_remote_env() {
  [[ -n "${DB_HOST:-}" ]] || die "DB_HOST is required in scripts/stg-uat-old-db-handoff.local.env"
  if is_local_db_host "$DB_HOST"; then
    die "DB_HOST is local; pscale connect / compose db are not allowed"
  fi
  [[ -n "${DB_USER:-}" ]] || die "DB_USER is required in scripts/stg-uat-old-db-handoff.local.env"
  [[ -n "${DB_PASSWORD:-}" ]] || die "DB_PASSWORD is required in scripts/stg-uat-old-db-handoff.local.env"
  [[ "${STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL:-}" == "YES_I_UNDERSTAND" ]] || \
    die "STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL=YES_I_UNDERSTAND is required"
}

resolve_bundle() {
  local dir="$1"
  python3 - "$dir" "$MASTER_DIR" <<'PY'
import csv
import json
import sys
from pathlib import Path

handoff_dir = Path(sys.argv[1])
master = Path(sys.argv[2])
manifest_path = handoff_dir / "manifest.json"
if not manifest_path.is_file():
    raise SystemExit(f"missing manifest: {manifest_path}")

manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
clinic = str(manifest.get("clinicCode") or "").strip()
run = str(manifest.get("sourceRunId") or manifest.get("migrationRunId") or "").strip()
try:
    ordinal = int(manifest.get("clinicOrdinal") or 0)
except (TypeError, ValueError) as exc:
    raise SystemExit(f"invalid clinicOrdinal: {exc}") from exc
if not clinic or not run or ordinal < 1:
    raise SystemExit("manifest clinicCode/sourceRunId/clinicOrdinal is incomplete")

def load(name):
    path = master / name
    if not path.is_file():
        raise SystemExit(f"missing 002_master csv: {path}")
    with path.open(newline="", encoding="utf-8") as handle:
        return list(csv.DictReader(handle))

def active(row):
    flag = (row.get("is_active") or "").strip().lower()
    return flag in {"t", "true", "1"}

def present(row):
    return not (row.get("deleted_at") or "").strip()

clinics = load("clinics.csv")
if not any(row.get("id") == str(ordinal) and active(row) for row in clinics):
    raise SystemExit(f"002_master clinics.csv has no active id={ordinal}")

species_rows = [row for row in load("animal_species.csv") if active(row)]
if not species_rows:
    raise SystemExit("002_master animal_species.csv has no active row")
species = species_rows[0]["id"]

def one(rows, label):
    if len(rows) != 1:
        raise SystemExit(f"expected 1 {label} for clinic_id={ordinal}, got {len(rows)}")
    ident = (rows[0].get("id") or "").strip()
    if not ident:
        raise SystemExit(f"{label} id is empty")
    return ident

exam = one(
    [
        row
        for row in load("exam_types.csv")
        if row.get("clinic_id") == str(ordinal) and row.get("name") == "検査" and present(row)
    ],
    "exam_types 検査",
)
trim = one(
    [
        row
        for row in load("reservation_types.csv")
        if row.get("clinic_id") == str(ordinal)
        and row.get("category") == "trimming"
        and active(row)
        and present(row)
    ],
    "reservation_types trimming",
)
cash = one(
    [
        row
        for row in load("payment_methods.csv")
        if row.get("clinic_id") == str(ordinal)
        and row.get("system_key") == "cash"
        and active(row)
        and present(row)
    ],
    "payment_methods cash",
)
card = one(
    [
        row
        for row in load("payment_methods.csv")
        if row.get("clinic_id") == str(ordinal)
        and row.get("system_key") == "credit_card"
        and active(row)
        and present(row)
    ],
    "payment_methods credit_card",
)

print(clinic)
print(run)
print(ordinal)
print(species)
print(exam)
print(trim)
print(cash)
print(card)
PY
}

require_remote_env

ran=0
for clinic_dir_name in "${clinics[@]}"; do
  dir="$HANDOFF_ROOT/$clinic_dir_name"
  if [[ ! -f "$dir/manifest.json" ]]; then
    echo "stg-uat-old-db-handoff: skip $clinic_dir_name (no staged manifest)" >&2
    continue
  fi

  bind="$(resolve_bundle "$dir")"
  clinic="$(printf '%s\n' "$bind" | sed -n '1p')"
  run="$(printf '%s\n' "$bind" | sed -n '2p')"
  ordinal="$(printf '%s\n' "$bind" | sed -n '3p')"
  species="$(printf '%s\n' "$bind" | sed -n '4p')"
  exam="$(printf '%s\n' "$bind" | sed -n '5p')"
  trim="$(printf '%s\n' "$bind" | sed -n '6p')"
  cash="$(printf '%s\n' "$bind" | sed -n '7p')"
  card="$(printf '%s\n' "$bind" | sed -n '8p')"
  sha="$(shasum -a 256 "$dir/manifest.json" | awk '{print $1}')"

  chmod 700 "$dir"
  find "$dir" -type d -exec chmod 700 {} +
  find "$dir" -type f -exec chmod 600 {} +

  export CSV_IMPORT_SOURCE_DIR="$dir"
  export CSV_MANIFEST_SHA256="$sha"
  export CLINIC_CODE="$clinic"
  export CLINIC_ORDINAL="$ordinal"
  export MIGRATION_RUN_ID="$run"
  export TARGET_CLINIC_ID="$ordinal"
  export FALLBACK_ANIMAL_SPECIES_ID="$species"
  export FALLBACK_EXAM_TYPE_ID="$exam"
  export TRIMMING_RESERVATION_TYPE_ID="$trim"
  export PAYMENT_METHOD_CASH_ID="$cash"
  export PAYMENT_METHOD_CREDIT_CARD_ID="$card"

  echo "stg-uat-old-db-handoff: $COMMAND clinic=$clinic dir=$clinic_dir_name run=$run ordinal=$ordinal clinic_id=$ordinal" >&2

  if [[ "${STG_UAT_HANDOFF_DRY_RUN:-}" == "1" ]]; then
    printf 'DRY_RUN target=%s clinic=%s run=%s ordinal=%s species=%s exam=%s trim=%s cash=%s card=%s sha256=%s\n' \
      "$make_target" "$clinic" "$run" "$ordinal" "$species" "$exam" "$trim" "$cash" "$card" "$sha"
    ran=1
    continue
  fi

  if [[ "$COMMAND" == "import" ]]; then
    report="$REPORT_DIR/${clinic}-${run}-stg-uat-apply.json"
    if [[ -f "$report" ]]; then
      status="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1], encoding="utf-8")).get("status",""))' "$report")"
      if [[ "$status" == "PASS" ]]; then
        echo "stg-uat-old-db-handoff: skip $clinic (apply report already PASS)" >&2
        ran=1
        continue
      fi
    fi
  fi

  make -C "$ROOT" "$make_target"
  ran=1
done

[[ "$ran" -eq 1 ]] || die "no staged handoff bundle was imported"
