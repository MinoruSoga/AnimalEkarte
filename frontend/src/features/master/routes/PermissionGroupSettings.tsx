import { useCallback } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { Lock } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { usePermission } from "@/hooks/use-permission";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { C, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import type { FilterProperty } from "@/components/shared/NotionFilter/types";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { PermissionGroupSidePanel } from "../components/PermissionGroupSidePanel";
import type { PermissionGroupFormData } from "../components/PermissionGroupSidePanelModel";
import {
  useGetPermissionGroups,
  useCreatePermissionGroup,
  useUpdatePermissionGroup,
  useDeletePermissionGroup,
  useUpdatePermissionGroupRules,
  useReorderPermissionGroups,
  type PermissionGroup,
  type CreatePermissionGroupRequest,
  type UpdatePermissionGroupRequest,
} from "../api/permission-groups";
import {
  buildPermissionGroupCreateRequest,
  buildPermissionGroupRulesRequest,
  buildPermissionGroupUpdateRequest,
} from "./PermissionGroupSettingsModel";
import { ResourceMasterPermission } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "グループ名", className: "flex-1" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const PERMISSION_GROUP_FILTER_PROPERTIES: FilterProperty[] = [
  MASTER_STATUS_FILTER,
];

export function PermissionGroupSettings() {
  const { canEdit } = usePermission(ResourceMasterPermission);
  const { data } = useGetPermissionGroups();
  const createMutation = useCreatePermissionGroup();
  const updateMutation = useUpdatePermissionGroup();
  const deleteMutation = useDeletePermissionGroup();
  const updateRulesMutation = useUpdatePermissionGroupRules();
  const reorderMutation = useReorderPermissionGroups();

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<PermissionGroup>({
    data,
    deleteMutation,
    entityLabel: "権限グループ",
    searchFilter: (g, lower) =>
      g.name.toLowerCase().includes(lower) ||
      g.description.toLowerCase().includes(lower),
    activeFilterApply: (item, filters) => {
      for (const f of filters) {
        if (f.key === "status" && typeof f.value === "string") {
          const want = f.value === "active";
          if (f.condition === "is" && item.isActive !== want) return false;
          if (f.condition === "is_not" && item.isActive === want) return false;
        }
      }
      return true;
    },
    dirtyGuard: dirty,
  });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const { orderedItems, sensors, handleDragEnd } = useSortableList({
    items: crud.filteredItems,
    onReorder: (newIds) => {
      reorderMutation.mutate(newIds);
    },
  });

  const { handleSave } = useMasterSave<
    PermissionGroup,
    PermissionGroupFormData,
    CreatePermissionGroupRequest,
    UpdatePermissionGroupRequest
  >({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => {
      if (!d.name.trim()) return "グループ名は必須です";
      if (!d.color) return "カラーは必須です";
      return null;
    },
    toCreateRequest: buildPermissionGroupCreateRequest,
    toUpdateRequest: buildPermissionGroupUpdateRequest,
    onSuccess: async (saved, formData) => {
      if (formData.rules.length > 0) {
        await updateRulesMutation.mutateAsync({
          id: saved.id,
          req: buildPermissionGroupRulesRequest(formData),
        });
      }
    },
  });

  return (
    <MasterCRUDPage
      title="権限グループマスタ"
      icon={<Lock className={`${ICON.page} ${C.text}`} />}
      resource={ResourceMasterPermission}
      entityLabel="グループ"
      searchPlaceholder="グループ名、説明で検索..."
      emptyMessage="権限グループが登録されていません"
      crud={crud}
      handleSave={handleSave}
      columns={COLUMNS}
      filterProperties={PERMISSION_GROUP_FILTER_PROPERTIES}
      renderRow={() => null}
      renderSidePanel={({ item, onClose, onSave, onDeleteRequest, readOnly }) => (
        <PermissionGroupSidePanel
          key={item?.id ?? "new"}
          item={item}
          onClose={onClose}
          onSave={onSave}
          onDeleteRequest={onDeleteRequest}
          readOnly={readOnly}
          onDirtyChange={handleDirtyChange}
        />
      )}
    >
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <SortableContext items={orderedItems.map((i) => i.id)} strategy={verticalListSortingStrategy}>
          <DataTable columns={COLUMNS} data={orderedItems} emptyMessage="権限グループが登録されていません"
            renderRow={(item) => (
              <SortableDataTableRow key={item.id} id={item.id} onClick={() => crud.handleEdit(item)}>
                <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
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
