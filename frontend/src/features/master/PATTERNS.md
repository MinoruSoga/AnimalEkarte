# マスタ設定ページ実装パターン

マスタ設定ページ（`features/master/routes/`）に共通するUIパターンの仕様書。
新規マスタページを実装する際はこのドキュメントを参照せよ。

> **参照実装**: `DiagnosisSettings.tsx`（基本パターン）/ `MedicineSettings.tsx`（アニメーション付き）

---

## 目次

1. [ページ構成](#1-ページ構成)
2. [テーブル仕様](#2-テーブル仕様)
3. [D&D（ドラッグ&ドロップ）](#3-dndドラッグドロップ)
4. [サイドピーク](#4-サイドピーク)
5. [ステータス表示](#5-ステータス表示)
6. [新規登録ボタン](#6-新規登録ボタン)
7. [PropertyFilter](#7-propertyfilter)
8. [タブ切り替え](#8-タブ切り替え)

---

## 1. ページ構成

```
PageLayout
└── TabsPrimitive.Root
    ├── TabsPrimitive.List      # タブナビゲーション
    └── TabsPrimitive.Content   # タブコンテンツ
        └── XxxTab（コンポーネント）
            ├── PropertyFilter + 新規登録ボタン
            ├── DndContext > SortableContext > DataTable
            └── サイドピーク（isEditing ? <XxxSidePanel /> : null）
```

### PageLayout のタイトル規則

- ページタイトルは `「○○マスタ」` 形式
- 例: `"診断病名マスタ"`, `"薬剤マスタ"`, `"診療種別マスタ"`
- アイコンは Lucide から適切なものを選択

```tsx
<PageLayout
  title="診断病名マスタ"
  icon={<ClipboardList className={`size-5 ${C.text}`} />}
  onBack={() => navigate("/settings")}
  maxWidth="max-w-full"
>
```

---

## 2. テーブル仕様

### 行の高さ

| 要素 | クラス | 高さ |
|------|--------|------|
| テーブル行（`DataTableRow`） | `h-12` | 48px |
| ヘッダ行 | `h-11` | 44px |
| セル垂直パディング | `py-2.5` | 上下10px |

**すべての TableCell に `py-2.5` を付与すること。**

### カラム定義パターン

```typescript
// ─────────────────────────────────────────────────
// Module-level constants（コンポーネント外に定義）
// ─────────────────────────────────────────────────

const COLUMNS = [
  { header: "", className: "w-[32px]" },              // D&Dハンドル（必須・先頭）
  { header: "名称" },                                  // メイン名称（flex-1）
  { header: "備考", className: "w-[240px]" },          // 補足（固定幅）
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];
```

**カラム幅のガイドライン:**

| カラム用途 | 推奨幅 |
|-----------|--------|
| D&Dハンドル | `w-[32px]` |
| 主要名称 | 幅なし（flex-1） |
| 所属カテゴリ | `w-[160px]` |
| 備考・説明 | `w-[240px]`（`truncate` + `max-w-[240px]`） |
| ステータス | `w-[100px]`、`align: "center"` |
| 操作ボタン | `w-[80px]`、`align: "right"` |

### セルのスタイル

```tsx
// ハンドルセル
<TableCell className={`w-[32px] py-2.5 ${C.text20} cursor-grab`} {...listeners}>
  <GripVertical className="size-4" />
</TableCell>

// 主要名称セル（太字）
<TableCell className={`font-medium text-sm ${C.text} py-2.5`}>
  {item.name}
</TableCell>

// 補足テキストセル（薄色・省略あり）
<TableCell className={`text-sm ${C.text70} py-2.5 truncate max-w-[240px]`}>
  {item.description || "-"}
</TableCell>

// ステータスセル
<TableCell className="text-center py-2.5">
  <StatusPill isActive={item.isActive} />
</TableCell>

// 操作ボタンセル
<TableCell className="text-right py-2.5">
  <RowActionButton onClick={onEdit} />
</TableCell>
```

---

## 3. D&D（ドラッグ&ドロップ）

### 使用ライブラリ

- `@dnd-kit/core` — `DndContext`, `closestCenter`
- `@dnd-kit/sortable` — `SortableContext`, `useSortable`, `verticalListSortingStrategy`
- `@dnd-kit/utilities` — `CSS`
- 共有 hook: `@/hooks/use-sortable-list`

### useSortableList

```typescript
// 引数
interface UseSortableListOptions<T extends { id: string }> {
  items: T[];
  onReorder: (newIds: string[]) => void; // ドラッグ完了時の新順序ID配列。ここでAPI呼び出し
}

// 戻り値
interface UseSortableListReturn<T> {
  orderedItems: T[];                    // 楽観的順序適用済みアイテム（これをレンダリングする）
  sensors: ReturnType<typeof useSensors>;
  activeId: string | null;              // ドラッグ中のID
  handleDragStart: (e) => void;
  handleDragEnd: (e) => void;
  handleDragCancel: () => void;
  resetOrder: () => void;               // API失敗時などに楽観的順序をリセット
}
```

**ポイント**: `orderedItems` が楽観的更新を含む最新順序。`items`（APIデータ）を直接レンダリングしてはいけない。

### DndContext 設定

```tsx
const { orderedItems, sensors, handleDragEnd } = useSortableList({
  items: rawData ?? [],
  onReorder: (newIds) => reorderMutation.mutate({ ids: newIds.map(Number) }),
});

// フィルタ後のリストに DnD を適用
const filteredItems = useMemo(() => ..., [orderedItems, searchTerm]);

<DndContext
  sensors={sensors}
  collisionDetection={closestCenter}
  onDragEnd={handleDragEnd}
>
  <SortableContext
    items={filteredItems.map((i) => i.id)}   // string[] であること
    strategy={verticalListSortingStrategy}
  >
    <DataTable
      columns={COLUMNS}
      data={filteredItems}
      renderRow={(item) => (
        <SortableRow
          key={item.id}
          item={item}
          canEdit={canEdit}
          onEdit={() => handleEdit(item)}
        />
      )}
    />
  </SortableContext>
</DndContext>
```

### SortableRow コンポーネント

```tsx
function SortableXxxRow({ item, canEdit, onEdit }: {
  item: Xxx;
  canEdit: boolean;
  onEdit: () => void;
}) {
  return (
    <SortableDataTableRow
      id={item.id}
      dragLabel={`並べ替え: 項目 ${item.name} (ID ${item.id})`}
      dragDisabled={!canEdit}
    >
      <TableCell>
        <DataTableRowButton
          aria-label={`詳細: 項目 ${item.name} (ID ${item.id})`}
          onClick={onEdit}
        >
          {item.name}
        </DataTableRowButton>
      </TableCell>
      {/* 残りのセル */}
    </SortableDataTableRow>
  );
}
```

**注意点:**
- 行全体へ `onClick` を付けない。表示・編集はセル内の native link / button から行う
- `SortableDataTableRow` は native 44px drag buttonへ `attributes` / `listeners` / `setActivatorNodeRef` を集約し、`setNodeRef` は測定対象の `<tr>` に保持する
- 並べ替えはwrite操作なので、edit権限がない場合は必ず `dragDisabled` にする
- `dragLabel` と表示・編集buttonのaccessible nameには、表示名とstable IDを含める

### DragOverlay（高度なパターン）

ドラッグ中にゴースト行を表示したい場合（`MedicineSettings.tsx` 参照）:

```tsx
// useSortableList の activeId を利用
const { orderedItems, sensors, activeId, handleDragStart, handleDragEnd, handleDragCancel } =
  useSortableList({ items, onReorder });

const activeItem = useMemo(
  () => orderedItems.find((i) => i.id === activeId) ?? null,
  [orderedItems, activeId],
);

<DndContext
  sensors={sensors}
  collisionDetection={closestCenter}
  onDragStart={handleDragStart}
  onDragEnd={handleDragEnd}
  onDragCancel={handleDragCancel}
>
  <SortableContext items={...} strategy={verticalListSortingStrategy}>
    {/* rows */}
  </SortableContext>

  <DragOverlay>
    {activeItem !== null ? (
      <table style={{ width: "100%" }}>
        <tbody>
          <XxxRow item={activeItem} isDragOverlay />
        </tbody>
      </table>
    ) : null}
  </DragOverlay>
</DndContext>
```

---

## 4. サイドピーク

### 寸法

| トークン | 値 |
|---------|-----|
| `LAYOUT.sidePeek.width` | `"w-[520px]"` |
| `LAYOUT.sidePeek.widthPx` | `520` |

### 基本構造

```tsx
// isEditing フラグで条件レンダー（&& 禁止、? : null を使う）
{isEditing ? (
  <XxxSidePanel
    key={selectedItem ? String(selectedItem.id) : "new"}  // key でリセット
    item={selectedItem}
    onClose={handleClose}
    onSave={handleSave}
    onDeleteRequest={() => setPendingDelete(selectedItem)}
  />
) : null}
```

### サイドピークの実装

```tsx
function XxxSidePanel({ item, onClose, onSave, onDeleteRequest }) {
  const [formData, setFormData] = useState({
    name: item?.name ?? "",
    isActive: item?.isActive ?? true,
    description: item?.description ?? "",
  });

  return (
    <div className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0`}>

      {/* ── Toolbar ── */}
      <div className={STYLE.sidePeekToolbar}>
        <span className={`text-xs ${C.text35} pl-1 select-none`}>
          {item !== null ? "編集" : "新規作成"}
        </span>
        <div className="flex items-center gap-1">
          {item !== null ? (
            <button
              type="button"
              onClick={onDeleteRequest}
              className={`${STYLE.sidePeekToolbarBtn} cursor-pointer ${C.danger} ${C.hoverBgDanger5}`}
            >
              <Trash2 className="size-4" />
            </button>
          ) : null}
          <button
            type="button"
            onClick={onClose}
            className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
            aria-label="閉じる"
          >
            <X className="size-4" />
          </button>
        </div>
      </div>

      {/* ── Body ── */}
      <div className={STYLE.sidePeekBody}>
        <div className="px-16 pb-8">

          {/* ページアイコン */}
          <div className="pt-4 pb-2">
            <div className={STYLE.pageIcon}>
              <SomeIcon className={LAYOUT.pageIcon.innerIcon} />
            </div>
          </div>

          {/* タイトル入力（30px Bold） */}
          <div className="pb-1 mb-4">
            <input
              type="text"
              className={`w-full bg-transparent ${C.text} placeholder:text-[rgba(55,53,47,0.15)] outline-none border-none p-0`}
              style={{
                fontSize: LAYOUT.pageTitle.fontSize,    // "30px"
                fontWeight: LAYOUT.pageTitle.fontWeight, // 700
                lineHeight: LAYOUT.pageTitle.lineHeight, // "1.2"
              }}
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="無題"
              autoFocus
            />
          </div>

          {/* セクション区切り */}
          <div className={`${STYLE.sectionDivider} mb-1`} />

          {/* プロパティ一覧 */}
          <div className="py-1">
            <PropertyRow label="ステータス">
              <button
                type="button"
                onClick={() => setFormData({ ...formData, isActive: !formData.isActive })}
                className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
              >
                <StatusPill isActive={formData.isActive} />
              </button>
            </PropertyRow>

            <PropertyRow label="備考">
              <PropInput
                value={formData.description}
                onChange={(v) => setFormData({ ...formData, description: v })}
                placeholder="補足情報など"
              />
            </PropertyRow>
          </div>
        </div>
      </div>

      {/* ── Footer ── */}
      <div className={STYLE.sidePeekFooter}>
        <button type="button" onClick={onClose} className={STYLE.sidePeekCancelBtn}>
          キャンセル
        </button>
        <button
          type="button"
          onClick={() => onSave(formData)}
          className={`px-5 py-[7px] text-base text-white ${C.bgBrand} ${C.hoverBgBrand} rounded-full transition-colors cursor-pointer ${STYLE.pillShadow}`}
        >
          保存
        </button>
      </div>
    </div>
  );
}
```

### PropertyRow

Notion スタイルのキーバリュー行。ラベル幅 `140px` 固定。

```tsx
function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px]`}>
      <div className={`w-[140px] shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}>
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
}
```

### PropInput（プロパティ行用インライン入力）

```tsx
function PropInput({ value, onChange, placeholder }: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
}) {
  return (
    <input
      type="text"
      className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder ?? "空"}
    />
  );
}
```

### サイドピークのトークン一覧

トークンの実体は `frontend/src/lib/design-tokens.ts` が SSOT。色を直書きせず、常に `C.*` / `STYLE.*` 経由で参照する。

| トークン | 定義（`design-tokens.ts` の合成元） |
|---------|------------|
| `STYLE.sidePeekPanel` | `flex flex-col h-full overflow-y-auto bg-white border-l ${C.borderLight} shadow-panel`（FE9-2: design-system.md §5.1 shadow-panel トークンへ移行。値は同一） |
| `STYLE.sidePeekToolbar` | `flex items-center justify-between h-[48px] px-3 shrink-0` |
| `STYLE.sidePeekToolbarBtn` | `size-9 flex items-center justify-center rounded-[3px] ${C.text45} ${C.hoverBgMedium} transition-colors` |
| `STYLE.sidePeekBody` | `flex-1 overflow-y-auto` |
| `STYLE.sidePeekFooter` | `flex items-center justify-end gap-2 px-4 py-3 border-t ${C.borderLight} shrink-0` |
| `STYLE.sidePeekCancelBtn` | `px-4 py-[7px] text-base ${C.text65} ${C.hoverBgLight} rounded-[3px] transition-colors cursor-pointer` |
| （保存ボタン） | 共通トークン化されていない。`px-5 py-[7px] text-base text-white ${C.bgBrand} ${C.hoverBgBrand} rounded-full transition-colors cursor-pointer ${STYLE.pillShadow}` を直書きする（`ShiftTemplateSettingsParts.tsx` の実例を参照。FE5-3 で未使用だった `STYLE.sidePeekSaveBtn` を削除済み） |
| `STYLE.pageIcon` | `size-[38px] flex items-center justify-center rounded-[3px] ${C.bgPage} ${C.text45}` |
| `LAYOUT.pageIcon.innerIcon` | `"size-5"` |

### アニメーション付きサイドピーク（高度なパターン）

`motion/react` で開閉アニメーションを付ける場合（`MedicineSettings.tsx` 参照）:

```tsx
import { AnimatePresence, motion } from "motion/react";

