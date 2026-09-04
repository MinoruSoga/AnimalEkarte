#!/usr/bin/env bash
# scripts/stg-uat-old-db-handoff.test.sh
#
# stg-uat-old-db-handoff.sh の契約テスト。実 STG / Docker import は走らせない。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SCRIPT="$SCRIPT_DIR/stg-uat-old-db-handoff.sh"

if [[ ! -f "$SCRIPT" ]]; then
  echo "FAIL  script not found at $SCRIPT"
  exit 1
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
failures=0

write_master() {
  local master="$1"
  mkdir -p "$master"
  printf '%s\n' 'id,company_id,name,is_active' >"$master/clinics.csv"
  printf '%s\n' '2,1,城東,t' >>"$master/clinics.csv"
  printf '%s\n' '3,1,敷島,t' >>"$master/clinics.csv"
  printf '%s\n' 'id,name,is_active' >"$master/animal_species.csv"
  printf '%s\n' '1,犬,t' >>"$master/animal_species.csv"
  printf '%s\n' 'id,clinic_id,name,deleted_at' >"$master/exam_types.csv"
  printf '%s\n' '11009,2,検査,' >>"$master/exam_types.csv"
  printf '%s\n' '11010,3,検査,' >>"$master/exam_types.csv"
  printf '%s\n' 'id,clinic_id,category,is_active,deleted_at' >"$master/reservation_types.csv"
  printf '%s\n' '2,2,trimming,t,' >>"$master/reservation_types.csv"
  printf '%s\n' '3,3,trimming,t,' >>"$master/reservation_types.csv"
  printf '%s\n' 'id,clinic_id,system_key,is_active,deleted_at' >"$master/payment_methods.csv"
  printf '%s\n' '5,2,cash,t,' >>"$master/payment_methods.csv"
  printf '%s\n' '6,2,credit_card,t,' >>"$master/payment_methods.csv"
  printf '%s\n' '9,3,cash,t,' >>"$master/payment_methods.csv"
  printf '%s\n' '10,3,credit_card,t,' >>"$master/payment_methods.csv"
}

write_handoff() {
  local dir="$1" code="$2" ordinal="$3" run="$4"
  mkdir -p "$dir"
  chmod 700 "$dir"
  python3 - "$dir/manifest.json" "$code" "$ordinal" "$run" <<'PY'
import json, sys
path, code, ordinal, run = sys.argv[1], sys.argv[2], int(sys.argv[3]), sys.argv[4]
json.dump(
    {"clinicCode": code, "clinicOrdinal": ordinal, "sourceRunId": run},
    open(path, "w", encoding="utf-8"),
)
PY
  chmod 600 "$dir/manifest.json"
}

run_case() {
  local name="$1"
  local expected_exit="$2"
  local want_text="$3"
  shift 3
  local out actual
  out="$(
    STG_UAT_HANDOFF_ROOT="$TMP/handoff" \
    STG_UAT_HANDOFF_MASTER_DIR="$TMP/master" \
    STG_UAT_HANDOFF_DRY_RUN=1 \
    STG_UAT_HANDOFF_SKIP_LOCAL_ENV=1 \
    DB_HOST="${CASE_DB_HOST-example.psdb.cloud}" \
    DB_USER="${CASE_DB_USER-role-user}" \
    DB_PASSWORD="${CASE_DB_PASSWORD-secret}" \
    APP_ENV="${CASE_APP_ENV-staging}" \
    bash "$SCRIPT" "$@" 2>&1
  )" && actual=0 || actual=$?
  if [[ "$actual" -ne "$expected_exit" ]]; then
    echo "FAIL  [$name] exit=$actual want=$expected_exit"
    printf '%s\n' "$out"
    failures=$((failures + 1))
    return
  fi
  if [[ -n "$want_text" ]] && ! printf '%s\n' "$out" | grep -q "$want_text"; then
    echo "FAIL  [$name] missing text: $want_text"
    printf '%s\n' "$out"
    failures=$((failures + 1))
    return
  fi
  echo "PASS  [$name]"
}

write_master "$TMP/master"
write_handoff "$TMP/handoff/jouto" jouto 2 jouto-intake-20260822-01
write_handoff "$TMP/handoff/shikishima" shikishima 3 jouto-intake-20260822-01
mkdir -p "$TMP/handoff/hachioji"

