import { useCallback } from "react";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { ShoppingBag } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
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
import { MerchandiseSidePanel } from "../components/MerchandiseSidePanel";
import { MerchandiseSortableTable } from "../components/MerchandiseSortableTable";
import type { MerchandiseFormData } from "../components/merchandise-side-panel-model";
import {
  buildMerchandiseCreateRequest,
  buildMerchandiseUpdateRequest,
} from "./merchandise-item-settings-model";
import { ResourceMasterMerchandise } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

// ─── Page ───

export function MerchandiseItemSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMerchandise);
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
    permissions: { canDelete },
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
      if (!canEdit) return;
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
    permissions: { canCreate, canEdit },
  });

  return (
    <>
    <MasterCRUDPage
      title="商品マスタ"
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
      <MerchandiseSortableTable
        items={sortedItems}
        sensors={sensors}
        activeId={activeId}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
        onDragCancel={handleDragCancel}
        canEdit={canEdit}
        onEdit={crud.handleEdit}
      />
    </MasterCRUDPage>
    {dirty.discardDialog}
    </>
  );
}
