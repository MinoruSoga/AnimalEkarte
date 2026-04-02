import { memo, useCallback, useDeferredValue, useMemo, useState, useTransition } from "react";
import type { ChangeEvent } from "react";
import { Shield } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton/RowActionButton";
import { MasterListPage } from "@/features/master/components/MasterListPage";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { C, LAYOUT, ICON, PALETTE } from "@/lib/design-tokens";
import { useAuth } from "@/features/auth";
import { PermissionRuleTable } from "@/features/master/components/PermissionRuleTable";
import { PERMISSION_RESOURCES } from "@/features/master/types/permission-resources";
import { useGetPermissionGroups, useCreatePermissionGroup, useUpdatePermissionGroup, useDeletePermissionGroup, useSetPermissionGroupRules } from "@/features/master/api/permission-groups/permission-groups";
import type { RuleInput } from "@/features/master/api/permission-groups/permission-groups";
import type { PermissionGroup, PermissionGroupRule } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────

function buildDefaultRules(): RuleInput[] {
  return PERMISSION_RESOURCES.map((r) => ({
    resource: r.key,
    can_view: false,
    can_create: false,
    can_edit: false,
    can_delete: false,
  }));
}

function groupRulesToRuleInputs(rules: PermissionGroupRule[] | undefined): RuleInput[] {
  const base = buildDefaultRules();
  if (!rules) return base;
  return base.map((r) => {
    const found = rules.find((gr) => gr.resource === r.resource);
    if (!found) return r;
    return {
      resource: r.resource,
      can_view: found.can_view,
      can_create: found.can_create,
      can_edit: found.can_edit,
      can_delete: found.can_delete,
    };
  });
}

// ─────────────────────────────────────────────────
// Table columns
// ─────────────────────────────────────────────────

