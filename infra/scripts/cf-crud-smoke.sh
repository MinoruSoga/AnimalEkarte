#!/usr/bin/env bash
# P4-6(試行11): Cloudflare Workers/Containers 経由 STG API に対する CRUD スモーク +
# 混在会計(payment_splits) API スモーク。PlanetScale 接続・Cookie 認証・主要 CRUD 導線・
# payment_splits が Container 経由で動作することを curl + jq で検証する。
#
# 参考: infra/scripts/cf-run-migrate.sh と同パターン（環境変数必須チェック・タイムアウト・
# exit code 契約・シークレット非出力）。
#
# 前提:
#   - STG_DEMO_EMAIL / STG_DEMO_PASSWORD 環境変数（system_admin デモアカウント）
#   - WORKER_URL 環境変数（省略時は試行9/10確定のworkers.dev URLを既定値とする）
#
# 使い方:
#   STG_DEMO_EMAIL=xxx STG_DEMO_PASSWORD=xxx ./infra/scripts/cf-crud-smoke.sh
#
# 注意: 本スクリプトは live STG API に対して実際に CRUD/PATCH を実行する（PlanetScale
# データ変更を伴う）。DB_RESET・secret ローテーション・wrangler deploy は一切行わない。

# NOTE: -e は意図的に外している。全 AC を継続実行し FAIL_COUNT で最終判定するため、
# 個々の curl/jq が非0を返しても後続 AC の検証を止めない設計（code-reviewer指摘 L-1）。
set -uo pipefail

WORKER_URL="${WORKER_URL:-https://animalekarte-stg-api.baritech-soga.workers.dev}"
API_V1="${WORKER_URL}/api/v1"

if [[ -z "${STG_DEMO_EMAIL:-}" || -z "${STG_DEMO_PASSWORD:-}" ]]; then
  echo "::error::STG_DEMO_EMAIL / STG_DEMO_PASSWORD が未設定です" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "::error::jq が見つかりません（JSON応答の検証に必須）" >&2
  exit 1
fi

HEALTH_TIMEOUT=15
API_TIMEOUT=30

COOKIE_JAR="$(mktemp)"
BAD_COOKIE_JAR="$(mktemp)"
cleanup_jars() {
  rm -f "${COOKIE_JAR}" "${BAD_COOKIE_JAR}"
}
trap cleanup_jars EXIT

# ---- 結果集計 -----------------------------------------------------------
AC_LOG=()
FAIL_COUNT=0

record() {
  local ac="$1" result="$2" note="$3"
  AC_LOG+=("[${result}] ${ac}: ${note}")
  echo "[${result}] ${ac}: ${note}"
  if [[ "${result}" == "FAIL" ]]; then
    FAIL_COUNT=$((FAIL_COUNT + 1))
  fi
}

# ---- HTTP helper ---------------------------------------------------------
# call_api METHOD PATH [JSON_BODY] [COOKIE_JAR]
# 結果は RESP_CODE / RESP_BODY にセットされる。
RESP_CODE=""
RESP_BODY=""
call_api() {
  local method="$1" path="$2" body="${3:-}" jar="${4:-${COOKIE_JAR}}"
  local args=(-s --max-time "${API_TIMEOUT}" -w '\n%{http_code}' -b "${jar}" -c "${jar}" -X "${method}" "${API_V1}${path}" -H "Accept: application/json")
  case "${method}" in
    GET|HEAD) ;;
    *) args+=(-H "X-Requested-With: XMLHttpRequest") ;;
  esac
  if [[ -n "${body}" ]]; then
    args+=(-H "Content-Type: application/json" --data "${body}")
  fi
  local raw curl_exit attempt=0
  while :; do
    raw="$(curl "${args[@]}")"
    curl_exit=$?
    if [[ "${curl_exit}" -ne 0 ]]; then
      # code-reviewer指摘 M-1: curl自体の失敗(タイムアウト/DNS等)とAPI異常を区別できるようにする
      RESP_CODE="curl_error_${curl_exit}"
      RESP_BODY=""
      return
    fi
    RESP_CODE="$(printf '%s' "${raw}" | tail -n1)"
    RESP_BODY="$(printf '%s' "${raw}" | sed '$d')"
    # デプロイ直後は Cloudflare Containers の非同期ロールアウトがスモーク実行中に
    # インスタンスを入れ替えることがあり、遷移中は 500 "container is not running" が返る
    # (2026-07-17 実測・数秒〜数十秒で自己回復)。この場合のみ最大5回リトライする。
    if [[ "${RESP_CODE}" == "500" ]] && printf '%s' "${RESP_BODY}" | grep -q "container is not running" && (( attempt < 5 )); then
      attempt=$((attempt + 1))
      echo "    (container rollout 遷移中 — retry ${attempt}/5 in 10s)"
      sleep 10
      continue
    fi
    return
  done
}

