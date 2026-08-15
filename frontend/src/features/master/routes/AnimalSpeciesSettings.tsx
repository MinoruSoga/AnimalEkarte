import { useCallback, useContext } from "react";
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
import { AuthContext } from "@/hooks/auth-context";
import { usePermission } from "@/hooks/use-permission";
import { useGetAnimalSpecies, useCreateAnimalSpecies, useUpdateAnimalSpecies, useDeleteAnimalSpecies, useReorderAnimalSpecies } from "../api/animal-species";
import type { AnimalSpecies, CreateAnimalSpeciesRequest, UpdateAnimalSpeciesRequest } from "../api/animal-species";
import {
  buildAnimalSpeciesCreateRequest,
  buildAnimalSpeciesUpdateRequest,
} from "./animal-species-settings-model";
import { ResourceMasterAnimalSpecies } from "@/types/generated/models";

export function AnimalSpeciesSettings() {
  // Read remains resource-view scoped (route + listing). Global animal_species
  // writes require isSystemAdmin (backend requireSystemAdminForGlobalMaster).
  usePermission(ResourceMasterAnimalSpecies);
  // Fail-closed: missing AuthProvider ⇒ no mutation affordances.
  const auth = useContext(AuthContext);
  const canMutate = auth?.user?.isSystemAdmin === true;
  const {
    data: animalSpecies = [],
    isPending,
    isError,
  } = useGetAnimalSpecies();
  const isSpeciesUnavailable = isError || isPending;
  const createMutation = useCreateAnimalSpecies();
  const updateMutation = useUpdateAnimalSpecies();
  const deleteMutation = useDeleteAnimalSpecies();
  const reorderMutation = useReorderAnimalSpecies();

  // BUG-380: 未保存変更の破棄確認 + beforeunload ガード
  const dirty = useSidePeekDirty();

  const crud = useMasterCRUD<AnimalSpecies>({
    data: isSpeciesUnavailable ? [] : animalSpecies,
    deleteMutation,
    entityLabel: "動物種類",
    dirtyGuard: dirty,
    permissions: { canDelete: canMutate && !isSpeciesUnavailable },
  });

  const handleDirtyChange = useCallback(
    (d: boolean) => {
      if (d) dirty.markDirty();
      else dirty.markClean();
    },
    [dirty],
  );

  const { orderedItems, sensors, handleDragEnd } = useSortableList({
    items: crud.filteredItems,
    onReorder: (newIds) => {
      if (!canMutate || isSpeciesUnavailable) return;
      reorderMutation.mutate({ ids: newIds.map(Number) });
    },
  });

  const { handleSave } = useMasterSave<AnimalSpecies, AnimalSpeciesFormData, CreateAnimalSpeciesRequest, UpdateAnimalSpeciesRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "動物種類名は必須です" : null),
    toCreateRequest: buildAnimalSpeciesCreateRequest,
    toUpdateRequest: buildAnimalSpeciesUpdateRequest,
    permissions: {
      canCreate: canMutate,
      canEdit: canMutate && !isSpeciesUnavailable,
    },
  });

  return (
    // resource を渡すと MasterCRUDPage が legacy resource-edit で mutation UI を出してしまう。
    // 未指定時 usePermission("") は isSystemAdmin のみ true（hasPermission バイパス）になるため、
    // 新規登録/side panel の mutation affordance も system admin に限定される。
    <MasterCRUDPage title="動物種類マスタ" icon={<PawPrint className={`${ICON.page} ${C.text}`} />}
      entityLabel="動物種類" searchPlaceholder="動物種類名で検索..." emptyMessage="動物種類が登録されていません"
      crud={crud} handleSave={handleSave} columns={ANIMAL_SPECIES_COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      deleteDescription={`「${crud.pendingDelete?.name}」を削除します。ペットで使用中の場合は削除できません。この操作は取り消せません。`}
      renderRow={() => null}
      renderSidePanel={({ readOnly, ...props }) => <AnimalSpeciesSidePanel key={props.item?.id ?? "new"} {...props} readOnly={readOnly || (isSpeciesUnavailable && props.item !== null)} onDirtyChange={handleDirtyChange} />}
    >
      {isError ? (
        <p role="alert" aria-atomic="true" className={`text-sm ${C.danger}`}>
          動物種の取得に失敗しました。
        </p>
      ) : isPending ? (
        <p
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`text-sm ${C.text50}`}
        >
          動物種を読み込み中です。
        </p>
      ) : animalSpecies.length === 0 ? (
        <p
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`text-sm ${C.text50}`}
        >
          動物種マスタが登録されていません。
        </p>
      ) : (
        <AnimalSpeciesSortableTable
          items={orderedItems}
          sensors={sensors}
          onDragEnd={handleDragEnd}
          canEdit={canMutate}
          onEdit={crud.handleEdit}
        />
      )}
    </MasterCRUDPage>
  );
}
