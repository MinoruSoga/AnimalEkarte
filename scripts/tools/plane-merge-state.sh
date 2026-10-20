#!/usr/bin/env bash
# Move Plane EMR work items when a GitHub pull request merges.
#
# main    + EMR-N                           -> Review
# staging + EMR-N now reaching staging      -> UAT, and only from Review
#
# EMR-123, [EMR-123], and EMR-282/325/326 all count. Parentheses are optional.
# Done, Cancelled, Duplicate, Needs Human, Blocked, and Phase2 never move.
# production is ignored.
#
# Auth: PLANE_GITHUB_AUTOMATION_KEY (Plane API key). This is not PLANE_API_KEY.
# The key is never printed. Scopes: projects.work_items:read,
# projects.work_items:write, projects.work_items.comments:write.
#
# Rate limiting: Plane API 429 answers are waited out honoring Retry-After
# (seconds or HTTP-date); missing or invalid headers fall back to capped
# exponential backoff, and every request is paced by a small throttle. All
# waits share one finite whole-run budget: when a requested wait exceeds it,
# the ticket is deferred and listed in the summary instead of being re-hit.
# Writes whose outcome is unknown (transport failure, 5xx) are re-verified
# against the live state / comment dedupe before any resend, so an executed
# write is never repeated. Permanent 4xx errors (401/403, bad input) are
# never retried.
set -euo pipefail

PLANE_API_ORIGIN="${PLANE_API_ORIGIN:-https://api.plane.so}"
PLANE_WORKSPACE_SLUG="${PLANE_WORKSPACE_SLUG:-baritechllc}"
PLANE_PROJECT_ID="${PLANE_PROJECT_ID:-6a009324-8a8a-455c-b16a-41b3f6207b0e}"
PLANE_PROJECT_IDENTIFIER="${PLANE_PROJECT_IDENTIFIER:-EMR}"
PLANE_MERGE_MAX_COMMIT_PAGES="${PLANE_MERGE_MAX_COMMIT_PAGES:-50}"

fold_state() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//'
}

state_kind() {
  case "$(fold_state "$1")" in
    "")
      printf '%s\n' unreadable
      ;;
    done|cancelled|duplicate|"needs human"|blocked|phase2)
      printf '%s\n' protected
      ;;
    review)
      printf '%s\n' review
      ;;
    uat)
      printf '%s\n' uat
      ;;
    backlog|ready|research|running|本番切替対応)
      printf '%s\n' main-source
      ;;
    *)
      printf '%s\n' unknown
      ;;
  esac
}

# stdout: target state name, or "SKIP <reason>"
decide_transition() {
  local base="$1"
  local current="$2"
  local head="${3:-}"
  local kind
  kind="$(state_kind "$current")"

  case "$base" in
    main|staging) ;;
    *)
      printf '%s\n' "SKIP unsupported-base"
      return 0
      ;;
  esac

  case "$kind" in
    unreadable|unknown)
      printf '%s\n' "SKIP unknown"
      return 0
      ;;
    protected)
      printf '%s\n' "SKIP protected"
      return 0
      ;;
  esac

  if [[ "$base" == "staging" && "$head" != "main" ]]; then
    printf '%s\n' "SKIP head-not-main"
    return 0
  fi

  if [[ "$base" == "main" ]]; then
    if [[ "$kind" == "review" ]]; then
      printf '%s\n' "SKIP already"
      return 0
    fi
    if [[ "$kind" == "uat" ]]; then
      printf '%s\n' "SKIP uat-or-later"
      return 0
    fi
    printf '%s\n' "Review"
    return 0
  fi

  if [[ "$kind" == "review" ]]; then
    printf '%s\n' "UAT"
    return 0
  fi
  printf '%s\n' "SKIP not-review"
}

# Expand EMR-282 and EMR-282/325/326 into one EMR-N per line.
expand_ticket_tokens() {
  local token head rest
  while IFS= read -r token; do
    [[ -z "$token" ]] && continue
    head="${token%%/*}"
    printf '%s\n' "$head"
    if [[ "$token" == */* ]]; then
      rest="${token#*/}"
      printf '%s\n' "$rest" | tr '/' '\n' | sed 's/^/EMR-/'
    fi
  done
}

# stdout: EMR-N, one per line, numerically unique. Reads stdin.
extract_ticket_ids() {
  local matches
  matches="$(grep -oE 'EMR-[1-9][0-9]*(/[1-9][0-9]*)*' || true)"
  if [[ -z "$matches" ]]; then
    return 0
  fi
  printf '%s\n' "$matches" | expand_ticket_tokens | LC_ALL=C sort -u -t- -k2,2n
}

# stdout: PR numbers referenced as "(#123)" or "Merge pull request #123".
extract_referenced_pr_numbers() {
  local matches
  matches="$(
    grep -oE '\(#[1-9][0-9]*\)|Merge pull request #[1-9][0-9]+' || true
  )"
  if [[ -z "$matches" ]]; then
    return 0
  fi
  printf '%s\n' "$matches" | sed -E 's/.*#([0-9]+).*/\1/' | LC_ALL=C sort -u -n
}

# stdout: yes or no
pr_qualifies_for_staging_tickets() {
  local base_ref="$1"
  local merged="$2"
  local pr_number="$3"
  local current_number="$4"
  if [[ "$base_ref" == "main" && "$merged" == "true" && "$pr_number" != "$current_number" ]]; then
    printf '%s\n' yes
    return 0
  fi
  printf '%s\n' no
}