body_preview() {
  # エラー時のみ本文の先頭を短く見せる（PII/秘匿情報が長文で残らないよう300文字で切る）
  printf '%s' "${RESP_BODY}" | head -c 300
}

# =========================================================================
# AC-0: Pre-flight health
# =========================================================================
echo "==> AC-0: GET /health"
HEALTH_RAW="$(curl -s --max-time "${HEALTH_TIMEOUT}" -w '\n%{http_code}' "${WORKER_URL}/health")"
HEALTH_CODE="$(printf '%s' "${HEALTH_RAW}" | tail -n1)"
HEALTH_BODY="$(printf '%s' "${HEALTH_RAW}" | sed '$d')"
if [[ "${HEALTH_CODE}" == "200" ]] && printf '%s' "${HEALTH_BODY}" | jq -e '.status == "ok"' >/dev/null 2>&1; then
  record "AC-0" PASS "GET /health -> 200 ${HEALTH_BODY}"
else
  record "AC-0" FAIL "GET /health -> ${HEALTH_CODE} ${HEALTH_BODY}"
fi

# =========================================================================
# AC-1b: Login failure on bad creds（先に失敗系を確認 — 正規cookieを汚さない）
# =========================================================================
echo "==> AC-1b: POST /login with wrong password"
WRONG_PASSWORD="${STG_DEMO_PASSWORD}_INVALID_$(date +%s)"
BAD_LOGIN_PAYLOAD="$(jq -nc --arg email "${STG_DEMO_EMAIL}" --arg password "${WRONG_PASSWORD}" '{email:$email,password:$password}')"
call_api POST "/login" "${BAD_LOGIN_PAYLOAD}" "${BAD_COOKIE_JAR}"
BAD_LOGIN_CODE="${RESP_CODE}"
if [[ "${BAD_LOGIN_CODE}" == "401" || "${BAD_LOGIN_CODE}" == "403" ]]; then
  call_api GET "/me" "" "${BAD_COOKIE_JAR}"
  if [[ "${RESP_CODE}" == "401" ]]; then
    record "AC-1b" PASS "wrong password -> ${BAD_LOGIN_CODE}, 後続 /me -> 401"
  else
    record "AC-1b" FAIL "wrong password -> ${BAD_LOGIN_CODE} だが後続 /me が ${RESP_CODE}（401期待）"
  fi
else
  record "AC-1b" FAIL "wrong password -> ${BAD_LOGIN_CODE}（401/403期待） body=$(body_preview)"
fi

# =========================================================================
# AC-1: Seed / login readiness
# =========================================================================
echo "==> AC-1: POST /login (correct creds) + GET /me"
LOGIN_PAYLOAD="$(jq -nc --arg email "${STG_DEMO_EMAIL}" --arg password "${STG_DEMO_PASSWORD}" '{email:$email,password:$password}')"
call_api POST "/login" "${LOGIN_PAYLOAD}"
LOGIN_CODE="${RESP_CODE}"
LOGIN_OK=false
if [[ "${LOGIN_CODE}" == "200" || "${LOGIN_CODE}" == "201" ]] && printf '%s' "${RESP_BODY}" | jq -e '.user.id' >/dev/null 2>&1; then
  if grep -qi 'access_token' "${COOKIE_JAR}" 2>/dev/null; then
    LOGIN_OK=true
  fi
fi

