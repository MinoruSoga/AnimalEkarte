# 作業台帳（docs/work）

## 入口

| 文書 | 役割 |
|------|------|
| **Linear** hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) | **実行 SoT**（状態・担当・ゲート） |
| [`../../todo.md`](../../todo.md) · [`../../todo-po.md`](../../todo-po.md) | **ポインタのみ** |
| CorpVault `50_Projects/ノア動物病院電子カルテ/` | 会社側索引・時点ログ |

| 補助 | 役割 |
|------|------|
| [decisions/](./decisions/README.md) | 採択済み方針の短いポインタ |
| [phase2-deferred.md](./phase2-deferred.md) | 今期外の短い索引 |

**削除統合済:** `STATUS.md` · `PO-todo.md` · 旧フル `todo.md`/`todo-po.md` · 直下 `phase2.html` / `research-cloudflare.html` / `codex-security-output/` · `reports/`（2026-08-14 UAT/agent 時点） · `residual-closeout-ledger.md` · 旧 STATUS 全文アーカイブ · `_archive/migration-cloudflare.md`

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
