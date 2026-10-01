#!/usr/bin/env bash
# scenarios/ acceptance coverage gate — read-only.
#
# Layer 1: document integrity
#   - scenarios/README.md index <-> S*/V* files (both directions)
#   - S/V template compliance (required sections, H1 id, depth=深い -> 異常系)
#   - "no execution results in scenario files" rule (UAT-254 checklist rule 5)
#   - relative link resolution
#
# Layer 2: execution coverage diff
#   - parses FORM-FIELD-INVENTORY.md tables (column layout detected per header row)
#   - diffs required (formId, fieldKey, F#) against recorded `formId.fieldKey.F#`
#     results under reports/uat-*/ (FIELD-LEVEL-PROTOCOL §3 recommended key).
#     0 formatted results => UNVERIFIED, never claimed as covered.
#
# A PASS here proves only consistency of enumerated items. It does NOT prove
# that all forms/fields are enumerated (inventory carries wildcard/要実測
# remainder), that runs happened, or that the product is defect-free.
set -euo pipefail

readonly EXIT_FAIL=1
ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
SCEN="$ROOT/docs/ops/testing/scenarios"
INVENTORY="$SCEN/FORM-FIELD-INVENTORY.md"
README="$SCEN/README.md"
REPORTS_ROOT="$ROOT/reports"

ok=0
fail=0
warn=0
pass() { echo "PASS  $1"; ok=$((ok + 1)); }
bad()  { echo "FAIL  $1"; fail=$((fail + 1)); }
wrn()  { echo "WARN  $1"; warn=$((warn + 1)); }
info() { echo "INFO  $1"; }

workdir="$(mktemp -d "${TMPDIR:-/tmp}/scenarios-coverage.XXXXXX")"
cleanup() { # shellcheck disable=SC2317  # invoked via trap
  [ -z "$workdir" ] || rm -rf -- "$workdir"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

REPORTS_DIR=""
while [ $# -gt 0 ]; do
  case "$1" in
    --reports)
      REPORTS_DIR="${2:?--reports requires a directory}"
      shift 2
      ;;
    *)
      echo "usage: $0 [--reports DIR]" >&2
      exit 2
      ;;
  esac
done

echo "=== scenarios/ coverage gate ==="
echo "root: $ROOT"
echo "time: $(date -Iseconds)"

# ---------------------------------------------------------------------------
# Layer 1: document integrity
# ---------------------------------------------------------------------------
echo "=== Layer 1: document integrity ==="

if [ ! -f "$README" ] || [ ! -f "$INVENTORY" ]; then
  bad "README.md / FORM-FIELD-INVENTORY.md missing under $SCEN"
  echo "=== summary: PASS=$ok FAIL=$fail WARN=$warn ==="
  exit "$EXIT_FAIL"
fi

# id \t file \t depth
awk -F'|' '/^\| \[(S|V)[0-9]+\]\(/ {
  cell = $2
  gsub(/[[:space:]]/, "", cell)
  id = cell; sub(/\].*/, "", id); sub(/^\[/, "", id)
  link = cell; sub(/^[^()]*\(/, "", link); sub(/\).*/, "", link)
  depth = $(NF - 1)
  gsub(/[[:space:]]/, "", depth)
  print id "\t" link "\t" depth
}' "$README" > "$workdir/index.tsv"

index_count=$(wc -l < "$workdir/index.tsv" | tr -d ' ')
if [ "$index_count" -eq 0 ]; then
  bad "no S/V index entries parsed from README.md — table format changed?"
fi

# A. index <-> files
missing_files=0
orphan_files=0
while IFS=$'\t' read -r id file depth; do
  if [ ! -f "$SCEN/$file" ]; then
    bad "index entry $id points to missing file: $file"
    missing_files=$((missing_files + 1))
  fi
done < "$workdir/index.tsv"

for f in "$SCEN"/S[0-9]*.md "$SCEN"/V[0-9]*.md; do
  [ -e "$f" ] || continue
  base=$(basename "$f")
  if ! grep -qF "$base" "$workdir/index.tsv"; then
    bad "scenario file not listed in README index: $base"
    orphan_files=$((orphan_files + 1))
  fi
done
if [ "$missing_files" -eq 0 ] && [ "$orphan_files" -eq 0 ]; then
  pass "index <-> files consistent ($index_count entries)"
fi

