# FE-023: NotionFilter — フィルタ条件・ルール行表示・AND/OR 切替

**Status**: Open
**Priority**: High
**Affects**: shared NotionFilter コンポーネント
**Date Created**: 2026-03-17
**Related**: TASK-006, FE-025

## Summary

NotionFilter のフィルタ追加フローに「条件」ステップを追加し、アクティブフィルタをピルではなくルール行で表示し、AND/OR の切替を可能にする。Notion 本家のフィルタUIに準拠。

## Notion テーブル参照（スクリーンショット: 22.44.23）

```
右上ツールバー:  ≡(フィルタ)  ↕(ソート)  🔍(検索)  ← 3つの独立アイコンボタン

ピル行:
[↓ 最新使用日 ∨]          ← オレンジ色ソートピル（クリックで編集 Popover）
[📊 ステータス: Approved ∨] ← ブルー色フィルタピル（クリックで条件編集 Popover）
+ フィルター               ← テキストリンク
```

**重要な差分:**
- ピルはクリックで**条件を再編集**できる（削除だけでなく値変更も）
- ソートピルとフィルタピルが**色で区別**されている（オレンジ vs ブルー）
- ピル末尾は `×` ではなく `∨`（ドロップダウン展開）

## 現状のコード

### 現在のフィルタ追加フロー

```
「+ フィルタを追加」→ プロパティ選択 → 値選択 → ピル表示
```

Notion 本家:
```
「+ フィルタルールを追加」→ プロパティ選択 → 条件選択（is/is not/contains...）→ 値選択 → ルール行表示
```

### 現在の型定義

```typescript
// frontend/src/components/shared/NotionFilter/types.ts:1-37
export type FilterType = "select" | "multi-select" | "date-range";

export interface ActiveFilter {
  key: string;
  value: string | string[] | { from?: string; to?: string };
  displayValue: string;
  // ← condition がない
}
```

### 現在のフィルタ表示（FilterPill）

```typescript
// frontend/src/components/shared/NotionFilter/FilterPill.tsx:17-46
// Badge（ピル）で横並び表示
<Badge variant="secondary" className="gap-1 pl-2 pr-1 h-7 ...">
  <span>{property.label}:</span>
  <span>{filter.displayValue}</span>
  <span onClick={onRemove}><X /></span>
</Badge>
```

## 必要な変更

### 1. 型定義の拡張

```typescript
// frontend/src/components/shared/NotionFilter/types.ts

// フィルタ条件の種類
export type FilterCondition =
  | "is"              // 一致
  | "is_not"          // 不一致
  | "contains"        // 含む（テキスト）
  | "does_not_contain" // 含まない（テキスト）
  | "is_before"       // 以前（日付）
  | "is_after"        // 以降（日付）
  | "is_between"      // 期間内（日付）
  | "is_empty"        // 空
  | "is_not_empty";   // 空でない

// プロパティ型ごとに使える条件を定義
export const FILTER_CONDITIONS: Record<FilterType, { value: FilterCondition; label: string }[]> = {
  "select": [
    { value: "is", label: "次と一致" },
    { value: "is_not", label: "次と不一致" },
    { value: "is_empty", label: "空" },
    { value: "is_not_empty", label: "空でない" },
  ],
  "multi-select": [
    { value: "contains", label: "含む" },
    { value: "does_not_contain", label: "含まない" },
    { value: "is_empty", label: "空" },
    { value: "is_not_empty", label: "空でない" },
  ],
  "date-range": [
    { value: "is", label: "次と一致" },
    { value: "is_before", label: "以前" },
    { value: "is_after", label: "以降" },
    { value: "is_between", label: "期間内" },
    { value: "is_empty", label: "空" },
    { value: "is_not_empty", label: "空でない" },
  ],
};

// ActiveFilter に condition を追加
export interface ActiveFilter {
  key: string;
  condition: FilterCondition;     // ← 追加
  value: string | string[] | { from?: string; to?: string };
  displayValue: string;
}

// フィルタグループの論理演算
export type FilterLogic = "and" | "or";

// NotionFilter Props に logic を追加
export interface NotionFilterProps {
  properties: FilterProperty[];
  activeFilters: ActiveFilter[];
  onFilterChange: (filters: ActiveFilter[]) => void;
  filterLogic?: FilterLogic;                    // ← 追加
  onFilterLogicChange?: (logic: FilterLogic) => void;  // ← 追加
  searchTerm?: string;
  onSearchChange?: (value: string) => void;
  searchPlaceholder?: string;
  count?: number;
}
```

### 2. FilterAddPopover — 3ステップフロー

```
Step 1: プロパティ選択（現状と同じ）
Step 2: 条件選択（新規）—「次と一致」「次と不一致」「含む」等
Step 3: 値選択（現状と同じ）
```

### 3. FilterRuleRow — ルール行表示（新規コンポーネント）

FilterPill を置き換え。Notion 本家のようにルール行で表示:

