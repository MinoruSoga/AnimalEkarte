# 作業台帳（docs/work）

## 入口

| 文書 | 役割 |
|------|------|
| **Linear** hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) | **実行 SoT**（状態・担当・ゲート） |
| [`../../todo-po.md`](../../todo-po.md) | **入口ポインタのみ** |
| [`../../todo.md`](../../todo.md) | repo に結び付く未完了作業の入口（受入残・USER ゲート・STG・deferred） |
| [`../../bug.md`](../../bug.md) | UAT 製品 FAIL の暫定例外台帳を含む |
| CorpVault `50_Projects/ノア動物病院電子カルテ/` | 会社側索引・時点ログ |

**競合・終了ルール:** 状態・担当・Done は Linear を正とする。`todo.md` は未完了作業だけを保持し、完了行を除く。STG Lane 4 の終了条件・記録方法は [STG 手順書](../ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) を参照する。`bug.md` の確認済み製品 FAIL は Linear と対応付け、受入未実施や環境 BLOCKED を製品 FAIL に混ぜない。

| 補助 | 役割 |
|------|------|
| [decisions/](./decisions/README.md) | 採択済み方針の短いポインタ |
| [phase2-deferred.md](./phase2-deferred.md) | 今期外の短い索引 |
| [linear-f1-f6-mapping.md](./linear-f1-f6-mapping.md) | repo の実装履歴と Linear の対応案（Linear 現在状態は UNKNOWN） |
| [skill-reeval-2026-09-06.md](./skill-reeval-2026-09-06.md) | 代表タスクごとの資料選択・停止判断 |

## 削除済み docs（復活防止）

同じ文書を作り直さない。復元は git 履歴。

| 削除したもの | 理由 | 後継 |
|---|---|---|
| `STATUS.md`・`PO-todo.md`・2026-08-20 時点の旧フル `todo.md` / `todo-po.md` | 実行台帳の統合・縮小 | Linear。現在の root 台帳の役割は上記入口を参照 |
| root `phase2.html` / `research-cloudflare.html` / `codex-security-output/` | 作業・調査資料の整理 | 当時の内容は git 履歴。今期外項目は `phase2-deferred.md` |
| root `reports/`（UAT/fable/kanban/walk） | 時点レポート | CorpVault `evidence/2026-08-20-docs-cleanup/` · 新規 UAT は gitignore の `reports/uat-YYYY-MM-DD/`（コミットしない） |
| `docs/work/residual-closeout-ledger.md` | U0–U6 完了ログ | 同 vault `decisions/` · 残ゲートは Linear |
| `docs/work/archives/STATUS-before-2026-08-13-slim.md` | 旧フル台帳 | git 履歴 |
| `docs/ops/infra/_archive/migration-cloudflare.md` | STG 移行の凍結実施記録 | 現行は `docs/ops/infra/architecture.md` · 経緯は git 履歴 |
| `docs/spec/line/01_曽我さん向け_*.md` | クライアント受領原本の二重管理 | 製品正本 `lstep-integration.md` · 原本は vault `client/` |
| `frontend/docs/design-audit-pages.md` | 監査表の複製 | `docs/spec/ui-design-compliance.md` |

**正本:** Linear · 製品 docs · 会社側索引は CorpVault

## 置かないもの

- シナリオ md への実行結果書き込み
- 秘密 · 臨床数値の発明
- root への作業台帳フル再構築
- UAT ランナー・スクショ・agent loop の git 追跡
