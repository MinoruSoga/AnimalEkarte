# FE-017: NotionFilter 共有コンポーネント新規作成

**Status**: Open
**Priority**: High
**Affects**: 全一覧ページ共通 — 検索フィルタUI
**Date Created**: 2026-03-17
**Related**: TASK-003, FE-018

## Summary

Notion のフィルタUIパターンを再現した共有コンポーネント `NotionFilter` を新規作成する。「フィルタ追加」ボタン → Popover → プロパティ選択 → 条件/値入力 → アクティブフィルタのピル表示、というフローを実装する。

## 現状のコード

### SearchFilterBar（置き換え対象）

```typescript
// frontend/src/components/shared/SearchFilterBar/SearchFilterBar.tsx:14-42
interface SearchFilterBarProps {
  searchTerm: string;
  onSearchChange: (value: string) => void;
  placeholder?: string;
  children?: React.ReactNode;  // 追加フィルタを slot で配置
  count?: number;
  className?: string;
}
// テキスト検索バー + children スロットでフィルタを差し込む設計
// Lucide Search アイコン + Input + 結果件数表示
```

### 使用している shadcn/ui コンポーネント（利用可能）

```
frontend/src/components/ui/
├── popover.tsx     ← Popover（フィルタ追加パネル）
├── command.tsx     ← Command（コマンドパレット風の検索付きリスト）
├── select.tsx      ← Select（条件選択）
├── badge.tsx       ← Badge（アクティブフィルタのピル表示）
├── button.tsx      ← Button
├── input.tsx       ← Input（テキスト入力）
├── calendar.tsx    ← Calendar（日付選択）
└── separator.tsx   ← Separator
```

## Notion フィルタ UI の仕様

### 1. フィルタバー（常時表示）

```
[🔍 テキスト検索...] [+ フィルタを追加] [件数: XX件]

// アクティブフィルタがある場合:
[🔍 テキスト検索...] [ステータス: 会計待ち ×] [日付: 2026-03-01〜03-31 ×] [+ フィルタを追加] [件数: XX件]
```

- テキスト検索は常時表示（Notion と同様）
- アクティブフィルタは Badge（ピル）で表示
- 各ピルに `×` ボタンで個別削除
- 「+ フィルタを追加」ボタンで Popover を開く

### 2. フィルタ追加 Popover

```
┌─────────────────────────┐
│ フィルタを追加           │
│ ─────────────────────── │
│ 📋 ステータス            │
│ 📅 日付                  │
│ 👤 担当医                │
│ 📦 カテゴリ              │
│ ...（ページ固有の選択肢） │
└─────────────────────────┘
```

- Command コンポーネントで検索可能なリスト表示
- プロパティを選択すると条件入力に遷移

### 3. 条件入力（プロパティ選択後）

```
// ステータス型: Select で値選択
┌─────────────────────────┐
│ ステータス               │
│ ─────────────────────── │
│ ○ 会計待ち               │
│ ○ 会計済                 │
│ ○ キャンセル             │
└─────────────────────────┘

// 日付型: DateRangePicker
┌─────────────────────────┐
│ 日付                     │
│ ─────────────────────── │
│ [開始日] 〜 [終了日]     │
└─────────────────────────┘
```

## 必要な変更

### 1. 型定義

```typescript
// frontend/src/components/shared/NotionFilter/types.ts（新規）

// フィルタプロパティの種類
type FilterType = "text" | "select" | "multi-select" | "date-range";

// フィルタプロパティの定義（各ページが定義）
interface FilterProperty {
  key: string;           // "status", "date", "doctor" 等
  label: string;         // "ステータス", "日付", "担当医"
  type: FilterType;
  options?: FilterOption[];  // select/multi-select の選択肢
  icon?: LucideIcon;        // プロパティアイコン
}

interface FilterOption {
  value: string;
  label: string;
}

// アクティブなフィルタ値
interface ActiveFilter {
  key: string;
  value: string | string[] | { from?: string; to?: string };
  displayValue: string;  // ピルに表示するテキスト
}

// コンポーネント Props
interface NotionFilterProps {
  properties: FilterProperty[];       // 利用可能なフィルタプロパティ
  activeFilters: ActiveFilter[];      // 現在アクティブなフィルタ
  onFilterChange: (filters: ActiveFilter[]) => void;  // フィルタ変更コールバック
  searchTerm?: string;                // テキスト検索（オプション）
  onSearchChange?: (value: string) => void;
  searchPlaceholder?: string;
  count?: number;                     // 結果件数
}
```

### 2. コンポーネント構成

```
frontend/src/components/shared/NotionFilter/
├── NotionFilter.tsx           ← メインコンポーネント（フィルタバー）
├── FilterAddPopover.tsx       ← 「+ フィルタを追加」Popover
├── FilterValuePopover.tsx     ← 値入力 Popover（Select/DateRange切替）
├── FilterPill.tsx             ← アクティブフィルタのピル表示
├── types.ts                   ← 型定義
└── index.ts                   ← 公開API
```

