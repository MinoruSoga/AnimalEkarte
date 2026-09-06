# Astra品質監査 — 未完了の後続

## このファイルの位置付け

Astra監査6件（F1〜F6）の実装は `origin/main` へ統合済み。本ファイルは未完了の後続だけを置く。実行状態の正本は Linear（[作業台帳ルール](docs/work/README.md)）。Linear 照会・Issue登録はこの更新でも未実施。

- 監査対象: `main` / `c41ba8b1c1aef7f8150457dc310fe7f61fca1a75`（2026-09-05）
- この更新の観測: 2026-09-06 / `main` / `9ed814fc0cd624fc19e30596708eb4646d24289c`（`origin/main` と一致）
- この更新の claim: `claim/TODO-NOW`（台帳更新。エージェントは削除しない）
- 削除した完了分: F1〜F6 の実装、F3 manual E2E auth smoke（run `33972458396`）、F4〜F6 の push 後通常 CI 記録。詳細は git 履歴。

## 未完了

| 優先 | 項目 | この記録上の状態 | Linear |
| ---- | ---- | ---------------- | ------ |
| 1    | Linear で F1〜F6 の重複確認と必要なら ID 紐付け | 未照会 | 未照会 |
| 2    | F3 performance / k6 の同一 run 完走 | 最新 run は workflow `failure`。k6 本体 step は success、aggregate validation のみ failure | 未照会 |
| 3    | 「診断カテゴリ」「診断病名」label と `SearchableSelect` trigger の接続 | 現 HEAD で未接続を再確認。実装未着手 | 未照会 |

### Linear

- [ ] Linear で F1〜F6 の重複 Issue・進行中作業を確認し、必要な ID だけ対応付ける。

この更新では Linear を照会していない。

### F3 performance / k6

最新の GitHub Actions `Performance Tests (Load & Profiling)`:

- run `34020760108`
- head SHA `9ed814fc0cd624fc19e30596708eb4646d24289c`
- event `workflow_dispatch`
- workflow conclusion `failure`
- job `Frontend Lighthouse Audit`: success
- job `Load Testing (k6)`: failure
  - `Run k6 API endpoints load test`: success
  - `Run k6 spike test`: success
  - `Upload load test results`: success
  - `Validate k6 summary aggregates`: failure
  - `Stop app stack`: success
- job `Performance Test Summary`: success

- [ ] 同一の Performance Tests run で、endpoint k6、spike k6、`Validate k6 summary aggregates`、`Stop app stack` がすべて success になった run ID と head SHA を記録する。

この run では共有 STG、PROD、full clinical/data-dependent E2E、release readiness は確認していない。

`todo.md` では同一後続を `CI-K6-SUMMARY-SCHEMA` と `CI-K6-RUNTIME-CLOSEOUT` として記載している。

### 診断セレクトの label 接続

現 HEAD `9ed814fc0` の `frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx` を読んで確認した。

- 「身体検査所見」「診断詳細」「治療方針」は `htmlFor` と textarea `id` が接続されている。
- 「診断カテゴリ」「診断病名」の `label` に `htmlFor` はない。
- 対応する `SearchableSelect` に `id` は渡していない。

`git branch --list 'claim/FE-CLINICAL-PLAN-SELECT-LABELS'` は空。`todo.md` では `FE-CLINICAL-PLAN-SELECT-LABELS` を未起票・未着手としている。

- [ ] 接続の実装と、visible label から trigger へ到達できる検証を行う。
