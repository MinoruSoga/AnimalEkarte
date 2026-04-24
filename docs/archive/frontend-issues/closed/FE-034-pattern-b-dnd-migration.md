# FE-034: パターンB 3ページ（DnD対応）を MasterCRUDPage + useMasterSave に移行

**Status**: Open
**Priority**: High
**Affects**: AnimalSpeciesSettings, ServiceTypeSettings, CageSettings
**Date Created**: 2026-03-18
**Related**: TASK-008, FE-031, FE-032, FE-033

## Summary

DnD（ドラッグ&ドロップ）対応の3ページを `useMasterSave` + `MasterCRUDPage` に移行する。DnD 部分は `MasterCRUDPage` の `children` prop で注入し、標準の DataTable レンダリングを上書きする。

## 対象ページと期待行数

| ページ | 現在 | 移行後（目安） | 削減率 |
|--------|------|--------------|--------|
| AnimalSpeciesSettings.tsx | 223行 | ~70行 | 69% |
| ServiceTypeSettings.tsx | 299行 | ~100行 | 67% |
| CageSettings.tsx | 420行 | ~130行 | 69% |
| **合計** | **942行** | **~300行** | **~68%** |

## 現状のコード

DnD ページは Standard CRUD + 以下の追加要素:

```typescript
// frontend/src/features/master/routes/AnimalSpeciesSettings.tsx

// 追加: DnD imports
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/useSortableList";

// 追加: reorder mutation
const reorderMutation = useReorderAnimalSpecies();

// 追加: useSortableList hook
const { orderedItems, sensors, handleDragEnd } = useSortableList({
  items: crud.filteredItems,
  onReorder: (newIds) => { reorderMutation.mutate({ ids: newIds.map(Number) }); },
});

// DataTable を DndContext + SortableContext で wrap
<DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
  <SortableContext items={orderedItems.map((i) => i.id)} strategy={verticalListSortingStrategy}>
    <DataTable columns={COLUMNS} data={orderedItems} renderRow={...}>
      {/* SortableDataTableRow を使用 */}
    </DataTable>
  </SortableContext>
</DndContext>
```

## 必要な変更

### MasterCRUDPage の `children` prop を活用

`MasterCRUDPage` の `children` prop が提供されている場合、デフォルトの DataTable の代わりに children をレンダリングする（FE-032 で実装済みの設計）。

### 移行パターン（AnimalSpeciesSettings）

**After (~70行):**

```typescript
export function AnimalSpeciesSettings() {
  const { data } = useGetAnimalSpecies();
  const createMutation = useCreateAnimalSpecies();
  const updateMutation = useUpdateAnimalSpecies();
  const deleteMutation = useDeleteAnimalSpecies();
  const reorderMutation = useReorderAnimalSpecies();

  const crud = useMasterCRUD<AnimalSpecies>({ data, deleteMutation, entityLabel: "動物種類" });

  const { orderedItems, sensors, handleDragEnd } = useSortableList({
    items: crud.filteredItems,
    onReorder: (newIds) => { reorderMutation.mutate({ ids: newIds.map(Number) }); },
  });

  const { handleSave } = useMasterSave({
    crud, createMutation, updateMutation,
    validate: (d: AnimalSpeciesFormData) => (!d.name.trim() ? "動物種類名は必須です" : null),
    toCreateRequest: (d) => ({ name: d.name, is_active: true, sort_order: 0 }),
    toUpdateRequest: (d): UpdateAnimalSpeciesRequest => ({ name: d.name, is_active: d.isActive }),
  });

  return (
    <MasterCRUDPage
      title="動物種類マスタ" icon={<PawPrint className="size-5 text-[#37352F]" />}
      entityLabel="動物種類" searchPlaceholder="動物種類名で検索..."
      emptyMessage="動物種類が登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      renderRow={(item, onEdit) => (
        <SortableDataTableRow key={item.id} id={item.id} onClick={() => onEdit(item)}>
          <TableCell className={`font-medium text-sm ${C.text}`}>{item.name}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right"><RowActionButton onClick={() => onEdit(item)} /></TableCell>
        </SortableDataTableRow>
      )}
      renderSidePanel={(props) => <AnimalSpeciesSidePanel key={props.item?.id ?? "new"} {...props} />}
    >
      {/* children: DnD ラッパー */}
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <SortableContext items={orderedItems.map((i) => i.id)} strategy={verticalListSortingStrategy}>
          <DataTable columns={COLUMNS} data={orderedItems} emptyMessage="動物種類が登録されていません"
            renderRow={(item) => (
              <SortableDataTableRow key={item.id} id={item.id} onClick={() => crud.handleEdit(item)}>
                <TableCell className={`font-medium text-sm ${C.text}`}>{item.name}</TableCell>
                <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
                <TableCell className="p-0 text-right"><RowActionButton onClick={() => crud.handleEdit(item)} /></TableCell>
              </SortableDataTableRow>
            )}
          />
        </SortableContext>
      </DndContext>
    </MasterCRUDPage>
  );
}
```

### ページ固有の注意点

#### ServiceTypeSettings
- `resetOrderRef` + `useEffect` で reorder 失敗時のロールバック
- `handleDragStart`, `handleDragCancel` も使用
- SidePanel: color（カラーピッカー）+ description
- テーブル行: カラードット + name 表示

#### CageSettings
- 最も複雑（420行）: DnD + size/type/price のリッチフォーム
- SidePanel: cageType（Select）、size（Select）、price（MoneyInput）、description
- テーブル列: タイプ、サイズ、料金、ステータス
- hoisted JSX: `CAGE_TYPE_SELECT_ITEMS`, `CAGE_SIZE_SELECT_ITEMS`

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useMasterSave` 経由）
- [ ] `useCallback` で安定化（`useMasterSave` 内部 + `handleReorder`）
- [ ] `memo()` で SidePanel メモ化
- [ ] `useSortableList` で DnD 状態管理
- [ ] Vercel Best Practices 全パターン準拠

## 依存関係

- FE-031（`useMasterSave` hook）が完了している必要がある
- FE-032（`MasterCRUDPage` コンポーネント）が完了している必要がある
- FE-033 と並行可（独立）

## 完了条件

- [ ] 3ページすべて移行完了
- [ ] DnD 機能が既存と同一の動作
- [ ] 並び替え + エラー時ロールバックが正常動作
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
- [ ] UI が既存と同一の動作（見た目・操作フロー変更なし）
