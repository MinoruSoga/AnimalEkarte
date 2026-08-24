# タスク台帳 — Linear が正本

| 項目 | 値 |
|------|-----|
| **実行 SoT** | Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)** |
| **直下整理** | **[BRT-105](https://linear.app/baritechllc/issue/BRT-105)** |
| **会社側ログ** | CorpVault `50_Projects/ノア動物病院電子カルテ/`（直下から移した時点レポートは `evidence/2026-08-20-root-docs/`） |
| **旧本文** | git 履歴 |

## 使い方

- 作業の状態・担当・次の一手 → **Linear Issue**
- 製品バグの新規 → Linear に起票
- 開発規約 → [`.claude/CLAUDE.md`](.claude/CLAUDE.md) · [`AGENTS.md`](AGENTS.md)
- 今期外の索引 → [`docs/work/phase2-deferred.md`](docs/work/phase2-deferred.md)

完了済みの検査機器連携メモは vault `evidence/2026-08-20-root-docs/todo-2026-08-20-lab-progress.md`。Linear 上の AE-LAB は BRT-94 / BRT-95〜98 / BRT-100。

## 移行要件（old_db / 城東）— 2026-08-24 PO

旧会計に一般的におかしな値があっても、**補正せずそのまま載せる**。old_db 側は符号を落とさない。AnimalEkarte が受け取れない制約はこちらで外す。

- [x] **AE-MIG-NEG-1** 2026-08-24: CSV cutover が負の請求・入金・split を受け入れる。`002_allow_negative_billing_amounts.sql` が `chk_billings_amounts` を DROP。ローカルは `make reset`（2026-08-24 17:19 JST）で適用済み。backend: `002` completed、`Migration key coverage missing=0 expected=5 recorded=5`。画面の返金表示は別途。commit `55f29ce5c`（未 push）。claim ブランチは統合確認済みで削除済み。
