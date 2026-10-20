#!/usr/bin/env bash
# plane-merge-state.sh の参照抽出と state 判定。ネットワークは使わない。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=plane-merge-state.sh
source "$SCRIPT_DIR/plane-merge-state.sh"

PASS_COUNT=0
FAIL_COUNT=0

ok() {
  PASS_COUNT=$((PASS_COUNT + 1))
  echo "PASS  $1"
}

ng() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  echo "FAIL  $1" >&2
}

assert_eq() {
  local label="$1" expected="$2" actual="$3"
  if [[ "$expected" == "$actual" ]]; then
    ok "$label"
  else
    ng "$label (expected $(printf '%q' "$expected") actual $(printf '%q' "$actual"))"
  fi
}

ids_of() {
  printf '%s' "$1" | extract_ticket_ids
}

prs_of() {
  printf '%s' "$1" | extract_referenced_pr_numbers
}

assert_eq "brackets, numeric order, unique" \
  $'EMR-3\nEMR-12' \
  "$(ids_of $'see [EMR-12] and [EMR-3] and [EMR-12]')"

assert_eq "bare identifiers count" $'EMR-3\nEMR-12' "$(ids_of 'EMR-12 and EMR-3')"
assert_eq "unclosed bracket still counts" "EMR-12" "$(ids_of '[EMR-12')"
assert_eq "lowercase is ignored" "" "$(ids_of '[emr-12]')"
assert_eq "zero and leading zero are ignored" "" "$(ids_of '[EMR-0] [EMR-01]')"
assert_eq "one bracketed identifier" "EMR-306" "$(ids_of 'prefix [EMR-306] suffix')"
assert_eq "fullwidth paren slash list" \
  $'EMR-282\nEMR-325\nEMR-326' \
  "$(ids_of 'feat: カルテ画面UX改善一式（EMR-282/325/326）')"
assert_eq "halfwidth paren slash list" \
  $'EMR-270\nEMR-274\nEMR-275\nEMR-318\nEMR-321' \
  "$(ids_of 'fix: 予約 (EMR-270/274/275/318/321)')"

assert_eq "squash and merge-commit PR numbers" \
  $'600\n601' \
  "$(prs_of $'feat: x (#600)\nMerge pull request #601 from org/branch')"
assert_eq "bare hash is not a PR reference" "" "$(prs_of 'see #602')"

assert_eq "main running becomes Review" "Review" "$(decide_transition main Running)"
assert_eq "main UAT never regresses" "SKIP uat-or-later" "$(decide_transition main UAT)"
assert_eq "main Backlog becomes Review" "Review" "$(decide_transition main Backlog)"
assert_eq "main Ready becomes Review" "Review" "$(decide_transition main Ready)"
assert_eq "main Research becomes Review" "Review" "$(decide_transition main Research)"
assert_eq "main cutover state becomes Review" "Review" "$(decide_transition main '本番切替対応')"
assert_eq "main already Review skips" "SKIP already" "$(decide_transition main Review)"
assert_eq "main Done stays" "SKIP protected" "$(decide_transition main Done)"
assert_eq "main Cancelled stays" "SKIP protected" "$(decide_transition main Cancelled)"
assert_eq "main Duplicate stays" "SKIP protected" "$(decide_transition main Duplicate)"
assert_eq "main Needs Human stays" "SKIP protected" "$(decide_transition main 'Needs Human')"
assert_eq "main Blocked stays" "SKIP protected" "$(decide_transition main Blocked)"
assert_eq "main Phase2 stays" "SKIP protected" "$(decide_transition main Phase2)"
assert_eq "main protected state is case-insensitive" "SKIP protected" "$(decide_transition main "done")"
assert_eq "main unknown state is not moved" "SKIP unknown" "$(decide_transition main Mystery)"
assert_eq "main unreadable state is not moved" "SKIP unknown" "$(decide_transition main '')"

assert_eq "staging Review becomes UAT when head is main" "UAT" "$(decide_transition staging Review main)"
assert_eq "staging UAT stays when head is main" "SKIP not-review" "$(decide_transition staging UAT main)"
assert_eq "staging Running stays" "SKIP not-review" "$(decide_transition staging Running main)"
assert_eq "staging Done stays protected" "SKIP protected" "$(decide_transition staging Done main)"
assert_eq "staging unknown state is not moved" "SKIP unknown" "$(decide_transition staging Mystery main)"
assert_eq "staging feature-head does not move to UAT" "SKIP head-not-main" "$(decide_transition staging Review feature-x)"
assert_eq "staging feature-head unknown state still fails closed" "SKIP unknown" "$(decide_transition staging Mystery feature-x)"
assert_eq "production is unsupported" "SKIP unsupported-base" "$(decide_transition production Review)"