run_case "usage-bad-command" 1 "usage:"
run_case "reject-clinic-arg" 1 "clinic selection is not supported" import jouto
CASE_DB_HOST=db run_case "refuse-local-host" 1 "local" import
CASE_APP_ENV=production run_case "refuse-production" 1 "production" import
run_case "dry-run-all-jouto" 0 "DRY_RUN target=stg-uat-import clinic=jouto" import
run_case "dry-run-all-shikishima" 0 "clinic=shikishima" import
run_case "dry-run-skip-missing" 0 "skip hakobuneco" import
run_case "dry-run-seed-ids" 0 "exam=11009" preflight

mkdir -p "$TMP/reports" "$TMP/bin"
printf '%s\n' '{"status":"PASS"}' >"$TMP/reports/jouto-jouto-intake-20260822-01-stg-uat-apply.json"
cat >"$TMP/bin/make" <<'EOF'
#!/bin/sh
echo MAKE_CALLED "$*"
EOF
chmod +x "$TMP/bin/make"
skip_out="$(
  PATH="$TMP/bin:$PATH" \
  STG_UAT_HANDOFF_ROOT="$TMP/handoff" \
  STG_UAT_HANDOFF_MASTER_DIR="$TMP/master" \
  STG_UAT_HANDOFF_REPORT_DIR="$TMP/reports" \
  STG_UAT_HANDOFF_SKIP_LOCAL_ENV=1 \
  DB_HOST="${CASE_DB_HOST-example.psdb.cloud}" \
  DB_USER="${CASE_DB_USER-role-user}" \
  DB_PASSWORD="${CASE_DB_PASSWORD-secret}" \
  APP_ENV=staging \
  bash "$SCRIPT" import 2>&1
)" && skip_actual=0 || skip_actual=$?
if [[ "$skip_actual" -eq 0 ]] && \
   printf '%s\n' "$skip_out" | grep -q 'skip jouto (apply report already PASS)' && \
   printf '%s\n' "$skip_out" | grep -q 'MAKE_CALLED' && \
   printf '%s\n' "$skip_out" | grep -q 'clinic=shikishima'; then
  echo "PASS  [skip-pass-report-continues]"
else
  echo "FAIL  [skip-pass-report-continues] exit=$skip_actual"
  printf '%s\n' "$skip_out"
  failures=$((failures + 1))
fi

missing_env="$TMP/missing.env"
out="$(
  STG_UAT_HANDOFF_ROOT="$TMP/handoff" \
  STG_UAT_HANDOFF_MASTER_DIR="$TMP/master" \
  STG_UAT_HANDOFF_DRY_RUN=1 \
  STG_UAT_HANDOFF_ENV_FILE="$missing_env" \
  bash "$SCRIPT" import 2>&1
)" && actual=0 || actual=$?
if [[ "$actual" -eq 1 ]] && printf '%s\n' "$out" | grep -q "missing"; then
  echo "PASS  [missing-local-env]"
else
  echo "FAIL  [missing-local-env] exit=$actual"
  printf '%s\n' "$out"
  failures=$((failures + 1))
fi

if grep -E '^[^#]*--allow-local-rehearsal' "$SCRIPT" >/dev/null; then
  echo "FAIL  [no-local-rehearsal-flag] script must not pass --allow-local-rehearsal"
  failures=$((failures + 1))
else
  echo "PASS  [no-local-rehearsal-flag]"
fi

if grep -Eq '^export STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL=YES_I_UNDERSTAND$' "$SCRIPT" && \
   grep -Eq '^export DB_SSL_MODE=verify-full$' "$SCRIPT"; then
  echo "PASS  [exports-in-script]"
else
  echo "FAIL  [exports-in-script]"
  failures=$((failures + 1))
fi

if grep -Eq 'stg-uat-old-db-handoff.sh import$' "$ROOT/Makefile" && \
   ! grep -Eq 'stg-uat-handoff-all:' "$ROOT/Makefile" && \
   ! grep -Eq 'stg-uat-old-db-handoff.sh import "\$\(CLINIC_CODE\)"' "$ROOT/Makefile"; then
  echo "PASS  [makefile-wires-script]"
else
  echo "FAIL  [makefile-wires-script]"
  failures=$((failures + 1))
fi

if grep -Eq 'scripts/stg-uat-old-db-handoff.local.env' "$ROOT/.gitignore"; then
  echo "PASS  [local-env-gitignored]"
else
  echo "FAIL  [local-env-gitignored]"
  failures=$((failures + 1))
fi

if [[ "$failures" -eq 0 ]]; then
  echo "OK  stg-uat-old-db-handoff contract"
  exit 0
fi
echo "FAIL  $failures check(s)"
exit 1