const panelDuration = useReducedMotion() ? 0 : 0.2;

<AnimatePresence>
  {isEditing ? (
    <motion.div
      key="side-peek"
      initial={{ width: 0, opacity: 0 }}
      animate={{ width: LAYOUT.sidePeek.widthPx, opacity: 1 }}
      exit={{ width: 0, opacity: 0 }}
      transition={{ duration: panelDuration, ease: [0.25, 0.1, 0.25, 1] }}
      className="shrink-0 min-h-0 overflow-hidden"
    >
      <XxxSidePanel ... />
    </motion.div>
  ) : null}
</AnimatePresence>
```

---

## 5. ステータス表示

マスタページのステータスは **StatusPill** を使う。`StatusBadge` は使わない。

```tsx
const STATUS_CONFIG = {
  active: {
    dot:   C.bgBrandDot,     // ブランドteal ドット
    label: "有効",
    bg:    C.bgBrandLight,   // 薄teal背景
    text:  C.textBrandDark,  // 濃teal テキスト
  },
  inactive: {
    dot:   C.bgPrimary10,    // グレードット
    label: "無効",
    bg:    C.bgInactive,     // グレー背景
    text:  C.text60,
  },
} as const;

function StatusPill({ isActive }: { isActive: boolean }) {
  const cfg = STATUS_CONFIG[isActive ? "active" : "inactive"];
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-sm text-xs ${cfg.bg} ${cfg.text}`}
    >
      <span className={`size-[7px] rounded-full ${cfg.dot}`} />
      {cfg.label}
    </span>
  );
}
```

**各ファイルにコピーして定義する**（共通コンポーネント化しない理由: マスタページ固有のデザイントークンを直書きしているため）。

---

## 6. 新規登録ボタン

テキストリンクスタイル。solid ボタン（`PrimaryButton`）は使わない。

```tsx
<button
  type="button"
  onClick={handleCreate}
  className={`inline-flex items-center gap-1 text-sm font-medium ${C.textBrand} ${C.hoverTextBrand} cursor-pointer transition-colors`}
>
  <Plus className="size-4" />
  新規登録
</button>
```

**配置**: PropertyFilter の右端（flex レイアウトで末尾に）

```tsx
<div className="flex items-center gap-3">
  <div className="flex-1 min-w-0">
    <PropertyFilter ... />
  </div>
  {/* 新規登録ボタン */}
  <button type="button" onClick={handleCreate} className="...">
    <Plus className="size-4" />
    新規登録
  </button>
</div>
```

---

## 7. PropertyFilter

```tsx
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";

<PropertyFilter
  properties={FILTER_PROPERTIES}
  activeFilters={activeFilters}
  onFilterChange={setActiveFilters}
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  searchPlaceholder="カテゴリ名で検索..."
  count={filteredItems.length}
/>
```

| Props | 型 | 説明 |
|-------|-----|------|
| `properties` | `FilterProperty[]` | フィルタ可能なプロパティ定義 |
| `activeFilters` | `ActiveFilter[]` | 現在アクティブなフィルタ |
| `onFilterChange` | `(filters: ActiveFilter[]) => void` | フィルタ変更ハンドラ |
| `filterLogic` | `"and" \| "or"?` | フィルタ論理（デフォルト: `"and"`） |
| `onFilterLogicChange` | `(logic: FilterLogic) => void?` | 論理切替ハンドラ |
| `searchTerm` | `string?` | 検索語（state） |
| `onSearchChange` | `(v: string) => void?` | 検索変更ハンドラ |
| `searchPlaceholder` | `string?` | 入力プレースホルダー |
| `count` | `number?` | 結果件数（`N 件` 表示） |
| `sortProperties` | `SortProperty[]?` | ソート可能なプロパティ |
| `activeSorts` | `ActiveSort[]?` | アクティブなソート |
| `onSortChange` | `(sorts: ActiveSort[]) => void?` | ソート変更ハンドラ |

**注意**: `count` は `filteredItems.length`（フィルタ後）を渡す。全件数は渡さない。

---

## 8. タブ切り替え

URL の `?tab=xxx` で状態を管理する（ブラウザバック・リロード対応）。

```tsx
import * as TabsPrimitive from "@radix-ui/react-tabs";
import { useSearchParams } from "react-router";

const TABS = [
  { value: "category", label: "カテゴリ" },
  { value: "item",     label: "アイテム" },
] as const;

export function XxxSettings() {
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") ?? "category"; // デフォルトは先頭タブ

  return (
    <PageLayout title="XXXマスタ" ...>
      <TabsPrimitive.Root
        value={activeTab}
        onValueChange={(tab) => setSearchParams({ tab })}
        className="flex flex-col gap-4"
      >
        <TabsPrimitive.List className={`flex h-9 border-b ${C.borderLight} gap-0`}>
          {TABS.map((tab) => (
            <TabsPrimitive.Trigger
              key={tab.value}
              value={tab.value}
              className={`h-9 border-b-2 border-b-transparent px-4 text-sm ${C.text60} outline-none transition-colors cursor-pointer
                ${C.dataActiveBorderB} ${C.dataActiveText} data-[state=active]:font-medium`}
            >
              {tab.label}
            </TabsPrimitive.Trigger>
          ))}
        </TabsPrimitive.List>

        <TabsPrimitive.Content value="category" className="mt-4">
          <CategoryTab />
        </TabsPrimitive.Content>
        <TabsPrimitive.Content value="item" className="mt-4">
          <ItemTab />
        </TabsPrimitive.Content>
      </TabsPrimitive.Root>
    </PageLayout>
  );
}
```

**タブ値の命名規則**: `snake_case`（例: `diagnosis_category`, `diagnosis_name`）

---

## チェックリスト

新規マスタページ実装時の確認事項:

- [ ] ページタイトルは `「○○マスタ」` 形式
- [ ] カラム定義はモジュールレベル定数（`const COLUMNS = [...]`）
- [ ] D&Dハンドルは先頭カラム `w-[32px]`、`GripVertical`、`{...listeners}` をセルに
- [ ] `orderedItems`（`useSortableList` 戻り値）をレンダリング（`items` 直接不可）
- [ ] テーブルセルに `py-2.5` を付与
- [ ] ステータスは `StatusPill`（`StatusBadge` 禁止）
- [ ] 新規登録ボタンはテキストリンクスタイル（`PrimaryButton` 禁止）
- [ ] サイドピーク幅は `LAYOUT.sidePeek.width`（`w-[520px]`）
- [ ] サイドピークは `isEditing ? <Panel /> : null` で条件レンダー（`&&` 禁止）
- [ ] `key` プロップでサイドピークをリセット（新規 vs 編集）
- [ ] URL `?tab=xxx` でタブ状態を管理（`useSearchParams`）
- [ ] `PropertyFilter` の `count` にはフィルタ後の件数を渡す
