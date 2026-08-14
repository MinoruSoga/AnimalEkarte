# bug.md — 受入バグ Open

| 項目 | 値 |
|------|-----|
| **更新** | 2026-08-14（サイドバーマスタ UAT 再実施完了） |
| **main tip** | `1386e1db0` |
| **方針** | 未対応コード欠陥のみ |

## Open

**なし。**

| 実施 | 結果 |
|------|------|
| サイドバーマスタ全ページ + フォーム項目 | [`reports/uat-2026-08-14-sidebar-masters/FINAL.md`](reports/uat-2026-08-14-sidebar-masters/FINAL.md) · **FAIL 0 · 起票 0** |
| scenarios 通し | [`reports/uat-2026-08-14/FINAL.md`](reports/uat-2026-08-14/FINAL.md) · FAIL 0 |

## メモ（バグではない）

- キャンペーン: 開始日・終了日必須
- 問診テンプレ: カテゴリ必須
- シフトテンプレ: 名は `テンプレート名`（`#master-title` ではない）
- 支払方法 C3-2 / 締め休診 C1 は自動化 PARTIAL（欠陥証拠なし）

## 同期

- 技術・PO タスク: **Linear**（[todo.md](todo.md) はポインタ · マップ [todo-docs-linear-map](reports/todo-walk-2026-08-14/todo-docs-linear-map.md)）
- 受入バグ Open: なし（新規は Linear / 必要なら GitHub）
