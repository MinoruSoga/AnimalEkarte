#!/usr/bin/env bash
# scripts/check-actions-version-drift.sh
#
# .github/workflows/*.yml 全体で、同一 action が複数バージョンで参照されて
# いないかを検査する（docs/ops/ci-policy.md 運用ルール2「同一 action は全ワーク
# フローで単一バージョンに統一」の機械的強制・#195 再発防止）。
#
# 背景:
#   #195 で setup-node 等のバージョン統一を達成したが、その後の新規ワーク
#   フロー追加（backend-deploy.yml, Cloudflare Phase 5）で @v4 が混入し回帰
#   した。actionlint は構文・式のみを検査しバージョンドリフトを検出しない
#   ため、検出手段が実在しない限り新規ワークフロー追加のたびに再発する。
#
# ルール:
#   `uses: <owner>/<repo>@<ref>` 形式の参照について、同一 <owner>/<repo>
#   （サブパス含む）が 2 種類以上の <ref> で参照されていたら FAIL。
#   ref がメジャータグか完全 semver か SHA かは問わない — 「単一に収束して
#   いるか」だけを見る（記法ポリシー自体は docs/ops/ci-policy.md とレビューの領分）。
#   ローカル参照（uses: ./...）と docker:// は対象外。
#
# このチェックは Docker を必要とせず、YAML のテキストだけを検査する。
#
# Usage: bash scripts/check-actions-version-drift.sh [workflows-dir]
# Exit codes:
#   0  全 action が単一バージョンに収束している
#   1  ドリフト検出（同一 action に複数バージョン）／解析失敗
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKFLOWS_DIR="${1:-$SCRIPT_DIR/../.github/workflows}"

if [[ ! -d "$WORKFLOWS_DIR" ]]; then
  echo "FAIL  workflows dir not found: $WORKFLOWS_DIR"
  exit 1
fi

# 対象: *.yml / *.yaml。nullglob で存在するファイルだけを集める
# （*.yaml が 1 件も無い場合に glob 文字列が grep へ渡り exit 2 になるのを防ぐ）。
shopt -s nullglob
workflow_files=("$WORKFLOWS_DIR"/*.yml "$WORKFLOWS_DIR"/*.yaml)
shopt -u nullglob

if [[ ${#workflow_files[@]} -eq 0 ]]; then
  echo "FAIL  no workflow files found under $WORKFLOWS_DIR"
  exit 1
fi

# grep -h でファイル名を落とし、uses 行から action@ref を抽出する。
# uses 行が 1 行も無い場合は grep(exit 1)と awk の seen_any ガード(exit 1)が
# 一致して FAIL になる（pipefail 下でも exit 1 のまま）。
grep -hE '^[[:space:]]*(-[[:space:]]+)?uses:[[:space:]]' "${workflow_files[@]}" |
  sed -E 's/^[[:space:]]*(-[[:space:]]+)?uses:[[:space:]]*//; s/[[:space:]]*(#.*)?$//' |
  grep -Ev '^(\./|docker://)' |
  awk '
    /@/ {
      seen_any = 1
      split($0, parts, "@")
      action = parts[1]
      ref = parts[2]
      key = action SUBSEP ref
      if (!(key in counted)) {
        counted[key] = 1
        refs[action] = (action in refs) ? refs[action] ", @" ref : "@" ref
        nrefs[action]++
      }
    }
    END {
      if (!seen_any) {
        print "FAIL  no uses: lines parsed (parser drift? check workflow files)"
        exit 1
      }
      errors = 0
      for (a in nrefs) {
        if (nrefs[a] > 1) {
          printf "FAIL  action \"%s\" is referenced with %d different versions: %s\n", a, nrefs[a], refs[a]
          errors++
        }
      }
      if (errors > 0) {
        print "----"
        printf "RESULT  %d action(s) drift across workflows — unify to a single version (docs/ops/ci-policy.md rule 2)\n", errors
        exit 1
      }
      print "PASS  every action resolves to a single version across all workflows"
    }
  '
