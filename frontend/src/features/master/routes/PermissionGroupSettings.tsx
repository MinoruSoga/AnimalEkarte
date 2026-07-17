import { useCallback } from "react";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { Lock } from "lucide-react";
import { usePermission } from "@/hooks/use-permission";
import { C, ICON } from "@/lib/design-tokens";
import { normalizeKana } from "@/lib/normalize-kana";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import type { FilterProperty } from "@/components/shared/PropertyFilter/types";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { PermissionGroupSidePanel } from "../components/PermissionGroupSidePanel";
import type { PermissionGroupFormData } from "../components/permission-group-side-panel-model";
import {
  PERMISSION_GROUP_COLUMNS,
  PermissionGroupSortableTable,
} from "../components/PermissionGroupSortableTable";
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
} from "./permission-group-settings-model";
import { ResourceMasterPermission } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

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
      normalizeKana(g.name).toLowerCase().includes(lower) ||
      normalizeKana(g.description).toLowerCase().includes(lower),
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
      columns={PERMISSION_GROUP_COLUMNS}
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
      <PermissionGroupSortableTable
        items={orderedItems}
        sensors={sensors}
        onDragEnd={handleDragEnd}
        canEdit={canEdit}
        onEdit={crud.handleEdit}
      />
    </MasterCRUDPage>
  );
}
