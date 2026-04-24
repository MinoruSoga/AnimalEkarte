# FE-038: NotionFilter ツールバーアイコンが16pxに縮小される問題の修正

**親タスク**: [TASK-010](../../docs/tasks/open/TASK-010-notion-filter-icon-size-enlarge.md)
**Status**: Open
**Priority**: High
**Affects**: NotionFilter（全リスト画面に影響）
**Date Created**: 2026-03-18
**Related**: FE-035（closed — 前回の2倍化対応）

## Summary

NotionFilter のツールバーアイコンに `h-6 w-6`（24px）を指定しているが、shadcn/ui の Button コンポーネントの CSS ルールにより **実際のレンダリングは 16x16px に強制縮小されている**。

## 根本原因

shadcn/ui の Button コンポーネントに以下の CSS ルールが含まれている:

```css
[&_svg:not([class*='size-'])]:size-4
```

このルールは「`size-` を含むクラスを持たない子 SVG に `size-4`（16px）を強制適用する」。
`h-6 w-6` は `size-` を含まないため、このルールに**上書きされて 16px になる**。

### Chrome DevTools での計測結果（実測値）

| 要素 | 指定クラス | 期待サイズ | **実際のレンダリング** |
|------|-----------|-----------|---------------------|
| Search アイコン（虫眼鏡） | `h-6 w-6` | 24px | **16px** |
| ArrowUpDown アイコン（ソート） | `h-6 w-6` | 24px | **16px** |
| Plus アイコン（フィルタ追加） | `h-5 w-5` | 20px | **16px** |

コンテナサイズ（`h-9 w-9` = 36px）は正しくレンダリングされている。

## 必要な変更

### 修正方法

`h-X w-X` を `size-X` に置き換える。`size-` クラスがあれば `[&_svg:not([class*='size-'])]:size-4` の条件から外れ、指定サイズが適用される。

### 1. NotionFilter.tsx

```typescript
// Line 140: ListFilter アイコン
// Before
<ListFilter className="h-6 w-6" />
// After
<ListFilter className="size-6" />

// Line 166: Search アイコン
// Before
<Search className="h-6 w-6" />
// After
<Search className="size-6" />
```

### 2. SortPopover.tsx

```typescript
// Line 177: ArrowUpDown アイコン
// Before
<ArrowUpDown className="h-6 w-6" />
// After
<ArrowUpDown className="size-6" />

// Line 86: ArrowUp アイコン
// Before
<ArrowUp className="h-5 w-5" />
// After
<ArrowUp className="size-5" />

// Line 88: ArrowDown アイコン
// Before
<ArrowDown className="h-5 w-5" />
// After
<ArrowDown className="size-5" />

// Line 100: X アイコン（削除）
// Before
<X className="h-5 w-5" />
// After
<X className="size-5" />

// Line 233: Plus アイコン（追加）
// Before
<Plus className="h-5 w-5" />
// After
<Plus className="size-5" />
```

### 3. FilterAddPopover.tsx

```typescript
// Line 294: Plus アイコン
// Before
<Plus className="h-5 w-5" />
// After
<Plus className="size-5" />

// Line 314: ChevronLeft アイコン
// Before
<ChevronLeft className="h-4 w-4" />
// After
<ChevronLeft className="size-4" />
```

### 4. FilterRuleRow.tsx

```typescript
// Line 128: ChevronDown アイコン
// Before
<ChevronDown className="h-5 w-5" />
// After
<ChevronDown className="size-5" />

// Line 455: X アイコン（削除）
// Before
<X className="h-5 w-5" />
// After
<X className="size-5" />
```

### 5. SortPill.tsx

```typescript
// Line 64: DirectionIcon
// Before
<DirectionIcon className="h-5 w-5" />
// After
<DirectionIcon className="size-5" />

// Line 68: ChevronDown
// Before
<ChevronDown className="h-5 w-5" />
// After
<ChevronDown className="size-5" />

// Line 80: ArrowDown
// Before
<ArrowDown className="h-5 w-5" />
// After
<ArrowDown className="size-5" />

// Line 82: ArrowUp
// Before
<ArrowUp className="h-5 w-5" />
// After
<ArrowUp className="size-5" />

// Line 122: X アイコン
// Before
<X className="h-5 w-5" />
// After
<X className="size-5" />
```

### 6. design-tokens.ts

```typescript
// Line 641-642: searchIcon
// Before
searchIcon: `absolute left-2.5 top-1/2 -translate-y-1/2 size-5 ${C.text30}`,
// After（これは既に size-5 なので変更不要）
```

## 変更ルール（まとめ）

NotionFilter 内の全 SVG アイコンクラスを以下のように置換:

| Before | After | 備考 |
|--------|-------|------|
| `h-6 w-6` | `size-6` | ツールバー主要アイコン |
| `h-5 w-5` | `size-5` | ポップオーバー内操作アイコン |
| `h-4 w-4` | `size-4` | 小型アイコン（ChevronLeft等） |

**注意**: `mr-2 h-5 w-5` のようにマージン付きの場合は `mr-2 size-5` に置換する。

## 影響ページ（全リスト画面）

NotionFilter は共有コンポーネントのため、以下の全ページに自動的に修正が反映される:
- 飼主・ペット一覧（/owners）
- カルテ一覧（/medical-records）
- 予約管理（/reservations）
- 検査管理（/examinations）
- 予防接種（/vaccinations）
- 会計管理（/accounting）
- 入院管理（/hospitalization）
- トリミング（/trimming）
- 在庫管理（/inventory）
- 見積書管理（/estimates）

## プロジェクトルール遵守チェック

- [x] `any` 型なし
- [x] `FC` / `forwardRef` なし
- [x] barrel index 経由 import なし
- [x] 条件レンダー `? ... : null`（`&&` 禁止）
- [x] CSSクラスの変更のみ（ロジック変更なし）

## 完了条件

- [ ] NotionFilter 内の全 SVG アイコンが `size-X` クラスを使用
- [ ] Chrome DevTools で Search アイコンが **24x24px** でレンダリングされることを確認
- [ ] Chrome DevTools で ArrowUpDown アイコンが **24x24px** でレンダリングされることを確認
- [ ] Chrome DevTools で Plus アイコンが **20x20px** でレンダリングされることを確認
- [ ] 全リスト画面で目視確認（レイアウト崩れなし）
- [ ] `docker compose exec frontend pnpm lint` エラーなし
- [ ] `docker compose exec frontend pnpm build` 成功