html_escape() {
  # Avoid the incompatible replacement quoting and patsub_replacement rules
  # between macOS Bash 3.2 and Linux Bash 5. Preserve every other character.
  local value="$1" escaped='' character
  while [[ -n "$value" ]]; do
    character="${value:0:1}"
    value="${value:1}"
    case "$character" in
      '&') escaped+='&amp;' ;;
      '<') escaped+='&lt;' ;;
      '>') escaped+='&gt;' ;;
      '"') escaped+='&quot;' ;;
      *) escaped+="$character" ;;
    esac
  done
  printf '%s' "$escaped"
}

build_comment_html() {
  local base="$1"
  local target="$2"
  local pr_url="$3"
  local safe_base safe_target safe_url
  safe_base="$(html_escape "$base")"
  safe_target="$(html_escape "$target")"
  safe_url="$(html_escape "$pr_url")"
  printf '<p>%s へのマージに合わせて %s に移動しました。</p><p>PR: <a href="%s">%s</a></p>' \
    "$safe_base" "$safe_target" "$safe_url" "$safe_url"
}

# Append commit messages and, for staging, qualifying main PR title/body.
tickets_from_graphql_payload() {
  local base="$1"
  local number="$2"
  local json_file="$3"
  jq -r '.data.repository.pullRequest.commits.nodes[]?.commit.message // empty' "$json_file"
  if [[ "$base" != "staging" ]]; then
    return 0
  fi
  jq -r --argjson number "$number" '
    .data.repository.pullRequest.commits.nodes[]?
    | .commit.associatedPullRequests.nodes[]?
    | select(.merged == true and .baseRefName == "main" and (.number != $number))
    | .title, (.body // "")
  ' "$json_file"
}

graphql_commit_query() {
  cat <<'EOF'
query($owner: String!, $name: String!, $number: Int!, $cursor: String) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $number) {
      commits(first: 100, after: $cursor) {
        pageInfo { hasNextPage endCursor }
        nodes {
          commit {
            message
            associatedPullRequests(first: 20) {
              pageInfo { hasNextPage }
              nodes { number merged baseRefName title body }
            }
          }
        }
      }
    }
  }
}
EOF
}

log_line() {
  printf '%s\n' "$1" >&2
  if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
    printf '%s\n' "$1" >>"$GITHUB_STEP_SUMMARY"
  fi
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    log_line "FAIL  required command not found: $1"
    exit 1
  fi
}

# Tunable finite-retry policy. Values are re-read at call time so tests and
# CI can scope them per run. All waits are charged to one whole-run budget.
#   PLANE_MERGE_STATE_MAX_ATTEMPTS      per-request attempt cap (default 4)
#   PLANE_MERGE_STATE_MAX_WAIT_SECONDS  whole-run wait budget (default 120s)
#   PLANE_MERGE_STATE_THROTTLE_MS       pacing sleep between requests (default 400ms)
#   PLANE_MERGE_STATE_BACKOFF_CAP_SECONDS  cap for headerless backoff (default 15s)
# Whole-run wait budget in milliseconds, derived from a sanitized integer
# second count so a malformed override cannot defeat the budget arithmetic.
plane_wait_budget_ms() {
  local seconds="${PLANE_MERGE_STATE_MAX_WAIT_SECONDS:-120}"
  case "$seconds" in ''|*[!0-9]*) seconds=120 ;; esac
  # Cap digit length so an absurd override cannot wrap int64 arithmetic.
  if (( ${#seconds} > 9 )); then
    seconds=120
  fi
  printf '%s\n' $(( seconds * 1000 ))
}

RUN_WAIT_BUDGET_REMAINING_MS="$(plane_wait_budget_ms)"
RUN_BUDGET_EXHAUSTED=0
RUN_ANY_REQUEST_MADE=""
# plane_http / plane_write_http store the response code here instead of a
# command substitution, so wait-budget accounting survives in this shell.
PLANE_HTTP_CODE="000"

plane_now_epoch() {
  date +%s
}

plane_sleep_seconds() {
  sleep "$1"
}

# Charge milliseconds against the whole-run wait budget. Refuses when the
# requested wait does not fit: the caller must stop instead of shortening
# the wait and re-hitting the API.
plane_budget_charge_ms() {
  local ms="$1"
  case "$ms" in ''|*[!0-9]*) return 1 ;; esac
  if (( ms > RUN_WAIT_BUDGET_REMAINING_MS )); then
    RUN_BUDGET_EXHAUSTED=1
    return 1
  fi
  RUN_WAIT_BUDGET_REMAINING_MS=$(( RUN_WAIT_BUDGET_REMAINING_MS - ms ))
  if (( ms > 0 )); then
    plane_sleep_seconds "$(printf '%d.%03d' $(( ms / 1000 )) $(( ms % 1000 )))"
  fi
  return 0
}

plane_throttle() {
  local ms="${PLANE_MERGE_STATE_THROTTLE_MS:-400}"
  case "$ms" in ''|*[!0-9]*) ms=400 ;; esac
  (( ms <= 0 )) && return 0
  if [[ -z "$RUN_ANY_REQUEST_MADE" ]]; then
    RUN_ANY_REQUEST_MADE=1
    return 0
  fi
  if (( ms > RUN_WAIT_BUDGET_REMAINING_MS )); then
    return 0
  fi
  RUN_WAIT_BUDGET_REMAINING_MS=$(( RUN_WAIT_BUDGET_REMAINING_MS - ms ))
  plane_sleep_seconds "$(printf '%d.%03d' $(( ms / 1000 )) $(( ms % 1000 )))"
}

