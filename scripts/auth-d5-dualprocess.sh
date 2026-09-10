#!/usr/bin/env bash
# D5 dual-process immediate-revoke and latency harness.
# Safe disposable DB only. Does not print emails, passwords, cookies, or tokens.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPORT_DIR="${D5_REPORT_DIR:-$ROOT/reports}"
WORKDIR="${D5_WORKDIR:-/tmp/auth-d5-dualprocess-$$}"
API_IMAGE="${D5_API_IMAGE:-animalekarte-backend:latest}"
NETWORK="${D5_NETWORK:-ekarte-network}"
PG_CONTAINER="${D5_PG_CONTAINER:-ae-auth-fix-disposable-pg}"

DB_HOST="${D5_DB_HOST:-ae-auth-fix-disposable-pg}"
DB_PORT="${D5_DB_PORT:-5432}"
DB_NAME="${D5_DB_NAME:-auth_d1_db}"
DB_USER="${D5_DB_USER:-auth_test}"
DB_PASSWORD="${D5_DB_PASSWORD:-auth_test_pass}"
DB_SSL_MODE="${D5_DB_SSL_MODE:-disable}"

PORT_A="${D5_PORT_A:-18081}"
PORT_B="${D5_PORT_B:-18082}"
NAME_A="${D5_NAME_A:-ae-auth-d5-api-a}"
NAME_B="${D5_NAME_B:-ae-auth-d5-api-b}"

ACTOR_EMAIL="${D5_ACTOR_EMAIL:-stg-staff-10000021@example.test}"
VICTIM_EMAIL="${D5_VICTIM_EMAIL:-stg-staff-10000003@example.test}"
# Public non-production demo credential (seedlogin.SharedPassword).
DEMO_PASSWORD="${D5_DEMO_PASSWORD:-password}"
NEW_PASSWORD="${D5_NEW_PASSWORD:-D5NewPassw0rd}"

VICTIM_STAFF_ID="${D5_VICTIM_STAFF_ID:-10000003}"
ACTOR_STAFF_ID="${D5_ACTOR_STAFF_ID:-10000021}"
CLINIC_KEEP="${D5_CLINIC_KEEP:-1}"
CLINIC_DROP="${D5_CLINIC_DROP:-2}"
CLINIC_FIXED_PATH="${D5_CLINIC_FIXED_PATH:-/api/v1/masters/staffs}"

WARMUP="${D5_WARMUP:-5}"
MEASURE_N="${D5_MEASURE_N:-30}"
INTERVAL_MS="${D5_INTERVAL_MS:-50}"

mkdir -p "$REPORT_DIR" "$WORKDIR"
EVIDENCE="$REPORT_DIR/auth-fix-d5-evidence.md"
RAW_LAT="$WORKDIR/latencies.txt"
STATUS_LOG="$WORKDIR/status.log"
: >"$RAW_LAT"
: >"$STATUS_LOG"

cleanup() {
  docker rm -f "$NAME_A" "$NAME_B" >/dev/null 2>&1 || true
}
trap cleanup EXIT

