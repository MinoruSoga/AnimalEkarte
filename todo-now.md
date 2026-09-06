# Astra品質監査 — 完了履歴と入口

Astra監査 F1〜F6 の実装は `origin/main` へ統合済み。未完了の後続は [`todo.md`](todo.md) を入口にする。実行状態の正本は Linear（[作業台帳ルール](docs/work/README.md)）。

- 監査対象: `main` / `c41ba8b1c1aef7f8150457dc310fe7f61fca1a75`（2026-09-05）
- 最終観測: 2026-09-06 / `b987729fd57b05b5f94f0a7e0d5860d515401421`
- この更新の claim: `claim/TODO-NOW` は不在だったため、`todo.md` の `LEDGER-TODO-NOW-POINTER` として入口ポインタ化した

## 履歴（詳細は git）

- F1〜F6 の実装と F3 manual E2E auth smoke（run `33972458396`）
- F4〜F6 の push 後通常 CI
- F3 k6 aggregate の原因は `metric.values.*` のみ参照。修正本体は `todo.md` の `CI-K6-SUMMARY-SCHEMA`
- 「診断カテゴリ」「診断病名」label 接続は `FE-CLINICAL-PLAN-SELECT-LABELS`

Linear の F1〜F6 対応案は [docs/work/linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md)。Linear 本体は UNKNOWN。Done は USER。
