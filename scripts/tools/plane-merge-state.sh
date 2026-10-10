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
    backlog|ready|research|running|uat|本番切替対応)
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

  if [[ "$base" == "main" ]]; then
    if [[ "$kind" == "review" ]]; then
      printf '%s\n' "SKIP already"
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
  local value="$1"
  value="${value//&/&amp;}"
  value="${value//</&lt;}"
  value="${value//>/&gt;}"
  value="${value//\"/&quot;}"
  printf '%s' "$value"
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

plane_http() {
  local method="$1"
  local url="$2"
  local payload="$3"
  local body_file="$4"
  local config_file http_code
  config_file="$(mktemp)"
  chmod 600 "$config_file"
  printf 'header = "X-Api-Key: %s"\nheader = "Accept: application/json"\n' \
    "$PLANE_GITHUB_AUTOMATION_KEY" >"$config_file"
  if [[ -n "$payload" ]]; then
    printf 'header = "Content-Type: application/json"\n' >>"$config_file"
    http_code="$(
      curl -sS --max-time 30 --config "$config_file" \
        -o "$body_file" -w '%{http_code}' -X "$method" --data "$payload" "$url" || true
    )"
  else
    http_code="$(
      curl -sS --max-time 30 --config "$config_file" \
        -o "$body_file" -w '%{http_code}' -X "$method" "$url" || true
    )"
  fi
  rm -f "$config_file"
  if [[ ! "$http_code" =~ ^[0-9]{3}$ ]]; then
    http_code="000"
  fi
  printf '%s\n' "$http_code"
}

