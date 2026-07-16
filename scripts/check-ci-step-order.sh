#!/usr/bin/env bash
# scripts/check-ci-step-order.sh
#
# .github/workflows/ci.yml の各 job 内で、Build（および再導入時の Lint/Type check）
# のような品質ゲート系ステップが、Test/ratchet のような機能回帰系ステップより後に
# 配置されていないかを検査する（feedback_ci_step_order_masks_lint 再発防止・
# FE-refactor.md「CI step順序ガードレール」）。
#
# 背景:
#   GitHub Actions は既定で「あるステップが失敗すると同一 job の後続ステップは
#   スキップされる」（fail-fast）。ratchet/test 系ステップが Build より前に
#   配置されると、それらが失敗した際に Build が一度も実行されないまま job が
#   red になり、Build 側の別の壊れが完全に見えなくなる
#   （PR #218 で実際に混入し 3fcbcaf4 で手動是正。過去にも backend 側で同型あり
#   = feedback_ci_step_order_masks_lint）。都度レビューに頼らず機械的に再発防止する。
#
# 2026-07 以降: CI の必須静的 lint（golangci / ESLint / type-check / knip）は
# ローカル make に寄せた。CI 上の主ゲートは Build。Lint / Type check 名は
# 再導入時の順序契約用に引き続きゲートとして認識する（fixture 互換）。
#
# ルール:
#   各 job 内で、ゲート系ステップ（名前が完全一致で Lint / Build / Type check の
#   いずれか）の最大インデックスは、リスク系ステップ（名前に大小文字無視で
#   Test または ratchet を含む）の最小インデックスより小さくなければならない。
#   ゲート系・リスク系のいずれか一方しか存在しない job には制約を課さない
#   （比較対象が無いため）。
#
# このチェックは Docker を必要とせず、ci.yml のテキストだけを検査する。
#
# Usage: bash scripts/check-ci-step-order.sh
# Exit codes:
#   0  全 job が順序契約を満たす
#   1  契約違反（ゲート系ステップがリスク系ステップより後にある）／解析失敗
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CI_YML="$SCRIPT_DIR/../.github/workflows/ci.yml"

if [[ ! -f "$CI_YML" ]]; then
  echo "FAIL  ci.yml not found at $CI_YML"
  exit 1
fi

awk '
  function is_gate(n) { return (n == "Lint" || n == "Build" || n == "Type check") }
  function is_risk(n) { return (tolower(n) ~ /test|ratchet/) }
  {
    sub(/\r$/, "")
  }
  /^jobs:[[:space:]]*$/ { injobs = 1; next }
  injobs && /^  [A-Za-z0-9_-]+:[[:space:]]*$/ {
    job = $0
    sub(/^  /, "", job)
    sub(/:[[:space:]]*$/, "", job)
    idx = 0
    next
  }
  injobs && /^      - name:[[:space:]]/ {
    if (job == "") { next }
    name = $0
    sub(/^      - name:[[:space:]]*/, "", name)
    sub(/[[:space:]]+$/, "", name)
    seen_any_job = 1
    if (is_gate(name)) {
      if (!(job in gate_max) || idx > gate_max[job]) { gate_max[job] = idx; gate_name[job] = name }
    }
    if (is_risk(name)) {
      if (!(job in risk_min) || idx < risk_min[job]) { risk_min[job] = idx; risk_name[job] = name }
    }
    idx++
    next
  }
  END {
    if (!seen_any_job) {
      print "FAIL  no job steps parsed (parser drift? check ci.yml indentation assumptions)"
      exit 1
    }
    errors = 0
    for (j in gate_max) {
      if (j in risk_min) {
        if (gate_max[j] >= risk_min[j]) {
          printf "FAIL  job \"%s\": gate step \"%s\" (idx %d) is not before risk step \"%s\" (idx %d)\n", j, gate_name[j], gate_max[j], risk_name[j], risk_min[j]
          errors++
        }
      }
    }
    if (errors > 0) {
      print "----"
      printf "RESULT  %d job(s) violate step order (gate step after risk step — a Test/ratchet failure would fail-fast-mask Lint/Build)\n", errors
      exit 1
    }
    print "PASS  all jobs keep gate steps (Build; Lint/Type check if present) before risk steps (Test/ratchet)"
  }
' "$CI_YML"
