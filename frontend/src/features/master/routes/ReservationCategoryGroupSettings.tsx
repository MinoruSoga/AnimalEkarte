import { memo, useState, useCallback, useRef, useEffect } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { Layers } from "lucide-react";
import { handleApiError } from "@/lib/handle-api-error";
import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, LAYOUT, ICON, PALETTE } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import { usePermission } from "@/features/auth";
import {
  useGetReservationCategoryGroups,
  useCreateReservationCategoryGroup,
  useUpdateReservationCategoryGroup,
  useDeleteReservationCategoryGroup,
  useReorderReservationCategoryGroups,
} from "@/features/master/api/reservation-category-groups";
import type { ReservationCategoryGroup } from "@/features/master/api/reservation-category-groups";
import type {
  CreateReservationCategoryGroupRequest,
  UpdateReservationCategoryGroupRequest,
} from "@/features/master/api/reservation-category-groups";
import { ResourceMasterReservationCategory } from "@/types/generated/models";

const COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "グループ名" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface GroupFormData {
  name: string;
  color: string;
  isActive: boolean;
}

const ReservationCategoryGroupSidePanel = memo(function ReservationCategoryGroupSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
}: {
  item: ReservationCategoryGroup | null;
  onClose: () => void;
  onSave: (d: GroupFormData) => void;
  onDeleteRequest?: (i: ReservationCategoryGroup) => void;
  readOnly?: boolean;
}) {
  const [f, setF] = useState<GroupFormData>(() => ({
    name: item?.name ?? "",
    color: item?.color ?? PALETTE.pickerDefaultBlue,
    isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  const handleTitleChange = useCallback((v: string) => {
    setF((p) => ({ ...p, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleColorPickerChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setF((p) => ({ ...p, color: e.target.value }));
    setIsDirty(true);
  }, []);

  const handleColorInputChange = useCallback((v: string) => {
    setF((p) => ({ ...p, color: v }));
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setF((p) => ({ ...p, isActive: !p.isActive }));
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    if (!f.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(f);
    setIsDirty(false);
  }, [f, onSave]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel
      isNew={item === null}
      title={f.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Layers className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={f.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="カラー">
        <div className="flex items-center gap-2">
          <input
            type="color"
            value={f.color}
            onChange={handleColorPickerChange}
            className="w-7 h-7 rounded cursor-pointer border-0 bg-transparent p-0"
          />
          <PropertyInput value={f.color} onChange={handleColorInputChange} placeholder="#3B82F6" />
        </div>
      </PropertyRow>
    </MasterSidePanel>
  );
});

export function ReservationCategoryGroupSettings() {
  const { canEdit } = usePermission(ResourceMasterReservationCategory);
  const { data } = useGetReservationCategoryGroups();
  const createMutation = useCreateReservationCategoryGroup();
  const updateMutation = useUpdateReservationCategoryGroup();
  const deleteMutation = useDeleteReservationCategoryGroup();
  const reorderMutation = useReorderReservationCategoryGroups();

  const crud = useMasterCRUD<ReservationCategoryGroup>({
    data,
    deleteMutation,
    entityLabel: "予約区分グループ",
  });

  const resetOrderRef = useRef<() => void>(() => {});
  const handleReorder = useCallback(
    (newIds: string[]) => {
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        {
          onError: (error: unknown) => {
            resetOrderRef.current();
            handleApiError(error, "並び替え");
          },
        },
      );
    },
    [reorderMutation],
  );

  const { orderedItems, sensors, handleDragStart, handleDragEnd, handleDragCancel, resetOrder } =
    useSortableList({ items: crud.filteredItems, onReorder: handleReorder });
  useEffect(() => {
    resetOrderRef.current = resetOrder;
  }, [resetOrder]);

  const { handleSave } = useMasterSave<
    ReservationCategoryGroup,
    GroupFormData,
    CreateReservationCategoryGroupRequest,
    UpdateReservationCategoryGroupRequest
  >({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: (d) => ({
      name: d.name,
      color: d.color || undefined,
      is_active: d.isActive,
    }),
    toUpdateRequest: (d) => ({
      name: d.name,
      color: d.color || undefined,
      is_active: d.isActive,
    }),
  });

  return (
    <MasterCRUDPage
      title="予約区分グループマスタ"
      icon={<Layers className={`${ICON.page} ${C.text}`} />}
      resource={ResourceMasterReservationCategory}
      entityLabel="予約区分グループ"
      searchPlaceholder="グループ名で検索..."
      emptyMessage="予約区分グループが登録されていません"
      crud={crud}
      handleSave={handleSave}
      columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={() => null}
      renderSidePanel={({ readOnly, ...props }) => (
        <ReservationCategoryGroupSidePanel
          key={props.item?.id ?? "new"}
          {...props}
          readOnly={readOnly}
        />
      )}
    >
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
        onDragCancel={handleDragCancel}
      >
        <SortableContext
          items={orderedItems.map((i) => i.id)}
          strategy={verticalListSortingStrategy}
        >
          <DataTable
            columns={COLUMNS}
            data={orderedItems}
            emptyMessage="予約区分グループが登録されていません"
            renderRow={(item) => (
              <SortableDataTableRow
                key={item.id}
                id={item.id}
                onClick={() => crud.handleEdit(item)}
              >
                <TableCell className={`font-medium text-base ${C.text}`}>
                  <div className="flex items-center gap-2">
                    <span
                      className="size-3 rounded-full shrink-0"
                      style={{ backgroundColor: item.color }}
                    />
                    {item.name}
                  </div>
                </TableCell>
                <TableCell className="text-center">
                  <NotionStatusPill isActive={item.isActive} />
                </TableCell>
                <TableCell className="p-0 text-right">
                  {canEdit ? <RowActionButton onClick={() => crud.handleEdit(item)} /> : null}
                </TableCell>
              </SortableDataTableRow>
            )}
          />
        </SortableContext>
      </DndContext>
    </MasterCRUDPage>
  );
}
