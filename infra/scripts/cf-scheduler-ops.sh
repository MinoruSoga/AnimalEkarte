#!/usr/bin/env bash
# BE9-3: Cloudflare durable scheduler の status / pause / resume /
# missing-slot catch-up を専用 credential で操作する。

set -euo pipefail

usage() {
  cat >&2 <<'USAGE'
Usage:
  cf-scheduler-ops.sh status [limit]
  cf-scheduler-ops.sh pause <expected-revision> <reason> [request-id]
  cf-scheduler-ops.sh resume <expected-revision> <reason> [request-id]
  cf-scheduler-ops.sh run <no_show|delivery|dormant> <scheduled-time-ms> <reason> [request-id]

Required environment:
  SCHEDULER_OPS_BASE_URL   HTTPS Worker base URL
  SCHEDULER_OPS_ALLOWED_HOST Exact Worker hostname allowlist
  SCHEDULER_OPS_SECRET     Dedicated scheduler operations secret
USAGE
  exit 2
}

is_valid_hostname() {
  local value="$1"
  local -a labels
  local label
  [[ "${#value}" -le 253 ]] || return 1
  IFS='.' read -r -a labels <<<"${value}"
  [[ "${#labels[@]}" -ge 2 ]] || return 1
  for label in "${labels[@]}"; do
    [[ "${#label}" -le 63 ]] || return 1
    [[ "${label}" =~ ^[a-z0-9]([a-z0-9-]*[a-z0-9])?$ ]] || return 1
  done
}

if [[ -z "${SCHEDULER_OPS_BASE_URL:-}" ]]; then
  echo "::error::SCHEDULER_OPS_BASE_URL が未設定です" >&2
  exit 1
fi
if [[ ! "${SCHEDULER_OPS_BASE_URL}" =~ ^https://([A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?)(:443)?/?$ ]]; then
  echo "::error::SCHEDULER_OPS_BASE_URL は path/query/credential を含まない HTTPS URL にしてください" >&2
  exit 1
fi
base_host="${BASH_REMATCH[1]}"
if [[ -z "${SCHEDULER_OPS_ALLOWED_HOST:-}" ]]; then
  echo "::error::SCHEDULER_OPS_ALLOWED_HOST が未設定です" >&2
  exit 1
fi
normalized_base_host=$(printf '%s' "${base_host}" | tr '[:upper:]' '[:lower:]')
normalized_allowed_host=$(printf '%s' "${SCHEDULER_OPS_ALLOWED_HOST}" | tr '[:upper:]' '[:lower:]')
if ! is_valid_hostname "${normalized_base_host}" || ! is_valid_hostname "${normalized_allowed_host}"; then
  echo "::error::scheduler operations host が不正です" >&2
  exit 1
fi
if [[ "${normalized_base_host}" != "${normalized_allowed_host}" ]]; then
  echo "::error::SCHEDULER_OPS_BASE_URL の host が allowlist と一致しません" >&2
  exit 1
fi
if [[ -z "${SCHEDULER_OPS_SECRET:-}" ]]; then
  echo "::error::SCHEDULER_OPS_SECRET が未設定です" >&2
  exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "::error::jq が必要です" >&2
  exit 1
fi

base_url="${SCHEDULER_OPS_BASE_URL%/}"
auth_header_file=$(mktemp)
cleanup() {
  : >"${auth_header_file}"
  rm -f "${auth_header_file}"
}
trap cleanup EXIT
chmod 600 "${auth_header_file}"
printf 'Authorization: Bearer %s\n' "${SCHEDULER_OPS_SECRET}" >"${auth_header_file}"
unset SCHEDULER_OPS_SECRET

new_request_id() {
  if ! command -v uuidgen >/dev/null 2>&1; then
    echo "::error::request-id を省略する場合は uuidgen が必要です" >&2
    exit 1
  fi
  uuidgen | tr '[:upper:]' '[:lower:]'
}

request_id_or_new() {
  local supplied="${1:-}"
  if [[ -n "${supplied}" ]]; then
    printf '%s\n' "${supplied}"
    return
  fi
  new_request_id
}

api_request() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local curl_args=(
    --silent
    --show-error
    --fail-with-body
    --connect-timeout 10
    --max-time 30
    --proto "=https"
    --proto-redir "=https"
    --max-redirs 0
    --request "${method}"
    --header "@${auth_header_file}"
  )
  if [[ -n "${body}" ]]; then
    curl_args+=(
      --header "Content-Type: application/json"
      --data-binary "${body}"
    )
  fi
  curl --disable "${curl_args[@]}" "${base_url}${path}" | jq .
}

command="${1:-}"
case "${command}" in
  status)
    [[ "$#" -le 2 ]] || usage
    limit="${2:-20}"
    [[ "${limit}" =~ ^[0-9]+$ ]] || usage
    (( limit >= 1 && limit <= 50 )) || usage
    api_request GET "/_internal/scheduler/status?limit=${limit}"
    ;;
  pause|resume)
    [[ "$#" -ge 3 && "$#" -le 4 ]] || usage
    expected_revision="$2"
    reason="$3"
    request_id=$(request_id_or_new "${4:-}")
    [[ "${expected_revision}" =~ ^[0-9]+$ ]] || usage
    paused=false
    if [[ "${command}" == "pause" ]]; then
      paused=true
    fi
    body=$(jq -cn \
      --argjson paused "${paused}" \
      --argjson expected_revision "${expected_revision}" \
      --arg request_id "${request_id}" \
      --arg reason "${reason}" \
      '{
        paused: $paused,
        expected_revision: $expected_revision,
        request_id: $request_id,
        reason: $reason
      }')
    api_request PUT "/_internal/scheduler/control" "${body}"
    ;;
  run)
    [[ "$#" -ge 4 && "$#" -le 5 ]] || usage
    job="$2"
    scheduled_time_ms="$3"
    reason="$4"
    request_id=$(request_id_or_new "${5:-}")
    case "${job}" in
      no_show|delivery|dormant) ;;
      *) usage ;;
    esac
    [[ "${scheduled_time_ms}" =~ ^[0-9]+$ ]] || usage
    body=$(jq -cn \
      --arg job "${job}" \
      --argjson scheduled_time_ms "${scheduled_time_ms}" \
      --arg request_id "${request_id}" \
      --arg reason "${reason}" \
      '{
        job: $job,
        scheduled_time_ms: $scheduled_time_ms,
        mode: "catch_up",
        request_id: $request_id,
        reason: $reason
      }')
    api_request POST "/_internal/scheduler/runs" "${body}"
    ;;
  *)
    usage
    ;;
esac
