# FE-024: NotionFilter — ソート統合・検索トグル

**Status**: Open
**Priority**: Medium
**Affects**: shared NotionFilter コンポーネント
**Date Created**: 2026-03-17
**Related**: TASK-006, FE-025

## Summary

NotionFilter に「並べ替え」ボタンと検索トグル（クリックで展開/非展開）を追加し、Notion のテーブルツールバーに準拠させる。現在は検索が常時表示で、ソートは各ページが個別に `useTableSort` で実装している。

## 現状のコード

### 検索（常時表示）

```typescript
// frontend/src/components/shared/NotionFilter/NotionFilter.tsx:59-70
{onSearchChange ? (
  <div className="relative flex-1 max-w-md ml-auto">
    <Search className={STYLE.searchIcon} />
    <input
      type="text"
      value={searchTerm ?? ""}
      onChange={(e) => onSearchChange(e.target.value)}
      placeholder={searchPlaceholder}
      className={STYLE.searchInput}
    />
  </div>
) : null}
```

### ソート（各ページ個別実装）

```typescript
// frontend/src/features/owners/routes/OwnersList.tsx 等
// useTableSort hook で個別にソート管理
// テーブルヘッダーにソートボタン（▲▼）
```

## 必要な変更

### 1. NotionFilter Props 拡張

```typescript
// frontend/src/components/shared/NotionFilter/types.ts に追加

export interface SortProperty {
  key: string;
  label: string;
  icon?: LucideIcon;
}

export interface ActiveSort {
  key: string;
  direction: "asc" | "desc";
}

export interface NotionFilterProps {
  // 既存
  properties: FilterProperty[];
  activeFilters: ActiveFilter[];
  onFilterChange: (filters: ActiveFilter[]) => void;
  filterLogic?: FilterLogic;
  onFilterLogicChange?: (logic: FilterLogic) => void;
  searchTerm?: string;
  onSearchChange?: (value: string) => void;
  searchPlaceholder?: string;
  count?: number;

  // 追加: ソート
  sortProperties?: SortProperty[];
  activeSorts?: ActiveSort[];
  onSortChange?: (sorts: ActiveSort[]) => void;
}
```

## Notion テーブル参照（スクリーンショット: 22.44.23）

```
右上ツールバー:  ≡(フィルタ)  ↕(ソート)  🔍(検索)

ピル行:
[↓ 最新使用日 ∨]   ← オレンジ色ソートピル
[📊 ステータス: Approved ∨]  ← ブルー色フィルタピル
+ フィルター
```

**ソートピルの仕様:**
- オレンジ色の Badge で `↓ プロパティ名` 表示
- `∨` クリックで昇順/降順切替、プロパティ変更、削除が可能
- フィルタピルと同じ行に並ぶ

### 2. ツールバー — 3つのアイコンボタン（Notion 準拠）

```
テーブルヘッダー右端:  ≡  ↕  🔍
```

- **≡** フィルタ: クリックでフィルタ Popover（プロパティ→条件→値）
- **↕** ソート: クリックでソート Popover（プロパティ→昇順/降順）
- **🔍** 検索: クリックで検索バー展開/非展開（トグル）

3つとも小さいアイコンボタン（テキストなし）。

### 3. SortPopover（新規コンポーネント）

```typescript
// frontend/src/components/shared/NotionFilter/SortPopover.tsx（新規）

// クリックで Popover 表示
// プロパティ選択 → 昇順/降順 選択
// 複数ソート可（Notion と同様）
// アクティブなソートがルール行で表示:
// [飼主名 ▼ 昇順 ×]
// [生年月日 ▼ 降順 ×]
```

### 4. 検索トグル

```typescript
// Before: 検索バー常時表示
// After: 🔍 アイコンボタンクリックで展開/非展開

const [searchOpen, setSearchOpen] = useState(false);

// ツールバー
<Button variant="ghost" size="sm" onClick={() => setSearchOpen(!searchOpen)}>
  <Search className="h-4 w-4" />
</Button>

// 検索バー（searchOpen 時のみ表示）
{searchOpen ? (
  <div className="relative ...">
    <input ... autoFocus />
  </div>
) : null}
```

## コンポーネント構成（変更後）

```
frontend/src/components/shared/NotionFilter/
├── NotionFilter.tsx           ← ツールバーレイアウト変更
├── FilterAddPopover.tsx       ← 既存（FE-023 で条件追加）
├── FilterRuleRow.tsx          ← FE-023 で新規作成
├── SortPopover.tsx            ← 新規: 並べ替え Popover
├── types.ts                   ← SortProperty, ActiveSort 追加
└── index.ts
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `memo()` で SortPopover を最適化
- [ ] `useCallback` でハンドラ安定化

## 依存関係

- Backend 変更不要
- FE-023 と並行実装可能（独立した機能追加）
- FE-025 で全ページの呼び出しを更新

## 完了条件

- [ ] 「並べ替え」ボタンが表示される
- [ ] クリックで SortPopover が開く
- [ ] プロパティ選択 → 昇順/降順が動作
- [ ] 複数ソートが可能
- [ ] 検索が🔍アイコンクリックでトグル展開/非展開
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
