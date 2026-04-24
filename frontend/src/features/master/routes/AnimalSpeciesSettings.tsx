import { memo, useState, useCallback, useEffect } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import PawPrint from "lucide-react/dist/esm/icons/paw-print";
import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { StatusToggleButton, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { usePermission } from "@/hooks/use-permission";
import { useGetAnimalSpecies, useCreateAnimalSpecies, useUpdateAnimalSpecies, useDeleteAnimalSpecies, useReorderAnimalSpecies } from "../api/animal-species";
import type { AnimalSpecies, UpdateAnimalSpeciesRequest } from "../api/animal-species";
import { ResourceMasterAnimalSpecies } from "@/types/generated/models";

const COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "動物種類名" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface FormData { name: string; isActive: boolean; }

const SidePanel = memo(function SidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly, onDirtyChange,
}: { item: AnimalSpecies | null; onClose: () => void; onSave: (d: FormData) => void; onDeleteRequest?: (i: AnimalSpecies) => void; readOnly?: boolean; onDirtyChange?: (dirty: boolean) => void; }) {
  const [formData, setFormData] = useState<FormData>(() => ({ name: item?.name ?? "", isActive: item?.isActive ?? true }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // BUG-380
  useEffect(() => { onDirtyChange?.(isDirty); }, [isDirty, onDirtyChange]);
  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleTitleChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: v }));
    if (v.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(formData);
    setIsDirty(false);
  }, [formData, onSave]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel isNew={item === null} title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose} action={readOnly ? undefined : handleAction} onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<PawPrint className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}>
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
    </MasterSidePanel>
  );
});

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

  const { handleSave } = useMasterSave<AnimalSpecies, FormData, { name: string; is_active: boolean; sort_order: number }, UpdateAnimalSpeciesRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "動物種類名は必須です" : null),
    toCreateRequest: (d) => ({ name: d.name, is_active: true, sort_order: 0 }),
    toUpdateRequest: (d) => ({ name: d.name, is_active: d.isActive }),
  });

  return (
    <MasterCRUDPage title="動物種類マスタ" icon={<PawPrint className={`${ICON.page} ${C.text}`} />} resource={ResourceMasterAnimalSpecies}
      entityLabel="動物種類" searchPlaceholder="動物種類名で検索..." emptyMessage="動物種類が登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      deleteDescription={`「${crud.pendingDelete?.name}」を削除します。ペットで使用中の場合は削除できません。この操作は取り消せません。`}
      renderRow={() => null}
      renderSidePanel={({ readOnly, ...props }) => <SidePanel key={props.item?.id ?? "new"} {...props} readOnly={readOnly} onDirtyChange={handleDirtyChange} />}
    >
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <SortableContext items={orderedItems.map((i) => i.id)} strategy={verticalListSortingStrategy}>
          <DataTable columns={COLUMNS} data={orderedItems} emptyMessage="動物種類が登録されていません"
            renderRow={(item) => (
              <SortableDataTableRow key={item.id} id={item.id} onClick={() => crud.handleEdit(item)}>
                <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
                <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
                <TableCell className="p-0 text-right">{canEdit ? <RowActionButton onClick={() => crud.handleEdit(item)} /> : null}</TableCell>
              </SortableDataTableRow>
            )}
          />
        </SortableContext>
      </DndContext>
    </MasterCRUDPage>
  );
}