assert_eq "merged main PR qualifies" "yes" "$(pr_qualifies_for_staging_tickets main true 600 700)"
assert_eq "current PR does not qualify" "no" "$(pr_qualifies_for_staging_tickets main true 600 600)"
assert_eq "unmerged PR does not qualify" "no" "$(pr_qualifies_for_staging_tickets main false 600 700)"
assert_eq "non-main base does not qualify" "no" "$(pr_qualifies_for_staging_tickets staging true 600 700)"

FIXTURE="$(mktemp)"
trap 'rm -f "$FIXTURE"' EXIT
cat >"$FIXTURE" <<'EOF'
{
  "data": {
    "repository": {
      "pullRequest": {
        "commits": {
          "pageInfo": { "hasNextPage": false, "endCursor": null },
          "nodes": [
            {
              "commit": {
                "message": "land [EMR-10]",
                "associatedPullRequests": {
                  "pageInfo": { "hasNextPage": false },
                  "nodes": [
                    {
                      "number": 5,
                      "merged": true,
                      "baseRefName": "main",
                      "title": "feat: [EMR-11]",
                      "body": "see EMR-12"
                    },
                    {
                      "number": 6,
                      "merged": true,
                      "baseRefName": "staging",
                      "title": "[EMR-13]",
                      "body": ""
                    },
                    {
                      "number": 7,
                      "merged": false,
                      "baseRefName": "main",
                      "title": "[EMR-14]",
                      "body": ""
                    },
                    {
                      "number": 9,
                      "merged": true,
                      "baseRefName": "main",
                      "title": "[EMR-15]",
                      "body": ""
                    }
                  ]
                }
              }
            }
          ]
        }
      }
    }
  }
}
EOF

assert_eq "staging page keeps new main tickets only" \
  $'EMR-10\nEMR-11\nEMR-12' \
  "$(tickets_from_graphql_payload staging 9 "$FIXTURE" | extract_ticket_ids)"
assert_eq "main page uses commit messages only" \
  "EMR-10" \
  "$(tickets_from_graphql_payload main 9 "$FIXTURE" | extract_ticket_ids)"

assert_eq "comment escapes markup" \
  '<p>main へのマージに合わせて Review に移動しました。</p><p>PR: <a href="https://example.com/a?b=&amp;c=&lt;d&gt;">https://example.com/a?b=&amp;c=&lt;d&gt;</a></p>' \
  "$(build_comment_html main Review 'https://example.com/a?b=&c=<d>')"

assert_eq "escape all attribute delimiters without backslashes" \
  'a&amp;b&lt;c&gt;&quot;d' "$(html_escape 'a&b<c>"d')"
assert_eq "escape repeated delimiters and preserve backslashes" \
  '\&amp;&amp;&lt;&lt;&gt;&gt;&quot;&quot;' "$(html_escape '\&&<<>>""')"

# Exercise the real orchestration with a file-backed HTTP fake. State writes
# survive command-substitution subshells; no network or credentials are used.
API_FIXTURE_ROOT="$(mktemp -d)"
RATE_ROOT="$(mktemp -d)"
trap 'rm -f "$FIXTURE"; rm -rf "$API_FIXTURE_ROOT" "$RATE_ROOT"' EXIT
# Deterministic policy for the orchestration fixtures below.
export PLANE_MERGE_STATE_THROTTLE_MS=0
plane_sleep_seconds() {
  # No real sleeping in tests; waits are charged to the budget counter.
  :
}
plane_http_once() {
  local method="$1" url="$2" payload="$3" dest="$4" headers_file="$5"
  printf '%s\n' "$method" >> "$CASE_DIR/calls"
  : > "$headers_file"
  case "$method:$url" in
    GET:*/work-items/EMR-999/\?expand=state)
      jq -n --arg state "$(cat "$CASE_DIR/state")" --arg project "$PLANE_PROJECT_ID" \
        '{identifier:"EMR-999", id:"fixture-item", project_id:$project, state:{name:$state}}' > "$dest"
      printf '200'
      ;;
    PATCH:*/work-items/fixture-item/)
      jq -r '.state' <<< "$payload" > "$CASE_DIR/state"
      printf '{}' > "$dest"
      printf '200'
      ;;
    GET:*/comments/\?*)
      cat "$CASE_DIR/comments" > "$dest"
      printf '%s' "$(cat "$CASE_DIR/comment_lookup_status")"
      ;;
    POST:*/comments/)
      if [[ "$(cat "$CASE_DIR/fail_once")" -gt 0 ]]; then
        printf '%s' "$(( $(cat "$CASE_DIR/fail_once") - 1 ))" > "$CASE_DIR/fail_once"
        printf '{"detail":"fixture transient failure"}' > "$dest"
        printf '503'
      else
        jq -n --argjson comment "$payload" '[$comment]' > "$CASE_DIR/comments"
        printf '{}' > "$dest"
        printf '201'
      fi
      ;;
    *)
      printf '{}' > "$dest"
      printf '500'
      ;;
  esac
}

