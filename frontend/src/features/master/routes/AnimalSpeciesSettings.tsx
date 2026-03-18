import { memo, useState } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import PawPrint from "lucide-react/dist/esm/icons/paw-print";
import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { C, LAYOUT } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import {
  useGetAnimalSpecies, useCreateAnimalSpecies, useUpdateAnimalSpecies, useDeleteAnimalSpecies, useReorderAnimalSpecies,
} from "@/features/master/api/animal-species";
import type { AnimalSpecies, UpdateAnimalSpeciesRequest } from "@/features/master/api/animal-species";

const COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "動物種類名" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface FormData { name: string; isActive: boolean; }

const SidePanel = memo(function SidePanel({
  item, onClose, onSave, onDeleteRequest,
}: { item: AnimalSpecies | null; onClose: () => void; onSave: (d: FormData) => void; onDeleteRequest: (i: AnimalSpecies) => void; }) {
  const [f, setF] = useState<FormData>(() => ({ name: item?.name ?? "", isActive: item?.isActive ?? true }));
  return (
    <MasterSidePanel isNew={item === null} title={f.name} onTitleChange={(v) => setF((p) => ({ ...p, name: v }))}
      onClose={onClose} onSave={() => onSave(f)} onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      icon={<PawPrint className={LAYOUT.pageIcon.innerIcon} />}>
      <StatusToggleButton isActive={f.isActive} onToggle={() => setF((p) => ({ ...p, isActive: !p.isActive }))} />
    </MasterSidePanel>
  );
});

export function AnimalSpeciesSettings() {
  const { data } = useGetAnimalSpecies();
  const createMutation = useCreateAnimalSpecies();
  const updateMutation = useUpdateAnimalSpecies();
  const deleteMutation = useDeleteAnimalSpecies();
  const reorderMutation = useReorderAnimalSpecies();

  const crud = useMasterCRUD<AnimalSpecies>({ data, deleteMutation, entityLabel: "動物種類" });

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
    <MasterCRUDPage title="動物種類マスタ" icon={<PawPrint className="size-5 text-[#37352F]" />}
      entityLabel="動物種類" searchPlaceholder="動物種類名で検索..." emptyMessage="動物種類が登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      deleteDescription={`「${crud.pendingDelete?.name}」を削除します。ペットで使用中の場合は削除できません。この操作は取り消せません。`}
      renderRow={() => null}
      renderSidePanel={(props) => <SidePanel key={props.item?.id ?? "new"} {...props} />}
    >
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <SortableContext items={orderedItems.map((i) => i.id)} strategy={verticalListSortingStrategy}>
          <DataTable columns={COLUMNS} data={orderedItems} emptyMessage="動物種類が登録されていません"
            renderRow={(item) => (
              <SortableDataTableRow key={item.id} id={item.id} onClick={() => crud.handleEdit(item)}>
                <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
                <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
                <TableCell className="p-0 text-right"><RowActionButton onClick={() => crud.handleEdit(item)} /></TableCell>
              </SortableDataTableRow>
            )}
          />
        </SortableContext>
      </DndContext>
    </MasterCRUDPage>
  );
}