const COLUMNS = [
  { header: "グループ名", className: "flex-1" },
  { header: "説明", className: "flex-1" },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

// ─────────────────────────────────────────────────
// Side panel
// ─────────────────────────────────────────────────

interface GroupSidePanelProps {
  group: PermissionGroup | null;
  clinicId: string;
  onClose: () => void;
  onDeleteRequest: (group: PermissionGroup) => void;
}

const GroupSidePanel = memo(function GroupSidePanel({
  group,
  clinicId,
  onClose,
  onDeleteRequest,
}: GroupSidePanelProps) {
  const isNew = group === null;

  const [name, setName] = useState(() => group?.name ?? "");
  const [description, setDescription] = useState(() => group?.description ?? "");
  const [color, setColor] = useState(() => group?.color ?? PALETTE.defaultGray);
  const [rules, setRules] = useState<RuleInput[]>(() =>
    groupRulesToRuleInputs(group?.rules),
  );
  const [isDirty, setIsDirty] = useState(false);
  // BUG-083: inline validation error for the name field
  const [nameError, setNameError] = useState("");

  const createMutation = useCreatePermissionGroup();
  const updateMutation = useUpdatePermissionGroup(clinicId);
  const setRulesMutation = useSetPermissionGroupRules(clinicId);

  const [isSavePending, startSaveTransition] = useTransition();

  const handleColorChange = useCallback((e: ChangeEvent<HTMLInputElement>) => {
    setColor(e.target.value);
    setIsDirty(true);
  }, []);

  const handleDescriptionChange = useCallback((e: ChangeEvent<HTMLTextAreaElement>) => {
    setDescription(e.target.value);
    setIsDirty(true);
  }, []);

  const handleRulesChange = useCallback((updated: RuleInput[]) => {
    setRules(updated);
    setIsDirty(true);
  }, []);

  const handleSave = useCallback(() => {
    if (!name.trim()) {
      // BUG-083: インラインエラー表示（toast に加えてフィールド下にも表示）
      setNameError("名称を入力してください");
      toast.error("グループ名を入力してください");
      return;
    }
    setNameError("");
    startSaveTransition(async () => {
      try {
        let groupId: number;
        if (isNew) {
          const created = await createMutation.mutateAsync({
            clinicId,
            name: name.trim(),
            description: description.trim(),
            color,
          });
          groupId = created.id;
        } else {
          await updateMutation.mutateAsync({
            groupId: group.id,
            input: { name: name.trim(), description: description.trim(), color },
          });
          groupId = group.id;
        }
        await setRulesMutation.mutateAsync({ groupId, rules });
        toast.success(isNew ? "権限グループを作成しました" : "権限グループを更新しました");
        setIsDirty(false);
        onClose();
      } catch {
        toast.error(isNew ? "グループの作成に失敗しました" : "グループの更新に失敗しました");
      }
    });
  }, [
    name,
    description,
    color,
    rules,
    clinicId,
    isNew,
    group,
    createMutation,
    updateMutation,
    setRulesMutation,
    onClose,
  ]);

  const handleDelete = useCallback(() => {
    if (group !== null) onDeleteRequest(group);
  }, [group, onDeleteRequest]);

  const handleNameChange = useCallback((v: string) => {
    setName(v);
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel
      isNew={isNew}
      title={name}
      onTitleChange={handleNameChange}
      onClose={handleClose}
      action={handleSave}
      onDelete={isNew ? undefined : handleDelete}
      icon={<Shield className={LAYOUT.pageIcon.innerIcon} />}
      isPending={isSavePending}
      isDirty={isDirty}
      titlePlaceholder="グループ名を入力"
      titleError={nameError}
      titleMaxLength={100}
    >
      <PropertyRow label="カラー">
        <div className="flex items-center gap-2">
          <input
            type="color"
            value={color}
            onChange={handleColorChange}
            className="w-7 h-7 rounded border cursor-pointer"
          />
          <span className="text-sm text-muted-foreground">{color}</span>
        </div>
      </PropertyRow>

      <PropertyRow label="説明">
        <textarea
          value={description}
          onChange={handleDescriptionChange}
          placeholder="このグループの役割・用途を入力"
          rows={2}
          className="w-full text-sm bg-transparent border-none outline-none resize-none placeholder:text-muted-foreground"
        />
      </PropertyRow>

      <div className="mt-4 border-t pt-4">
        <p className="text-xs font-medium text-muted-foreground mb-3">ページ権限</p>
        <PermissionRuleTable rules={rules} onChange={handleRulesChange} />
      </div>
    </MasterSidePanel>
  );
});

// ─────────────────────────────────────────────────
// Table row
// ─────────────────────────────────────────────────

interface GroupRowProps {
  group: PermissionGroup;
  onEdit: (group: PermissionGroup) => void;
}

const GroupRow = memo(function GroupRow({ group, onEdit }: GroupRowProps) {
  const handleEdit = useCallback(
    (e: { stopPropagation: () => void }) => {
      e.stopPropagation();
      onEdit(group);
    },
    [group, onEdit],
  );

  const handleRowClick = useCallback(() => {
    onEdit(group);
  }, [group, onEdit]);

  return (
    <DataTableRow className="cursor-pointer" onClick={handleRowClick}>
      <TableCell>
        <div className="flex items-center gap-2">
          <div
            className="w-3 h-3 rounded-full flex-shrink-0"
            style={{ backgroundColor: group.color ?? PALETTE.defaultGray }}
          />
          <span className="text-sm font-medium">{group.name}</span>
        </div>
      </TableCell>
      <TableCell className="text-sm text-muted-foreground">
        {group.description ? group.description : "-"}
      </TableCell>
      <TableCell className="text-right">
        <RowActionButton onClick={handleEdit} />
      </TableCell>
    </DataTableRow>
  );
});

// ─────────────────────────────────────────────────
// Main page
// ─────────────────────────────────────────────────

export function PermissionGroupSettings() {
  const { currentClinicId } = useAuth();
  const clinicId = currentClinicId ?? "";

  const { data: groups = [], isLoading } = useGetPermissionGroups(clinicId);
  const deleteMutation = useDeletePermissionGroup(clinicId);

  // undefined = 閉じている, null = 新規作成, PermissionGroup = 編集中
  const [panelGroup, setPanelGroup] = useState<PermissionGroup | null | undefined>(undefined);
  const [pendingDelete, setPendingDelete] = useState<PermissionGroup | null>(null);
  const [, startDeleteTransition] = useTransition();

  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredGroups = useMemo(
    () =>
      deferredSearch.trim() === ""
        ? groups
        : groups.filter((g) =>
            g.name.toLowerCase().includes(deferredSearch.toLowerCase()),
          ),
    [groups, deferredSearch],
  );

  const handleNew = useCallback(() => {
    setPanelGroup(null);
  }, []);

  const handleEdit = useCallback((group: PermissionGroup) => {
    setPanelGroup(group);
  }, []);

  const handleClose = useCallback(() => {
    setPanelGroup(undefined);
  }, []);

  const handleDeleteRequest = useCallback((group: PermissionGroup) => {
    setPendingDelete(group);
    setPanelGroup(undefined);
  }, []);

  const handleDeleteConfirm = useCallback(() => {
    if (!pendingDelete) return;
    startDeleteTransition(async () => {
      try {
        await deleteMutation.mutateAsync(pendingDelete.id);
        toast.success("権限グループを削除しました");
      } catch (error) {
        handleApiError(error, "権限グループの削除");
      } finally {
        setPendingDelete(null);
      }
    });
  }, [pendingDelete, deleteMutation]);

  const handleDeleteCancel = useCallback(() => {
    setPendingDelete(null);
  }, []);

  return (
    <MasterListPage
      title="権限グループ管理"
      icon={<Shield className={`${ICON.page} ${C.text}`} />}
      searchTerm={searchTerm}
      onSearchChange={setSearchTerm}
      searchPlaceholder="グループ名を検索..."
      count={filteredGroups.length}
      onNew={handleNew}
      sidePanel={
        panelGroup !== undefined ? (
          <GroupSidePanel
            group={panelGroup}
            clinicId={clinicId}
            onClose={handleClose}
            onDeleteRequest={handleDeleteRequest}
          />
        ) : null
      }
      deleteOpen={pendingDelete !== null}
      deleteTitle="権限グループを削除しますか？"
      deleteDescription={`「${pendingDelete?.name ?? ""}」を削除します。この操作は取り消せません。`}
      onDeleteConfirm={handleDeleteConfirm}
      onDeleteCancel={handleDeleteCancel}
    >
      {isLoading ? (
        <div className="text-sm text-muted-foreground py-8 text-center">読み込み中...</div>
      ) : (
        <DataTable
          columns={COLUMNS}
          data={filteredGroups}
          emptyMessage="権限グループがありません。「新規登録」から追加してください。"
          renderRow={(g: PermissionGroup) => (
            <GroupRow key={g.id} group={g} onEdit={handleEdit} />
          )}
        />
      )}
    </MasterListPage>
  );
}
