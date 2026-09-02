import { useCallback } from "react";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { Building2 } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { useGetAllCages, useCreateCage, useUpdateCage, useDeleteCage, useReorderCages } from "../api/cages";
import type { Cage, CreateCageRequest, UpdateCageRequest } from "../api/cages";
import { CageSidePanel } from "../components/CageSidePanel";
import { CageSortableTable } from "../components/CageSortableTable";
import type { CageFormData } from "../components/cage-side-panel-model";
import {
  buildCageCreateRequest,
  buildCageUpdateRequest,
} from "./cage-settings-model";
import { ResourceMasterHospitalization } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

// ─── Page ───
export function CageSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterHospitalization);
  const { data } = useGetAllCages();
  const createMutation = useCreateCage();
  const updateMutation = useUpdateCage();
  const deleteMutation = useDeleteCage();
  const reorderMutation = useReorderCages();

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Cage>({
    data,
    deleteMutation,
    entityLabel: "ケージ",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const { orderedItems: sortedCages, sensors, activeId, handleDragStart, handleDragCancel, handleDragEnd, resetOrder } =
    useSortableList({
      items: crud.filteredItems,
      onReorder: (newIds) => {
        if (!canEdit) return;
        reorderMutation.mutate(
          { ids: newIds.map(Number) },
          { onSuccess: resetOrder },
        );
      },
    });

  const { handleSave } = useMasterSave<Cage, CageFormData, CreateCageRequest, UpdateCageRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "ケージ名は必須です" : null),
    toCreateRequest: buildCageCreateRequest,
    toUpdateRequest: buildCageUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  return (
    <>
    <MasterCRUDPage title="ケージマスタ" icon={<Building2 className={`${ICON.page} ${C.text}`} />} resource={ResourceMasterHospitalization}
      entityLabel="ケージ" searchPlaceholder="ケージ名で検索..." emptyMessage="ケージが登録されていません"
      crud={crud} handleSave={handleSave}
      filterProperties={[MASTER_STATUS_FILTER]}
      columns={[]} renderRow={() => null}
      renderSidePanel={({ readOnly, ...props }) => <CageSidePanel key={props.item?.id ?? "new"} {...props} readOnly={readOnly} onDirtyChange={handleDirtyChange} />}
    >
      <CageSortableTable
        items={sortedCages}
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