# B/C. template compliance + H1 id match
template_bad=0
checked_files=0
while IFS=$'\t' read -r id file depth; do
  f="$SCEN/$file"
  [ -f "$f" ] || continue
  checked_files=$((checked_files + 1))
  h1=$(head -n 1 "$f")
  if [[ "$h1" != "# $id:"* ]]; then
    bad "$id H1 does not match filename (got: $h1)"
    template_bad=$((template_bad + 1))
  fi
  if [[ "$id" == S* ]]; then
    for sec in "## 前提条件" "## 手順と期待結果" "## 確認観点"; do
      if ! grep -qF "$sec" "$f"; then
        bad "$id missing required section: $sec"
        template_bad=$((template_bad + 1))
      fi
    done
    if [[ "$depth" == *深い* ]] && ! grep -q '^## 異常系' "$f"; then
      bad "$id depth=深い but has no 異常系 section"
      template_bad=$((template_bad + 1))
    fi
  else
    if ! grep -qF "共通チェック手順" "$f"; then
      bad "$id missing 共通チェック手順 (C1/C2/C3 entry point)"
      template_bad=$((template_bad + 1))
    fi
    for c in C1 C2 C3; do
      if ! grep -qF "$c" "$f"; then
        bad "$id missing $c check definition"
        template_bad=$((template_bad + 1))
      fi
    done
  fi
done < "$workdir/index.tsv"
if [ "$template_bad" -eq 0 ]; then
  pass "template compliance: $checked_files S/V files"
fi

# D. no execution results inside scenario files (checklist rule 5)
result_bad=0
for f in "$SCEN"/S[0-9]*.md "$SCEN"/V[0-9]*.md; do
  [ -e "$f" ] || continue
  hits=$(grep -nE '^[[:space:]]*(PASS|FAIL|PARTIAL|BLOCKED)[[:space:][:punct:]]|\*\*(PASS|FAIL|PARTIAL|BLOCKED)\*\*|([0-9]{4}-[0-9]{2}-[0-9]{2}).*(PASS|FAIL)|(PASS|FAIL).*([0-9]{4}-[0-9]{2}-[0-9]{2})' "$f" 2>/dev/null || true)
  if [ -n "$hits" ]; then
    # Expected-result instructions ("〜場合は FAIL") are not recorded results.
    hits=$(printf '%s\n' "$hits" | grep -vE '(場合は ?(PASS|FAIL)|たら ?(PASS|FAIL)|なら ?(PASS|FAIL)|ときは ?(PASS|FAIL))' || true)
  fi
  if [ -n "$hits" ]; then
    bad "$(basename "$f") looks like it records execution results (rule: results only in reports/uat-*):"
    printf '%s\n' "$hits" | sed 's/^/      /'
    result_bad=$((result_bad + 1))
  fi
done
if [ "$result_bad" -eq 0 ]; then
  pass "no execution results embedded in S/V files"
fi

