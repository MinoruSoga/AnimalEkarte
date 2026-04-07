import { memo, useState, useCallback } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { Lock } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { usePermission } from "@/features/auth/hooks/use-permission";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS, MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import type { FilterProperty } from "@/components/shared/NotionFilter/types";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import {
  useGetPermissionGroups,
  useCreatePermissionGroup,
  useUpdatePermissionGroup,
  useDeletePermissionGroup,
  useSetPermissionGroupRules,
  useReorderPermissionGroups,
  type PermissionGroup,
  type CreatePermissionGroupRequest,
  type UpdatePermissionGroupRequest,
  type SetPermissionGroupRulesRequest,
} from "@/features/master/api/permission-groups";
import { PermissionRuleTable, type PermissionRule } from "@/features/master/components/PermissionRuleTable";
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

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface PermissionGroupFormData {
  name: string;
  description: string;
  color: string;
  isActive: boolean;
  rules: PermissionRule[];
}

// ─────────────────────────────────────────────────
// PermissionGroupSidePanel
// ─────────────────────────────────────────────────

interface PermissionGroupSidePanelProps {
  item: PermissionGroup | null;
  onClose: () => void;
  onSave: (d: PermissionGroupFormData) => void;
  onDeleteRequest?: (i: PermissionGroup) => void;
}

const PermissionGroupSidePanel = memo(function PermissionGroupSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: PermissionGroupSidePanelProps) {
  const isNew = item === null;

  const [f, setF] = useState<PermissionGroupFormData>(() => ({
    name: item?.name ?? "",
    description: item?.description ?? "",
    color: item?.color ?? "#6B7280",
    isActive: item?.isActive ?? true,
    rules: item?.rules?.map((r) => ({
      resource: r.resource,
      canView: r.canView,
      canCreate: r.canCreate,
      canEdit: r.canEdit,
      canDelete: r.canDelete,
    })) ?? [],
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // ── Handlers ─────────────────────────────────────
  const handleSave = useCallback(() => {
    if (!f.name.trim()) {
      setNameError("グループ名を入力してください");
      return;
    }
    setNameError("");
    onSave(f);
    setIsDirty(false);
  }, [f, onSave]);

  const handleNameChange = useCallback((v: string) => {
    setF((p) => ({ ...p, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleDescriptionChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setF((p) => ({ ...p, description: e.target.value }));
    setIsDirty(true);
  }, []);

  const handleColorChange = useCallback((v: string) => {
    setF((p) => ({ ...p, color: v }));
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setF((p) => ({ ...p, isActive: !p.isActive }));
    setIsDirty(true);
  }, []);

  const handleRuleChange = useCallback((resource: string, field: keyof Omit<PermissionRule, "resource">, value: boolean) => {
    setF((p) => {
      const existingRule = p.rules.find((r) => r.resource === resource);
      if (existingRule) {
        // 既存ルール更新
        return {
          ...p,
          rules: p.rules.map((r) =>
            r.resource === resource ? { ...r, [field]: value } : r
          ),
        };
      } else {
        // 新規ルール作成
        const newRule: PermissionRule = {
          resource,
          canView: field === "canView" ? value : false,
          canCreate: field === "canCreate" ? value : false,
          canEdit: field === "canEdit" ? value : false,
          canDelete: field === "canDelete" ? value : false,
        };
        return {
          ...p,
          rules: [...p.rules, newRule],
        };
      }
    });
    setIsDirty(true);
  }, []);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel
      isNew={isNew}
      title={f.name}
      onTitleChange={handleNameChange}
      onClose={handleClose}
      action={handleSave}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Lock className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
    >
      <StatusToggleButton
        isActive={f.isActive}
        onToggle={handleToggleActive}
      />

      <PropertyRow label="説明">
        <textarea
          className={`${MASTER_INPUT_CLASS} resize-none`}
          value={f.description}
          onChange={handleDescriptionChange}
          placeholder="グループの説明"
          rows={3}
        />
      </PropertyRow>

      <PropertyRow label="カラー">
        <div className="flex items-center gap-3">
          <input
            type="color"
            className="w-12 h-12 rounded border"
            value={f.color}
            onChange={(e) => handleColorChange(e.target.value)}
          />
          <span className={`text-sm ${C.text50}`}>{f.color}</span>
        </div>
      </PropertyRow>

      <PermissionRuleTable
        group={item}
        rules={f.rules}
        onRuleChange={handleRuleChange}
      />
    </MasterSidePanel>
  );
});

// ─────────────────────────────────────────────────
// PermissionGroupSettings (page)
// ─────────────────────────────────────────────────

export function PermissionGroupSettings() {
  const { canEdit } = usePermission(ResourceMasterPermission);
  const { data } = useGetPermissionGroups();
  const createMutation = useCreatePermissionGroup();
  const updateMutation = useUpdatePermissionGroup();
  const deleteMutation = useDeletePermissionGroup();
  const setRulesMutation = useSetPermissionGroupRules();
  const reorderMutation = useReorderPermissionGroups();

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
  });

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
    toCreateRequest: (d) => ({
      name: d.name,
      description: d.description,
      color: d.color,
      is_active: d.isActive,
    }),
    toUpdateRequest: (d) => ({
      name: d.name,
      description: d.description,
      color: d.color,
      is_active: d.isActive,
    }),
    onSuccess: async (saved, formData) => {
      // Set rules if any
      if (formData.rules.length > 0) {
        const rulesReq: SetPermissionGroupRulesRequest = {
          rules: formData.rules.map((r) => ({
            resource: r.resource,
            can_view: r.canView,
            can_create: r.canCreate,
            can_edit: r.canEdit,
            can_delete: r.canDelete,
          })),
        };
        await setRulesMutation.mutateAsync({ id: saved.id, req: rulesReq });
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
      renderSidePanel={({ item, onClose, onSave, onDeleteRequest }) => (
        <PermissionGroupSidePanel
          key={item?.id ?? "new"}
          item={item}
          onClose={onClose}
          onSave={onSave}
          onDeleteRequest={onDeleteRequest}
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
