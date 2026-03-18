# TASK-010: NotionFilter ツールバーアイコン・ボタンサイズ拡大

**作成日**: 2026-03-18
**ステータス**: Open
**依頼元**: ユーザー（Chromeブラウザでの目視確認）

---

## 概要

FE-035（2倍化）実施済みだが、NotionFilter ツールバーのアイコン（ソート・検索・フィルタ追加の Plus）が依然として小さく、視認性・操作性が不十分。アイコンとボタンコンテナをさらに拡大する。

## 依頼内容（原文）

> 検索フィルタのアイコンが小さいです。chromeブラウザで確認して、大きくするタスクを作成して。

## 現状（Chromeブラウザで確認済み）

### ツールバー右側のアイコンボタン
- **Search アイコン**: `h-5 w-5`（20px）/ コンテナ `h-8 w-8`（32px）
- **ArrowUpDown アイコン**（ソート）: `h-5 w-5`（20px）/ コンテナ `h-8 w-8`（32px）
- **ListFilter アイコン**（アクティブフィルタ表示）: `h-5 w-5`（20px）/ コンテナ `h-8 w-8`（32px）

### フィルタ追加ボタン
- **Plus アイコン**: `h-4 w-4`（16px）/ テキスト `text-sm`（14px）

### ポップオーバー内操作アイコン
- **X / ChevronDown / ChevronLeft / Arrow**: `h-4 w-4`（16px）
- **プロパティリスト内アイコン**: `h-5 w-5`（20px）

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | NotionFilter ツールバーアイコン・ボタンサイズ拡大 | FE | FE-038 | - |

## 影響範囲

### DB / Backend
- 変更なし

### Frontend
- `frontend/src/components/shared/NotionFilter/NotionFilter.tsx` — ツールバーアイコン・コンテナサイズ拡大
- `frontend/src/components/shared/NotionFilter/FilterAddPopover.tsx` — Plus アイコン・テキストサイズ拡大
- `frontend/src/components/shared/NotionFilter/SortPopover.tsx` — ソートボタンアイコン・コンテナサイズ拡大
- `frontend/src/components/shared/NotionFilter/FilterRuleRow.tsx` — ルール行内アイコンサイズ拡大
- `frontend/src/components/shared/NotionFilter/SortPill.tsx` — ソートピル内アイコンサイズ拡大

## 実装順序

1. FE-038（NotionFilter 全コンポーネントのアイコンサイズ拡大）

## 関連イシュー

- FE-038: [NotionFilter アイコン・ボタンサイズ拡大](../../frontend/issues/open/FE-038-notion-filter-icon-enlarge.md)
- FE-035（closed）: 前回の2倍化対応（TASK-009）