# E. relative links resolve
link_bad=0
for f in "$SCEN"/*.md; do
  dir=$(dirname "$f")
  while IFS= read -r p; do
    [ -n "$p" ] || continue
    case "$p" in
      http* | mailto:* | data:*) continue ;;
    esac
    if [ ! -e "$dir/$p" ]; then
      bad "$(basename "$f"): broken relative link -> $p"
      link_bad=$((link_bad + 1))
    fi
  done < <(grep -oE '\]\([^)]+\)' "$f" 2>/dev/null | sed -E 's/^\]\(//; s/\)$//; s/#.*$//' || true)
done
if [ "$link_bad" -eq 0 ]; then
  pass "all relative links in scenarios/*.md resolve"
fi

# ---------------------------------------------------------------------------
# Layer 2: execution coverage (inventory x reports)
# ---------------------------------------------------------------------------
echo "=== Layer 2: execution coverage ==="

# Inventory parser. Column layout is detected from each header row, because the
# inventory mixes at least 4 table schemas:
#   | UI/system state | wire key | R/O | 型 | 制約・特記 | F 重点 |
#   | fieldKey | ラベル概要 | R/O | 型 | 制約・特記 | F 重点 |
#   | fieldKey | R/O | 型 | F 重点 |
#   | fieldKey | R/O | F 重点 |
# Emits TSV: SUMMARY\tform\texact\tsystem\tnonexact / KEY\tform\tkey /
#            REQ\tform\tkey\tF# / NA\tform\tkey
awk '
/^###+ / {  # ### section or #### sub-form (standard masters use #### per formId)
  if (formid != "") printf "SUMMARY\t%s\t%d\t%d\t%d\n", formid, exact, syst, nonx
  sec = $0; sub(/^#+ /, "", sec)
  m = split(sec, a, /[ \t]/)
  formid = a[1]
  exact = 0; syst = 0; nonx = 0
  keyidx = 0; roidx = 0; fidx = 0; fidcol = 0; intable = 0
  next
}
/^$/ { intable = 0; next }
! /^\|/ { intable = 0; next }
{
  line = $0
  gsub(/`/, "", line)
  probe = line
  gsub(/[:| \-]/, "", probe)
  if (probe == "") next

  n = split(line, c, "|")
  ki = 0; ri = 0; fi = 0; fc = 0; isheader = 0
  for (i = 1; i <= n; i++) {
    cell = c[i]; gsub(/[[:space:]]/, "", cell)
    if (cell == "formId") { fc = i; isheader = 1 }
    else if (cell ~ /fieldKey$/ || cell == "wirekey") { ki = i; isheader = 1 }
    else if (cell == "R/O") { ri = i }
    else if (cell ~ /^F重点/) { fi = i; isheader = 1 }
  }
  if (isheader && ki > 0 && (ri > 0 || fc > 0)) {
    keyidx = ki; roidx = ri; fidx = fi; fidcol = fc; intable = 1
    next
  }
  if (!intable || keyidx == 0 || formid == "") next

  # per-row formId override (| formId | 必須 fieldKey | その他 | F 重点 | schema)
  if (fidcol && fidcol <= n) {
    rf = c[fidcol]; gsub(/[[:space:]]/, "", rf)
    if (rf != "") formid = rf
  }

  key = (keyidx <= n) ? c[keyidx] : ""
  gsub(/[[:space:]]/, "", key)
  if (key == "" || key == "fieldKey" || key == "wirekey") next
  ro = (roidx && roidx <= n) ? c[roidx] : ""
  gsub(/[[:space:]]/, "", ro)
  fs = (fidx && fidx <= n) ? c[fidx] : ""

  if (ro == "S" || ro == "C/S" || ro == "S/C") {
    syst++; printf "NA\t%s\t%s\n", formid, key
    next
  }
  # dynamic/shape rows (bracketed or runtime-enumerated) stay non-enumerated
  if (key ~ /[*（(]/ || key ~ /列挙/) { nonx++; next }
  # comma/slash-joined cells (e.g. | pet_id / owner_id / visit_date |) are key lists
  cnt = split(key, kp, /[、,\/]/)
  for (j = 1; j <= cnt; j++) {
    k = kp[j]; gsub(/[[:space:]]/, "", k)
    if (k == "") continue
    exact++
    printf "KEY\t%s\t%s\n", formid, k
    req = 0
    m = split(fs, ft, /[[:space:]・、()（）\/]+/)
    for (i = 1; i <= m; i++) {
      t = ft[i]
      if (t ~ /^F[0-6]/) {
        sub(/^F([0-6]).*/, "F\\1", t)
        req++; printf "REQ\t%s\t%s\t%s\n", formid, k, t
      }
    }
    if (req == 0) printf "REQ\t%s\t%s\tF0\n", formid, k  # F0 applies to every field (protocol §2)
  }
}
END {
  if (formid != "") printf "SUMMARY\t%s\t%d\t%d\t%d\n", formid, exact, syst, nonx
}
' "$INVENTORY" > "$workdir/inv.tsv"

total_keys=$(grep -c '^KEY' "$workdir/inv.tsv" || true)
total_req=$(grep -c '^REQ' "$workdir/inv.tsv" || true)
total_na=$(grep -c '^NA' "$workdir/inv.tsv" || true)
total_forms=$(grep -c '^SUMMARY' "$workdir/inv.tsv" || true)
nonexact_rows=$(awk -F'\t' '$1=="SUMMARY"{s+=$5} END{print s+0}' "$workdir/inv.tsv")
if [ "$total_keys" -eq 0 ]; then
  bad "inventory parse produced 0 field keys — table format changed?"
else
  pass "inventory parsed: $total_forms form sections, $total_keys exact fieldKeys, $total_req required F checks, $total_na system(N/A) fields"