for base in main staging; do
  CASE_DIR="$API_FIXTURE_ROOT/$base"
  mkdir -p "$CASE_DIR"
  initial=Running
  target=Review
  [[ "$base" == "staging" ]] && { initial=Review; target=UAT; }
  printf '%s' "$initial" > "$CASE_DIR/state"
  printf '[]' > "$CASE_DIR/comments"
  printf '1' > "$CASE_DIR/fail_once"
  printf '200' > "$CASE_DIR/comment_lookup_status"
  : > "$CASE_DIR/calls"
  if sync_one_ticket EMR-999 "$base" 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
    ok "$base transient comment failure heals through the dedupe re-check"
  else
    ng "$base transient comment failure heals through the dedupe re-check"
  fi
  assert_eq "$base state PATCH succeeded before comment failure" "$target" "$(cat "$CASE_DIR/state")"
  assert_eq "$base in-run comment retry does not repeat PATCH" '1' "$(grep -c '^PATCH$' "$CASE_DIR/calls")"
  assert_eq "$base in-run comment retry posts exactly twice" '2' "$(grep -c '^POST$' "$CASE_DIR/calls")"
  assert_eq "$base comment identity is preserved across the retry" "gh-merge:606:$base:$target" \
    "$(jq -r '.[0].external_id' "$CASE_DIR/comments")"
  if sync_one_ticket EMR-999 "$base" 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
    ok "$base rerun is idempotent after the in-run recovery"
  else
    ng "$base rerun is idempotent after the in-run recovery"
  fi
  assert_eq "$base rerun does not repeat PATCH" '1' "$(grep -c '^PATCH$' "$CASE_DIR/calls")"
  if sync_one_ticket EMR-999 "$base" 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
    ok "$base completed reconciliation is idempotent"
  else
    ng "$base completed reconciliation is idempotent"
  fi
  assert_eq "$base completed rerun does not repeat POST" '2' "$(grep -c '^POST$' "$CASE_DIR/calls")"

  # A failed comment lookup must not be treated as evidence of absence.
  printf '503' > "$CASE_DIR/comment_lookup_status"
  if sync_one_ticket EMR-999 "$base" 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
    ng "$base comment lookup failure is reported"
  else
    ok "$base comment lookup failure is reported"
  fi
  assert_eq "$base lookup failure does not risk duplicate POST" '2' "$(grep -c '^POST$' "$CASE_DIR/calls")"
  printf '200' > "$CASE_DIR/comment_lookup_status"
  for body in '{"unexpected":"not a collection"}' '<html>fixture error</html>' '["invalid comment"]'; do
    printf '%s' "$body" > "$CASE_DIR/comments"
    if sync_one_ticket EMR-999 "$base" 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
      ng "$base invalid collection is rejected"
    else
      ok "$base invalid collection is rejected"
    fi
    assert_eq "$base invalid collection does not risk duplicate POST" '2' "$(grep -c '^POST$' "$CASE_DIR/calls")"
  done
done

CASE_DIR="$API_FIXTURE_ROOT/staging-feature-head"
mkdir -p "$CASE_DIR"
printf 'UAT' > "$CASE_DIR/state"
printf '[]' > "$CASE_DIR/comments"
printf '0' > "$CASE_DIR/fail_once"
printf '200' > "$CASE_DIR/comment_lookup_status"
: > "$CASE_DIR/calls"
if sync_one_ticket EMR-999 staging 606 'https://github.com/example/repo/pull/606' feature-x >/dev/null; then
  ok "staging feature-head UAT skips cleanly"
else
  ng "staging feature-head UAT skips cleanly"
fi
assert_eq "staging feature-head UAT sends no PATCH" '0' "$(grep -c '^PATCH$' "$CASE_DIR/calls")"
assert_eq "staging feature-head UAT sends no POST" '0' "$(grep -c '^POST$' "$CASE_DIR/calls")"

