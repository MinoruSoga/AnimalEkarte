# 作業台帳（docs/work）

## 入口

| 文書 | 役割 |
|------|------|
| **Linear** hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) | **実行 SoT**（状態・担当・ゲート） |
| [`../../todo-po.md`](../../todo-po.md) | **入口ポインタのみ** |
| [`../../todo.md`](../../todo.md) | STG 実データ運用テストの暫定例外台帳を含む |
| [`../../bug.md`](../../bug.md) | UAT 製品 FAIL の暫定例外台帳を含む |
| CorpVault `50_Projects/ノア動物病院電子カルテ/` | 会社側索引・時点ログ |

**競合・終了ルール:** 状態が競合した場合は Linear を正とし、root 例外台帳は同一変更で同期する。`todo.md` の STG 例外は、同ファイルに定めた「必須 4 業務を連続 5 営業日、上限 8 週」の終了判定後に閉じ、遅くともその 8 週 window の終了時に Linear / 承認済み evidence へ結果を移して `todo.md` を入口ポインタへ戻す。`bug.md` の UAT FAIL も Linear で解決・延期を確定した同じ変更で例外行を除く。

| 補助 | 役割 |
|------|------|
| [decisions/](./decisions/README.md) | 採択済み方針の短いポインタ |
| [phase2-deferred.md](./phase2-deferred.md) | 今期外の短い索引 |

**削除統合済:** `STATUS.md` · `PO-todo.md` · 2026-08-20 時点の旧フル `todo.md`/`todo-po.md`（その後 `todo.md` へ STG 例外台帳を再導入） · 直下 `phase2.html` / `research-cloudflare.html` / `codex-security-output/` · `reports/`（2026-08-14 UAT/agent 時点） · `residual-closeout-ledger.md` · 旧 STATUS 全文アーカイブ · `_archive/migration-cloudflare.md`

## 削除済み docs（復活防止）

同じ文書を作り直さない。復元は git 履歴。

| 削除したもの | 理由 | 後継 |
|---|---|---|
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