if [[ "${LOGIN_OK}" == "true" ]]; then
  IS_SYSTEM_ADMIN="$(printf '%s' "${RESP_BODY}" | jq -r '.is_system_admin')"
  call_api GET "/me"
  if [[ "${RESP_CODE}" == "200" ]] && printf '%s' "${RESP_BODY}" | jq -e '.id' >/dev/null 2>&1; then
    record "AC-1" PASS "login -> ${LOGIN_CODE} (is_system_admin=${IS_SYSTEM_ADMIN}) + Cookie付き /me -> 200"
  else
    record "AC-1" FAIL "login成功したが /me が ${RESP_CODE} body=$(body_preview)"
  fi
else
  record "AC-1" FAIL "login -> ${LOGIN_CODE}（200 + Set-Cookie access_token 期待） body=$(body_preview)"
fi

# 以降、AC-1 が FAIL の場合は認証必須の全 AC が実行不能なため BLOCKED として記録して終了する。
if [[ "${LOGIN_OK}" != "true" ]]; then
  for ac in AC-2 AC-3 AC-4 AC-5 AC-6 AC-7 AC-8; do
    record "${ac}" FAIL "AC-1(login)失敗のため未実行"
  done
  record "AC-11" BLOCKED "UI混在会計(§3〜§10)は本試行スコープ外(frontendはAWS API向きのため) — P7-3へdefer"
  echo "=========================================="
  echo "SUMMARY (FAIL=${FAIL_COUNT})"
  printf '%s\n' "${AC_LOG[@]}"
  exit 1
fi

# =========================================================================
# AC-2: CRUD A-1 clinics list
# =========================================================================
echo "==> AC-2: GET /clinics"
call_api GET "/clinics"
# code-reviewer指摘 L-3: 配列/{"data":[...]}どちらの形式でも件数を取れるようにする
CLINIC_COUNT="$(printf '%s' "${RESP_BODY}" | jq 'if type=="array" then length else (.data // [] | length) end' 2>/dev/null || echo 0)"
if [[ "${RESP_CODE}" == "200" ]] && [[ "${CLINIC_COUNT}" -ge 1 ]]; then
  record "AC-2" PASS "GET /clinics -> 200, ${CLINIC_COUNT} clinics"
else
  record "AC-2" FAIL "GET /clinics -> ${RESP_CODE} body=$(body_preview)"
fi

# =========================================================================
# AC-3: CRUD B-1/B-2/B-3 permission groups
# =========================================================================
echo "==> AC-3: permission-groups POST -> protected DELETE 409 -> test DELETE 204"
# PIDを付与し同一秒の並行実行でもリソース名が衝突しないようにする（code-reviewer指摘 L-4）
TS="$(date +%s)-$$"
PG_PAYLOAD="$(jq -nc --arg name "TEST-GROUP-${TS}" '{name:$name,description:"Smoke test temporary group (P4-6 trial11)",color:"#0000FF",is_active:true,sort_order:99}')"
call_api POST "/masters/permission-groups" "${PG_PAYLOAD}"
PG_CREATE_CODE="${RESP_CODE}"
TEST_GROUP_ID=""
if [[ "${PG_CREATE_CODE}" == "201" ]]; then
  TEST_GROUP_ID="$(printf '%s' "${RESP_BODY}" | jq -r '.id')"
fi

if [[ -n "${TEST_GROUP_ID}" && "${TEST_GROUP_ID}" != "null" ]]; then
  # NOTE(code-reviewer指摘 L-2): id=1(執行グループ/clinic_id=1)がSTG seedで
  # staff_id=1割当済み(使用中)であることを前提に409を期待している。seedデータが
  # 変わると本チェックは機能しなくなる — 003_seed_demo.sql 7d節 staff_permission_groups 依存。
  call_api DELETE "/masters/permission-groups/1"
  PROTECTED_DELETE_CODE="${RESP_CODE}"

  call_api DELETE "/masters/permission-groups/${TEST_GROUP_ID}"
  TEST_DELETE_CODE="${RESP_CODE}"

  if [[ "${PROTECTED_DELETE_CODE}" == "409" && "${TEST_DELETE_CODE}" == "204" ]]; then
    record "AC-3" PASS "POST -> 201 (id=${TEST_GROUP_ID}), DELETE id=1 -> 409, DELETE test -> 204"
  else
    record "AC-3" FAIL "POST -> 201 (id=${TEST_GROUP_ID}) だが DELETE id=1 -> ${PROTECTED_DELETE_CODE}（409期待）/ DELETE test -> ${TEST_DELETE_CODE}（204期待）"
  fi
