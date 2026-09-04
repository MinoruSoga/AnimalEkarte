# CI ポリシー — ゲート分担と Actions ピン記法

> `.github/workflows/*.yml` と `scripts/run-local-ci.sh` が実装の正本。この文書は契約を要約し、手作業の action/step inventory は持たない。

## リモートとローカルの分担

| 区分 | 契約 | 実装の正本 |
|---|---|---|
| Remote | path-filtered backend/frontend build・test・coverage、secret scan、worker、codegen、main PR の migration verify | `.github/workflows/ci.yml` |
| AgentShield | main 向け PR で agent config が変わった場合、または manual dispatch で `force_fail_on_findings` を有効にした場合だけ findings を fail 扱い。他の branch/trigger は report-only | `.github/workflows/security-scan.yml` |
| Local process policy | push/PR 前に `make ci` を実行する。GitHub はこのローカル実行を強制も証明もしない | `scripts/run-local-ci.sh` |
| E2E | `workflow_dispatch` のみ。自動 push/PR gate ではない | `.github/workflows/e2e.yml` |
| Performance | schedule と manual dispatch。push trigger はない | `.github/workflows/performance-tests.yml` |

`make ci` の正確な gate 一覧と順序は `scripts/run-local-ci.sh` の `begin_step` 呼び出しを参照する。現在は inventory/guardrail、A4 rehearsal isolation、design-system audit、lint/type/codegen、backend/frontend build/test などを含む。件数や列挙をこの文書へ複製しない。

## Remote CI の要点

- Backend test は独立 PostgreSQL を持つ 4 shard。coverage profile は `scripts/merge_go_coverprofiles.py` で統合する。
- Frontend test は 2 Vitest shard。blob report を native merge して coverage ratchet を実行する。
- `main`、`staging`、`production` 向け PR は、path filter に該当する層の build/test を行う。
- frontend install は frozen lockfile を使う。
- Backend は console 出力を制限し、gzip の full log artifact を 7 日保持する。
- Frontend は `vitest-full.log` を削除する。Vitest blob を artifact にし、失敗時は bounded `vitest-tail.log` のみ追加 upload する。Backend と同じ full-log 契約ではない。
- job timeout で暴走を止める。

## Actions ピン方針

| 対象 | 記法 |
|---|---|
| GitHub 公式 `actions/*` | major tag **または exact semver** |
| ベンダー公式 action | major tag または exact semver |
| その他の third-party action | commit SHA pin と version comment |
| remote script/artifact | pipe-to-shell 禁止。version と SHA-256 を固定 |

実際の action version は workflow の `uses:` が正本。`scripts/check-actions-version-drift.sh` が同一 action の混在を検出する。古い version inventory は保持しない。

## 静的チェックの enforcement 状態

- `scripts/check-workflow-remote-exec-policy.sh`
- `scripts/check-agent-security-policy.sh`

上記は HEAD では `make ci` や workflow から呼ばれない **manual checks** である。実行せずに「reject/fail される」とは表現しない。enforced gate にする場合はスクリプトと regression test を `scripts/run-local-ci.sh` または workflow へ接続する。

## Historical decision record

2026-07 に remote CI を軽量化し、再現可能な inventory/guardrail を local process policy へ移した。これは当時の設計判断であり、現在の step/action inventory ではない。現在値は必ず workflow と `begin_step` から確認する。
