#!/usr/bin/env bash
# After local `make reset` postflight, import every staged old_db handoff under
# backend/migrations/seeds/_old_db_handoff/<clinic>/<run>/ into the local DB.
#
# Uses csv-import --allow-local-rehearsal (REHEARSAL_ONLY allowed). Never for
# staging/production. Invoked only from scripts/local-db-reset-contract.sh.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

HANDOFF_ROOT="$ROOT/backend/migrations/seeds/_old_db_handoff"
if [[ ! -d "$HANDOFF_ROOT" ]]; then
  echo "INFO  no old_db handoff root; skip import"
  exit 0
fi

APP_ENV_VAL="$(grep -E '^APP_ENV=' .env.local 2>/dev/null | head -1 | cut -d= -f2- | tr -d '"' || true)"
APP_ENV_VAL="${APP_ENV_VAL:-${APP_ENV:-}}"
case "$(printf '%s' "$APP_ENV_VAL" | tr '[:upper:]' '[:lower:]')" in
  development|local|dev|test) ;;
  *)
    echo "INFO  APP_ENV=$APP_ENV_VAL is not local; skip old_db handoff import"
    exit 0
    ;;
esac

DB_NAME_VAL="$(grep -E '^DB_NAME=' .env.local 2>/dev/null | head -1 | cut -d= -f2- | tr -d '"' || true)"
DB_NAME_VAL="${DB_NAME_VAL:-${DB_NAME:-ekarte_db}}"

psql_scalar() {
  docker compose -p animalekarte exec -T db \
    psql -U ekarte_user -d "$DB_NAME_VAL" -tAc "$1" | tr -d '[:space:]'
}

ensure_clinic_seed_graph() {
  local clinic_id="$1"
  # Prefer the clinic that matches producer ordinal so (clinic_id,name) unique
  # indexes do not collide with demo clinic 1 rows.
  local exists
  exists="$(psql_scalar "SELECT 1 FROM clinics WHERE id=${clinic_id} AND is_active IS TRUE LIMIT 1")"
  if [[ "$exists" != "1" ]]; then
    echo "FAIL  clinics.id=${clinic_id} is missing/inactive" >&2
    return 1
  fi
  # Ensure exam type "検査"
  if [[ -z "$(psql_scalar "SELECT id FROM exam_types WHERE clinic_id=${clinic_id} AND name='検査' AND deleted_at IS NULL LIMIT 1")" ]]; then
    psql_scalar "INSERT INTO exam_types (clinic_id, name, is_active, description, sort_order, is_non_insurance)
      VALUES (${clinic_id}, '検査', true, 'local old_db handoff seed', 0, false) RETURNING id" >/dev/null \
      || { echo "FAIL  insert exam_types for clinic ${clinic_id}" >&2; return 1; }
    echo "INFO  inserted exam_types 検査 for clinic_id=${clinic_id}" >&2
  fi
  # Ensure one trimming reservation type
  if [[ -z "$(psql_scalar "SELECT id FROM reservation_types WHERE clinic_id=${clinic_id} AND category='trimming' AND deleted_at IS NULL AND is_active IS TRUE LIMIT 1")" ]]; then
    psql_scalar "INSERT INTO reservation_types (
        clinic_id, name, is_active, description, color, sort_order,
        reservation_display_name, duration_minutes, short_name, show_short_name,
        reservation_visible, reservation_comment, reservation_image_url,
        reservation_day_option, is_internal, category
      ) VALUES (
        ${clinic_id}, 'トリミング', true, 'local old_db handoff seed', '#3B82F6', 0,
        'トリミング', 30, 'トリミング', false,
        true, '', '',
        'none', false, 'trimming'
      ) RETURNING id" >/dev/null \
      || { echo "FAIL  insert reservation_types for clinic ${clinic_id}" >&2; return 1; }
    echo "INFO  inserted reservation_types trimming for clinic_id=${clinic_id}" >&2
  fi
}

resolve_seeds_for_clinic() {
  local clinic_id="$1"
  local species exam trim cash card
  ensure_clinic_seed_graph "$clinic_id" || return 1
  species="$(psql_scalar "SELECT id FROM animal_species WHERE is_active IS TRUE ORDER BY id LIMIT 1")"
  exam="$(psql_scalar "SELECT id FROM exam_types WHERE clinic_id=${clinic_id} AND name='検査' AND deleted_at IS NULL ORDER BY id LIMIT 1")"
  trim="$(psql_scalar "SELECT id FROM reservation_types WHERE clinic_id=${clinic_id} AND category='trimming' AND deleted_at IS NULL AND is_active IS TRUE ORDER BY id LIMIT 1")"
  cash="$(psql_scalar "SELECT id FROM payment_methods WHERE clinic_id=${clinic_id} AND system_key='cash' AND deleted_at IS NULL ORDER BY id LIMIT 1")"
  card="$(psql_scalar "SELECT id FROM payment_methods WHERE clinic_id=${clinic_id} AND system_key='credit_card' AND deleted_at IS NULL ORDER BY id LIMIT 1")"
  if [[ -z "$species" || -z "$exam" || -z "$trim" || -z "$cash" || -z "$card" ]]; then
    echo "FAIL  could not resolve local seed IDs for clinic_id=${clinic_id}" >&2
    return 1
  fi
  printf '%s %s %s %s %s %s\n' "$clinic_id" "$species" "$exam" "$trim" "$cash" "$card"
}

