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

assert_eq "brackets only, numeric order, unique" \
  $'EMR-3\nEMR-12' \
  "$(ids_of $'see [EMR-12] and [EMR-3] and [EMR-12]')"

assert_eq "bare identifier is ignored" "" "$(ids_of 'EMR-12 and EMR-3')"
assert_eq "unclosed bracket is ignored" "" "$(ids_of '[EMR-12')"
assert_eq "lowercase is ignored" "" "$(ids_of '[emr-12]')"
assert_eq "zero and leading zero are ignored" "" "$(ids_of '[EMR-0] [EMR-01]')"
assert_eq "one bracketed identifier" "EMR-306" "$(ids_of 'prefix [EMR-306] suffix')"

assert_eq "squash and merge-commit PR numbers" \
  $'600\n601' \
  "$(prs_of $'feat: x (#600)\nMerge pull request #601 from org/branch')"
assert_eq "bare hash is not a PR reference" "" "$(prs_of 'see #602')"

assert_eq "main running becomes Review" "Review" "$(decide_transition main Running)"
assert_eq "main UAT returns to Review" "Review" "$(decide_transition main UAT)"
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

assert_eq "staging Review becomes UAT" "UAT" "$(decide_transition staging Review)"
assert_eq "staging UAT stays" "SKIP not-review" "$(decide_transition staging UAT)"
assert_eq "staging Running stays" "SKIP not-review" "$(decide_transition staging Running)"
assert_eq "staging Done stays protected" "SKIP protected" "$(decide_transition staging Done)"
assert_eq "staging unknown state is not moved" "SKIP unknown" "$(decide_transition staging Mystery)"
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
  $'EMR-10\nEMR-11' \
  "$(tickets_from_graphql_payload staging 9 "$FIXTURE" | extract_ticket_ids)"
assert_eq "main page uses commit messages only" \
  "EMR-10" \
  "$(tickets_from_graphql_payload main 9 "$FIXTURE" | extract_ticket_ids)"

assert_eq "comment escapes markup" \
  '<p>main へのマージに合わせて Review に移動しました。</p><p>PR: <a href="https://example.com/a?b=&amp;c=&lt;d&gt;">https://example.com/a?b=&amp;c=&lt;d&gt;</a></p>' \
  "$(build_comment_html main Review 'https://example.com/a?b=&c=<d>')"

echo "----"
echo "PASS ${PASS_COUNT}  FAIL ${FAIL_COUNT}"
if [[ "$FAIL_COUNT" -ne 0 ]]; then
  exit 1
fi