CASE_DIR="$API_FIXTURE_ROOT/staging-uat"
mkdir -p "$CASE_DIR"
printf 'UAT' > "$CASE_DIR/state"
printf '[]' > "$CASE_DIR/comments"
printf '0' > "$CASE_DIR/fail_once"
printf '200' > "$CASE_DIR/comment_lookup_status"
: > "$CASE_DIR/calls"
if out="$(sync_one_ticket EMR-999 staging 606 'https://github.com/example/repo/pull/606' main 2>&1 >/dev/null)"; then
  ok "staging UAT head=main backfills missing comment"
else
  ng "staging UAT head=main backfills missing comment"
fi
assert_eq "staging UAT head=main sends no PATCH" '0' "$(grep -c '^PATCH$' "$CASE_DIR/calls")"
assert_eq "staging UAT head=main posts one audit comment" '1' "$(grep -c '^POST$' "$CASE_DIR/calls")"
assert_eq "staging UAT reconcile keeps UAT comment identity" "gh-merge:606:staging:UAT" \
  "$(jq -r '.[0].external_id' "$CASE_DIR/comments")"
assert_eq "staging UAT reconcile logs AUDIT state=unchanged" \
  "AUDIT  EMR-999  current=UAT target=UAT state=unchanged comment=backfilled" "$out"
assert_eq "staging UAT reconcile logs no MOVE line" '0' "$(printf '%s' "$out" | grep -c 'MOVE')"

CASE_DIR="$API_FIXTURE_ROOT/main-uat"
mkdir -p "$CASE_DIR"
printf 'UAT' > "$CASE_DIR/state"
printf '[]' > "$CASE_DIR/comments"
printf '0' > "$CASE_DIR/fail_once"
printf '200' > "$CASE_DIR/comment_lookup_status"
: > "$CASE_DIR/calls"
if out="$(sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main 2>&1 >/dev/null)"; then
  ok "main UAT backfills missing comment"
else
  ng "main UAT backfills missing comment"
fi
assert_eq "main UAT sends no PATCH" '0' "$(grep -c '^PATCH$' "$CASE_DIR/calls")"
assert_eq "main UAT posts one audit comment" '1' "$(grep -c '^POST$' "$CASE_DIR/calls")"
assert_eq "main UAT keeps Review comment identity" "gh-merge:606:main:Review" \
  "$(jq -r '.[0].external_id' "$CASE_DIR/comments")"
assert_eq "main UAT reconcile logs AUDIT state=unchanged" \
  "AUDIT  EMR-999  current=UAT target=Review state=unchanged comment=backfilled" "$out"
assert_eq "main UAT reconcile logs no MOVE line" '0' "$(printf '%s' "$out" | grep -c 'MOVE')"
assert_eq "main UAT audit comment names the actual state" '1' \
  "$(jq -r '.[0].comment_html' "$CASE_DIR/comments" | grep -c '現在の UAT')"
assert_eq "main UAT audit comment states no state write" '1' \
  "$(jq -r '.[0].comment_html' "$CASE_DIR/comments" | grep -c '状態変更なし')"
if out="$(sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main 2>&1 >/dev/null)"; then
  ok "main UAT rerun is a clean no-op"
else
  ng "main UAT rerun is a clean no-op"
fi
assert_eq "main UAT rerun posts no duplicate comment" '1' "$(grep -c '^POST$' "$CASE_DIR/calls")"
assert_eq "main UAT rerun logs AUDIT comment=already" \
  "AUDIT  EMR-999  current=UAT target=Review state=unchanged comment=already" "$out"
assert_eq "main UAT rerun logs no MOVE line" '0' "$(printf '%s' "$out" | grep -c 'MOVE')"

printf '503' > "$CASE_DIR/comment_lookup_status"
if out="$(sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main 2>&1 >/dev/null)"; then
  ng "main UAT comment lookup failure is reported"
else
  ok "main UAT comment lookup failure is reported"
fi
assert_eq "main UAT lookup failure logs audit reconciliation wording" '1' \
  "$(printf '%s' "$out" | grep -c 'audit reconciliation current=UAT target=Review state=unchanged comment HTTP 503')"
assert_eq "main UAT lookup failure logs no moved wording" '0' "$(printf '%s' "$out" | grep -c 'moved')"
assert_eq "main UAT lookup failure leaves state unchanged" 'UAT' "$(cat "$CASE_DIR/state")"
printf '200' > "$CASE_DIR/comment_lookup_status"