else
  record "AC-3" FAIL "POST /masters/permission-groups -> ${PG_CREATE_CODE} body=$(body_preview)"
fi

# =========================================================================
# AC-4: CRUD C-1/C-4 staff
# =========================================================================
echo "==> AC-4: staffs POST -> DELETE"
STAFF_PAYLOAD="$(jq -nc --arg name "TEST-STAFF-${TS}" '{name:$name,sort_order:99}')"
call_api POST "/masters/staffs" "${STAFF_PAYLOAD}"
STAFF_CREATE_CODE="${RESP_CODE}"
TEST_STAFF_ID=""
if [[ "${STAFF_CREATE_CODE}" == "201" ]]; then
  TEST_STAFF_ID="$(printf '%s' "${RESP_BODY}" | jq -r '.id')"
fi

if [[ -n "${TEST_STAFF_ID}" && "${TEST_STAFF_ID}" != "null" ]]; then
  call_api DELETE "/masters/staffs/${TEST_STAFF_ID}"
  STAFF_DELETE_CODE="${RESP_CODE}"
  if [[ "${STAFF_DELETE_CODE}" == "204" ]]; then
    record "AC-4" PASS "POST -> 201 (id=${TEST_STAFF_ID}), DELETE -> 204"
  else
    record "AC-4" FAIL "POST -> 201 (id=${TEST_STAFF_ID}) だが DELETE -> ${STAFF_DELETE_CODE}（204期待）"
  fi
else
  record "AC-4" FAIL "POST /masters/staffs -> ${STAFF_CREATE_CODE} body=$(body_preview)"
fi

# =========================================================================
# AC-5 / AC-7: waiting 会計を探索（id=3 優先、無ければ status=waiting を検索）
# =========================================================================
echo "==> waiting billing 探索（AC-5/AC-7用）"
BILLING_ID=""
BILLING_AMOUNT=""
call_api GET "/accountings/3"
if [[ "${RESP_CODE}" == "200" ]] && [[ "$(printf '%s' "${RESP_BODY}" | jq -r '.status')" == "waiting" ]]; then
  BILLING_ID="3"
  BILLING_AMOUNT="$(printf '%s' "${RESP_BODY}" | jq -r '.total_amount')"
else
  call_api GET "/accountings?status=waiting&limit=5"
  if [[ "${RESP_CODE}" == "200" ]]; then
    BILLING_ID="$(printf '%s' "${RESP_BODY}" | jq -r '.data[0].id // empty')"
    BILLING_AMOUNT="$(printf '%s' "${RESP_BODY}" | jq -r '.data[0].total_amount // empty')"
  fi
fi

if [[ -z "${BILLING_ID}" || "${BILLING_ID}" == "null" ]]; then
  # フルデモ等、waiting 会計を含まないデータセットではフィクスチャ前提が満たされないため
  # FAIL ではなく SKIP とする（2026-07-17: フルデモは pending/completed のみで waiting 0 件を実測）。
  # 新規作成→混在精算の実検証は AC-6 が担う。
  record "AC-5" SKIP "waiting会計がデータセットに存在しないためSKIP（AC-6で新規作成系は検証済み）"
  record "AC-7" SKIP "waiting会計がデータセットに存在しないためSKIP"
