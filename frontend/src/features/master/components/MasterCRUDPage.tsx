import { memo, type ReactNode } from "react";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { MasterListPage } from "@/features/master/components/MasterListPage";
import { usePermission } from "@/features/auth/hooks/use-permission";
import type { UseMasterCRUDReturn } from "@/features/master/hooks/use-master-crud";
import type { FilterProperty, SortProperty } from "@/components/shared/NotionFilter/types";
import type { Resource } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface MasterEntity {
  id: string;
}

interface Column {
  header: ReactNode;
  className?: string;
  align?: "left" | "center" | "right";
}

interface SidePanelRenderProps<T> {
  item: T | null;
  onClose: () => void;
  onSave: (data: never) => void;
  onDeleteRequest: (item: T) => void;
  /** BUG-158: true の場合、保存・削除ボタンを非表示にする */
  readOnly?: boolean;
}

interface MasterCRUDPageProps<T extends MasterEntity> {
  /** Page title (e.g. "役職マスタ") */
  title: string;
  /** Page icon element */
  icon: ReactNode;
  /** 権限バッジ表示用リソース */
  resource?: Resource;
  /** Entity label for delete dialog (e.g. "役職") */
  entityLabel: string;
  /** Search placeholder */
  searchPlaceholder: string;
  /** Empty table message */
  emptyMessage: string;

  /** CRUD state from useMasterCRUD */
  crud: UseMasterCRUDReturn<T>;
  /** Save handler from useMasterSave */
  handleSave: (data: never) => void;

  /** Table columns */
  columns: Column[];
  /** Table row renderer. Called with (item, onEdit). */
  renderRow: (item: T, onEdit: (item: T) => void) => ReactNode;

  /** SidePanel render prop */
  renderSidePanel: (props: SidePanelRenderProps<T>) => ReactNode;

  /** Override default DataTable with custom content (e.g. DnD wrapper) */
  children?: ReactNode;

  /** Field key used for delete dialog display name. Defaults to "name". */
  deleteNameField?: string;

  /** Custom delete description (overrides auto-generated one) */
  deleteDescription?: string;

  /** NotionFilter filter properties */
  filterProperties?: FilterProperty[];

  /** NotionFilter sort properties */
  sortProperties?: SortProperty[];
}

// ─────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────

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
  deleteNameField = "name",
  deleteDescription,
  filterProperties,
  sortProperties,
  resource,
}: MasterCRUDPageProps<T>) {
  // BUG-158: edit/delete 権限で保存・削除ボタンの表示を制御
  const { canEdit, canDelete } = usePermission(resource ?? "");

  const deleteName = crud.pendingDelete
    ? String((crud.pendingDelete as Record<string, unknown>)[deleteNameField] ?? "")
    : "";

  return (
    <MasterListPage
      title={title}
      icon={icon}
      resource={resource}
      searchTerm={crud.searchTerm}
      onSearchChange={crud.setSearchTerm}
      searchPlaceholder={searchPlaceholder}
      count={crud.filteredItems.length}
      onNew={crud.handleNew}
      filterProperties={filterProperties}
      activeFilters={crud.activeFilters}
      onFilterChange={crud.setActiveFilters}
      sortProperties={sortProperties}
      activeSorts={crud.activeSorts}
      onSortChange={crud.handleSortChange}
      sidePanel={
        crud.isEditing
          ? renderSidePanel({
              item: crud.panelItem,
              onClose: crud.handleClose,
              onSave: handleSave,
              onDeleteRequest: crud.setPendingDelete,
              readOnly: !canEdit,
            })
          : null
      }
      deleteOpen={crud.pendingDelete !== null}
      deleteTitle={`${entityLabel}を削除しますか？`}
      deleteDescription={
        deleteDescription ??
        `「${deleteName}」を削除します。この操作は取り消せません。`
      }
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