# Single un-retried HTTP request. Headers are dumped so the retry loop can
# honor Retry-After. Never logs the request, headers, key, or response body.
plane_http_once() {
  local method="$1"
  local url="$2"
  local payload="$3"
  local body_file="$4"
  local headers_file="$5"
  local config_file http_code
  config_file="$(mktemp)"
  chmod 600 "$config_file"
  printf 'header = "X-Api-Key: %s"\nheader = "Accept: application/json"\n' \
    "$PLANE_GITHUB_AUTOMATION_KEY" >"$config_file"
  if [[ -n "$payload" ]]; then
    printf 'header = "Content-Type: application/json"\n' >>"$config_file"
    http_code="$(
      curl -sS --max-time 30 --config "$config_file" \
        -o "$body_file" -D "$headers_file" -w '%{http_code}' -X "$method" --data "$payload" "$url" || true
    )"
  else
    http_code="$(
      curl -sS --max-time 30 --config "$config_file" \
        -o "$body_file" -D "$headers_file" -w '%{http_code}' -X "$method" "$url" || true
    )"
  fi
  rm -f "$config_file"
  if [[ ! "$http_code" =~ ^[0-9]{3}$ ]]; then
    http_code="000"
  fi
  printf '%s\n' "$http_code"
}

plane_retry_after_from_headers() {
  awk 'tolower($0) ~ /^retry-after:/ { sub(/^[^:]*:[ \t]*/, ""); sub(/\r$/, ""); print; exit }' "$1" 2>/dev/null
}

# RFC 7231 HTTP-date to epoch seconds, across GNU and BSD date dialects.
# Prints nothing when the value cannot be parsed.
plane_httpdate_to_epoch() {
  local value="$1" epoch=""
  epoch="$(date -u -d "$value" +%s 2>/dev/null || true)"
  case "$epoch" in ''|*[!0-9]*)
    epoch="$(date -u -j -f '%a, %d %b %Y %H:%M:%S %Z' "$value" +%s 2>/dev/null || true)"
    ;;
  esac
  case "$epoch" in ''|*[!0-9]*) return 0 ;; esac
  printf '%s\n' "$epoch"
}

# Retry-After raw header value to seconds. Empty output means "no usable
# value" (missing or invalid) and the caller falls back to capped backoff.
plane_retry_after_seconds() {
  local raw="$1" epoch now
  raw="${raw#"${raw%%[![:space:]]*}"}"
  raw="${raw%"${raw##*[![:space:]]}"}"
  if [[ -z "$raw" ]]; then
    return 0
  fi
  case "$raw" in
    *[!0-9]*)
      epoch="$(plane_httpdate_to_epoch "$raw")"
      if [[ -z "$epoch" ]]; then
        return 0
      fi
      now="$(plane_now_epoch)"
      case "$now" in ''|*[!0-9]*) return 0 ;; esac
      if (( epoch > now )); then
        printf '%s\n' $(( epoch - now ))
      fi
      ;;
    *)
      printf '%s\n' "$raw"
      ;;
  esac
}

# Capped exponential backoff for missing/invalid Retry-After and transient
# failures: 2s, 4s, 8s, ... up to the cap.
plane_backoff_wait_ms() {
  local step="$1" cap="$2" wait_ms=2000 d=1
  while (( d < step )); do
    wait_ms=$(( wait_ms * 2 ))
    d=$(( d + 1 ))
  done
  if (( wait_ms > cap * 1000 )); then
    wait_ms=$(( cap * 1000 ))
  fi
  printf '%s\n' "$wait_ms"
}

# Core retry loop. allow_transient_retry=1 (GET lookups) retries transport
# failures and 5xx with capped backoff; =0 (PATCH/POST writes) returns them
# immediately so the caller can reconcile what actually landed. 429 is always
# waited out (Retry-After respected, headerless capped backoff) within the
# run budget; when the requested wait exceeds the budget the loop stops
# without re-hitting. Permanent 4xx codes return immediately.
plane_http_inner() {
  local allow_transient="$1"
  shift
  local method="$1" url="$2" payload="$3" body_file="$4"
  local max_attempts="${PLANE_MERGE_STATE_MAX_ATTEMPTS:-4}"
  local backoff_cap="${PLANE_MERGE_STATE_BACKOFF_CAP_SECONDS:-15}"
  case "$max_attempts" in ''|*[!0-9]*) max_attempts=4 ;; esac
  case "$backoff_cap" in ''|*[!0-9]*) backoff_cap=15 ;; esac
  local attempt=0 code raw_wait wait_ms backoff_step=1 headers
  headers="$(mktemp)"
  while :; do
    attempt=$(( attempt + 1 ))
    plane_throttle
    code="$(plane_http_once "$method" "$url" "$payload" "$body_file" "$headers")"
    case "$code" in
      429)
        raw_wait="$(plane_retry_after_from_headers "$headers")"
        wait_ms="$(plane_retry_after_seconds "$raw_wait")"
        if [[ -n "$wait_ms" ]]; then
          wait_ms=$(( wait_ms * 1000 ))
        else
          wait_ms="$(plane_backoff_wait_ms "$backoff_step" "$backoff_cap")"
          backoff_step=$(( backoff_step + 1 ))
        fi
        ;;
      000|5??)
        if [[ "$allow_transient" != "1" ]]; then
          rm -f "$headers"
          PLANE_HTTP_CODE="$code"
          return 0
        fi
        wait_ms="$(plane_backoff_wait_ms "$backoff_step" "$backoff_cap")"
        backoff_step=$(( backoff_step + 1 ))
        ;;
      *)
        rm -f "$headers"
        PLANE_HTTP_CODE="$code"
        return 0
        ;;
    esac
    if (( attempt >= max_attempts )); then
      rm -f "$headers"
      PLANE_HTTP_CODE="$code"
      return 0
    fi
    if ! plane_budget_charge_ms "$wait_ms"; then
      rm -f "$headers"
      PLANE_HTTP_CODE="$code"
      return 0
    fi
  done
}

plane_http() {
  plane_http_inner 1 "$@"
}

# Write variant: uncertain outcomes (transport failure, 5xx) are surfaced to
# the caller for state/comment re-verification instead of blind resend.
plane_write_http() {
  plane_http_inner 0 "$@"
}