refuse_unsafe_dsn() {
  local host="$1" port="$2" name="$3"
  local lower
  lower="$(printf '%s' "$host/$name" | tr '[:upper:]' '[:lower:]')"
  if [[ "$name" == "ekarte_db" ]]; then
    echo "REFUSE: DB_NAME ekarte_db is shared app DB" >&2
    exit 2
  fi
  if [[ "$port" == "15432" ]]; then
    echo "REFUSE: port 15432 is reserved/unsafe for this harness" >&2
    exit 2
  fi
  if [[ "$host" == "db" ]]; then
    echo "REFUSE: host 'db' points at compose shared DB; use ae-auth-fix-disposable-pg" >&2
    exit 2
  fi
  if [[ "$host" == "animalekarte-db-1" || "$host" == "old-db-postgres" ]]; then
    echo "REFUSE: host $host is not the disposable auth DB" >&2
    exit 2
  fi
  if [[ "$lower" == *stg* || "$lower" == *prod* || "$lower" == *production* ]]; then
    echo "REFUSE: host/name looks like stg/prod" >&2
    exit 2
  fi
  if [[ "$host" != "ae-auth-fix-disposable-pg" && "$host" != "127.0.0.1" && "$host" != "localhost" ]]; then
    echo "REFUSE: unexpected DB_HOST=$host (allowlist: ae-auth-fix-disposable-pg|127.0.0.1|localhost)" >&2
    exit 2
  fi
  if [[ "$name" != "auth_d1_db" && "$name" != "auth_fix_db" && "$name" != "auth_fix_db_test" ]]; then
    echo "REFUSE: unexpected DB_NAME=$name" >&2
    exit 2
  fi
}

log_status() {
  printf '%s\n' "$*" | tee -a "$STATUS_LOG" >&2
}

require_cmd() {
  command -v "$1" >/dev/null || { echo "missing command: $1" >&2; exit 1; }
}

http_code() {
  # args: cookie_jar url [extra curl args...]
  local jar="$1" url="$2"
  shift 2
  curl -sS -o /dev/null -w '%{http_code}' \
    -b "$jar" -c "$jar" \
    -H 'X-Requested-With: XMLHttpRequest' \
    "$@" "$url"
}

http_code_timed() {
  local jar="$1" url="$2"
  shift 2
  local out code t
  out="$(curl -sS -o /dev/null -w '%{http_code} %{time_total}' \
    -b "$jar" -c "$jar" \
    -H 'X-Requested-With: XMLHttpRequest' \
    "$@" "$url")"
  code="${out%% *}"
  t="${out##* }"
  printf '%s %s\n' "$code" "$t"
}

login() {
  local base="$1" jar="$2" email="$3" password="$4"
  local code
  code="$(curl -sS -o "$WORKDIR/login_body.json" -w '%{http_code}' \
    -c "$jar" -b "$jar" \
    -H 'Content-Type: application/json' \
    -H 'X-Requested-With: XMLHttpRequest' \
    -X POST "$base/api/v1/login" \
    --data "{\"email\":\"$email\",\"password\":\"$password\"}")"
  # Never print body (may contain PII). Record status only.
  log_status "login base=$(basename_host "$base") status=$code"
  [[ "$code" == "200" ]]
}

basename_host() {
  printf '%s' "$1" | sed -E 's#https?://##'
}

percentile() {
  # stdin: floats one per line; arg: percentile 50/95
  # awk implementation for macOS bash 3.2 compatibility
  local p="$1"
  awk -v p="$p" '
    NF { a[++n] = $1 + 0 }
    END {
      if (n == 0) { print "NaN"; exit }
      for (i = 1; i <= n; i++) for (j = i + 1; j <= n; j++) if (a[i] > a[j]) { t = a[i]; a[i] = a[j]; a[j] = t }
      if (p == 50) idx = int((n + 1) / 2)
      else idx = int((n * p + 99) / 100)
      if (idx < 1) idx = 1
      if (idx > n) idx = n
      print a[idx]
    }'
}

measure_endpoint() {
  local label="$1" jar="$2" url="$3"
  shift 3
  local i code t
  local -a times=()
  local errors=0
  for ((i = 1; i <= WARMUP; i++)); do
    read -r code t < <(http_code_timed "$jar" "$url" "$@")
    sleep "$(awk -v ms="$INTERVAL_MS" 'BEGIN{printf "%.3f", ms/1000}')"
  done
  for ((i = 1; i <= MEASURE_N; i++)); do
    read -r code t < <(http_code_timed "$jar" "$url" "$@")
    printf '%s %s %s\n' "$label" "$code" "$t" >>"$RAW_LAT"
    times+=("$t")
    if [[ "$code" != "200" ]]; then
      errors=$((errors + 1))
    fi
    sleep "$(awk -v ms="$INTERVAL_MS" 'BEGIN{printf "%.3f", ms/1000}')"
  done
  local p50 p95
  p50="$(printf '%s\n' "${times[@]}" | percentile 50)"
  p95="$(printf '%s\n' "${times[@]}" | percentile 95)"
  log_status "measure label=$label n=$MEASURE_N errors=$errors p50_s=$p50 p95_s=$p95"
  printf '%s %s %s %s\n' "$label" "$p50" "$p95" "$errors"
}