plane_detail() {
  local body_file="$1"
  jq -r '.detail // .message // empty' "$body_file" 2>/dev/null | head -n 1
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
  local item_body patch_body comment_body
  local http_code identifier project_id item_uuid current_state decision
  local external_id comment_html list_url patch_url comment_url payload
  local reconcile_only=0

  item_body="$(mktemp)"
  patch_body="$(mktemp)"
  comment_body="$(mktemp)"

  http_code="$(plane_http GET \
    "${PLANE_API_ORIGIN}/api/v2/workspaces/${PLANE_WORKSPACE_SLUG}/work-items/${ticket_id}/?expand=state" \
    "" "$item_body")"
  if [[ "$http_code" != "200" ]]; then
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
    return 1
  fi
  if [[ -n "$project_id" && "$project_id" != "$PLANE_PROJECT_ID" ]]; then
    log_line "FAIL  ${ticket_id}  work item is outside project ${PLANE_PROJECT_IDENTIFIER}"
    rm -f "$item_body" "$patch_body" "$comment_body"
    return 1
  fi

  decision="$(decide_transition "$base" "$current_state")"
  # A previous run may have committed the state PATCH but failed to post its
  # comment. Reconcile the comment at the target without another state write.
  if [[ "$base" == "main" && "$decision" == "SKIP already" ]]; then
    decision=Review
    reconcile_only=1
  elif [[ "$base" == "staging" && "$(fold_state "$current_state")" == "uat" ]]; then
    decision=UAT
    reconcile_only=1
  fi
  case "$decision" in
    SKIP\ *)
      log_line "SKIP  ${ticket_id}  ${decision#SKIP }  current=${current_state:-unreadable}"
      rm -f "$item_body" "$patch_body" "$comment_body"
      if [[ "$decision" == "SKIP unknown" ]]; then
        return 1
      fi
      return 0
      ;;
    Review|UAT) ;;
    *)
      log_line "FAIL  ${ticket_id}  unexpected decision"
      rm -f "$item_body" "$patch_body" "$comment_body"
      return 1
      ;;
  esac

  if [[ "${PLANE_MERGE_STATE_DRY_RUN:-}" == "1" ]]; then
    log_line "DRY   ${ticket_id}  ${current_state} -> ${decision}"
    rm -f "$item_body" "$patch_body" "$comment_body"
    return 0
  fi

  if [[ "$reconcile_only" == "0" ]]; then
    payload="$(jq -n --arg state "$decision" '{state:$state}')"
    patch_url="${PLANE_API_ORIGIN}/api/v2/workspaces/${PLANE_WORKSPACE_SLUG}/projects/${PLANE_PROJECT_ID}/work-items/${item_uuid}/"
    http_code="$(plane_http PATCH "$patch_url" "$payload" "$patch_body")"
    if [[ "$http_code" != "200" ]]; then
      log_line "FAIL  ${ticket_id}  state update HTTP ${http_code} $(plane_detail "$patch_body")"
      rm -f "$item_body" "$patch_body" "$comment_body"
      return 1
    fi
  fi

  external_id="gh-merge:${pr_number}:${base}:${decision}"
  comment_html="$(build_comment_html "$base" "$decision" "$pr_url")"
  if [[ "$reconcile_only" == "1" ]]; then
    comment_html="<p>$(html_escape "$base") へのマージ後、現在の $(html_escape "$decision") 状態を確認し、このPRの監査コメントを補完しました。</p><p>PR: <a href=\"$(html_escape "$pr_url")\">$(html_escape "$pr_url")</a></p>"
  fi
  list_url="${PLANE_API_ORIGIN}/api/v2/workspaces/${PLANE_WORKSPACE_SLUG}/projects/${PLANE_PROJECT_ID}/work-items/${item_uuid}/comments/?external_source=github&external_id=$(jq -nr --arg v "$external_id" '$v|@uri')"
  http_code="$(plane_http GET "$list_url" "" "$comment_body")"
  if [[ "$http_code" != "200" ]]; then
    log_line "FAIL  ${ticket_id}  comment lookup HTTP ${http_code} $(plane_detail "$comment_body")"
    rm -f "$item_body" "$patch_body" "$comment_body"
    return 1
  fi
  if comment_already_posted "$comment_body" "$external_id"; then
    log_line "MOVE  ${ticket_id}  ${current_state} -> ${decision}  comment=already"
    rm -f "$item_body" "$patch_body" "$comment_body"
    return 0
  fi

  payload="$(jq -n \
    --arg html "$comment_html" \
    --arg external_id "$external_id" \
    '{comment_html:$html, access:"INTERNAL", external_source:"github", external_id:$external_id}')"
  comment_url="${PLANE_API_ORIGIN}/api/v2/workspaces/${PLANE_WORKSPACE_SLUG}/projects/${PLANE_PROJECT_ID}/work-items/${item_uuid}/comments/"
  http_code="$(plane_http POST "$comment_url" "$payload" "$comment_body")"
  if [[ "$http_code" != "200" && "$http_code" != "201" ]]; then
    log_line "FAIL  ${ticket_id}  moved to ${decision} but comment HTTP ${http_code} $(plane_detail "$comment_body")"
    rm -f "$item_body" "$patch_body" "$comment_body"
    return 1
  fi

  log_line "MOVE  ${ticket_id}  ${current_state} -> ${decision}"
  rm -f "$item_body" "$patch_body" "$comment_body"
  return 0
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

main() {
  local event_path="${GITHUB_EVENT_PATH:-}"
  local repository="${GITHUB_REPOSITORY:-}"
  local merged base number title body pr_url owner name
  local sources tickets_file ticket_id failures=0 rc

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

  while IFS= read -r ticket_id; do
    [[ -z "$ticket_id" ]] && continue
    # A conditional call would disable set -e inside the updater. Run it in a
    # subshell so one ticket failure is recorded and the rest still run.
    set +e
    (
      set -euo pipefail
      sync_one_ticket "$ticket_id" "$base" "$number" "$pr_url"
    )
    rc=$?
    set -e
    if [[ "$rc" -ne 0 ]]; then
      failures=$((failures + 1))
    fi
  done <"$tickets_file"

  if [[ "$failures" -ne 0 ]]; then
    log_line "FAIL  ${failures} work item update(s) failed"
    exit 1
  fi
  log_line "PASS  Plane states checked for ${base} PR #${number}"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
