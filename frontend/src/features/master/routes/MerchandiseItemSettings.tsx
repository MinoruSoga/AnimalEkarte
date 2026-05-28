import { useCallback } from "react";
import { DndContext, DragOverlay, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { Plus, ShoppingBag } from "lucide-react";
import { Table, TableBody, TableHeader, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { useGetAllMerchandiseItems, useCreateMerchandiseItem, useUpdateMerchandiseItem, useDeleteMerchandiseItem, useReorderMerchandiseItems } from "../api/merchandise-items";
import type {
  FrontendMerchandiseItem,
  CreateMerchandiseItemRequest,
  UpdateMerchandiseItemRequest,
} from "../api/merchandise-items";
import { MerchandiseRowOverlay } from "../components/MerchandiseRowOverlay";
import { MerchandiseSidePanel } from "../components/MerchandiseSidePanel";
import {
  formatMerchandiseTaxRate,
  MERCHANDISE_CATEGORY_LABELS,
  type MerchandiseFormData,
} from "../components/MerchandiseSidePanelModel";
import {
  buildMerchandiseCreateRequest,
  buildMerchandiseUpdateRequest,
} from "./MerchandiseItemSettingsModel";
import { ResourceMasterMerchandise } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

// ─── Columns ───

const TABLE_COLUMNS = [
  { key: "grip", className: "w-8 px-0" },
  { key: "name", label: "品目名", className: "pl-3" },
  { key: "category", label: "カテゴリ", className: "w-[90px] text-center" },
  { key: "price", label: "単価(税込)", className: "w-[120px] text-right pr-4" },
  { key: "taxRate", label: "税率", className: "w-[80px] text-center" },
  { key: "status", label: "ステータス", className: "w-[90px] text-center" },
  { key: "action", label: "操作", className: "w-[80px] text-right pr-2" },
];

// ─── Page ───

export function MerchandiseItemSettings() {
  const { canCreate, canEdit } = usePermission(ResourceMasterMerchandise);
  const { data } = useGetAllMerchandiseItems();
  const createMutation = useCreateMerchandiseItem();
  const updateMutation = useUpdateMerchandiseItem();
  const deleteMutation = useDeleteMerchandiseItem();
  const reorderMutation = useReorderMerchandiseItems();

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<FrontendMerchandiseItem>({
    data,
    deleteMutation,
    entityLabel: "品目",
    dirtyGuard: dirty,
  });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const {
    orderedItems: sortedItems,
    sensors,
    activeId,
    handleDragStart,
    handleDragCancel,
    handleDragEnd,
    resetOrder,
  } = useSortableList({
    items: crud.filteredItems,
    onReorder: (newIds) => {
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        { onSuccess: resetOrder },
      );
    },
  });

  const { handleSave } = useMasterSave<
    FrontendMerchandiseItem,
    MerchandiseFormData,
    CreateMerchandiseItemRequest,
    UpdateMerchandiseItemRequest
  >({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => (!d.name.trim() ? "品目名は必須です" : null),
    toCreateRequest: buildMerchandiseCreateRequest,
    toUpdateRequest: buildMerchandiseUpdateRequest,
  });

  return (
    <MasterCRUDPage
      title="物販・その他マスタ"
      icon={<ShoppingBag className={`${ICON.page} ${C.text}`} />}
      resource={ResourceMasterMerchandise}
      entityLabel="品目"
      searchPlaceholder="品目名で検索..."
      emptyMessage="品目が登録されていません"
      crud={crud}
      handleSave={handleSave}
      filterProperties={[MASTER_STATUS_FILTER]}
      columns={[]}
      renderRow={() => null}
      renderSidePanel={({ readOnly, ...props }) => (
        <MerchandiseSidePanel key={props.item?.id ?? "new"} {...props} readOnly={readOnly} onDirtyChange={handleDirtyChange} />
      )}
    >
      <div className={STYLE.tableContainer}>
        <div className="flex-1 overflow-auto relative">
          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
            onDragCancel={handleDragCancel}
          >
            <Table>
              <TableHeader className="sticky top-0 z-10">
                <TableRow className={STYLE.tableHeaderRow}>
                  {TABLE_COLUMNS.map((col) => (
                    <TableHead key={col.key} className={`${STYLE.tableHeaderCell} ${col.className}`}>
                      {col.label ?? ""}
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {sortedItems.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className={STYLE.tableEmpty}>
                      品目が登録されていません
                    </TableCell>
                  </TableRow>
                ) : null}
                <SortableContext items={sortedItems.map((m) => m.id)} strategy={verticalListSortingStrategy}>
                  {sortedItems.map((item) => (
                    <SortableDataTableRow key={item.id} id={item.id} onClick={() => crud.handleEdit(item)}>
                      <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
                      <TableCell className={`text-base ${C.text70} text-center`}>
                        {MERCHANDISE_CATEGORY_LABELS[item.category] ?? item.category}
                      </TableCell>
                      <TableCell className={`text-right font-mono text-base ${C.text} pr-4`}>
                        ¥{item.unitPrice.toLocaleString()}
                      </TableCell>
                      <TableCell className={`text-base ${C.text70} text-center`}>
                        {formatMerchandiseTaxRate(item.taxRate)}
                      </TableCell>
                      <TableCell className="text-center">
                        <NotionStatusPill isActive={item.isActive} />
                      </TableCell>
                      <TableCell className="p-0 text-right pr-2">
                        {canEdit ? <RowActionButton onClick={() => crud.handleEdit(item)} /> : null}
                      </TableCell>
                    </SortableDataTableRow>
                  ))}
                </SortableContext>
              </TableBody>
            </Table>
            <DragOverlay dropAnimation={null}>
              {activeId
                ? (() => {
                    const m = sortedItems.find((x) => x.id === activeId);
                    return m ? <MerchandiseRowOverlay item={m} /> : null;
                  })()
                : null}
            </DragOverlay>
          </DndContext>
        </div>
        {canCreate ? (
          <button
            type="button"
            onClick={crud.handleNew}
            className={STYLE.inlineAddBtn}
          >
            <Plus className={`${ICON.xs}`} />
            新しい品目を追加...
          </button>
        ) : null}
      </div>
    </MasterCRUDPage>
  );
}