import_one() {
  local dir="$1"
  local manifest="$dir/manifest.json"
  [[ -f "$manifest" ]] || return 0

  local meta clinic run ordinal sha
  meta="$(python3 - "$manifest" <<'PY'
import json, sys
m = json.load(open(sys.argv[1], encoding="utf-8"))
print(m.get("clinicCode") or "")
print(m.get("sourceRunId") or m.get("migrationRunId") or "")
print(int(m.get("clinicOrdinal") or 0))
PY
)"
  clinic="$(printf '%s\n' "$meta" | sed -n '1p')"
  run="$(printf '%s\n' "$meta" | sed -n '2p')"
  ordinal="$(printf '%s\n' "$meta" | sed -n '3p')"
  sha="$(shasum -a 256 "$manifest" | awk '{print $1}')"

  local seeds seed_clinic species exam trim cash card
  seeds="$(resolve_seeds_for_clinic "$ordinal")"
  seed_clinic="$(printf '%s\n' "$seeds" | awk '{print $1}')"
  species="$(printf '%s\n' "$seeds" | awk '{print $2}')"
  exam="$(printf '%s\n' "$seeds" | awk '{print $3}')"
  trim="$(printf '%s\n' "$seeds" | awk '{print $4}')"
  cash="$(printf '%s\n' "$seeds" | awk '{print $5}')"
  card="$(printf '%s\n' "$seeds" | awk '{print $6}')"

  echo "INFO  importing old_db handoff clinic=$clinic run=$run ordinal=$ordinal -> clinic_id=$seed_clinic"

  # csv-import preflight requires owner-only modes (dir 0700, files 0600).
  chmod 700 "$dir"
  find "$dir" -type d -exec chmod 700 {} +
  find "$dir" -type f -exec chmod 600 {} +

  # Local rehearsal only: remove existing demo rows for this clinic so producer
  # unique indexes (clinic_id,name/phone/...) cannot collide. Keep the seed
  # graph we just ensured (exam_types / reservation_types / payment_methods).
  docker compose -p animalekarte exec -T db psql -U ekarte_user -d "$DB_NAME_VAL" -v ON_ERROR_STOP=1 <<SQL >/dev/null
