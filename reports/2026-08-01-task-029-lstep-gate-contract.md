# TASK-029: Lステップ deploy/clinic gate 契約の docs 同期

- **Date**: 2026-08-01
- **Scope**: docs-only（3 spec + audit report + todo ledger）。backend / worker / wrangler は read-only 証拠源。
- **Claim**: `claim/TASK-029`（USER が統合後に解放。本 unit は claim を削除しない）
- **Runtime**: 本 unit は runtime / STG / production / cron 自然発火 / 実送信を一切実行していない。runtime green を主張しない。

## 契約の根拠（current source）

| Gate | Source | 契約 |
|:---|:---|:---|
| Deploy gate | `backend/internal/infra/lstep/client.go:22-25,72-85` | `LSTEP_WRITE_API_ENABLED` は exact `"true"` のみ有効。無効時は HTTP 未送信かつ `ErrWriteDisabled`（`nil` 成功にしない） |
| Clinic gate | `backend/internal/lstep/lstep_delivery_trigger_client.go:35-56` | `is_sync_enabled=false` または API キー未設定は `buildClient` が `nil, nil`（意図的スキップ） |
| Scheduler wiring | `backend/wrangler.jsonc:97-102`, `backend/worker/scheduled-jobs.ts:30-34` | cron 式と job 対応は配線済み。STG/production 自然発火・実送信は未実測 |
| Enable/stop/rollback 正本 | `docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md` | 手順の唯一の正本。spec は link のみ（本 unit は runbook を変更していない） |

## RED 相当（修正前・逐語）

```
docs/spec/screens/31-lstep-integration.md:64: ...一時停止中（noop）...
docs/spec/screens/34-lstep-delivery-monitor.md:8: ...一時停止中（noop）でも...
docs/spec/screens/34-lstep-delivery-monitor.md:61: ...外部タグ write が noop でも...
docs/spec/line/cost-analysis.md:12: ...現在一時停止中（noop）のため、現状の実配信量・実 API 呼び出し量はほぼ 0 件...
```

追加検出: 3 doc には `is_sync_enabled` / `ErrWriteDisabled` が欠落（clinic/deploy 分離未記述）。`cost-analysis` の「ほぼ 0 件」は未実測の数量主張だった。

## Drift 対応表

| # | 是正前 file:line | 問題 | 是正後 file:line | 根拠 source file:line |
|---:|:---|:---|:---|:---|
| 1 | `docs/spec/screens/31-lstep-integration.md:64` | deploy gate OFF を noop と記述 | `docs/spec/screens/31-lstep-integration.md:64`（Deploy gate / `ErrWriteDisabled` + HTTP 未送信） | `backend/internal/infra/lstep/client.go:77-83` |
| 2 | `docs/spec/screens/34-lstep-delivery-monitor.md:8` | 同上 | `docs/spec/screens/34-lstep-delivery-monitor.md:8` | `backend/internal/infra/lstep/client.go:77-83` |
| 3 | `docs/spec/screens/34-lstep-delivery-monitor.md:61` | 同上 | `docs/spec/screens/34-lstep-delivery-monitor.md:62` | `backend/internal/infra/lstep/client.go:77-83` |
| 4 | `docs/spec/line/cost-analysis.md:12` | 同上 + 未実測の「ほぼ 0 件」 | `docs/spec/line/cost-analysis.md:12-16` | `backend/internal/infra/lstep/client.go:22-25,72-85` |
| 5 | （欠落）3 doc に clinic gate 未分離 | clinic を deploy と混同し得る | `31:65`, `34:9`/`34:62`, `cost-analysis:15` | `backend/internal/lstep/lstep_delivery_trigger_client.go:35-56` |
| 6 | `31:65` は配線と未実測を同一箇条書き内で既に分離していたが、主語が「Write API」混同リスク | バッチ同期箇条書きを cron 明示 + 別事実強調 | `docs/spec/screens/31-lstep-integration.md:66` | `backend/wrangler.jsonc:97-102`, `backend/worker/scheduled-jobs.ts:30-34` |

## GREEN 相当（修正後ゲート）

- `rg -n 'noop|no-op'` on 3 doc → deploy gate を主語とする noop 行 **0**（clinic を noop と呼ぶ行も 0；clinic は「意図的スキップ / `nil, nil`」表記）
- `rg -n 'ErrWriteDisabled'` → 3 ファイルすべてヒット、HTTP 未送信 / `nil` 成功にしない を併記
- `rg -n 'is_sync_enabled'` → 3 ファイルすべてで deploy と別箇条書き
- `rg -n 'LSTEP_WRITE_API_PAUSE'` → 3 ファイルすべて link あり。enable/stop/rollback 手順・環境実値の新規複製なし
- `bash scripts/check-docs-symbol-drift.sh` → exit 0（Completion Report に逐語出力を転記）

## 変更ファイル

- `docs/spec/screens/31-lstep-integration.md`
- `docs/spec/screens/34-lstep-delivery-monitor.md`
- `docs/spec/line/cost-analysis.md`
- `reports/2026-08-01-task-029-lstep-gate-contract.md`（本ファイル）
- `todo.md`（TASK-029 結果追記のみ。foreign WIP hunk は stage しない）

## Remaining risks / follow-ups

- `docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md:44` に歴史的「paused no-op（`return nil`）…は廃止済み」記述がある。本 unit の A4 により runbook は変更せず。誤解が残る場合は別 TASK 候補。
- STG/production での cron 自然発火・実送信・環境変数実値は USER/OPS gate。
- GitHub Issue #259 の close / コメント / ラベルは USER 専権（本 unit は GitHub 書き込みなし）。
- `claim/TASK-029` の解放は USER 専権。

## Orchestration

- Mode: native Workflow tool (`task-029-investigate`) + parent synthesizer/writer
- Investigation: parallel read-only probes（deploy gate / clinic gate / cron / drift docs / todo+rules）
- Implementation: single parent writer（共有 tree の foreign WIP 保護のため 3 doc は serial write）
