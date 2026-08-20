# 作業台帳（docs/work）

## 入口

| 文書 | 役割 |
|------|------|
| **Linear** hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) | **実行 SoT**（状態・担当・ゲート） |
| [`../../todo.md`](../../todo.md) · [`../../todo-po.md`](../../todo-po.md) | **ポインタのみ**（本文は Linear へ移行済み） |
| [`../../reports/todo-walk-2026-08-14/todo-docs-linear-map.md`](../../reports/todo-walk-2026-08-14/todo-docs-linear-map.md) | docs Open → Linear マップ |
| [`../../reports/todo-walk-2026-08-14/github-linear-map.md`](../../reports/todo-walk-2026-08-14/github-linear-map.md) | GH Open → Linear マップ |

| 補助 | 役割 |
|------|------|
| [decisions/](./decisions/README.md) | 採択済み方針 |
| [residual-closeout-ledger.md](./residual-closeout-ledger.md) | opaque_ref 縦ログ |
| [archives/](./archives/) | 旧 STATUS 全文など |
| `phase2.html` | 今期外 |

**削除統合済:** `STATUS.md` · `PO-todo.md` · 旧フル `todo.md`/`todo-po.md` 本文（git 履歴）  

## 削除済み docs（復活防止）

同じ文書を作り直さないための記録。復元は `git checkout <sha>^ -- <path>`。

| 削除したもの | 理由 | 後継 | 削除 commit |
|---|---|---|---|
| `docs/architecture/be9-0-legacy-gate-inventory.md` | 2026-07-19 の時点 snapshot。追跡対象 `internal/{handler,service,repository}` は削除済み | live lint は `backend/internal/lintscan/` | `1bd219ff9` |
| `docs/ops/deploy/ANIMALEKARTE_CSV_IMPORT_COMPLETION.md` | 自ら「履歴記録」と宣言し後継を明示していた | `CLINIC_CSV_IMPORT.md` · `SEED_MIGRATION_OPERATIONS.md` | `1bd219ff9` |
| `docs/ops/testing/BUG-007-regression-memo.md` | 期待挙動の契約は実コードに存在 | `TestUnpaidIncludesCreditCorrectionResidual_BUG007` · `TestOutstandingAmount` | `1bd219ff9` |
| `docs/ops/testing/BUG-011-regression-memo.md` | 同上 | `resolveCompleteMedicalRecordID` + `use-accounting-completion-action.ts` の BUG-011 注記 | `1bd219ff9` |
| `docs/ops/testing/scenarios/reports/` 3 本 | 2026-07 時点の UAT snapshot | 証跡の置き場は root `reports/uat-YYYY-MM-DD/` | `1bd219ff9` |
| `docs/work/hermes-scenarios-team-board.md` | 2026-08-07 に完了した実行ボード | 本 README の「置かないもの」規律どおり作らない | `1bd219ff9` |

**再作成の条件**: 上記はいずれも時点記録である。同種の記録が必要になった場合も docs/ には置かず、root `reports/` か Linear に置く。

**正本:** Linear · マップ reports · ポインタ root `todo*.md`

## 置かないもの

- シナリオ md への実行結果書き込み
- 秘密 · 臨床数値の発明
- root への作業台帳フル再構築（Linear を増やす）