else
  echo "    waiting billing: id=${BILLING_ID} amount=${BILLING_AMOUNT}"

  # ---- AC-7: バリデーション — 重複method（waiting会計を完了させない） ----
  echo "==> AC-7: PATCH /accountings/${BILLING_ID} with duplicate method splits (invalid)"
  DUP_PAYLOAD="$(jq -nc --argjson amt "${BILLING_AMOUNT}" '{payment_splits:[{method:"cash",amount:$amt,received_amount:$amt,change_amount:0},{method:"cash",amount:$amt,received_amount:$amt,change_amount:0}]}')"
  call_api PATCH "/accountings/${BILLING_ID}" "${DUP_PAYLOAD}"
  DUP_CODE="${RESP_CODE}"
  # NOTE(spec-implementation差異): MIXED-PAYMENT-SMOKE-TEST.md §5 は HTTP 422 を期待するが、
  # 実装(backend/internal/service/accounting_service_builders.go validatePaymentSplits ->
  # apperrors.WrapInvalidInput -> handler/response.go RespondError)は HTTP 400 を返す。
  # これは本トライアル(Cloudflare移行)に起因しない既存のdoc drift。重複検出自体が正しく
  # 機能していること(4xxで会計が completed にならないこと)を本質的な合格条件とする。
  if [[ "${DUP_CODE}" == "422" || "${DUP_CODE}" == "400" ]]; then
    call_api GET "/accountings/${BILLING_ID}"
    STILL_WAITING="$(printf '%s' "${RESP_BODY}" | jq -r '.status')"
    if [[ "${STILL_WAITING}" == "waiting" ]]; then
      NOTE="重複method -> HTTP ${DUP_CODE}, 会計は waiting のまま(未完了)"
      if [[ "${DUP_CODE}" == "422" ]]; then
        record "AC-7" PASS "${NOTE}"
      else
        record "AC-7" PASS "${NOTE} [NOTE: doc期待値422だが実装は400 — spec-implementation差異、既存doc drift]"
      fi
    else
      record "AC-7" FAIL "重複method -> HTTP ${DUP_CODE} だが会計status=${STILL_WAITING}（waiting維持を期待）"
    fi
  else
    record "AC-7" FAIL "重複method -> HTTP ${DUP_CODE}（400/422期待） body=$(body_preview)"
  fi

  # ---- AC-5: 単一支払い（現金のみ） ----
  echo "==> AC-5: PATCH /accountings/${BILLING_ID} single cash payment"
  SINGLE_PAYLOAD="$(jq -nc --argjson amt "${BILLING_AMOUNT}" '{status:"completed",billing_amount:$amt,received_amount:$amt,change_amount:0,payment_splits:[{method:"cash",amount:$amt,received_amount:$amt,change_amount:0}]}')"
  call_api PATCH "/accountings/${BILLING_ID}" "${SINGLE_PAYLOAD}"
  SINGLE_PATCH_CODE="${RESP_CODE}"
  if [[ "${SINGLE_PATCH_CODE}" == "200" ]]; then
    call_api GET "/accountings/${BILLING_ID}"
    SPLIT_LEN="$(printf '%s' "${RESP_BODY}" | jq '.payment_splits | length')"
    SPLIT_METHOD="$(printf '%s' "${RESP_BODY}" | jq -r '.payment_splits[0].method')"
    SPLIT_AMOUNT="$(printf '%s' "${RESP_BODY}" | jq -r '.payment_splits[0].amount')"
    if [[ "${SPLIT_LEN}" == "1" && "${SPLIT_METHOD}" == "cash" && "${SPLIT_AMOUNT}" == "${BILLING_AMOUNT}" ]]; then
      record "AC-5" PASS "PATCH -> 200, GET payment_splits=[cash:${SPLIT_AMOUNT}] (billing_amount=${BILLING_AMOUNT})"
    else
      record "AC-5" FAIL "PATCH -> 200 だが GET payment_splits 不整合 (len=${SPLIT_LEN} method=${SPLIT_METHOD} amount=${SPLIT_AMOUNT})"
    fi
  else
    record "AC-5" FAIL "PATCH /accountings/${BILLING_ID} (single cash) -> ${SINGLE_PATCH_CODE} body=$(body_preview)"
  fi
fi