plane_detail() {
  local body_file="$1"
  jq -r '.detail // .message // empty' "$body_file" 2>/dev/null | head -n 1
}

valid_comment_collection() {
  jq -e '
    if type == "array" then .
    elif type == "object" and (.data | type) == "array" then .data
    elif type == "object" and (.results | type) == "array" then .results
    else error("unsupported comment collection") end
    | all(.[]; type == "object")
  ' "$1" >/dev/null 2>&1
}

comment_already_posted() {
  local body_file="$1"
  local external_id="$2"
  jq -e --arg id "$external_id" '
    def items:
      if type == "array" then .
      elif (.data | type) == "array" then .data
      elif (.results | type) == "array" then .results
      else [] end;
    items | map(select(.external_id == $id)) | length > 0
  ' "$body_file" >/dev/null 2>&1
}

sync_one_ticket() {
  local ticket_id="$1"
  local base="$2"
  local pr_number="$3"
  local pr_url="$4"
  local head="${5:-}"
  local item_body patch_body comment_body
  local http_code identifier project_id item_uuid current_state decision
  local external_id comment_html list_url patch_url comment_url payload
  local reconcile_only=0
  local max_attempts="${PLANE_MERGE_STATE_MAX_ATTEMPTS:-4}"
  case "$max_attempts" in ''|*[!0-9]*) max_attempts=4 ;; esac
  # Outcome kind for the run-level retry pass: ok / skip / permanent /
  # transient / unknown / deferred. rc: 0 ok, 1 failed, 2 unresolved
  # (unknown or deferred by the wait budget).
  SYNC_OUTCOME_KIND="transient"

  item_body="$(mktemp)"
  patch_body="$(mktemp)"
  comment_body="$(mktemp)"

  plane_http GET \
    "${PLANE_API_ORIGIN}/api/v2/workspaces/${PLANE_WORKSPACE_SLUG}/work-items/${ticket_id}/?expand=state" \
    "" "$item_body"
  http_code="$PLANE_HTTP_CODE"
  if [[ "$http_code" != "200" ]]; then
    case "$http_code" in
      429|000|5??)
        if [[ "$RUN_BUDGET_EXHAUSTED" == "1" ]]; then
          log_line "WAIT  ${ticket_id}  lookup deferred: wait budget exhausted (HTTP ${http_code})"
          rm -f "$item_body" "$patch_body" "$comment_body"
          SYNC_OUTCOME_KIND="deferred"
          return 2
        fi
        SYNC_OUTCOME_KIND="transient"
        ;;
      *) SYNC_OUTCOME_KIND="permanent" ;;
    esac
    log_line "FAIL  ${ticket_id}  lookup HTTP ${http_code} $(plane_detail "$item_body")"
    rm -f "$item_body" "$patch_body" "$comment_body"
    return 1
  fi

  identifier="$(jq -r '.identifier // empty' "$item_body")"
  project_id="$(jq -r '.project_id // empty' "$item_body")"
  item_uuid="$(jq -r '.id // empty' "$item_body")"
  current_state="$(jq -r '.state.name // empty' "$item_body")"
  if [[ "$(fold_state "$identifier")" != "$(fold_state "$ticket_id")" || -z "$item_uuid" ]]; then
    log_line "FAIL  ${ticket_id}  lookup response did not match the identifier"
    rm -f "$item_body" "$patch_body" "$comment_body"
    SYNC_OUTCOME_KIND="permanent"
    return 1
  fi
  if [[ -n "$project_id" && "$project_id" != "$PLANE_PROJECT_ID" ]]; then
    log_line "FAIL  ${ticket_id}  work item is outside project ${PLANE_PROJECT_IDENTIFIER}"
    rm -f "$item_body" "$patch_body" "$comment_body"
    SYNC_OUTCOME_KIND="permanent"
    return 1
  fi

  decision="$(decide_transition "$base" "$current_state" "$head")"
  # A previous run may have committed the state PATCH but failed to post its
  # comment. Reconcile the comment at the target without another state write.
  if [[ "$base" == "main" && "$decision" == "SKIP already" ]]; then
    decision=Review
    reconcile_only=1
  elif [[ "$base" == "main" && "$decision" == "SKIP uat-or-later" ]]; then
    decision=Review
    reconcile_only=1
  elif [[ "$base" == "staging" && "$head" == "main" && "$(fold_state "$current_state")" == "uat" ]]; then
    decision=UAT
    reconcile_only=1
  fi
  case "$decision" in
    SKIP\ *)
      log_line "SKIP  ${ticket_id}  ${decision#SKIP }  current=${current_state:-unreadable}"
      rm -f "$item_body" "$patch_body" "$comment_body"
      if [[ "$decision" == "SKIP unknown" ]]; then
        SYNC_OUTCOME_KIND="permanent"
        return 1
      fi
      SYNC_OUTCOME_KIND="skip"
      return 0
      ;;
    Review|UAT) ;;
    *)
      log_line "FAIL  ${ticket_id}  unexpected decision"
      rm -f "$item_body" "$patch_body" "$comment_body"
      SYNC_OUTCOME_KIND="permanent"
      return 1
      ;;
  esac

  if [[ "${PLANE_MERGE_STATE_DRY_RUN:-}" == "1" ]]; then
    log_line "DRY   ${ticket_id}  ${current_state} -> ${decision}"
    rm -f "$item_body" "$patch_body" "$comment_body"
    SYNC_OUTCOME_KIND="ok"
    return 0
  fi

  # State write with uncertainty reconciliation. A transport failure or 5xx
  # does not tell whether the PATCH landed: re-read the current state and
  # only resend when the lookup confirms the write never happened.
  if [[ "$reconcile_only" == "0" ]]; then
    local pre_state="$current_state"
    local verify_code verify_state patch_attempt=0 state_written=0
    patch_url="${PLANE_API_ORIGIN}/api/v2/workspaces/${PLANE_WORKSPACE_SLUG}/projects/${PLANE_PROJECT_ID}/work-items/${item_uuid}/"
    payload="$(jq -n --arg state "$decision" '{state:$state}')"
    while :; do
      patch_attempt=$(( patch_attempt + 1 ))
      plane_write_http PATCH "$patch_url" "$payload" "$patch_body"
      http_code="$PLANE_HTTP_CODE"
      if [[ "$http_code" == "200" ]]; then
        state_written=1
        break
      fi
      case "$http_code" in
        429)
          if [[ "$RUN_BUDGET_EXHAUSTED" == "1" ]]; then
            log_line "WAIT  ${ticket_id}  state update deferred: wait budget exhausted (HTTP 429)"
            rm -f "$item_body" "$patch_body" "$comment_body"
            SYNC_OUTCOME_KIND="deferred"
            return 2
          fi
          # A 429 answer means the state was not changed; the bounded
          # attempts are used up, so leave this ticket to the retry pass.
          log_line "FAIL  ${ticket_id}  state update HTTP 429 after ${patch_attempt} attempt(s)"
          rm -f "$item_body" "$patch_body" "$comment_body"
          SYNC_OUTCOME_KIND="transient"
          return 1
          ;;
        000|5??)
          plane_http GET \
            "${PLANE_API_ORIGIN}/api/v2/workspaces/${PLANE_WORKSPACE_SLUG}/work-items/${ticket_id}/?expand=state" \
            "" "$item_body"
          verify_code="$PLANE_HTTP_CODE"
          if [[ "$verify_code" != "200" ]]; then
            if [[ "$RUN_BUDGET_EXHAUSTED" == "1" ]]; then
              log_line "WAIT  ${ticket_id}  state update uncertain (HTTP ${http_code}); verify deferred: wait budget exhausted"
              rm -f "$item_body" "$patch_body" "$comment_body"
              SYNC_OUTCOME_KIND="deferred"
              return 2
            fi
            log_line "UNKNOWN  ${ticket_id}  state update uncertain (HTTP ${http_code}); verify lookup HTTP ${verify_code}"
            rm -f "$item_body" "$patch_body" "$comment_body"
            SYNC_OUTCOME_KIND="unknown"
            return 2
          fi
          verify_state="$(jq -r '.state.name // empty' "$item_body")"
          if [[ "$(fold_state "$verify_state")" == "$(fold_state "$decision")" ]]; then
            # The write landed despite the uncertain answer; no resend.
            current_state="$verify_state"
            state_written=1
            break
          fi
          if [[ "$(fold_state "$verify_state")" != "$(fold_state "$pre_state")" ]]; then
            log_line "UNKNOWN  ${ticket_id}  state update uncertain: state is '${verify_state}', expected '${decision}' or pre-update '${pre_state}'"
            rm -f "$item_body" "$patch_body" "$comment_body"
            SYNC_OUTCOME_KIND="unknown"
            return 2
          fi
          if (( patch_attempt >= max_attempts )); then
            log_line "FAIL  ${ticket_id}  state update HTTP ${http_code} after ${patch_attempt} attempt(s), state confirmed unchanged"
            rm -f "$item_body" "$patch_body" "$comment_body"
            SYNC_OUTCOME_KIND="transient"
            return 1
          fi
          if ! plane_budget_charge_ms 2000; then
            log_line "WAIT  ${ticket_id}  state update deferred: wait budget exhausted before resend"
            rm -f "$item_body" "$patch_body" "$comment_body"
            SYNC_OUTCOME_KIND="deferred"
            return 2
          fi
          ;;
        *)
          log_line "FAIL  ${ticket_id}  state update HTTP ${http_code} $(plane_detail "$patch_body")"
          rm -f "$item_body" "$patch_body" "$comment_body"
          SYNC_OUTCOME_KIND="permanent"
          return 1
          ;;
      esac
    done
    if [[ "$state_written" != "1" ]]; then
      # Defensive: unreachable, but never fall through without a decision.
      rm -f "$item_body" "$patch_body" "$comment_body"
      SYNC_OUTCOME_KIND="unknown"
      return 2
    fi
  fi

  external_id="gh-merge:${pr_number}:${base}:${decision}"
  comment_html="$(build_comment_html "$base" "$decision" "$pr_url")"
  if [[ "$reconcile_only" == "1" ]]; then
    comment_html="<p>$(html_escape "$base") へのマージ後、現在の $(html_escape "$current_state") 状態を確認し、状態変更なしでこのPRの監査コメントを補完しました。</p><p>PR: <a href=\"$(html_escape "$pr_url")\">$(html_escape "$pr_url")</a></p>"
  fi
  list_url="${PLANE_API_ORIGIN}/api/v2/workspaces/${PLANE_WORKSPACE_SLUG}/projects/${PLANE_PROJECT_ID}/work-items/${item_uuid}/comments/?external_source=github&external_id=$(jq -nr --arg v "$external_id" '$v|@uri')"
  payload="$(jq -n \
    --arg html "$comment_html" \
    --arg external_id "$external_id" \
    '{comment_html:$html, access:"INTERNAL", external_source:"github", external_id:$external_id}')"
  comment_url="${PLANE_API_ORIGIN}/api/v2/workspaces/${PLANE_WORKSPACE_SLUG}/projects/${PLANE_PROJECT_ID}/work-items/${item_uuid}/comments/"

  # Comment phase. Each loop round re-reads the comment collection first, so
  # a comment that landed despite an uncertain POST answer is detected by
  # the external_id dedupe and never posted twice.
  local comment_attempt=0
  while :; do
    comment_attempt=$(( comment_attempt + 1 ))
    plane_http GET "$list_url" "" "$comment_body"
    http_code="$PLANE_HTTP_CODE"
    if [[ "$http_code" != "200" ]]; then
      case "$http_code" in
        429|000|5??)
          if [[ "$RUN_BUDGET_EXHAUSTED" == "1" ]]; then
            log_line "WAIT  ${ticket_id}  comment lookup deferred: wait budget exhausted (HTTP ${http_code})"
            rm -f "$item_body" "$patch_body" "$comment_body"
            SYNC_OUTCOME_KIND="deferred"
            return 2
          fi
          SYNC_OUTCOME_KIND="transient"
          ;;
        *) SYNC_OUTCOME_KIND="permanent" ;;
      esac
      if [[ "$reconcile_only" == "1" ]]; then
        log_line "FAIL  ${ticket_id}  audit reconciliation current=${current_state} target=${decision} state=unchanged comment HTTP ${http_code} $(plane_detail "$comment_body")"
      else
        log_line "FAIL  ${ticket_id}  comment lookup HTTP ${http_code} $(plane_detail "$comment_body")"
      fi
      rm -f "$item_body" "$patch_body" "$comment_body"
      return 1
    fi
    if ! valid_comment_collection "$comment_body"; then
      log_line "FAIL  ${ticket_id}  comment lookup returned an invalid collection"
      rm -f "$item_body" "$patch_body" "$comment_body"
      SYNC_OUTCOME_KIND="transient"
      return 1
    fi
    if comment_already_posted "$comment_body" "$external_id"; then
      if [[ "$reconcile_only" == "1" ]]; then
        log_line "AUDIT  ${ticket_id}  current=${current_state} target=${decision} state=unchanged comment=already"
      else
        log_line "MOVE  ${ticket_id}  ${current_state} -> ${decision}  comment=already"
      fi
      rm -f "$item_body" "$patch_body" "$comment_body"
      SYNC_OUTCOME_KIND="ok"
      return 0
    fi

    plane_write_http POST "$comment_url" "$payload" "$comment_body"
    http_code="$PLANE_HTTP_CODE"
    if [[ "$http_code" == "200" || "$http_code" == "201" ]]; then
      if [[ "$reconcile_only" == "1" ]]; then
        log_line "AUDIT  ${ticket_id}  current=${current_state} target=${decision} state=unchanged comment=backfilled"
      else
        log_line "MOVE  ${ticket_id}  ${current_state} -> ${decision}"
      fi
      rm -f "$item_body" "$patch_body" "$comment_body"
      SYNC_OUTCOME_KIND="ok"
      return 0
    fi
    case "$http_code" in
      429)
        if [[ "$RUN_BUDGET_EXHAUSTED" == "1" ]]; then
          log_line "WAIT  ${ticket_id}  comment post deferred: wait budget exhausted (HTTP 429)"
          rm -f "$item_body" "$patch_body" "$comment_body"
          SYNC_OUTCOME_KIND="deferred"
          return 2
        fi
        log_line "FAIL  ${ticket_id}  comment post HTTP 429 after ${comment_attempt} attempt(s)"
        rm -f "$item_body" "$patch_body" "$comment_body"
        SYNC_OUTCOME_KIND="transient"
        return 1
        ;;
      000|5??)
        # Uncertain: the next round re-reads the collection and dedupes, so
        # a landed comment is never resent. Bound the retries.
        if (( comment_attempt >= max_attempts )); then
          if [[ "$reconcile_only" == "1" ]]; then
            log_line "FAIL  ${ticket_id}  audit reconciliation current=${current_state} target=${decision} state=unchanged comment HTTP ${http_code} $(plane_detail "$comment_body")"
          else
            log_line "FAIL  ${ticket_id}  moved to ${decision} but comment HTTP ${http_code} $(plane_detail "$comment_body")"
          fi
          rm -f "$item_body" "$patch_body" "$comment_body"
          SYNC_OUTCOME_KIND="transient"
          return 1
        fi
        if ! plane_budget_charge_ms 2000; then
          log_line "WAIT  ${ticket_id}  comment post deferred: wait budget exhausted"
          rm -f "$item_body" "$patch_body" "$comment_body"
          SYNC_OUTCOME_KIND="deferred"
          return 2
        fi
        ;;
      *)
        if [[ "$reconcile_only" == "1" ]]; then
          log_line "FAIL  ${ticket_id}  audit reconciliation current=${current_state} target=${decision} state=unchanged comment HTTP ${http_code} $(plane_detail "$comment_body")"
        else
          log_line "FAIL  ${ticket_id}  moved to ${decision} but comment HTTP ${http_code} $(plane_detail "$comment_body")"
        fi
        rm -f "$item_body" "$patch_body" "$comment_body"
        SYNC_OUTCOME_KIND="permanent"
        return 1
        ;;
    esac
  done
}

