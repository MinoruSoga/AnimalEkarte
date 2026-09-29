# 作業台帳（docs/work）

## 入口

| 文書 | 役割 |
|------|------|
| **Plane** workspace `baritechllc` / Project `EMR`（hub `EMR-1`。旧 Linear `BRT-4` は case_id として継承） | **実行 SoT**（状態・担当・ゲート）。2026-09-29 訂正: Linear は 2026-09-16 に閉鎖済みで履歴参照のみ。旧 `BRT-*` / `LINMIG-*` ID は履歴別名 |
| [todo.md](../../todo.md) | 確認済み製品 FAIL・Slack 出典対応表・維持制約・受入/運用の履歴。未完了タスク状態の正本は Plane（2026-09-23 移行。[receipt](./plane-md-migration-20260923-receipt.md)） |
| [todo.md 検証節](../../todo.md#verification-ledger) | 開発検証・測定・受入・外部照合 |
| [todo.md 運用節](../../todo.md#operations-ledger) | STG・本番・納品などの運用・外部実行 |
| CorpVault `50_Projects/ノア動物病院電子カルテ/` | 会社側索引・時点ログ |

**競合・終了ルール:** 状態・担当・Done は Plane `EMR` を正とする（2026-09-29 訂正: 旧記述「Linear を正とする」→ Linear は 2026-09-16 閉鎖済み）。`todo.md` の実行キューには未完了作業だけを保持し、完了行を除く。統合元の完了履歴・維持制約は別節に分離し、現在の受入や release 判定には使わない。STG Lane 4 の終了条件・記録方法は [運用計画](../../todo.md#ops-lane-4) と [受入計画](../../todo.md#ver-todo-v-stg-data) を参照する。[製品 FAIL 節](../../todo.md#product-bugs) は Plane と対応付け、受入未実施や環境 BLOCKED を製品 FAIL に混ぜない。

```mermaid
flowchart TB
    L["Plane EMR（実行 SoT）<br>状態・担当・Done は Plane を正とする<br>（旧 Linear hub。2026-09-16 閉鎖）"]
    subgraph Repo["repo 側の入口"]
        T["todo.md<br>未完了作業キュー"]
        TV["todo.md 検証節<br>検証・受入・外部照合"]
        TO["todo.md 運用節<br>STG・本番・納品の運用"]
    end
    W["docs/work 補助票<br>decisions/・phase2-deferred・各 campaign 票ほか<br>判断根拠の保存（実行 SoT ではない）"]
    CV["CorpVault<br>会社側索引・時点ログ"]
    L -->|競合時の正| Repo
    W -.->|根拠・索引| Repo
    CV -.->|時点ログ| L
```

| 補助 | 役割 |
|------|------|
| [decisions/](./decisions/README.md) | 採択済み方針の短いポインタ |
| [phase2-deferred.md](./phase2-deferred.md) | 今期外の短い索引 |
| [linear-f1-f6-mapping.md](./linear-f1-f6-mapping.md) | repo の実装履歴と旧 Linear の対応案（2026-09-11 時点の記録。Linear は閉鎖済みで F1–F6 対応は未特定のまま） |
| [skill-reeval-2026-09-06.md](./skill-reeval-2026-09-06.md) | 代表タスクごとの資料選択・停止判断 |
| [development-task-decisions.md](./development-task-decisions.md) | 旧10開発候補の採否と具体的な着手範囲の根拠。実行キューはtodo.md |
| [stg-uat-clinic-feedback-q1-q4.md](./stg-uat-clinic-feedback-q1-q4.md) | STG UAT 医院Q1–Q4の送付用回答と修正タスク。Q1/Q4保険/Q2履歴は STG デプロイ済み（PR #411）。実行 SoT ではない |
| [linmig-campaign-20260919/](./linmig-campaign-20260919/) | P1–P5/P8・検査機器・実LINE・migrate・件数・フォント・締め時間・bundle のローカル準備票。実行完了ではない |
| [remaining-campaign-20260920/](./remaining-campaign-20260920/) | P6/P7 と Slack PO/evidence 16件の判断材料・ケース票。PO裁定と実機/STGは未了 |
| [todo-campaign-20260918/](./todo-campaign-20260918/) | UAT-R2 / Q2・Q4 集計 / 処置移行 / 死亡日 / Linear 照合の準備票 |
| [todo-campaign-20260919-ready17/](./todo-campaign-20260919-ready17/) | Slack READY 13件の設計票 |
| [todo-ledger-20260922/](./todo-ledger-20260922/) | 2026-09-22 台帳の分類表・実行 receipt・訂正ドラフト（DEV-V-OWNER-DB・UAT-Q3-GENDER-MAP・CSV 契約差分・EXCLUSIVE-LOCK） |
| [plane-md-migration-20260923-receipt.md](./plane-md-migration-20260923-receipt.md) | 2026-09-23 Markdown→Plane 移行 receipt。旧 ID → `EMR-*` / `MIG-*` の crosswalk 正本 |
| [plane-ready-triage-20260927/](./plane-ready-triage-20260927/) | 2026-09-27 Plane Needs Human/Blocked 65 件の ready 判定 verdicts（ready=0、verdicts.json あり） |
| [emr-89-karte-copy-verification.md](./emr-89-karte-copy-verification.md) · [emr-180-note-staff-starttime-rdt-investigation.md](./emr-180-note-staff-starttime-rdt-investigation.md) · [emr-200-note2-coverage/](./emr-200-note2-coverage/) · [emr-211-synthetic-teardown-audit-logs.md](./emr-211-synthetic-teardown-audit-logs.md) | 各 Plane チケットの調査・検証記録（証跡票） |
| [docs-refresh-20260929/](./docs-refresh-20260929/) | 本キャンペーン（docs 全件照合 2026-09-29）の子レポート置き場 |
| [docs-perfection/README.md](./docs-perfection/README.md) | **Astra 調査パッケージ入口**（棚卸し・RQ・子 goal・証拠）。製品の実行 SoT は上記 Plane |
| [docs-perfection/REPAIR-QUEUE.md](./docs-perfection/REPAIR-QUEUE.md) | 調査時点の優先修復キュー（RQ-001〜） |
| [docs-perfection/ROLE-MAP.md](./docs-perfection/ROLE-MAP.md) | docs 保守の Living Docs 役割割当 |

## 削除済み docs（復活防止）

同じ文書を作り直さない。復元は git 履歴。

> 2026-09-29 注記: 下表の「後継」列に残る `Linear` 参照は削除当時の正本。Linear は 2026-09-16 に閉鎖済みで、現在の実行 SoT は Plane `EMR`（hub `EMR-1`）。個別行の改竄はせず、この注記で年代を固定する。

| 削除したもの | 理由 | 後継 |
|---|---|---|
| root `todo-issue.md`・`todo-verification.md`・`todo-operations.md`・`todo-performance.md`（2026-09-26） | ユーザー依頼により todo.md へ統合。別台帳として再作成しない | [todo.md](../../todo.md) の Issue 出典・検証・運用・性能の各節。統合前の原文と行番号参照は Git |
| root `todo-check-auth.md` / `todo-fix-auth.md`（2026-09-15） | 認証の参照用メモと残件の重複管理を解消 | 契約は [認証設計](../architecture/auth.md)、D1・メール・Linear の残件は [統合検証 TODO](../../todo.md#認証認可の外部境界) |
| root `fe-refactor.md`（2026-09-15） | 完了した監査の独立メモ | 維持判断は [裁定記録](./development-task-decisions.md#frontend-監査から引き継ぐ維持判断2026-09-15)、完了履歴は Git |
| root `readiness-report.md`（2026-09-15） | 9月4日時点の環境評価。現在の品質・実行状態の証拠として使わない | 現行規約は [.claude/CLAUDE.md](../../.claude/CLAUDE.md)、検証条件は [agent harness](../ops/agent-harness.md)。当時の評価は Git |
| root `bug.md`・`todo-now.md`・`todo-po.md`・`todo-refactor.md`（2026-09-08） | ユーザー依頼による5台帳の統合。別台帳として再作成しない | [todo.md](../../todo.md) の製品 FAIL・PO・Astra 履歴・FE 履歴。削除前の原文は Git。2026-09-29 訂正: `bug.md` は Plane 移行期に製品欠陥台帳として再作成された後、2026-09-27 の台帳整理で中身を削除され 0 バイトで残存（git `1b4b421bb`）。現在の不具合追跡は Plane `EMR`。旧台帳の再構築は引き続き禁止 |
| `STATUS.md`・`PO-todo.md`・2026-08-20 時点の旧フル `todo.md` / `todo-po.md` | 実行台帳の統合・縮小 | Linear。現在の root 台帳の役割は上記入口を参照 |
| root `phase2.html` / `research-cloudflare.html` / `codex-security-output/` | 作業・調査資料の整理 | 当時の内容は git 履歴。今期外項目は `phase2-deferred.md` |
| root `reports/`（UAT/fable/kanban/walk） | 時点レポート | CorpVault `evidence/2026-08-20-docs-cleanup/` · 新規 UAT は gitignore の `reports/uat-YYYY-MM-DD/`（コミットしない） |
| `docs/work/residual-closeout-ledger.md` | U0–U6 完了ログ | 同 vault `decisions/` · 残ゲートは Linear |
| `docs/work/archives/STATUS-before-2026-08-13-slim.md` | 旧フル台帳 | git 履歴 |
| `docs/ops/infra/_archive/migration-cloudflare.md` | STG 移行の凍結実施記録 | 現行は `docs/ops/infra/architecture.md` · 経緯は git 履歴 |
| `docs/spec/line/01_曽我さん向け_*.md` | クライアント受領原本の二重管理 | 製品正本 `lstep-integration.md` · 原本は vault `client/` |
| `frontend/docs/design-audit-pages.md` | 監査表の複製 | `docs/spec/ui-design-compliance.md` |

**正本:** Plane `EMR`（2026-09-29 訂正: 旧記述「Linear」→ Linear 閉鎖済み）· 製品 docs · 会社側索引は CorpVault

## 置かないもの

- シナリオ md への実行結果書き込み
- 秘密 · 臨床数値の発明
- root への作業台帳フル再構築
- UAT ランナー・スクショ・agent loop の git 追跡
