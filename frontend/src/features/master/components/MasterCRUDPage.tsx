import { memo, type ReactNode } from "react";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { MasterListPage } from "../components/MasterListPage";
import { usePermission } from "@/hooks/use-permission";
import type { UseMasterCRUDReturn } from "../hooks/use-master-crud";
import type { FilterProperty, SortProperty } from "@/components/shared/PropertyFilter/types";
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

interface SidePanelRenderProps<T, TForm = Record<string, unknown>> {
  item: T | null;
  onClose: () => void;
  onSave: (data: TForm) => Promise<boolean> | boolean | void;
  onDeleteRequest: ((item: T) => void) | undefined;
  /** BUG-158: true の場合、保存・削除ボタンを非表示にする */
  readOnly?: boolean;
}

interface MasterCRUDPageProps<T extends MasterEntity, TForm = Record<string, unknown>> {
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
  handleSave: (data: TForm) => Promise<boolean> | boolean | void;

  /** Table columns */
  columns: Column[];
  /** Table row renderer. Called with (item, onEdit, canEdit). */
  renderRow: (item: T, onEdit: (item: T) => void, canEdit: boolean) => ReactNode;

  /** SidePanel render prop */
  renderSidePanel: (props: SidePanelRenderProps<T, TForm>) => ReactNode;

  /** Override default DataTable with custom content (e.g. DnD wrapper) */
  children?: ReactNode;

  /** Field key used for delete dialog display name. Defaults to "name". */
  deleteNameField?: string;

  /** Custom delete description (overrides auto-generated one) */
  deleteDescription?: string;

  /** PropertyFilter filter properties */
  filterProperties?: FilterProperty[];

  /** PropertyFilter sort properties */
  sortProperties?: SortProperty[];
}

// ─────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────

export const MasterCRUDPage = memo(function MasterCRUDPage<T extends MasterEntity, TForm = Record<string, unknown>>({
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
}: MasterCRUDPageProps<T, TForm>) {
  // BUG-158: action-specific 権限で保存・削除ボタンの表示を制御
  // FE6-2: resource は任意。フック呼び出し順序維持のための sentinel（"" は未定義扱い）。
  const { canCreate, canEdit, canDelete } = usePermission((resource ?? "") as Resource);

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
              onDeleteRequest: canDelete ? crud.setPendingDelete : undefined,
              readOnly: crud.editTarget === "new" ? canCreate !== true : canEdit !== true,
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
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={columns}
          data={crud.filteredItems}
          emptyMessage={emptyMessage}
          renderRow={(item) => renderRow(item, crud.handleEdit, canEdit)}
        />
      )}
    </MasterListPage>
  );
}) as <T extends MasterEntity, TForm = Record<string, unknown>>(props: MasterCRUDPageProps<T, TForm>) => ReactNode;