# =========================================================================
# AC-6: 2種混在 — 新規会計を作成して独立に検証（seedのwaiting会計を消費しない）
# =========================================================================
echo "==> AC-6: POST /accountings (new waiting) -> PATCH 2-way split"
SCHEDULED_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
NEW_BILLING_AMOUNT=1100
# 実在の pet/owner を動的取得する。旧実装の owner_id=2/pet_id=3 固定値は小デモの ID 体系前提で、
# フルデモ等 ID 体系が異なるデータセットでは 404 になる（2026-07-17 実測）。
call_api GET "/pets?limit=1"
SMOKE_PET_ID="$(printf '%s' "${RESP_BODY}" | jq -r '.data[0].id // empty' 2>/dev/null || true)"
SMOKE_OWNER_ID="$(printf '%s' "${RESP_BODY}" | jq -r '.data[0].owner_id // empty' 2>/dev/null || true)"
CREATE_CODE=""
NEW_BILLING_ID=""
if [[ -n "${SMOKE_PET_ID}" && -n "${SMOKE_OWNER_ID}" ]]; then
  CREATE_PAYLOAD="$(jq -nc --argjson owner "${SMOKE_OWNER_ID}" --argjson pet "${SMOKE_PET_ID}" --argjson total "${NEW_BILLING_AMOUNT}" --arg date "${SCHEDULED_DATE}" '{owner_id:$owner,pet_id:$pet,subtotal:1000,tax_total:100,total_amount:$total,has_insurance:false,status:"waiting",scheduled_date:$date,memo:"P4-6 trial11 smoke (mixed payment two-way split)"}')"
  call_api POST "/accountings" "${CREATE_PAYLOAD}"
  CREATE_CODE="${RESP_CODE}"
  if [[ "${CREATE_CODE}" == "201" ]]; then
    NEW_BILLING_ID="$(printf '%s' "${RESP_BODY}" | jq -r '.id')"
  fi
fi

if [[ -n "${NEW_BILLING_ID}" && "${NEW_BILLING_ID}" != "null" ]]; then
  CASH_AMT=$((NEW_BILLING_AMOUNT / 2))
  CARD_AMT=$((NEW_BILLING_AMOUNT - CASH_AMT))
  SPLIT_PAYLOAD="$(jq -nc --argjson total "${NEW_BILLING_AMOUNT}" --argjson cash "${CASH_AMT}" --argjson card "${CARD_AMT}" \
    '{status:"completed",billing_amount:$total,payment_splits:[{method:"cash",amount:$cash,received_amount:$cash,change_amount:0},{method:"credit_card",amount:$card}]}')"
  call_api PATCH "/accountings/${NEW_BILLING_ID}" "${SPLIT_PAYLOAD}"
  SPLIT_PATCH_CODE="${RESP_CODE}"
  if [[ "${SPLIT_PATCH_CODE}" == "200" ]]; then
    call_api GET "/accountings/${NEW_BILLING_ID}"
    SPLITS_LEN="$(printf '%s' "${RESP_BODY}" | jq '.payment_splits | length')"
    SPLITS_SUM="$(printf '%s' "${RESP_BODY}" | jq '[.payment_splits[].amount] | add')"
    # NOTE(code-reviewer指摘 M-3): payments は accounting_response.go で omitempty のため
    # 新規作成会計では存在しないことがある。legacy method は参考ログのみに留め、合否判定は
    # payment_splits の件数・合計のみで行う（偽陰性防止）。
    LEGACY_METHOD="$(printf '%s' "${RESP_BODY}" | jq -r '.payments[0].method // "(omitted)"')"
    if [[ "${SPLITS_LEN}" == "2" && "${SPLITS_SUM}" == "${NEW_BILLING_AMOUNT}" ]]; then
      record "AC-6" PASS "新規会計 id=${NEW_BILLING_ID} PATCH -> 200, payment_splits=2件 合計=${SPLITS_SUM} (legacy payments.method=${LEGACY_METHOD}, 参考値)"
    else
      record "AC-6" FAIL "PATCH -> 200 だが GET 不整合 (splits_len=${SPLITS_LEN} sum=${SPLITS_SUM} legacy_method=${LEGACY_METHOD})"
    fi
  else
    record "AC-6" FAIL "PATCH /accountings/${NEW_BILLING_ID} (2-way split) -> ${SPLIT_PATCH_CODE} body=$(body_preview)"
  fi