BEGIN;
DELETE FROM payment_splits WHERE clinic_id = ${seed_clinic};
DELETE FROM payments WHERE clinic_id = ${seed_clinic};
DELETE FROM billing_items WHERE billing_id IN (SELECT id FROM billings WHERE clinic_id = ${seed_clinic});
DELETE FROM billing_refunds WHERE clinic_id = ${seed_clinic};
DELETE FROM billings WHERE clinic_id = ${seed_clinic};
DELETE FROM estimate_items WHERE estimate_id IN (SELECT id FROM estimates WHERE clinic_id = ${seed_clinic});
DELETE FROM estimates WHERE clinic_id = ${seed_clinic};
DELETE FROM exam_results WHERE exam_id IN (SELECT id FROM exams WHERE clinic_id = ${seed_clinic});
DELETE FROM exams WHERE clinic_id = ${seed_clinic};
DELETE FROM vaccinations WHERE clinic_id = ${seed_clinic};
DELETE FROM vaccines WHERE clinic_id = ${seed_clinic};
DELETE FROM appointment_trimming_details WHERE clinic_id = ${seed_clinic};
DELETE FROM appointments WHERE clinic_id = ${seed_clinic};
DELETE FROM vital_records WHERE clinic_id = ${seed_clinic};
DELETE FROM inquiries WHERE medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ${seed_clinic});
DELETE FROM clinical_plans WHERE medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ${seed_clinic});
DELETE FROM medical_record_addenda WHERE clinic_id = ${seed_clinic};
DELETE FROM prescriptions WHERE clinic_id = ${seed_clinic};
DELETE FROM medical_records WHERE clinic_id = ${seed_clinic};
DELETE FROM checkups WHERE clinic_id = ${seed_clinic};
DELETE FROM hospitalizations WHERE clinic_id = ${seed_clinic};
DELETE FROM pet_chronic_conditions WHERE clinic_id = ${seed_clinic};
DELETE FROM pets WHERE clinic_id = ${seed_clinic};
DELETE FROM shared_files WHERE clinic_id = ${seed_clinic};
DELETE FROM line_send_logs WHERE clinic_id = ${seed_clinic};
DELETE FROM line_link_tokens WHERE clinic_id = ${seed_clinic};
DELETE FROM lstep_tag_cache WHERE clinic_id = ${seed_clinic};
DELETE FROM lstep_delivery_trigger_log WHERE clinic_id = ${seed_clinic};
DELETE FROM lstep_migration_progress WHERE clinic_id = ${seed_clinic};
DELETE FROM lstep_sync_error_counters WHERE clinic_id = ${seed_clinic};
DELETE FROM owners WHERE clinic_id = ${seed_clinic};
DELETE FROM procedures WHERE clinic_id = ${seed_clinic};
DELETE FROM merchandise_items WHERE clinic_id = ${seed_clinic};
-- Keep existing demo staffs (RESTRICT FKs from audit/shift tables). Producer
-- staff IDs are in the clinic band and do not collide on name for jouto.
--
-- The delete list above must stay closed under RESTRICT foreign keys: every
-- table that references a deleted table with ON DELETE RESTRICT has to be
-- deleted first, or the whole transaction aborts.
-- backend/internal/lintscan/handoff_delete_closure_lint_test.go derives that
-- closure from backend/migrations/001_init.sql and fails when a new RESTRICT
-- reference is added, or when a child is ordered after its parent here.
COMMIT;
SQL
  echo "INFO  cleared existing clinic_id=${seed_clinic} clinical/owner/catalog rows before import" >&2

  export CSV_IMPORT_SOURCE_DIR="$dir"
  export CSV_MANIFEST_SHA256="$sha"
  export CLINIC_CODE="$clinic"
  export CLINIC_ORDINAL="$ordinal"
  export MIGRATION_RUN_ID="$run"
  export TARGET_CLINIC_ID="$seed_clinic"
  export FALLBACK_ANIMAL_SPECIES_ID="$species"
  export FALLBACK_EXAM_TYPE_ID="$exam"
  export TRIMMING_RESERVATION_TYPE_ID="$trim"
  export PAYMENT_METHOD_CASH_ID="$cash"
  export PAYMENT_METHOD_CREDIT_CARD_ID="$card"
  export TARGET_DB_NAME="$DB_NAME_VAL"
  export CSV_IMPORT_ALLOW_LOCAL_REHEARSAL=1
  export APP_ENV="${APP_ENV_VAL:-development}"
  export CSV_IMPORT_EXTRA_ARGS='--allow-local-rehearsal'

  # apply refuses to overwrite an existing report path.
  mkdir -p -m 700 "$ROOT/sensitive-local/csv-import-reports"
  rm -f "$ROOT/sensitive-local/csv-import-reports/${clinic}-${run}-apply.json"

  make csv-import-preflight
  make csv-import TARGET_DB_NAME="$DB_NAME_VAL"
  make csv-import-verify
  echo "INFO  imported $clinic/$run"
}

# Prefer <run>-local over <run> when both are staged for the same clinic.
# Producer CSV may fail uk_owners_clinic_phone; local sanitized bundles are
# named <sourceRunId>-local and keep the same manifest sourceRunId.
select_handoff_dirs() {
  python3 - "$HANDOFF_ROOT" <<'PY'
import json, os, sys
from pathlib import Path

root = Path(sys.argv[1])
candidates = []
for manifest in sorted(root.glob("*/*/manifest.json")):
    try:
        m = json.loads(manifest.read_text(encoding="utf-8"))
    except Exception:
        continue
    clinic = m.get("clinicCode") or manifest.parent.parent.name
    run = m.get("sourceRunId") or m.get("migrationRunId") or manifest.parent.name
    dirname = manifest.parent.name
    is_local = dirname.endswith("-local") or dirname == f"{run}-local"
    candidates.append((clinic, run, is_local, str(manifest.parent)))

chosen = {}
for clinic, run, is_local, path in candidates:
    key = (clinic, run)
    prev = chosen.get(key)
    if prev is None:
        chosen[key] = (is_local, path)
        continue
    prev_local, _ = prev
    if is_local and not prev_local:
        chosen[key] = (is_local, path)

for (_clinic, _run), (_is_local, path) in sorted(chosen.items()):
    print(path)
PY
}

found=0
while IFS= read -r dir; do
  [[ -n "$dir" ]] || continue
  found=1
  import_one "$dir"
done < <(select_handoff_dirs)

if [[ "$found" -eq 0 ]]; then
  echo "INFO  no staged old_db handoff bundles; skip import"
fi