printf '4' > "$CASE_DIR/fail_once"
printf '[]' > "$CASE_DIR/comments"
if out="$(sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main 2>&1 >/dev/null)"; then
  ng "main UAT comment POST failure is reported"
else
  ok "main UAT comment POST failure is reported"
fi
assert_eq "main UAT POST failure logs audit reconciliation wording" '1' \
  "$(printf '%s' "$out" | grep -c 'audit reconciliation current=UAT target=Review state=unchanged comment HTTP 503')"
assert_eq "main UAT POST failure logs no moved wording" '0' "$(printf '%s' "$out" | grep -c 'moved')"
assert_eq "main UAT POST failure leaves state unchanged" 'UAT' "$(cat "$CASE_DIR/state")"

# The workflow uses one repository-wide queue. Exercise the real HTTP fake
# for distinct PRs sharing a ticket in each serialized main/staging order.
for order in main-first staging-first; do
  CASE_DIR="$API_FIXTURE_ROOT/serialized-$order"
  mkdir -p "$CASE_DIR"
  initial=Running
  [[ "$order" == staging-first ]] && initial=Review
  printf '%s' "$initial" > "$CASE_DIR/state"
  printf '[]' > "$CASE_DIR/comments"
  printf '0' > "$CASE_DIR/fail_once"
  printf '200' > "$CASE_DIR/comment_lookup_status"
  : > "$CASE_DIR/calls"
  if [[ "$order" == main-first ]]; then
    sync_one_ticket EMR-999 main 607 'https://example.test/pr/607' feature-a >/dev/null
  fi
  sync_one_ticket EMR-999 staging 608 'https://example.test/pr/608' main >/dev/null
  sync_one_ticket EMR-999 main 609 'https://example.test/pr/609' feature-b >/dev/null
  assert_eq "$order serialized merges keep UAT" UAT "$(cat "$CASE_DIR/state")"
  expected_patches=2
  [[ "$order" == staging-first ]] && expected_patches=1
  assert_eq "$order later main event performs no regressive PATCH" "$expected_patches" \
    "$(grep -c '^PATCH$' "$CASE_DIR/calls")"
done

# ---------------------------------------------------------------------------
# EMR-442: finite retry, whole-run wait budget, Retry-After handling, and
# write-result reconciliation. The plan-driven fake replays exact request
# sequences: each consumed line is "status|retry-after-value|body-kind".
RATE_EPOCH_NOW=1760000000
plane_now_epoch() { printf '%s\n' "$RATE_EPOCH_NOW"; }
SLEEP_LOG="$RATE_ROOT/sleeps"
plane_sleep_seconds() {
  printf '%s\n' "$1" >> "$SLEEP_LOG"
}
plane_http_once() {
  local method="$1" url="$2" payload="$3" dest="$4" headers_file="$5"
  local line status rahdr bodykind state_match
  printf '%s %s\n' "$method" \
    "$(printf '%s' "$url" | sed -E 's#https?://[^/]+##; s#^/api/v2/##; s#\?.*##')" >> "$CASE_DIR/calls"
  line="$(head -n 1 "$CASE_DIR/plan" 2>/dev/null || true)"
  if [[ -z "$line" ]]; then
    printf '{}' > "$dest"
    : > "$headers_file"
    printf '500'
    return 0
  fi
  tail -n +2 "$CASE_DIR/plan" > "$CASE_DIR/plan.next" && mv "$CASE_DIR/plan.next" "$CASE_DIR/plan"
  IFS='|' read -r status rahdr bodykind <<< "$line"
  local ticket_from_url
  ticket_from_url="$(printf '%s' "$url" | sed -E 's#.*/work-items/([^/]+)/.*#\1#')"
  case "$bodykind" in
    item:*)
      state_match="${bodykind#item:}"
      jq -n --arg id "$ticket_from_url" --arg state "$state_match" --arg project "$PLANE_PROJECT_ID" \
        '{identifier:$id, id:"fixture-item", project_id:$project, state:{name:$state}}' > "$dest"
      ;;
    item)
      jq -n --arg id "$ticket_from_url" --arg state "$(cat "$CASE_DIR/state")" --arg project "$PLANE_PROJECT_ID" \
        '{identifier:$id, id:"fixture-item", project_id:$project, state:{name:$state}}' > "$dest"
      ;;
    comments-posted)
      jq -n --arg e "$(cat "$CASE_DIR/posted_ext")" '[{external_id:$e}]' > "$dest"
      ;;
    comments)
      cat "$CASE_DIR/comments" > "$dest"
      ;;
    detail)
      printf '{"detail":"fixture"}' > "$dest"
      ;;
    *)
      printf '{}' > "$dest"
      ;;
  esac
  if [[ "$status" == "200" && "$method" == "PATCH" ]]; then
    jq -r '.state' <<< "$payload" > "$CASE_DIR/state"
  fi
  if [[ "$status" == "201" && "$method" == "POST" ]]; then
    jq -n --argjson comment "$payload" '[$comment]' > "$CASE_DIR/comments"
    jq -r '.external_id' <<< "$payload" > "$CASE_DIR/posted_ext"
  fi
  if [[ -n "$rahdr" ]]; then
    printf 'Retry-After: %s\n' "$rahdr" > "$headers_file"
  else
    : > "$headers_file"
  fi
  printf '%s' "$status"
}

