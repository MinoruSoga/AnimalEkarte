# FE-032: MasterCRUDPage 高レベルラッパーコンポーネント

**Status**: Open
**Priority**: High
**Affects**: 全マスタ設定ページ（9ページ）
**Date Created**: 2026-03-18
**Related**: TASK-008, FE-031

## Summary

`MasterListPage` + DataTable + SidePanel + DeleteDialog の組み立てを1つのコンポーネントに統合する。各ページで繰り返される ~40行の props ボイラープレートを排除する。

## 現状のコード

**9ページすべてで以下のパターンが重複:**

```typescript
// frontend/src/features/master/routes/JobTitleSettings.tsx:164-213
return (
  <MasterListPage
    title="役職マスタ"
    icon={<Briefcase className="size-5 text-[#37352F]" />}
    searchTerm={crud.searchTerm}
    onSearchChange={crud.setSearchTerm}
    searchPlaceholder="役職名で検索..."
    count={crud.filteredItems.length}
    onNew={crud.handleNew}
    sidePanel={
      crud.isEditing ? (
        <JobTitleSidePanel
          key={crud.panelItem ? String(crud.panelItem.id) : "new-job-title"}
          item={crud.panelItem}
          onClose={crud.handleClose}
          onSave={handleSave}
          onDeleteRequest={crud.setPendingDelete}
        />
      ) : null
    }
    deleteOpen={crud.pendingDelete !== null}
    deleteTitle="役職を削除しますか？"
    deleteDescription={`「${crud.pendingDelete?.name}」を削除します。この操作は取り消せません。`}
    onDeleteConfirm={crud.handleDeleteConfirm}
    onDeleteCancel={crud.handleDeleteCancel}
  >
    <DataTable columns={COLUMNS} data={crud.filteredItems} emptyMessage="..." renderRow={...} />
  </MasterListPage>
);
```

**ページ間の差分（これだけが異なる）:**
1. メタデータ: title, icon, searchPlaceholder, entityLabel（deleteTitle/Description 生成に使用）
2. SidePanel の内部フォーム（`renderSidePanel` で注入）
3. DataTable の `columns` と `renderRow`
4. DnD 対応の有無

## 必要な変更

### 1. コンポーネント作成

```typescript
// frontend/src/features/master/components/MasterCRUDPage.tsx

import { memo, type ReactNode } from "react";
import { TableCell } from "@/components/ui/table";
import { DataTable, type Column } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { MasterListPage } from "@/features/master/components/MasterListPage";
import { C } from "@/lib/design-tokens";
import type { UseMasterCRUDReturn } from "@/features/master/hooks/use-master-crud";

interface MasterEntity {
  id: string;
  name: string;
  isActive: boolean;
}

// SidePanel render prop に渡す props
interface SidePanelRenderProps<T> {
  item: T | null;
  onClose: () => void;
  onSave: (data: unknown) => void;
  onDeleteRequest: (item: T) => void;
}

interface MasterCRUDPageProps<T extends MasterEntity> {
  // メタデータ
  title: string;
  icon: ReactNode;
  entityLabel: string;
  searchPlaceholder: string;
  emptyMessage: string;

  // CRUD state
  crud: UseMasterCRUDReturn<T>;
  handleSave: (data: never) => void;

  // テーブル
  columns: Column[];
  renderRow: (item: T, onEdit: (item: T) => void) => ReactNode;

  // サイドパネル
  renderSidePanel: (props: SidePanelRenderProps<T>) => ReactNode;

  // DnD（オプション）
  children?: ReactNode;

  // 削除ダイアログのカスタムテキスト（オプション）
  deleteNameField?: keyof T;
}

export const MasterCRUDPage = memo(function MasterCRUDPage<T extends MasterEntity>({
  title,
  icon,
  entityLabel,
  searchPlaceholder,
  emptyMessage,
  crud,
  handleSave,
  columns,
  renderRow,
  renderSidePanel,
  children,
  deleteNameField = "name" as keyof T,
}: MasterCRUDPageProps<T>) {
  const deleteName = crud.pendingDelete
    ? String(crud.pendingDelete[deleteNameField])
    : "";

  return (
    <MasterListPage
      title={title}
      icon={icon}
      searchTerm={crud.searchTerm}
      onSearchChange={crud.setSearchTerm}
      searchPlaceholder={searchPlaceholder}
      count={crud.filteredItems.length}
      onNew={crud.handleNew}
      sidePanel={
        crud.isEditing
          ? renderSidePanel({
              item: crud.panelItem,
              onClose: crud.handleClose,
              onSave: handleSave,
              onDeleteRequest: crud.setPendingDelete,
            })
          : null
      }
      deleteOpen={crud.pendingDelete !== null}
      deleteTitle={`${entityLabel}を削除しますか？`}
      deleteDescription={`「${deleteName}」を削除します。この操作は取り消せません。`}
      onDeleteConfirm={crud.handleDeleteConfirm}
      onDeleteCancel={crud.handleDeleteCancel}
    >
      {children ?? (
        <DataTable
          columns={columns}
          data={crud.filteredItems}
          emptyMessage={emptyMessage}
          renderRow={(item) => renderRow(item, crud.handleEdit)}
        />
      )}
    </MasterListPage>
  );
}) as <T extends MasterEntity>(props: MasterCRUDPageProps<T>) => ReactNode;
```

