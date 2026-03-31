# TASK-010: NotionFilter ツールバーアイコンが16pxに縮小される問題の修正

**作成日**: 2026-03-18
**ステータス**: Closed
**依頼元**: ユーザー（Chromeブラウザでの目視確認 + DevTools実測）

---

## 概要

NotionFilter のツールバーアイコンに `h-6 w-6`（24px）を指定しているが、shadcn/ui Button の CSS ルール `[&_svg:not([class*='size-'])]:size-4` により **16px に強制縮小されている**。`h-X w-X` を `size-X` に置き換えることで解決する。

## 依頼内容（原文）

> 検索フィルタのアイコンが小さいです。chromeブラウザで確認して、大きくするタスクを作成して。

## 根本原因

shadcn/ui の Button コンポーネントが子 SVG に対して以下のルールを適用:

```css
[&_svg:not([class*='size-'])]:size-4
```

`h-6 w-6` は `size-` を含まないため、`size-4`（16px）に上書きされる。

### Chrome DevTools 実測値

| 要素 | 指定クラス | 期待値 | **実際のレンダリング** |
|------|-----------|--------|---------------------|
| Search（虫眼鏡） | `h-6 w-6` | 24px | **16px** |
| ArrowUpDown（ソート） | `h-6 w-6` | 24px | **16px** |
| Plus（フィルタ追加） | `h-5 w-5` | 20px | **16px** |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | NotionFilter 全アイコンの `h-X w-X` → `size-X` 置換 | FE | FE-038 | - |

## 影響範囲

### DB / Backend
- 変更なし

### Frontend
- `frontend/src/components/shared/NotionFilter/NotionFilter.tsx`
- `frontend/src/components/shared/NotionFilter/FilterAddPopover.tsx`
- `frontend/src/components/shared/NotionFilter/FilterRuleRow.tsx`
- `frontend/src/components/shared/NotionFilter/SortPopover.tsx`
- `frontend/src/components/shared/NotionFilter/SortPill.tsx`

## 実装順序

1. FE-038（全5ファイルの `h-X w-X` → `size-X` 一括置換）

## 関連イシュー

- FE-038: [NotionFilter アイコン size-X 修正](../../frontend/issues/closed/FE-038-notion-filter-icon-enlarge.md)
- FE-035（closed）: 前回のサイズ拡大対応（値は変更済みだが、Button の CSS に上書きされていた）