rate_case() {
  CASE_DIR="$RATE_ROOT/$1"
  mkdir -p "$CASE_DIR"
  printf 'Running' > "$CASE_DIR/state"
  printf '[]' > "$CASE_DIR/comments"
  : > "$CASE_DIR/calls"
  : > "$SLEEP_LOG"
  RUN_WAIT_BUDGET_REMAINING_MS=120000
  # shellcheck disable=SC2034  # read by plane-merge-state.sh at runtime
  RUN_BUDGET_EXHAUSTED=0
  # shellcheck disable=SC2034  # read by plane-merge-state.sh at runtime
  RUN_ANY_REQUEST_MADE=""
  PLANE_MERGE_STATE_MAX_ATTEMPTS=4
}

rate_calls() { grep -c "$1" "$CASE_DIR/calls" || true; }

# lookup 429 (Retry-After: 2) then 200: respected wait, then success.
rate_case "429-then-200"
printf '%s\n' '429|2|detail' '200||item' '200||empty' '200||comments' '201||empty' > "$CASE_DIR/plan"
if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
  ok "lookup 429 with Retry-After recovers and completes"
else
  ng "lookup 429 with Retry-After recovers and completes"
fi
assert_eq "lookup 429 retries the same ticket finitely" '2' "$(rate_calls '^GET .*work-items/EMR-999')"
assert_eq "lookup 429 honors the Retry-After wait" '1' "$(grep -c '^2\.000$' "$SLEEP_LOG")"
assert_eq "lookup 429 budget spends only the honored wait" '118000' "$RUN_WAIT_BUDGET_REMAINING_MS"

# continuous 429 exhausts the bounded attempts and fails the ticket.
rate_case "continuous-429"
printf '%s\n' '429|2|detail' '429|2|detail' '429|2|detail' '429|2|detail' > "$CASE_DIR/plan"
if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
  ng "continuous 429 fails the ticket after bounded attempts"
else
  ok "continuous 429 fails the ticket after bounded attempts"
fi
assert_eq "continuous 429 stops at the attempt cap" '4' "$(rate_calls '^GET .*work-items/EMR-999')"
assert_eq "continuous 429 sends no PATCH" '0' "$(rate_calls '^PATCH')"
assert_eq "continuous 429 sends no POST" '0' "$(rate_calls '^POST')"

# missing Retry-After falls back to capped exponential backoff.
rate_case "headerless-429"
printf '%s\n' '429||detail' '429||detail' '429||detail' '429||detail' > "$CASE_DIR/plan"
sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null 2>&1 || true
assert_eq "headerless 429 backs off exponentially" '1' "$(grep -c '^2\.000$' "$SLEEP_LOG")"
assert_eq "headerless 429 second step doubles" '1' "$(grep -c '^4\.000$' "$SLEEP_LOG")"
assert_eq "headerless 429 third step doubles again" '1' "$(grep -c '^8\.000$' "$SLEEP_LOG")"

# invalid Retry-After value falls back to the same backoff path.
rate_case "invalid-retry-after"
printf '%s\n' '429|soon|detail' '429|soon|detail' '200||item' '200||empty' '200||comments' '201||empty' > "$CASE_DIR/plan"
if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
  ok "invalid Retry-After falls back to capped backoff and recovers"
else
  ng "invalid Retry-After falls back to capped backoff and recovers"
fi
assert_eq "invalid Retry-After waits the backoff step" '1' "$(grep -c '^2\.000$' "$SLEEP_LOG")"

