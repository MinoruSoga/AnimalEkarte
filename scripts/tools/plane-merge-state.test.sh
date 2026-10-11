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

# Exercise the real orchestration with a file-backed HTTP fake. State writes
# survive command-substitution subshells; no network or credentials are used.
API_FIXTURE_ROOT="$(mktemp -d)"
trap 'rm -f "$FIXTURE"; rm -rf "$API_FIXTURE_ROOT"' EXIT
plane_http() {
  local method="$1" url="$2" payload="$3" dest="$4"
  printf '%s\n' "$method" >> "$CASE_DIR/calls"
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
      if [[ "$(cat "$CASE_DIR/fail_once")" == "1" ]]; then
        printf '0' > "$CASE_DIR/fail_once"
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
    ng "$base partial comment failure is reported"
  else
    ok "$base partial comment failure is reported"
  fi
  assert_eq "$base state PATCH succeeded before comment failure" "$target" "$(cat "$CASE_DIR/state")"
  if sync_one_ticket EMR-999 "$base" 606 'https://github.com/example/repo/pull/606' main >/dev/null; then
    ok "$base rerun recovers missing comment"
  else
    ng "$base rerun recovers missing comment"
  fi
  assert_eq "$base recovery does not repeat PATCH" '1' "$(grep -c '^PATCH$' "$CASE_DIR/calls")"
  assert_eq "$base recovery posts the intended comment identity" "gh-merge:606:$base:$target" \
    "$(jq -r '.[0].external_id' "$CASE_DIR/comments")"
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

printf '1' > "$CASE_DIR/fail_once"
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

echo "----"
echo "PASS ${PASS_COUNT}  FAIL ${FAIL_COUNT}"
if [[ "$FAIL_COUNT" -ne 0 ]]; then
  exit 1
fi