assert_no_truncated_associations() {
  local json_file="$1"
  local truncated
  truncated="$(jq -r '
    [
      .data.repository.pullRequest.commits.nodes[]?
      | .commit.associatedPullRequests
      | select(.pageInfo.hasNextPage == true)
    ] | length
  ' "$json_file")"
  if [[ "$truncated" != "0" ]]; then
    log_line "FAIL  a commit has more associated pull requests than the query page"
    return 1
  fi
}

append_graphql_page() {
  local owner="$1"
  local name="$2"
  local number="$3"
  local base="$4"
  local cursor_mode="$5"
  local cursor="$6"
  local dest="$7"
  local payload_file query
  payload_file="$(mktemp)"
  query="$(graphql_commit_query)"
  if [[ "$cursor_mode" == "null" ]]; then
    if ! gh api graphql \
      -f query="$query" \
      -f owner="$owner" \
      -f name="$name" \
      -F number="$number" \
      -F cursor=null >"$payload_file"; then
      log_line "FAIL  GitHub GraphQL request failed"
      rm -f "$payload_file"
      return 1
    fi
  elif ! gh api graphql \
    -f query="$query" \
    -f owner="$owner" \
    -f name="$name" \
    -F number="$number" \
    -f cursor="$cursor" >"$payload_file"; then
    log_line "FAIL  GitHub GraphQL request failed"
    rm -f "$payload_file"
    return 1
  fi
  if jq -e '(.errors // []) | length > 0' "$payload_file" >/dev/null; then
    log_line "FAIL  GitHub GraphQL error"
    jq -r '.errors[].message' "$payload_file" >&2
    rm -f "$payload_file"
    return 1
  fi
  if ! jq -e '.data.repository.pullRequest' "$payload_file" >/dev/null; then
    log_line "FAIL  GitHub pull request was not returned"
    rm -f "$payload_file"
    return 1
  fi
  if ! assert_no_truncated_associations "$payload_file"; then
    rm -f "$payload_file"
    return 1
  fi
  if ! tickets_from_graphql_payload "$base" "$number" "$payload_file" >>"$dest"; then
    log_line "FAIL  could not read pull request commits"
    rm -f "$payload_file"
    return 1
  fi
  if [[ "$base" == "staging" ]]; then
    if ! jq -r --argjson number "$number" '
      .data.repository.pullRequest.commits.nodes[]?
      | .commit.associatedPullRequests.nodes[]?
      | select(.merged == true and .baseRefName == "main" and (.number != $number))
      | .number
    ' "$payload_file" >>"${dest}.seen-prs"; then
      log_line "FAIL  could not read associated pull requests"
      rm -f "$payload_file"
      return 1
    fi
  fi
  if ! jq -r '.data.repository.pullRequest.commits.pageInfo | [.hasNextPage, (.endCursor // "")] | @tsv' "$payload_file"; then
    log_line "FAIL  could not read the commit page cursor"
    rm -f "$payload_file"
    return 1
  fi
  rm -f "$payload_file"
}