```
[AND ▼] [ステータス ▼] [次と一致 ▼] [会計待ち ▼] [×]
[       ] [日付      ▼] [期間内   ▼] [今週      ▼] [×]
```

- 最初の行は AND/OR 切替ドロップダウン
- 2行目以降は AND/OR ラベルのみ（変更は最初の行で一括）
- 各セル（プロパティ/条件/値）はクリックで変更可能
- × で行削除

### 4. NotionFilter.tsx — レイアウト変更

```typescript
// Before: flex-wrap でピル横並び
<div className="flex flex-wrap items-center gap-2">
  {count}
  {activeFilters.map(FilterPill)}
  <FilterAddPopover />
  {search}
</div>

// After: フィルタルール行を縦積み + ツールバー
<div className="flex flex-col gap-1">
  {/* ツールバー行 */}
  <div className="flex items-center gap-2">
    {count}
    <FilterAddPopover />
    {search}
  </div>
  {/* フィルタルール行 */}
  {activeFilters.length > 0 ? (
    <div className="flex flex-col gap-1 pl-1">
      {activeFilters.map((filter, i) => (
        <FilterRuleRow
          key={filter.key}
          filter={filter}
          property={...}
          isFirst={i === 0}
          logic={filterLogic}
          onLogicChange={onFilterLogicChange}
          onConditionChange={...}
          onValueChange={...}
          onRemove={...}
        />
      ))}
    </div>
  ) : null}
</div>
```

## コンポーネント構成（変更後）

```
frontend/src/components/shared/NotionFilter/
├── NotionFilter.tsx           ← レイアウト変更（ピル → ルール行）
├── FilterAddPopover.tsx       ← 3ステップフロー（プロパティ→条件→値）
├── FilterRuleRow.tsx          ← 新規: ルール行表示
├── FilterPill.tsx             ← 廃止（FilterRuleRow に置換）
├── types.ts                   ← FilterCondition, FilterLogic 追加
└── index.ts
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `memo()` で FilterRuleRow を最適化
- [ ] `useCallback` でハンドラ安定化

## 依存関係

- Backend 変更不要
- FilterPill.tsx は FilterRuleRow.tsx に置換（後方互換性なし）
- FE-025 で全ページの呼び出しを更新する必要あり

## Notion 日付フィルタ UI 参照（スクリーンショット: /Users/minoru/Downloads/スクリーンショット 2026-03-17 22.42.15.png）

Notion の日付フィルタは以下の構成:

```
┌─────────────────────────────────────┐
│ [開始日 ∨]  [今日と相対日付 ∨]      │  ← フィールド選択 + 条件タイプ
│                                     │
│ [今 ∨]        [週 ∨]                │  ← 相対日付: 時点 × 単位
│                                     │
│        2026年3月        < >         │  ← カレンダー
│  日  月  火  水  木  金  土          │
│  ...                                │
│ [15] 16  17  18  19  20 [21]        │  ← 今週がハイライト
│  ...                                │
│                                     │
│ フィルターは最新の日付に合わせて更新... │
└─────────────────────────────────────┘

下部ピル: [更新日時: 今 週 ∨]  + フィルター
```

### 日付条件タイプ（「今日と相対日付 ∨」ドロップダウン）

```typescript
type DateConditionType =
  | "relative"        // 今日と相対日付（今/前/来 × 日/週/月/年）
  | "exact"           // 完全一致（特定の日付）
  | "before"          // 以前
  | "after"           // 以降
  | "on_or_before"    // 以前（当日含む）
  | "on_or_after"     // 以降（当日含む）
  | "between"         // 期間内（カスタム範囲）
  | "is_empty"        // 空
  | "is_not_empty";   // 空でない
```

### 相対日付セレクタ（2つのドロップダウン組合せ）

```typescript
// 時点セレクタ
type RelativePoint = "this" | "last" | "next";  // 今 / 前 / 来

// 単位セレクタ
type RelativeUnit = "day" | "week" | "month" | "year";  // 日 / 週 / 月 / 年

// 組合せ例:
// 「今」×「週」= 今週（3/15〜3/21）
// 「前」×「月」= 先月（2月）
// 「来」×「日」= 明日
```

この相対日付セレクタは現在の DATE_PRESETS（固定リスト）より柔軟だ。DATE_PRESETS を廃止し、ドロップダウン組合せに置き換える。

## 完了条件

- [ ] フィルタ追加時に条件選択ステップが表示される
- [ ] select 型: 「次と一致」「次と不一致」「空」「空でない」
- [ ] date-range 型: 条件タイプ切替（相対日付/完全一致/以前/以降/期間内/空/空でない）
- [ ] 相対日付: 時点（今/前/来）× 単位（日/週/月/年）のドロップダウン組合せ
- [ ] カレンダーで選択範囲がハイライト表示
- [ ] アクティブフィルタがルール行で表示される
- [ ] AND/OR 切替が動作する
- [ ] 各ルール行の × で個別削除が動作する
- [ ] フィルタピルに「更新日時: 今 週」形式で表示
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