### 3. NotionFilter.tsx（メインコンポーネント）

```typescript
// frontend/src/components/shared/NotionFilter/NotionFilter.tsx

export function NotionFilter({
  properties,
  activeFilters,
  onFilterChange,
  searchTerm,
  onSearchChange,
  searchPlaceholder = "検索...",
  count,
}: NotionFilterProps) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      {/* テキスト検索 */}
      {onSearchChange ? (
        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            value={searchTerm}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder={searchPlaceholder}
            className="pl-8 h-8 w-[200px]"
          />
        </div>
      ) : null}

      {/* アクティブフィルタのピル */}
      {activeFilters.map((filter) => (
        <FilterPill
          key={filter.key}
          filter={filter}
          property={properties.find((p) => p.key === filter.key)}
          onRemove={() => handleRemoveFilter(filter.key)}
          onChange={(newValue) => handleUpdateFilter(filter.key, newValue)}
        />
      ))}

      {/* フィルタ追加ボタン */}
      <FilterAddPopover
        properties={properties}
        activeFilters={activeFilters}
        onAdd={handleAddFilter}
      />

      {/* 件数表示 */}
      {count !== undefined ? (
        <span className="text-xs text-muted-foreground ml-auto">
          {count}件
        </span>
      ) : null}
    </div>
  );
}
```

### 4. FilterPill.tsx（ピル表示）

```typescript
// アクティブフィルタ1件分のピル表示
// クリックで値変更 Popover、× で削除

<Badge variant="secondary" className="gap-1 pl-2 pr-1 h-7 cursor-pointer">
  <span className="text-xs text-muted-foreground">{property.label}:</span>
  <span className="text-xs font-medium">{filter.displayValue}</span>
  <button onClick={onRemove} className="ml-1 hover:bg-muted rounded-sm p-0.5">
    <X className="h-3 w-3" />
  </button>
</Badge>
```

### 5. FilterAddPopover.tsx（フィルタ追加）

```typescript
// Popover + Command でプロパティ一覧を検索可能に表示
// 既にアクティブなプロパティは非表示（同じフィルタの重複追加を防止）
// プロパティ選択後、値入力 Popover に遷移

<Popover>
  <PopoverTrigger asChild>
    <Button variant="outline" size="sm" className="h-7 gap-1">
      <Plus className="h-3 w-3" />
      フィルタを追加
    </Button>
  </PopoverTrigger>
  <PopoverContent className="w-[200px] p-0">
    <Command>
      <CommandInput placeholder="フィルタを検索..." />
      <CommandList>
        {availableProperties.map((prop) => (
          <CommandItem key={prop.key} onSelect={() => handleSelect(prop)}>
            {prop.icon ? <prop.icon className="mr-2 h-4 w-4" /> : null}
            {prop.label}
          </CommandItem>
        ))}
      </CommandList>
    </Command>
  </PopoverContent>
</Popover>
```

### 6. FilterValuePopover.tsx（値入力）

```typescript
// フィルタ種別に応じて値入力UIを切替:
// - "select": ラジオボタンリスト
// - "multi-select": チェックボックスリスト
// - "date-range": DateRangePicker
```

## 使用例（各ページでの呼び出し）

```typescript
// 例: 会計一覧
const ACCOUNTING_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    options: [
      { value: "waiting", label: "会計待ち" },
      { value: "completed", label: "会計済" },
      { value: "cancelled", label: "キャンセル" },
    ],
  },
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
];

// コンポーネント内:
<NotionFilter
  properties={ACCOUNTING_FILTER_PROPERTIES}
  activeFilters={activeFilters}
  onFilterChange={setActiveFilters}
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  searchPlaceholder="飼主名、ペット名..."
  count={filteredRecords.length}
/>
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし（直接ファイル import）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `memo()` で FilterPill, FilterAddPopover を最適化
- [ ] `useCallback` でハンドラ安定化
- [ ] 型は明示的に定義（types.ts）
- [ ] shadcn/ui コンポーネントを活用（Popover, Command, Badge, Button）

## 依存関係

- Backend 変更不要
- shadcn/ui の Popover, Command, Badge が既に利用可能

## 完了条件

- [ ] `NotionFilter/` ディレクトリ新規作成（5ファイル）
- [ ] テキスト検索が動作
- [ ] select 型フィルタが Popover で選択・ピル表示・削除可能
- [ ] date-range 型フィルタが DateRangePicker で選択・ピル表示・削除可能
- [ ] 複数フィルタの同時適用が可能
- [ ] 「+ フィルタを追加」で未使用プロパティのみ表示
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