elif [[ -z "${SMOKE_PET_ID}" || -z "${SMOKE_OWNER_ID}" ]]; then
  record "AC-6" SKIP "実在の pet/owner を動的取得できずAC-6を実行できません（データセット前提不足）"
else
  record "AC-6" FAIL "POST /accountings (new waiting) -> ${CREATE_CODE} body=$(body_preview)"
fi

# =========================================================================
# AC-8: Cleanup 確認
# =========================================================================
echo "==> AC-8: cleanup 確認"
CLEANUP_OK=true
CLEANUP_NOTES=()

if [[ -n "${TEST_GROUP_ID:-}" && "${TEST_GROUP_ID}" != "null" ]]; then
  call_api GET "/masters/permission-groups/${TEST_GROUP_ID}"
  if [[ "${RESP_CODE}" != "404" ]]; then
    # code-reviewer指摘 M-2: AC-3本体のDELETEが失敗していた場合のbest-effort再試行（結果は無視）
    call_api DELETE "/masters/permission-groups/${TEST_GROUP_ID}"
    call_api GET "/masters/permission-groups/${TEST_GROUP_ID}"
  fi
  if [[ "${RESP_CODE}" != "404" ]]; then
    CLEANUP_OK=false
    CLEANUP_NOTES+=("TEST permission group id=${TEST_GROUP_ID} が GET ${RESP_CODE} で残存（再DELETE試行済み）")
  else
    CLEANUP_NOTES+=("TEST permission group id=${TEST_GROUP_ID} -> 404 (削除確認済み)")
  fi
fi

if [[ -n "${TEST_STAFF_ID:-}" && "${TEST_STAFF_ID}" != "null" ]]; then
  call_api GET "/masters/staffs/${TEST_STAFF_ID}"
  if [[ "${RESP_CODE}" != "404" ]]; then
    # code-reviewer指摘 M-2: AC-4本体のDELETEが失敗していた場合のbest-effort再試行（結果は無視）
    call_api DELETE "/masters/staffs/${TEST_STAFF_ID}"
    call_api GET "/masters/staffs/${TEST_STAFF_ID}"
  fi
  if [[ "${RESP_CODE}" != "404" ]]; then
    CLEANUP_OK=false
    CLEANUP_NOTES+=("TEST staff id=${TEST_STAFF_ID} が GET ${RESP_CODE} で残存（再DELETE試行済み）")
  else
    CLEANUP_NOTES+=("TEST staff id=${TEST_STAFF_ID} -> 404 (削除確認済み)")
  fi
fi

if [[ -n "${BILLING_ID:-}" ]]; then
  CLEANUP_NOTES+=("seed会計 id=${BILLING_ID} を completed に変更（STG許容・元waitingへは戻していない）")
fi
if [[ -n "${NEW_BILLING_ID:-}" && "${NEW_BILLING_ID}" != "null" ]]; then
  CLEANUP_NOTES+=("smoke新規作成 会計 id=${NEW_BILLING_ID} は削除APIが存在しないため残存（billingsにDELETEルートなし）。memoにtrial11 smokeと明記済み")
fi

if [[ "${CLEANUP_OK}" == "true" ]]; then
  record "AC-8" PASS "$(printf '%s; ' "${CLEANUP_NOTES[@]}")"
else
  record "AC-8" FAIL "$(printf '%s; ' "${CLEANUP_NOTES[@]}")"
fi

# =========================================================================
# AC-11: UI mixed payment（明示 BLOCKED — スコープ外）
# =========================================================================
record "AC-11" BLOCKED "UI混在会計(§3〜§10)は本試行スコープ外。frontendのvercel.jsonプロキシがapi.stg.noah-karte.com(AWS)向きのためworkers.dev未接続 — P7-3へdefer"

# =========================================================================
# Summary
# =========================================================================
echo "=========================================="
echo "SUMMARY (FAIL=${FAIL_COUNT})"
printf '%s\n' "${AC_LOG[@]}"

if [[ "${FAIL_COUNT}" -gt 0 ]]; then
  exit 1
fi
exit 0
