import { memo, useState, useCallback, useRef, useEffect } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { Activity } from "lucide-react";
import { toast } from "sonner";
import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { PropertyInput } from "@/components/shared/SidePeek/PropertyInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import {
  useGetServiceTypes, useCreateServiceType, useUpdateServiceType, useDeleteServiceType, useReorderServiceTypes,
} from "@/features/master/api/service-types";
import type { ServiceType } from "@/features/master/api/service-types";
import type { CreateServiceTypeRequest, UpdateServiceTypeRequest } from "@/types/service-type";

const COLUMNS = [
  { header: "", className: "w-[32px]" }, { header: "名称" },
  { header: "備考", className: "w-[240px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface ServiceTypeFormData { name: string; description: string; color: string; isActive: boolean; }

const ServiceTypeSidePanel = memo(function ServiceTypeSidePanel({
  item, onClose, onSave, onDeleteRequest,
}: { item: ServiceType | null; onClose: () => void; onSave: (d: ServiceTypeFormData) => void; onDeleteRequest: (i: ServiceType) => void; }) {
  const [f, setF] = useState<ServiceTypeFormData>(() => ({
    name: item?.name ?? "", description: item?.description ?? "", color: item?.color ?? "#3B82F6", isActive: item?.isActive ?? true,
  }));
  return (
    <MasterSidePanel isNew={item === null} title={f.name} onTitleChange={(v) => setF((p) => ({ ...p, name: v }))}
      onClose={onClose} onSave={() => onSave(f)} onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      icon={<Activity className={LAYOUT.pageIcon.innerIcon} />}>
      <StatusToggleButton isActive={f.isActive} onToggle={() => setF((p) => ({ ...p, isActive: !p.isActive }))} />
      <PropertyRow label="カラー">
        <div className="flex items-center gap-2">
          <input type="color" value={f.color} onChange={(e) => setF((p) => ({ ...p, color: e.target.value }))}
            className="w-7 h-7 rounded cursor-pointer border-0 bg-transparent p-0" />
          <PropertyInput value={f.color} onChange={(v) => setF((p) => ({ ...p, color: v }))} placeholder="#3B82F6" />
        </div>
      </PropertyRow>
      <PropertyRow label="備考">
        <PropertyInput value={f.description} onChange={(v) => setF((p) => ({ ...p, description: v }))} placeholder="補足情報など" />
      </PropertyRow>
    </MasterSidePanel>
  );
});

export function ServiceTypeSettings() {
  const { data } = useGetServiceTypes();
  const createMutation = useCreateServiceType();
  const updateMutation = useUpdateServiceType();
  const deleteMutation = useDeleteServiceType();
  const reorderMutation = useReorderServiceTypes();

  const crud = useMasterCRUD<ServiceType>({ data, deleteMutation, entityLabel: "予約区分" });

  const resetOrderRef = useRef<() => void>(() => {});
  const handleReorder = useCallback((newIds: string[]) => {
    reorderMutation.mutate({ ids: newIds.map(Number) }, {
      onError: () => { resetOrderRef.current(); toast.error("並び替えに失敗しました"); },
    });
  }, [reorderMutation]);

  const { orderedItems, sensors, handleDragStart, handleDragEnd, handleDragCancel, resetOrder } =
    useSortableList({ items: crud.filteredItems, onReorder: handleReorder });
  useEffect(() => { resetOrderRef.current = resetOrder; }, [resetOrder]);

  const { handleSave } = useMasterSave<ServiceType, ServiceTypeFormData, CreateServiceTypeRequest, UpdateServiceTypeRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: (d) => ({ name: d.name, description: d.description || undefined, color: d.color || undefined, is_active: true }),
    toUpdateRequest: (d) => ({ name: d.name, description: d.description || undefined, color: d.color || undefined, is_active: d.isActive }),
  });

  return (
    <MasterCRUDPage title="予約区分マスタ" icon={<Activity className={`${ICON.page} text-[#37352F]`} />}
      entityLabel="予約区分" searchPlaceholder="予約区分名で検索..." emptyMessage="予約区分が登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={() => null}
      renderSidePanel={(props) => <ServiceTypeSidePanel key={props.item?.id ?? "new"} {...props} />}
    >
      <DndContext sensors={sensors} collisionDetection={closestCenter}
        onDragStart={handleDragStart} onDragEnd={handleDragEnd} onDragCancel={handleDragCancel}>
        <SortableContext items={orderedItems.map((i) => i.id)} strategy={verticalListSortingStrategy}>
          <DataTable columns={COLUMNS} data={orderedItems} emptyMessage="予約区分が登録されていません"
            renderRow={(item) => (
              <SortableDataTableRow key={item.id} id={item.id} onClick={() => crud.handleEdit(item)}>
                <TableCell className={`font-medium text-base ${C.text}`}>
                  <div className="flex items-center gap-2">
                    <span className="size-3 rounded-full shrink-0" style={{ backgroundColor: item.color }} />
                    {item.name}
                  </div>
                </TableCell>
                <TableCell className={`text-base ${C.text70} truncate max-w-[240px]`}>{item.description || "-"}</TableCell>
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
