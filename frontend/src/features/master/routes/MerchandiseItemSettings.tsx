import { useCallback } from "react";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";

import { ErrorFallback, LoadingFallback } from "@/components/shared/DataStates";
import { LAYOUT } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import {
  useGetAllMerchandiseItems,
  useCreateMerchandiseItem,
  useUpdateMerchandiseItem,
  useDeleteMerchandiseItem,
  useReorderMerchandiseItems,
} from "../api/merchandise-items";
import type {
  FrontendMerchandiseItem,
  CreateMerchandiseItemRequest,
  UpdateMerchandiseItemRequest,
} from "../api/merchandise-items";
import { MerchandiseSidePanel } from "../components/MerchandiseSidePanel";
import { MerchandiseSortableTable } from "../components/MerchandiseSortableTable";
import { moveItemInOrder } from "../lib/merchandise-reorder";
import type { MerchandiseFormData } from "../lib/merchandise-side-panel-model";
import {
  buildMerchandiseCreateRequest,
  buildMerchandiseUpdateRequest,
} from "./merchandise-item-settings-model";
import { ResourceMasterMerchandise } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

// ─── Page ───

export function MerchandiseItemSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMerchandise);
  const { data, isPending, isError } = useGetAllMerchandiseItems();
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
  // 部分リストを再採番すると非表示行と sort_order が衝突するため、絞り込み中は並べ替えない。
  const canReorder = canEdit && crud.filteredItems.length === (data?.length ?? 0);
  const handleDirtyChange = useCallback(
    (d: boolean) => {
      if (d) dirty.markDirty();
      else dirty.markClean();
    },
    [dirty],
  );

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
      // PR665: 並べ替え API 実行中はドラッグ由来のコールバックも無視する。
      if (!canReorder || reorderMutation.isPending) return;
      reorderMutation.mutate({ ids: newIds.map(Number) }, { onSuccess: resetOrder });
    },
  });

  // PR665 / EMR-401 C2-5: 50 行ページングを維持したままページをまたげるよう、
  // 明示的な前/後移動ボタンから全件順序の並べ替えをドラッグと同一経路で行う。
  const handleMoveItem = useCallback(
    (item: FrontendMerchandiseItem, direction: "up" | "down") => {
      // PR665: 並べ替え API 実行中は追加クリックを無視する。
      if (!canReorder || reorderMutation.isPending) return;
      const ids = sortedItems.map((i) => i.id);
      const newIds = moveItemInOrder(ids, item.id, direction);
      if (!newIds) return;
      reorderMutation.mutate({ ids: newIds.map(Number) }, { onSuccess: resetOrder });
    },
    [canReorder, sortedItems, reorderMutation, resetOrder],
  );

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
        title="物販・フードマスタ"
        resource={ResourceMasterMerchandise}
        entityLabel="品目"
        searchPlaceholder="品目名で検索..."
        emptyMessage="品目が登録されていません"
        crud={crud}
        handleSave={handleSave}
        filterProperties={[MASTER_STATUS_FILTER]}
        columns={[]}
        maxWidth={LAYOUT.pageContentMaxWidth.full}
        renderRow={() => null}
        renderSidePanel={({ readOnly, ...props }) => (
          <MerchandiseSidePanel
            key={props.item?.id ?? "new"}
            {...props}
            readOnly={readOnly}
            onDirtyChange={handleDirtyChange}
          />
        )}
      >
        {isError ? (
          <ErrorFallback message="品目の取得に失敗しました" />
        ) : isPending ? (
          <LoadingFallback />
        ) : (
          <MerchandiseSortableTable
            items={sortedItems}
            sensors={sensors}
            activeId={activeId}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
            onDragCancel={handleDragCancel}
            canEdit={canEdit}
            canReorder={canReorder}
            onEdit={crud.handleEdit}
            onMoveItem={handleMoveItem}
            isReordering={reorderMutation.isPending}
          />
        )}
      </MasterCRUDPage>
      {dirty.discardDialog}
    </>
  );
}