fi

# Non-enumerated remainder must stay visible (never claimed as covered).
yajissoku=$(grep -c '要実測' "$INVENTORY" 2>/dev/null || true)
wildcard=$(grep -ci 'wildcard' "$INVENTORY" 2>/dev/null || true)
info "non-enumerated remainder: 要実測 mentions=$yajissoku, wildcard mentions=$wildcard, non-exact rows=$nonexact_rows — these are NOT counted as covered"

# Resolve reports dir: --reports wins, else latest uat-* by name.
if [ -z "$REPORTS_DIR" ]; then
  latest=$(find "$REPORTS_ROOT" -maxdepth 1 -type d -name 'uat-*' 2>/dev/null | sort | tail -n 1 || true)
  if [ -n "$latest" ]; then REPORTS_DIR="$latest"; fi
fi

recorded=0
if [ -n "$REPORTS_DIR" ] && [ -d "$REPORTS_DIR" ]; then
  info "scanning results in: ${REPORTS_DIR#"$ROOT"/}"
  grep -rhoE '[A-Za-z0-9][A-Za-z0-9-]*\.[A-Za-z0-9_]+\.F[0-6]' "$REPORTS_DIR" 2>/dev/null | sort -u > "$workdir/results.txt" || true
  recorded=$(wc -l < "$workdir/results.txt" | tr -d ' ')
else
  wrn "no reports/uat-* directory found — execution coverage UNVERIFIED"
fi

if [ "$recorded" -eq 0 ]; then
  wrn "0 results recorded as formId.fieldKey.Fx — V-series execution coverage UNVERIFIED (FIELD-LEVEL-PROTOCOL §3 key format)"
else
  # required checks not yet recorded
  awk -F'\t' '$1=="REQ"{print $2"."$3"."$4}' "$workdir/inv.tsv" | sort -u > "$workdir/required.txt"
  comm -23 "$workdir/required.txt" "$workdir/results.txt" > "$workdir/missing.txt"
  missing=$(wc -l < "$workdir/missing.txt" | tr -d ' ')
  req_total=$(wc -l < "$workdir/required.txt" | tr -d ' ')
  if [ "$missing" -eq 0 ]; then
    pass "all $req_total required (form.field.F) checks have recorded results"
  else
    wrn "missing recorded results: $missing / $req_total required checks (showing up to 40):"
    head -n 40 "$workdir/missing.txt" | sed 's/^/      /'
    if [ "$missing" -gt 40 ]; then echo "      ... and $((missing - 40)) more"; fi
  fi
  # recorded tokens that match nothing in the inventory
  keys_file="$workdir/keys.txt"
  awk -F'\t' '$1=="KEY"{print $2 "." $3}' "$workdir/inv.tsv" | sort -u > "$keys_file"
  unmatched=$(grep -cvFf "$keys_file" "$workdir/results.txt" 2>/dev/null || true)
  if [ "$unmatched" -gt 0 ]; then
    wrn "$unmatched recorded tokens do not match any inventory form.field — check formId spelling:"
    grep -vFf "$keys_file" "$workdir/results.txt" 2>/dev/null | head -n 20 | sed 's/^/      /'
  fi
fi

# Scope note: S-series run status lives in free-form run reports + UAT-DOMAIN-STATUS.md
# and is intentionally NOT parsed here (results are never written to scenario files).
info "scope: S-series execution status is in reports/uat-*/ run reports and UAT-DOMAIN-STATUS.md — not mechanically diffed here"

echo "=== summary: PASS=$ok FAIL=$fail WARN=$warn ==="
if [ "${DEBUG:-0}" = "1" ]; then
  info "DEBUG inv.tsv dump (SUMMARY/KEY/REQ/NA rows):"
  cat "$workdir/inv.tsv"
fi
if [ "$fail" -gt 0 ]; then
  echo "scenarios coverage gate: FAIL (exit $EXIT_FAIL)"
  exit "$EXIT_FAIL"
fi
echo "scenarios coverage gate: document integrity OK; execution coverage: $([ "$recorded" -gt 0 ] && echo "diffed ($recorded recorded results)" || echo "UNVERIFIED (no formatted results)")"
echo "this gate proves consistency of enumerated items only — not run completion, not full-form coverage, not product quality"
exit 0