collect_source_text() {
  local owner="$1"
  local name="$2"
  local number="$3"
  local base="$4"
  local title="$5"
  local body="$6"
  local dest="$7"
  local page=0
  local cursor_mode="null"
  local cursor=""
  local page_info has_next next_cursor

  : >"$dest"
  : >"${dest}.seen-prs"
  printf '%s\n%s\n' "$title" "$body" >>"$dest"

  while true; do
    page=$((page + 1))
    if [[ "$page" -gt "$PLANE_MERGE_MAX_COMMIT_PAGES" ]]; then
      log_line "FAIL  pull request commit pages exceeded ${PLANE_MERGE_MAX_COMMIT_PAGES}"
      return 1
    fi
    if ! page_info="$(append_graphql_page "$owner" "$name" "$number" "$base" "$cursor_mode" "$cursor" "$dest")"; then
      return 1
    fi
    has_next="${page_info%%$'\t'*}"
    next_cursor="${page_info#*$'\t'}"
    if [[ "$has_next" != "true" ]]; then
      break
    fi
    if [[ -z "$next_cursor" ]]; then
      log_line "FAIL  GitHub commit page is missing endCursor"
      return 1
    fi
    cursor="$next_cursor"
    cursor_mode="set"
  done

  if [[ "$base" != "staging" ]]; then
    return 0
  fi

  local pr_numbers_file seen_file pr_number pr_json base_ref merged
  pr_numbers_file="$(mktemp)"
  seen_file="${dest}.seen-prs"
  # Commit messages are already in dest. Recover PR numbers the association page missed.
  extract_referenced_pr_numbers <"$dest" >"$pr_numbers_file"
  while IFS= read -r pr_number; do
    [[ -z "$pr_number" ]] && continue
    if [[ "$pr_number" == "$number" ]]; then
      continue
    fi
    if grep -qx "$pr_number" "$seen_file"; then
      continue
    fi
    pr_json="$(mktemp)"
    if ! gh api "repos/${owner}/${name}/pulls/${pr_number}" >"$pr_json"; then
      log_line "SKIP  PR #${pr_number}  could not be read"
      rm -f "$pr_json"
      continue
    fi
    base_ref="$(jq -r '.base.ref // empty' "$pr_json")"
    merged="$(jq -r '.merged // false' "$pr_json")"
    if [[ "$(pr_qualifies_for_staging_tickets "$base_ref" "$merged" "$pr_number" "$number")" == "yes" ]]; then
      jq -r '.title, (.body // "")' "$pr_json" >>"$dest"
      printf '%s\n' "$pr_number" >>"$seen_file"
    fi
    rm -f "$pr_json"
  done <"$pr_numbers_file"
  rm -f "$pr_numbers_file"
}