# HTTP-date Retry-After converts to the remaining seconds. The date string is
# parsed back through the script's own dialect-falling-back parser so the
# expectation holds on both GNU and BSD date hosts.
rate_case "httpdate-retry-after"
HTTPDATE_FIXTURE="Wed, 21 Oct 2026 12:00:00 GMT"
RATE_EPOCH_NOW=$(( $(plane_httpdate_to_epoch "$HTTPDATE_FIXTURE") - 37 ))
printf '%s\n' "429|$HTTPDATE_FIXTURE|detail" '200||item' '200||empty' '200||comments' '201||empty' > "$CASE_DIR/plan"
if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
  ok "HTTP-date Retry-After converts to a wait"
else
  ng "HTTP-date Retry-After converts to a wait"
fi
assert_eq "HTTP-date Retry-After waits the delta" '1' "$(grep -c '^37\.000$' "$SLEEP_LOG")"

# a Retry-After beyond the whole-run budget stops without re-hitting.
rate_case "budget-exhausted"
RUN_WAIT_BUDGET_REMAINING_MS=5000
printf '%s\n' '429|3600|detail' > "$CASE_DIR/plan"
if out="$(sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main 2>&1 >/dev/null)"; then
  ng "retry wait beyond the budget defers the ticket"
else
  ok "retry wait beyond the budget defers the ticket"
fi
assert_eq "budget-exhausted defers without re-hitting the API" '1' "$(rate_calls '^GET .*work-items/EMR-999')"
assert_eq "budget-exhausted logs the deferred wording" '1' \
  "$(printf '%s' "$out" | grep -c 'lookup deferred: wait budget exhausted')"

# 401/403 are permanent and never retried.
for code in 401 403; do
  rate_case "permanent-$code"
  printf '%s\n' "$code||detail" > "$CASE_DIR/plan"
  if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
    ng "HTTP $code lookup fails without retry"
  else
    ok "HTTP $code lookup fails without retry"
  fi
  assert_eq "HTTP $code sends exactly one request" '1' "$(rate_calls '^GET .*work-items/EMR-999')"
done

# transport failure on a lookup retries within the same bounded attempts.
rate_case "lookup-transport-failure"
printf '%s\n' '000||detail' '200||item' '200||empty' '200||comments' '201||empty' > "$CASE_DIR/plan"
if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
  ok "lookup transport failure recovers within bounded attempts"
else
  ng "lookup transport failure recovers within bounded attempts"
fi
assert_eq "lookup transport failure retries once" '2' "$(rate_calls '^GET .*work-items/EMR-999')"

# PATCH uncertainty: verify shows the target state; no duplicate state write.
rate_case "patch-landed-despite-uncertainty"
printf '%s\n' '200||item' '000||detail' '200||item:Review' '200||comments' '201||empty' > "$CASE_DIR/plan"
if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
  ok "uncertain PATCH that landed is not repeated"
else
  ng "uncertain PATCH that landed is not repeated"
fi
assert_eq "uncertain-but-landed PATCH sends no second write" '1' "$(rate_calls '^PATCH')"
assert_eq "uncertain-but-landed PATCH still posts the comment" '1' "$(rate_calls '^POST')"

# PATCH uncertainty: verify shows the old state; the write is retried once.
rate_case "patch-not-executed"
printf '%s\n' '200||item' '000||detail' '200||item:Running' '200||empty' '200||comments' '201||empty' > "$CASE_DIR/plan"
if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
  ok "unexecuted PATCH is retried after verification"
else
  ng "unexecuted PATCH is retried after verification"
fi
assert_eq "verified-unexecuted PATCH is retried exactly once" '2' "$(rate_calls '^PATCH')"

# PATCH uncertainty whose verify lookup also fails stops as UNKNOWN.
rate_case "patch-verify-unavailable"
printf '%s\n' '200||item' '000||detail' '000||detail' '000||detail' '000||detail' '000||detail' > "$CASE_DIR/plan"
if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
  ng "unverifiable PATCH stops as UNKNOWN without posting"
else
  ok "unverifiable PATCH stops as UNKNOWN without posting"
fi
assert_eq "unverifiable PATCH sends no comment POST" '0' "$(rate_calls '^POST')"
assert_eq "unverifiable PATCH leaves the state file unchanged" 'Running' "$(cat "$CASE_DIR/state")"

# comment POST uncertainty: the dedupe re-check finds it already posted.
rate_case "post-landed-despite-uncertainty"
printf 'gh-merge:606:main:Review' > "$CASE_DIR/posted_ext"
printf '%s\n' '200||item' '200||empty' '200||comments' '000||detail' '200||comments-posted' > "$CASE_DIR/plan"
if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
  ok "uncertain comment POST that landed is not repeated"
