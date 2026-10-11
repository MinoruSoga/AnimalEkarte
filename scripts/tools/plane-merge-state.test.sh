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
assert_eq "main already Review settles" "SETTLED Review" "$(decide_transition main Review)"
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
assert_eq "staging UAT settles when head is main" "SETTLED UAT" "$(decide_transition staging UAT main)"
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

MOCK_LOG="$(mktemp)"
MOCK_TICKET="EMR-1"

plane_http() {
  local method="$1" url="$2" payload="$3" body_file="$4"
  printf '%s %s\n' "$method" "$url" >>"$MOCK_LOG"
  case "$method $url" in
    "GET "*comments/*)
      if [[ "$MOCK_CGET_HAS" == "yes" ]]; then
        printf '[{"external_id":"%s"}]\n' "$MOCK_COMMENT_ID" >"$body_file"
      else
        printf '[]\n' >"$body_file"
      fi
      printf '%s\n' "$MOCK_CGET_CODE"
      ;;
    "POST "*comments/*)
      printf '{}\n' >"$body_file"
      printf '%s\n' "$MOCK_CPOST_CODE"
      ;;
    "PATCH "*)
      printf '{}\n' >"$body_file"
      printf '%s\n' "$MOCK_PATCH_CODE"
      ;;
    "GET "*work-items/*)
      printf '%s\n' "$MOCK_ITEM_BODY" >"$body_file"
      printf '%s\n' "$MOCK_ITEM_CODE"
      ;;
    *)
      printf 'unexpected plane_http call: %s %s\n' "$method" "$url" >&2
      printf '000\n'
      ;;
  esac
}

mock_state() {
  MOCK_ITEM_CODE=200
  MOCK_PATCH_CODE=200
  MOCK_CGET_CODE=200
  MOCK_CGET_HAS=no
  MOCK_CPOST_CODE=201
  MOCK_COMMENT_ID=""
  : >"$MOCK_LOG"
  MOCK_ITEM_BODY="{\"identifier\":\"${MOCK_TICKET}\",\"project_id\":\"${PLANE_PROJECT_ID}\",\"id\":\"uuid-1\",\"state\":{\"name\":\"$1\"}}"
}

call_count() {
  awk -v m="$1" '$1 == m' "$MOCK_LOG" | wc -l | tr -d ' '
}

comment_get_count() {
  grep -c 'comments/' "$MOCK_LOG" || true
}

run_sync() {
  local rc=0
  sync_one_ticket "$@" || rc=$?
  printf '%s' "$rc"
}

mock_state Review
assert_eq "feature-head staging skip is clean" "0" \
  "$(run_sync "$MOCK_TICKET" staging 600 https://example.test/pr/600 feature-x)"
assert_eq "feature-head staging sends no PATCH" "0" "$(call_count PATCH)"
assert_eq "feature-head staging sends no comment POST" "0" "$(call_count POST)"
assert_eq "feature-head staging does not reconcile comments" "0" "$(comment_get_count)"

mock_state Review
assert_eq "main-head staging Review moves to UAT" "0" \
  "$(run_sync "$MOCK_TICKET" staging 600 https://example.test/pr/600 main)"
assert_eq "main-head staging sends one PATCH" "1" "$(call_count PATCH)"
assert_eq "main-head staging posts one audit comment" "1" "$(call_count POST)"

mock_state UAT
assert_eq "main event on UAT does not regress" "0" \
  "$(run_sync "$MOCK_TICKET" main 600 https://example.test/pr/600 main)"
assert_eq "main event on UAT sends no PATCH" "0" "$(call_count PATCH)"
assert_eq "main event on UAT posts no comment" "0" "$(call_count POST)"

mock_state Mystery
assert_eq "unknown state fails closed" "1" \
  "$(run_sync "$MOCK_TICKET" main 600 https://example.test/pr/600 main)"
assert_eq "unknown state sends no PATCH" "0" "$(call_count PATCH)"

mock_state Running
MOCK_CGET_CODE=500
assert_eq "PATCH ok + comment lookup failure is a truthful failure" "1" \
  "$(run_sync "$MOCK_TICKET" main 600 https://example.test/pr/600 main)"
assert_eq "comment lookup failure still PATCHed state" "1" "$(call_count PATCH)"
assert_eq "comment lookup failure does not blind-POST" "0" "$(call_count POST)"

mock_state Running
MOCK_CPOST_CODE=500
assert_eq "PATCH ok + comment create failure is a truthful failure" "1" \
  "$(run_sync "$MOCK_TICKET" main 600 https://example.test/pr/600 main)"
assert_eq "comment create failure attempted one POST" "1" "$(call_count POST)"

mock_state UAT
assert_eq "rerun backfills missing comment on advanced state" "0" \
  "$(run_sync "$MOCK_TICKET" staging 600 https://example.test/pr/600 main)"
assert_eq "rerun on advanced state sends no PATCH" "0" "$(call_count PATCH)"
assert_eq "rerun on advanced state posts exactly one comment" "1" "$(call_count POST)"

mock_state UAT
MOCK_CGET_CODE=500
assert_eq "comment lookup failure on advanced state is a truthful failure" "1" \
  "$(run_sync "$MOCK_TICKET" staging 600 https://example.test/pr/600 main)"

mock_state UAT
MOCK_CGET_HAS=yes
MOCK_COMMENT_ID="gh-merge:600:staging:UAT"
assert_eq "existing comment marker is a clean no-op on rerun" "0" \
  "$(run_sync "$MOCK_TICKET" staging 600 https://example.test/pr/600 main)"
assert_eq "existing comment marker sends no PATCH" "0" "$(call_count PATCH)"
assert_eq "existing comment marker posts no duplicate" "0" "$(call_count POST)"

mock_state Review
MOCK_CGET_HAS=yes
MOCK_COMMENT_ID="gh-merge:600:main:Review"
assert_eq "main rerun with existing comment is a clean no-op" "0" \
  "$(run_sync "$MOCK_TICKET" main 600 https://example.test/pr/600 main)"
assert_eq "main rerun posts no duplicate comment" "0" "$(call_count POST)"

mock_state Review
assert_eq "main rerun backfills missing comment without PATCH" "0" \
  "$(run_sync "$MOCK_TICKET" main 600 https://example.test/pr/600 main)"
assert_eq "main rerun sends no PATCH when already Review" "0" "$(call_count PATCH)"
assert_eq "main rerun posts exactly one comment" "1" "$(call_count POST)"

echo "----"
echo "PASS ${PASS_COUNT}  FAIL ${FAIL_COUNT}"
if [[ "$FAIL_COUNT" -ne 0 ]]; then
  exit 1
fi