# Run the per-ticket sync over every referenced ticket, then one bounded
# retry pass for tickets that did not resolve. Permanent failures (401/403,
# malformed lookups, protected states) are never resent. Every wait —
# throttle, backoff, Retry-After — is charged to the same whole-run budget,
# and tickets the budget deferred are listed in the summary instead of
# being re-hit with a shortened wait.
run_ticket_passes() {
  local tickets_file="$1"
  local base="$2"
  local number="$3"
  local pr_url="$4"
  local head="$5"
  local ticket_id rc outcome_file rerun_file kept_file
  local ok=0 skip=0 failed=0 unknown=0 deferred=0 total=0

  outcome_file="$(mktemp)"
  : >"$outcome_file"
  while IFS= read -r ticket_id; do
    [[ -z "$ticket_id" ]] && continue
    total=$((total + 1))
    SYNC_OUTCOME_KIND="transient"
    rc=0
    # A conditional call keeps one ticket failure from aborting the pass
    # without a subshell, so wait-budget accounting persists across tickets.
    sync_one_ticket "$ticket_id" "$base" "$number" "$pr_url" "$head" || rc=$?
    printf '%s\t%s\t%s\n' "$ticket_id" "$rc" "$SYNC_OUTCOME_KIND" >>"$outcome_file"
  done <"$tickets_file"

  rerun_file="$(mktemp)"
  awk -F'\t' '$2 != 0 && $3 != "permanent" { print $1 }' "$outcome_file" >"$rerun_file"
  if [[ -s "$rerun_file" ]]; then
    kept_file="$(mktemp)"
    awk -F'\t' '$2 == 0 || $3 == "permanent"' "$outcome_file" >"$kept_file"
    while IFS= read -r ticket_id; do
      [[ -z "$ticket_id" ]] && continue
      total=$((total + 1))
      SYNC_OUTCOME_KIND="transient"
      rc=0
      sync_one_ticket "$ticket_id" "$base" "$number" "$pr_url" "$head" || rc=$?
      printf '%s\t%s\t%s\n' "$ticket_id" "$rc" "$SYNC_OUTCOME_KIND" >>"$kept_file"
    done <"$rerun_file"
    rm -f "$outcome_file"
    outcome_file="$kept_file"
  fi
  rm -f "$rerun_file"

  while IFS=$'\t' read -r _ticket_id _rc kind; do
    [[ -z "$kind" ]] && continue
    case "$kind" in
      ok) ok=$((ok + 1)) ;;
      skip) skip=$((skip + 1)) ;;
      permanent|transient) failed=$((failed + 1)) ;;
      unknown) unknown=$((unknown + 1)) ;;
      deferred) deferred=$((deferred + 1)) ;;
    esac
  done <"$outcome_file"
  rm -f "$outcome_file"

  log_line "SUMMARY  ok=${ok} skip=${skip} fail=${failed} unknown=${unknown} deferred=${deferred} requests=${total} wait-budget-left-ms=${RUN_WAIT_BUDGET_REMAINING_MS}"
  if (( failed + unknown + deferred > 0 )); then
    log_line "FAIL  $(( failed + unknown + deferred )) work item update(s) unresolved (fail=${failed} unknown=${unknown} deferred=${deferred}) for ${base} PR #${number}"
    return 1
  fi
  log_line "PASS  Plane states checked for ${base} PR #${number}"
}

