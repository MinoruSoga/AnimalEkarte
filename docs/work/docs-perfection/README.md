# docs-perfection — Astra 調査パッケージ（文書保守）

> **目的**: Codex (GPT-6 Astra) が実施した `docs/` 棚卸し・矛盾候補・修復キューを、チャット外のリポジトリ正本として置く。  
> **読者**: docs 保守を続けるエージェント / 人間。  
> **製品の実行 SoT ではない**: 状態・担当・Done は Linear（[BRT-4](https://linear.app/baritechllc/issue/BRT-4)）。本ディレクトリは **文書ドリフト調査と修復計画** 専用。

## 読む順序

1. 本 README（役割と境界）
2. [ROLE-MAP.md](ROLE-MAP.md) — Constitution / Map / Status / History の割当
3. [REPAIR-QUEUE.md](REPAIR-QUEUE.md) — 優先修復 RQ（調査時点の両側根拠付き）
4. [INVENTORY.md](INVENTORY.md) — 全 Markdown の一次分類（全命題の実装一致認定ではない）
5. [CHILD-GOALS/](CHILD-GOALS/) — 先行カテゴリ修復の実行契約（今回の全件照合は下記レポート）
6. [LEDGER.md](LEDGER.md) / [PROGRESS.md](PROGRESS.md) / [GOAL.yaml](GOAL.yaml) — 進行・親 goal
7. [EVIDENCE/](EVIDENCE/) — drift・検証・独立レビューの生証拠

## 調査の出所（Astra）

| 成果 | セッション | 備考 |
|------|------------|------|
| コーディネータ調査一式（ROLE-MAP / INVENTORY / REPAIR-QUEUE / CHILD-GOALS / EVIDENCE） | Codex Astra — prompt `agent-docs-perfection-coordinator-codex-astra-v2.md` | 本文カテゴリの大規模修復は **しない** 契約で完了 |
| product-philosophy（RQ-010）1行修復 | Codex Astra — child prompt `agent-docs-perfection-child-product-philosophy-codex-astra.md` | Mode 3 照合 PASS |

調査の機械ゲート記録: [EVIDENCE/docs-symbol-drift.txt](EVIDENCE/docs-symbol-drift.txt)（coordinator 時点で exit 0）。旧 `EVIDENCE/verify.py` と `checks-*.txt` は coordinator-only の範囲・全子 pending を検査した履歴であり、後続修復の現在ゲートではない。

2026-09-07、`DOCS-PERFECT-ASTRA-LOOP` の Phase A で入口から調査成果への到達を確認した。本文下書きと先行完了記録の採否・現在の検証は [LEDGER](LEDGER.md#docs-perfect-astra-loop) を参照する。調査の配置確認を本文修復の完了根拠にしない。

## 境界（重要）

- **調査ドキュメント** = 本ディレクトリ。先にここに置き、索引から辿れる状態にする。
- **先行本文修復** = 各 `CHILD-GOALS/*.md` の allowlist。2026-09-07 の追加依頼「docs/ 配下のドキュメントをすべて最新化」は [Issue照合レポート](ISSUE-RECONCILIATION.md) の独立した docs-only scope で扱う。
- **生成側セッションが調査配置の前に本文を直した場合**は下書き扱いとし、Astra 子 unit の再監査前に「調査に基づく正式修復」とみなさない。
- RQ-002（frontend manual）の修復差分は今回の開始時点で既に存在し、[REPAIR-QUEUE](REPAIR-QUEUE.md) は resolved。今回は会計作成の説明を実装と再照合した。コード修復・runtime 成功を今回の成果に含めない。

## ファイル一覧

| パス | Living Docs 役割 | 内容 |
|------|------------------|------|
| [ROLE-MAP.md](ROLE-MAP.md) | Map | 4役割と canonical owner |
| [INVENTORY.md](INVENTORY.md) | Map / Status 入口 | 199 Markdown の一次トリアージ（coordinator 時点198件＋本README） |
| [REPAIR-QUEUE.md](REPAIR-QUEUE.md) | Status | RQ-001〜011 と根拠 |
| [CHILD-GOALS/](CHILD-GOALS/) | Status（実行契約） | カテゴリ別 done_condition / verification |
| [GOAL.yaml](GOAL.yaml) | Status | 親 goal（durable） |
| [LEDGER.md](LEDGER.md) | Status | unit 台帳 |
| [PROGRESS.md](PROGRESS.md) | History | checkpoint |
| [EVIDENCE/](EVIDENCE/) | History / 証拠 | 生ログ・検証スクリプト |

## 現在の状態と残件

最新の全件照合は [Issue照合レポート（2026-09-07）](ISSUE-RECONCILIATION.md)。既存199 MarkdownとGOAL.yamlを確認し、35件の証跡・補助ファイルは履歴として保存する。GitHub Issue本文・コメントとコードを照合し、外部実行状態は別に扱う。

先行 `DOCS-PERFECT-ASTRA-LOOP` の独立レビューと検証は [LEDGER](LEDGER.md#docs-perfect-astra-loop) と [当時の最終検証ログ](EVIDENCE/astra-loop-checks-final.txt) に保存する。現在のGOAL.yamlの `status: completed` は今回の開始時点からある先行状態で、今回の人間受入を表さない。今回の受入状態は `issue_refresh.acceptance_state: awaiting-human-verification`。

RQ-002は修復差分あり。D-254のFAQ・スクリーンショット最終同期、Linear現在値、外部証跡・UAT・STG/PROD稼働は未確認のまま残す。claimは統合または明示的放棄後にUSERが解放する。