else
  ng "uncertain comment POST that landed is not repeated"
fi
assert_eq "uncertain-but-landed POST sends no duplicate" '1' "$(rate_calls '^POST')"

# comment POST uncertainty with a confirmed absence retries exactly once.
rate_case "post-not-executed"
printf '%s\n' '200||item' '200||empty' '200||comments' '000||detail' '200||comments' '201||empty' > "$CASE_DIR/plan"
if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
  ok "unexecuted comment POST is retried after the re-check"
else
  ng "unexecuted comment POST is retried after the re-check"
fi
assert_eq "verified-unexecuted POST is retried exactly once" '2' "$(rate_calls '^POST')"

# partial success then rerun: the rerun must not PATCH or POST again.
rate_case "partial-success-rerun"
printf '%s\n' '429|1|detail' '200||item' '200||empty' '200||comments' '000||detail' '200||comments' '201||empty' > "$CASE_DIR/plan"
sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null 2>&1 || true
before_patches="$(rate_calls '^PATCH')"
before_posts="$(rate_calls '^POST')"
printf '%s\n' '200||item' '200||comments-posted' > "$CASE_DIR/plan"
if sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
  ok "rerun after partial success completes idempotently"
else
  ng "rerun after partial success completes idempotently"
fi
assert_eq "partial-success rerun sends no new PATCH" '0' "$(( $(rate_calls '^PATCH') - before_patches ))"
assert_eq "partial-success rerun sends no new POST" '0' "$(( $(rate_calls '^POST') - before_posts ))"

# request pacing: every request after the first is throttled.
rate_case "throttle-pacing"
PLANE_MERGE_STATE_THROTTLE_MS=400
printf '%s\n' '200||item' '200||empty' '200||comments' '201||empty' > "$CASE_DIR/plan"
sync_one_ticket EMR-999 main 606 'https://github.com/example/repo/pull/606' main >/dev/null 2>&1 || true
assert_eq "inter-request throttle sleeps between requests" '3' "$(grep -c '^0\.400$' "$SLEEP_LOG")"
PLANE_MERGE_STATE_THROTTLE_MS=0

# 25-ticket equivalent: shared budget forces a finite stop with a summary.
# The fake shares one comment collection across tickets, so after the first
# full success the next three tickets resolve through the external_id dedupe
# (no PATCH/POST), and the rest are deferred by the exhausted budget — the
# point is the finite stop and the summary, not the exact per-ticket path.
rate_case "twentyfive-tickets"
RUN_WAIT_BUDGET_REMAINING_MS=8000
# shellcheck disable=SC2034  # read by plane-merge-state.sh at runtime
PLANE_MERGE_STATE_MAX_ATTEMPTS=2
{
  printf '%s\n' '429||detail' '200||item' '200||empty' '200||comments' '201||empty'
  i=0
  while (( i < 3 )); do
    printf '%s\n' '429||detail' '200||item' '200||comments'
    i=$((i + 1))
  done
  i=0
  while (( i < 42 )); do
    printf '%s\n' '429||detail'
    i=$((i + 1))
  done
} > "$CASE_DIR/plan"
tickets_file="$CASE_DIR/tickets"
seq 1 25 | sed 's/^/EMR-/' > "$tickets_file"
if run_ticket_passes "$tickets_file" main 606 'https://github.com/example/repo/pull/606' main > "$CASE_DIR/summary.out" 2>&1; then
  ng "25-ticket 429 mix stops finitely with unresolved tickets"
else
  ok "25-ticket 429 mix stops finitely with unresolved tickets"
fi
assert_eq "25-ticket run reports the summary line" '1' \
  "$(grep -c '^SUMMARY  ok=4 skip=0 fail=0 unknown=0 deferred=21' "$CASE_DIR/summary.out")"
assert_eq "25-ticket run stays within the bounded request count" '1' \
  "$(( $(wc -l < "$CASE_DIR/calls" | tr -d ' ') <= 90 ? 1 : 0 ))"
assert_eq "25-ticket successful tickets are never re-written" '1' "$(rate_calls '^PATCH')"
assert_eq "25-ticket comment is never posted twice" '1' "$(rate_calls '^POST')"
assert_eq "25-ticket run never spends a negative budget" '1' \
  "$(( RUN_WAIT_BUDGET_REMAINING_MS >= 0 ? 1 : 0 ))"

echo "----"
echo "PASS ${PASS_COUNT}  FAIL ${FAIL_COUNT}"
if [[ "$FAIL_COUNT" -ne 0 ]]; then
  exit 1
fi