main() {
  local event_path="${GITHUB_EVENT_PATH:-}"
  local repository="${GITHUB_REPOSITORY:-}"
  local merged base head number title body pr_url owner name
  local sources tickets_file ticket_id

  require_command jq
  require_command gh
  require_command curl

  if [[ -z "${PLANE_GITHUB_AUTOMATION_KEY:-}" ]]; then
    log_line "FAIL  PLANE_GITHUB_AUTOMATION_KEY is not set"
    exit 1
  fi
  if [[ "$PLANE_GITHUB_AUTOMATION_KEY" == *$'\n'* || "$PLANE_GITHUB_AUTOMATION_KEY" == *'"'* || "$PLANE_GITHUB_AUTOMATION_KEY" == *'\'* ]]; then
    log_line "FAIL  PLANE_GITHUB_AUTOMATION_KEY contains a character this runner rejects"
    exit 1
  fi
  if [[ -z "$event_path" || ! -f "$event_path" ]]; then
    log_line "FAIL  GITHUB_EVENT_PATH is missing"
    exit 1
  fi
  if [[ "$repository" != */* ]]; then
    log_line "FAIL  GITHUB_REPOSITORY is missing"
    exit 1
  fi

  merged="$(jq -r '.pull_request.merged // false' "$event_path")"
  base="$(jq -r '.pull_request.base.ref // empty' "$event_path")"
  head="$(jq -r '.pull_request.head.ref // empty' "$event_path")"
  number="$(jq -r '.pull_request.number // empty' "$event_path")"
  title="$(jq -r '.pull_request.title // ""' "$event_path")"
  body="$(jq -r '.pull_request.body // ""' "$event_path")"
  pr_url="$(jq -r '.pull_request.html_url // empty' "$event_path")"
  owner="${repository%%/*}"
  name="${repository#*/}"

  if [[ "$merged" != "true" ]]; then
    log_line "SKIP  pull request is not merged"
    exit 0
  fi
  if [[ "$base" != "main" && "$base" != "staging" ]]; then
    log_line "SKIP  base ${base:-unknown} is not main or staging"
    exit 0
  fi
  if [[ "$base" == "staging" && "$head" != "main" ]]; then
    log_line "SKIP  base staging head.ref ${head:-unknown} is not main"
    exit 0
  fi
  if [[ -z "$number" || -z "$pr_url" ]]; then
    log_line "FAIL  pull request number or URL is missing"
    exit 1
  fi

  sources="$(mktemp)"
  tickets_file="$(mktemp)"
  # shellcheck disable=SC2064
  trap "rm -f '$sources' '${sources}.seen-prs' '$tickets_file'" EXIT
  collect_source_text "$owner" "$name" "$number" "$base" "$title" "$body" "$sources"
  extract_ticket_ids <"$sources" >"$tickets_file"
  if [[ ! -s "$tickets_file" ]]; then
    log_line "SKIP  no EMR-N references on ${base} PR #${number}"
    exit 0
  fi

  RUN_WAIT_BUDGET_REMAINING_MS="$(plane_wait_budget_ms)"
  RUN_BUDGET_EXHAUSTED=0
  RUN_ANY_REQUEST_MADE=""
  run_ticket_passes "$tickets_file" "$base" "$number" "$pr_url" "$head"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