start_api() {
  local name="$1" host_port="$2"
  local jwt bin
  jwt="$(cat "${D5_JWT_SECRET_FILE:-/tmp/d5_jwt_secret.txt}")"
  bin="${D5_API_BIN:-/tmp/ae-d5-api-main}"
  if [[ ! -x "$bin" ]]; then
    echo "REFUSE: missing API binary at $bin (docker cp from running backend /app/tmp/main)" >&2
    exit 2
  fi
  docker rm -f "$name" >/dev/null 2>&1 || true
  # Bypass image entrypoint (migrate+air). Run extracted API binary only.
  docker run -d --name "$name" --network "$NETWORK" \
    --entrypoint /d5/api \
    -v "$bin:/d5/api:ro" \
    -p "127.0.0.1:${host_port}:8080" \
    -e APP_ENV=development \
    -e GIN_MODE=debug \
    -e PORT=8080 \
    -e DB_HOST="$DB_HOST" \
    -e DB_PORT="$DB_PORT" \
    -e DB_USER="$DB_USER" \
    -e DB_PASSWORD="$DB_PASSWORD" \
    -e DB_NAME="$DB_NAME" \
    -e DB_SSL_MODE="$DB_SSL_MODE" \
    -e JWT_SECRET="$jwt" \
    -e COOKIE_DOMAIN=localhost \
    -e COOKIE_CROSS_DOMAIN=false \
    -e CORS_ALLOWED_ORIGIN=http://localhost:3003 \
    -e LIFF_MOCK=true \
    -e LOG_LEVEL=error \
    "$API_IMAGE" >/dev/null
}

wait_health() {
  local base="$1" i code
  for ((i = 1; i <= 60; i++)); do
    code="$(curl -sS -o /dev/null -w '%{http_code}' "$base/health" || true)"
    if [[ "$code" == "200" ]]; then
      log_status "health ok base=$(basename_host "$base")"
      return 0
    fi
    sleep 1
  done
  log_status "health FAIL base=$(basename_host "$base") last=$code"
  return 1
}

restore_victim_active() {
  docker exec "$PG_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 \
    -c "UPDATE staffs SET is_active=true, updated_at=now() WHERE id=$VICTIM_STAFF_ID;" >/dev/null
}

restore_actor_assignments() {
  docker exec -i "$PG_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<SQL >/dev/null
INSERT INTO staff_clinic_assignments (staff_id, clinic_id, is_main)
SELECT $ACTOR_STAFF_ID, c.id, (c.id = 1)
FROM clinics c
WHERE c.id IN (1,2,3,4)
ON CONFLICT (staff_id, clinic_id) DO UPDATE
SET deleted_at = NULL, updated_at = now(), is_main = EXCLUDED.is_main;
UPDATE staff_clinic_assignments
SET deleted_at = NULL, updated_at = now(), is_main = (clinic_id = 1)
WHERE staff_id = $ACTOR_STAFF_ID AND clinic_id IN (1,2,3,4);
SQL
}

