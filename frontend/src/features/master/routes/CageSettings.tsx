import { useCallback } from "react";
import { DndContext, DragOverlay, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { Plus, Building2 } from "lucide-react";
import { Table, TableBody, TableHeader, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { useGetAllCages, useCreateCage, useUpdateCage, useDeleteCage, useReorderCages } from "../api/cages";
import type { Cage, CreateCageRequest, UpdateCageRequest } from "../api/cages";
import { CageRowOverlay } from "../components/CageRowOverlay";
import { CageSidePanel } from "../components/CageSidePanel";
import {
  CAGE_SIZE_LABELS,
  CAGE_TYPE_LABELS,
  formatCagePrice,
  type CageFormData,
} from "../components/CageSidePanelModel";
import {
  buildCageCreateRequest,
  buildCageUpdateRequest,
} from "./CageSettingsModel";
import { ResourceMasterHospitalization } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

// ─── Columns (custom table, not DataTable) ───
const TABLE_COLUMNS = [
  { key: "grip", className: "w-8 px-0" },
  { key: "name", label: "ケージ名", className: "pl-3" },
  { key: "type", label: "エリア", className: "w-[100px]" },
  { key: "size", label: "サイズ", className: "w-[90px]" },
  { key: "price", label: "単価(税込)", className: "w-[120px] text-right pr-4" },
  { key: "status", label: "ステータス", className: "w-[90px] text-center" },
  { key: "action", label: "操作", className: "w-[80px] text-right pr-2" },
];

// ─── Page ───
export function CageSettings() {
  const { canCreate, canEdit } = usePermission(ResourceMasterHospitalization);
  const { data } = useGetAllCages();
  const createMutation = useCreateCage();
  const updateMutation = useUpdateCage();
  const deleteMutation = useDeleteCage();
  const reorderMutation = useReorderCages();

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Cage>({ data, deleteMutation, entityLabel: "ケージ", dirtyGuard: dirty });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const { orderedItems: sortedCages, sensors, activeId, handleDragStart, handleDragCancel, handleDragEnd, resetOrder } =
    useSortableList({
      items: crud.filteredItems,
      onReorder: (newIds) => { reorderMutation.mutate({ ids: newIds.map(Number) }, { onSuccess: resetOrder }); },
    });

  const { handleSave } = useMasterSave<Cage, CageFormData, CreateCageRequest, UpdateCageRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "ケージ名は必須です" : null),
    toCreateRequest: buildCageCreateRequest,
    toUpdateRequest: buildCageUpdateRequest,
  });

  // CageSettings uses custom table (not DataTable) for DnD + DragOverlay + bottom "add" button
  return (
    <MasterCRUDPage title="ケージマスタ" icon={<Building2 className={`${ICON.page} ${C.text}`} />} resource={ResourceMasterHospitalization}
      entityLabel="ケージ" searchPlaceholder="ケージ名で検索..." emptyMessage="ケージが登録されていません"
      crud={crud} handleSave={handleSave}
      filterProperties={[MASTER_STATUS_FILTER]}
      columns={[]} renderRow={() => null}
      renderSidePanel={({ readOnly, ...props }) => <CageSidePanel key={props.item?.id ?? "new"} {...props} readOnly={readOnly} onDirtyChange={handleDirtyChange} />}
    >
      <div className={STYLE.tableContainer}>
        <div className="flex-1 overflow-auto relative">
          <DndContext sensors={sensors} collisionDetection={closestCenter}
            onDragStart={handleDragStart} onDragEnd={handleDragEnd} onDragCancel={handleDragCancel}>
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
                {sortedCages.length === 0 ? (
                  <TableRow><TableCell colSpan={7} className={STYLE.tableEmpty}>ケージが登録されていません</TableCell></TableRow>
                ) : null}
                <SortableContext items={sortedCages.map((m) => m.id)} strategy={verticalListSortingStrategy}>
                  {sortedCages.map((item) => (
                    <SortableDataTableRow key={item.id} id={item.id} onClick={() => crud.handleEdit(item)}>
                      <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
                      <TableCell className={`text-base ${C.text70}`}>{CAGE_TYPE_LABELS[item.cageType] || item.cageType}</TableCell>
                      <TableCell className={`text-base ${C.text70}`}>{CAGE_SIZE_LABELS[item.cageSize] || item.cageSize}</TableCell>
                      <TableCell className={`text-right font-mono text-base ${C.text} pr-4`}>{formatCagePrice(item.price)}</TableCell>
                      <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
                      <TableCell className="p-0 text-right pr-2">{canEdit ? <RowActionButton onClick={() => crud.handleEdit(item)} /> : null}</TableCell>
                    </SortableDataTableRow>
                  ))}
                </SortableContext>
              </TableBody>
            </Table>
            <DragOverlay dropAnimation={null}>
              {activeId ? (() => { const m = sortedCages.find((x) => x.id === activeId); return m ? <CageRowOverlay cage={m} /> : null; })() : null}
            </DragOverlay>
          </DndContext>
        </div>
        {canCreate ? (
          <button type="button" onClick={crud.handleNew}
            className={STYLE.inlineAddBtn}>
            <Plus className={`${ICON.xs}`} />新しいケージを追加...
          </button>
        ) : null}
      </div>
    </MasterCRUDPage>
  );
}