### 2. INPUT_CLASS 定数の共有化

5ページで重複定義されている `INPUT_CLASS` を共有定数に移動:

```typescript
// frontend/src/features/master/constants/styles.ts
import { C } from "@/lib/design-tokens";

export const MASTER_INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`;
```

## 使用例（移行後の JobTitleSettings）

```typescript
export function JobTitleSettings() {
  const { data } = useGetAllJobTitles();
  const createMutation = useCreateJobTitle();
  const updateMutation = useUpdateJobTitle();
  const deleteMutation = useDeleteJobTitle();

  const crud = useMasterCRUD<JobTitle>({ data, deleteMutation, entityLabel: "役職" });

  const { handleSave } = useMasterSave({ crud, createMutation, updateMutation, validate, toCreateRequest, toUpdateRequest });

  return (
    <MasterCRUDPage
      title="役職マスタ"
      icon={<Briefcase className="size-5 text-[#37352F]" />}
      entityLabel="役職"
      searchPlaceholder="役職名で検索..."
      emptyMessage="役職が登録されていません"
      crud={crud}
      handleSave={handleSave}
      columns={COLUMNS}
      renderRow={(item, onEdit) => (
        <DataTableRow key={item.id} onClick={() => onEdit(item)}>
          <TableCell className={`font-medium text-sm ${C.text}`}>{item.name}</TableCell>
          <TableCell className={`text-sm ${C.text}`}>{item.description || "-"}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right"><RowActionButton onClick={() => onEdit(item)} /></TableCell>
        </DataTableRow>
      )}
      renderSidePanel={({ item, onClose, onSave, onDeleteRequest }) => (
        <JobTitleSidePanel key={item ? item.id : "new"} item={item} onClose={onClose} onSave={onSave} onDeleteRequest={onDeleteRequest} />
      )}
    />
  );
}
```

## プロジェクトルール遵守チェック

- [x] `any` 型なし — ジェネリクスで型安全
- [x] `FC` / `forwardRef` なし — `memo()` + 関数宣言
- [x] barrel index 経由 import なし
- [x] 条件レンダー `? ... : null`（`&&` 禁止）
- [x] `memo()` で再レンダー最適化（`rerender-memo`）
- [x] 静的 JSX はモジュール定数に巻き上げ不要（props 駆動）

## 完了条件

- [ ] `MasterCRUDPage.tsx` 作成
- [ ] `MASTER_INPUT_CLASS` 共有定数作成
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend npm run lint` パス）
- [ ] 既存の9ページから呼び出し可能（FE-033/034 で移行）
