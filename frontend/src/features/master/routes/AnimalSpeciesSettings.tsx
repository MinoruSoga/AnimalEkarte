import { useCallback } from "react";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import PawPrint from "lucide-react/dist/esm/icons/paw-print";
import { C, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { AnimalSpeciesSidePanel } from "../components/AnimalSpeciesSidePanel";
import type { AnimalSpeciesFormData } from "../components/animal-species-side-panel-model";
import {
  ANIMAL_SPECIES_COLUMNS,
  AnimalSpeciesSortableTable,
} from "../components/AnimalSpeciesSortableTable";
import { usePermission } from "@/hooks/use-permission";
import { useGetAnimalSpecies, useCreateAnimalSpecies, useUpdateAnimalSpecies, useDeleteAnimalSpecies, useReorderAnimalSpecies } from "../api/animal-species";
import type { AnimalSpecies, CreateAnimalSpeciesRequest, UpdateAnimalSpeciesRequest } from "../api/animal-species";
import {
  buildAnimalSpeciesCreateRequest,
  buildAnimalSpeciesUpdateRequest,
} from "./animal-species-settings-model";
import { ResourceMasterAnimalSpecies } from "@/types/generated/models";

export function AnimalSpeciesSettings() {
  const { canEdit } = usePermission(ResourceMasterAnimalSpecies);
  const { data } = useGetAnimalSpecies();
  const createMutation = useCreateAnimalSpecies();
  const updateMutation = useUpdateAnimalSpecies();
  const deleteMutation = useDeleteAnimalSpecies();
  const reorderMutation = useReorderAnimalSpecies();

  // BUG-380: 未保存変更の破棄確認 + beforeunload ガード
  const dirty = useSidePeekDirty();

  const crud = useMasterCRUD<AnimalSpecies>({ data, deleteMutation, entityLabel: "動物種類", dirtyGuard: dirty });

  const handleDirtyChange = useCallback(
    (d: boolean) => {
      if (d) dirty.markDirty();
      else dirty.markClean();
    },
    [dirty],
  );

  const { orderedItems, sensors, handleDragEnd } = useSortableList({
    items: crud.filteredItems,
    onReorder: (newIds) => { reorderMutation.mutate({ ids: newIds.map(Number) }); },
  });

  const { handleSave } = useMasterSave<AnimalSpecies, AnimalSpeciesFormData, CreateAnimalSpeciesRequest, UpdateAnimalSpeciesRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "動物種類名は必須です" : null),
    toCreateRequest: buildAnimalSpeciesCreateRequest,
    toUpdateRequest: buildAnimalSpeciesUpdateRequest,
  });

  return (
    <MasterCRUDPage title="動物種類マスタ" icon={<PawPrint className={`${ICON.page} ${C.text}`} />} resource={ResourceMasterAnimalSpecies}
      entityLabel="動物種類" searchPlaceholder="動物種類名で検索..." emptyMessage="動物種類が登録されていません"
      crud={crud} handleSave={handleSave} columns={ANIMAL_SPECIES_COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      deleteDescription={`「${crud.pendingDelete?.name}」を削除します。ペットで使用中の場合は削除できません。この操作は取り消せません。`}
      renderRow={() => null}
      renderSidePanel={({ readOnly, ...props }) => <AnimalSpeciesSidePanel key={props.item?.id ?? "new"} {...props} readOnly={readOnly} onDirtyChange={handleDirtyChange} />}
    >
      <AnimalSpeciesSortableTable
        items={orderedItems}
        sensors={sensors}
        onDragEnd={handleDragEnd}
        canEdit={canEdit}
        onEdit={crud.handleEdit}
      />
    </MasterCRUDPage>
  );
}