restore_victim_password_demo() {
  # Re-hash SharedPassword onto victim account so later cases stay usable.
  # Hash generated offline in disposable path only.
  local hash
  hash="$(docker exec animalekarte-backend-1 sh -c 'go run /tmp/genhash.go "'"$DEMO_PASSWORD"'"' 2>/dev/null || true)"
  if [[ ${#hash} -ne 60 ]]; then
    docker exec -w /app animalekarte-backend-1 sh -c \
      'go run /tmp/genhash.go "'"$DEMO_PASSWORD"'" > /tmp/d5_demo_hash.txt'
    hash="$(docker exec animalekarte-backend-1 cat /tmp/d5_demo_hash.txt)"
  fi
  docker exec -i "$PG_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<SQL >/dev/null
UPDATE accounts a
SET password_hash = '$hash', updated_at = now()
FROM staffs s
WHERE s.account_id = a.id AND s.id = $VICTIM_STAFF_ID;
SQL
}

main() {
  require_cmd docker
  require_cmd curl
  require_cmd awk
  require_cmd openssl

  refuse_unsafe_dsn "$DB_HOST" "$DB_PORT" "$DB_NAME"

  if [[ ! -f "${D5_JWT_SECRET_FILE:-/tmp/d5_jwt_secret.txt}" ]]; then
    openssl rand -hex 32 >"${D5_JWT_SECRET_FILE:-/tmp/d5_jwt_secret.txt}"
  fi

  # Confirm disposable container + DB identity (non-secret).
  local db_check
  db_check="$(docker exec "$PG_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -Atc \
    "SELECT current_database()||'|'||session_user||'|'||count(*)::text FROM accounts")"
  log_status "db_identity=$db_check container=$PG_CONTAINER host=$DB_HOST port=$DB_PORT"

  start_api "$NAME_A" "$PORT_A"
  start_api "$NAME_B" "$PORT_B"
  local BASE_A="http://127.0.0.1:${PORT_A}"
  local BASE_B="http://127.0.0.1:${PORT_B}"
  if ! wait_health "$BASE_A"; then
    docker logs "$NAME_A" 2>&1 | tail -50 | tee -a "$STATUS_LOG" || true
    exit 1
  fi
  if ! wait_health "$BASE_B"; then
    docker logs "$NAME_B" 2>&1 | tail -50 | tee -a "$STATUS_LOG" || true
    exit 1
  fi

  local JAR_A="$WORKDIR/jar_a.txt"
  local JAR_B="$WORKDIR/jar_b.txt"
  local JAR_ACTOR="$WORKDIR/jar_actor.txt"
  : >"$JAR_A"; : >"$JAR_B"; : >"$JAR_ACTOR"

  restore_victim_active
  restore_actor_assignments
  restore_victim_password_demo

  # --- Baseline latency: 1 process ---
  login "$BASE_A" "$JAR_A" "$VICTIM_EMAIL" "$DEMO_PASSWORD"
  local m1
  m1="$(measure_endpoint baseline_1p "$JAR_A" "$BASE_A/api/v1/me")"
  measure_endpoint baseline_1p_clinic "$JAR_A" "$BASE_A$CLINIC_FIXED_PATH" \
    -H "X-Clinic-ID: $CLINIC_KEEP" >/dev/null

  # --- Dual process latency ---
  login "$BASE_B" "$JAR_B" "$VICTIM_EMAIL" "$DEMO_PASSWORD"
  local m2a m2b
  m2a="$(measure_endpoint dual_a "$JAR_A" "$BASE_A/api/v1/me")"
  m2b="$(measure_endpoint dual_b "$JAR_B" "$BASE_B/api/v1/me")"

  # --- Case: staff deactivate immediate revoke ---
  restore_victim_active
  : >"$JAR_A"; : >"$JAR_B"; : >"$JAR_ACTOR"
  login "$BASE_A" "$JAR_A" "$VICTIM_EMAIL" "$DEMO_PASSWORD"
  login "$BASE_B" "$JAR_B" "$VICTIM_EMAIL" "$DEMO_PASSWORD"
  login "$BASE_A" "$JAR_ACTOR" "$ACTOR_EMAIL" "$DEMO_PASSWORD"
  local pre_a pre_b deact_code post_a post_b
  pre_a="$(http_code "$JAR_A" "$BASE_A/api/v1/me")"
  pre_b="$(http_code "$JAR_B" "$BASE_B/api/v1/me")"
  deact_code="$(curl -sS -o /dev/null -w '%{http_code}' \
    -b "$JAR_ACTOR" -c "$JAR_ACTOR" \
    -H 'Content-Type: application/json' \
    -H 'X-Requested-With: XMLHttpRequest' \
    -H "X-Clinic-ID: $CLINIC_KEEP" \
    -X PATCH "$BASE_A/api/v1/masters/staffs/$VICTIM_STAFF_ID" \
    --data '{"is_active":false}')"
  post_a="$(http_code "$JAR_A" "$BASE_A/api/v1/me")"
  post_b="$(http_code "$JAR_B" "$BASE_B/api/v1/me")"
  local clinic_post_a clinic_post_b
  clinic_post_a="$(http_code "$JAR_A" "$BASE_A$CLINIC_FIXED_PATH" -H "X-Clinic-ID: $CLINIC_KEEP")"
  clinic_post_b="$(http_code "$JAR_B" "$BASE_B$CLINIC_FIXED_PATH" -H "X-Clinic-ID: $CLINIC_KEEP")"
  log_status "deactivate pre_a=$pre_a pre_b=$pre_b deact=$deact_code post_a=$post_a post_b=$post_b clinic_a=$clinic_post_a clinic_b=$clinic_post_b"
  local deactivate_pass=0
  if [[ "$pre_a" == "200" && "$pre_b" == "200" && "$deact_code" == "200" \
    && ( "$post_a" == "401" || "$post_a" == "403" ) \
    && ( "$post_b" == "401" || "$post_b" == "403" ) \
    && ( "$clinic_post_a" == "401" || "$clinic_post_a" == "403" ) \
    && ( "$clinic_post_b" == "401" || "$clinic_post_b" == "403" ) ]]; then
    deactivate_pass=1
  fi
  if (( deactivate_pass == 0 )); then
    log_status "STOP: deactivate immediate revoke failed"
  fi
  restore_victim_active

  # --- Case: unassign clinic; keep other clinic ---
  restore_actor_assignments
  : >"$JAR_A"; : >"$JAR_B"; : >"$JAR_ACTOR"
  login "$BASE_A" "$JAR_A" "$ACTOR_EMAIL" "$DEMO_PASSWORD"
  login "$BASE_B" "$JAR_B" "$ACTOR_EMAIL" "$DEMO_PASSWORD"
  local u_pre_drop u_pre_keep
  u_pre_drop="$(http_code "$JAR_A" "$BASE_A$CLINIC_FIXED_PATH" -H "X-Clinic-ID: $CLINIC_DROP")"
  u_pre_keep="$(http_code "$JAR_A" "$BASE_A$CLINIC_FIXED_PATH" -H "X-Clinic-ID: $CLINIC_KEEP")"
  local unassign_code
  # Keep clinics 1,3,4 — drop 2
  unassign_code="$(curl -sS -o /dev/null -w '%{http_code}' \
    -b "$JAR_A" -c "$JAR_A" \
    -H 'Content-Type: application/json' \
    -H 'X-Requested-With: XMLHttpRequest' \
    -H "X-Clinic-ID: $CLINIC_KEEP" \
    -X PUT "$BASE_A/api/v1/masters/staffs/$ACTOR_STAFF_ID/clinics" \
    --data '{"clinic_ids":[1,3,4]}')"
  local u_post_drop_a u_post_keep_a u_post_drop_b u_post_keep_b u_me_a
  u_post_drop_a="$(http_code "$JAR_A" "$BASE_A$CLINIC_FIXED_PATH" -H "X-Clinic-ID: $CLINIC_DROP")"
  u_post_keep_a="$(http_code "$JAR_A" "$BASE_A$CLINIC_FIXED_PATH" -H "X-Clinic-ID: $CLINIC_KEEP")"
  u_post_drop_b="$(http_code "$JAR_B" "$BASE_B$CLINIC_FIXED_PATH" -H "X-Clinic-ID: $CLINIC_DROP")"
  u_post_keep_b="$(http_code "$JAR_B" "$BASE_B$CLINIC_FIXED_PATH" -H "X-Clinic-ID: $CLINIC_KEEP")"
  u_me_a="$(http_code "$JAR_A" "$BASE_A/api/v1/me")"
  log_status "unassign pre_drop=$u_pre_drop pre_keep=$u_pre_keep code=$unassign_code post_drop_a=$u_post_drop_a post_keep_a=$u_post_keep_a post_drop_b=$u_post_drop_b post_keep_b=$u_post_keep_b me_a=$u_me_a"
  local unassign_pass=0
  if [[ "$u_pre_drop" == "200" && "$u_pre_keep" == "200" && "$unassign_code" == "200" \
    && ( "$u_post_drop_a" == "401" || "$u_post_drop_a" == "403" ) \
    && ( "$u_post_drop_b" == "401" || "$u_post_drop_b" == "403" ) \
    && "$u_post_keep_a" == "200" && "$u_post_keep_b" == "200" \
    && "$u_me_a" == "200" ]]; then
    unassign_pass=1
  fi
  if (( unassign_pass == 0 )); then
    log_status "STOP: unassign immediate revoke / keep-clinic assert failed"
  fi
  restore_actor_assignments

  # --- Case: password / epoch update ---
  restore_victim_active
  restore_victim_password_demo
  : >"$JAR_A"; : >"$JAR_B"
  login "$BASE_A" "$JAR_A" "$VICTIM_EMAIL" "$DEMO_PASSWORD"
  login "$BASE_B" "$JAR_B" "$VICTIM_EMAIL" "$DEMO_PASSWORD"
  local e_pre_a e_pre_b pw_code e_post_a e_post_b re_login
  e_pre_a="$(http_code "$JAR_A" "$BASE_A/api/v1/me")"
  e_pre_b="$(http_code "$JAR_B" "$BASE_B/api/v1/me")"
  pw_code="$(curl -sS -o /dev/null -w '%{http_code}' \
    -b "$JAR_A" -c "$JAR_A" \
    -H 'Content-Type: application/json' \
    -H 'X-Requested-With: XMLHttpRequest' \
    -H "X-Clinic-ID: $CLINIC_KEEP" \
    -X PUT "$BASE_A/api/v1/users/me/password" \
    --data "{\"current_password\":\"$DEMO_PASSWORD\",\"new_password\":\"$NEW_PASSWORD\"}")"
  e_post_a="$(http_code "$JAR_A" "$BASE_A/api/v1/me")"
  e_post_b="$(http_code "$JAR_B" "$BASE_B/api/v1/me")"
  : >"$WORKDIR/jar_relogin.txt"
  if login "$BASE_A" "$WORKDIR/jar_relogin.txt" "$VICTIM_EMAIL" "$NEW_PASSWORD"; then
    re_login=200
  else
    re_login=fail
  fi
  local relogin_me
  relogin_me="$(http_code "$WORKDIR/jar_relogin.txt" "$BASE_A/api/v1/me" || true)"
  log_status "epoch pre_a=$e_pre_a pre_b=$e_pre_b pw=$pw_code post_a=$e_post_a post_b=$e_post_b relogin=$re_login relogin_me=$relogin_me"
  local epoch_pass=0
  if [[ "$e_pre_a" == "200" && "$e_pre_b" == "200" && "$pw_code" == "200" \
    && ( "$e_post_a" == "401" || "$e_post_a" == "403" ) \
    && ( "$e_post_b" == "401" || "$e_post_b" == "403" ) \
    && "$re_login" == "200" && "$relogin_me" == "200" ]]; then
    epoch_pass=1
  fi
  restore_victim_password_demo
  restore_victim_active

  # Compose evidence (redacted)
  local cid_a cid_b
  cid_a="$(docker inspect -f '{{.Id}}' "$NAME_A" | cut -c1-12)"
  cid_b="$(docker inspect -f '{{.Id}}' "$NAME_B" | cut -c1-12)"

  {
    echo "# D5 Dual-Process Immediate Revoke Evidence"
    echo
    echo "Date: $(date '+%Y-%m-%d %H:%M:%S %z')"
    echo "Worktree: \`$ROOT\`"
    echo "DB: container=\`$PG_CONTAINER\` name=\`$DB_NAME\` host=\`$DB_HOST\` port=\`$DB_PORT\` identity=\`$db_check\`"
    echo "API image: \`$API_IMAGE\`"
    echo "Processes: A=\`$NAME_A\` id=\`$cid_a\` host_port=\`$PORT_A\`; B=\`$NAME_B\` id=\`$cid_b\` host_port=\`$PORT_B\`"
    echo "Endpoints: \`GET /api/v1/me\`, clinic-fixed \`$CLINIC_FIXED_PATH\`"
    echo "Load: warmup=$WARMUP measure=$MEASURE_N interval_ms=$INTERVAL_MS concurrency=1"
    echo "DB statements/request: **UNKNOWN** (pg_stat_statements unavailable / not queried)"
    echo
    echo "## Safety gate"
    echo
    echo "- Refused ekarte_db / host db / 15432 / stg|prod names"
    echo "- Connected only to disposable allowlisted DB"
    echo
    echo "## Latency (server curl time_total seconds)"
    echo
    echo "| label | p50_s | p95_s | errors |"
    echo "| --- | ---: | ---: | ---: |"
    echo "$m1" | awk '{printf "| %s | %s | %s | %s |\n",$1,$2,$3,$4}'
    echo "$m2a" | awk '{printf "| %s | %s | %s | %s |\n",$1,$2,$3,$4}'
    echo "$m2b" | awk '{printf "| %s | %s | %s | %s |\n",$1,$2,$3,$4}'
    echo
    echo "## Immediate revoke cases"
    echo
    echo "| case | pass | notes |"
    echo "| --- | --- | --- |"
    echo "| staff deactivate | $([[ $deactivate_pass -eq 1 ]] && echo PASS || echo FAIL) | pre_a=$pre_a pre_b=$pre_b deact=$deact_code post_a=$post_a post_b=$post_b clinic_a=$clinic_post_a clinic_b=$clinic_post_b |"
    echo "| clinic unassign | $([[ $unassign_pass -eq 1 ]] && echo PASS || echo FAIL) | drop clinic $CLINIC_DROP denied; keep clinic $CLINIC_KEEP allowed; unassign=$unassign_code |"
    echo "| password/epoch | $([[ $epoch_pass -eq 1 ]] && echo PASS || echo FAIL) | pw=$pw_code post_a=$e_post_a post_b=$e_post_b relogin_me=$relogin_me |"
    echo
    echo "## Overall"
    echo
    if (( deactivate_pass == 1 && unassign_pass == 1 && epoch_pass == 1 )); then
      echo "**D5 status: PASS**"
    else
      echo "**D5 status: FAIL** (see case table). Harness did not wait for TTL; failures are immediate-check results."
    fi
    echo
    echo "Raw latency file (local workdir, not committed secrets): \`$RAW_LAT\`"
    echo "Status log: \`$STATUS_LOG\`"
  } >"$EVIDENCE"

  log_status "evidence=$EVIDENCE"
  if (( deactivate_pass == 1 && unassign_pass == 1 && epoch_pass == 1 )); then
    exit 0
  fi
  exit 1
}

main "$@"
