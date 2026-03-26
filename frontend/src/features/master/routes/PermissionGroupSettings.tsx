import { memo, useCallback, useState, useTransition, useMemo } from "react";
import { Shield, Plus, Pencil, Trash2, ChevronUp } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { useAuth } from "@/features/auth/hooks/use-auth";
import { PermissionRuleTable } from "@/features/master/components/PermissionRuleTable";
import { PERMISSION_RESOURCES } from "@/features/master/types/permission-resources";
import {
  useGetPermissionGroups,
  useCreatePermissionGroup,
  useUpdatePermissionGroup,
  useDeletePermissionGroup,
  useSetPermissionGroupRules,
} from "@/features/master/api/permission-groups/permission-groups";
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
// Form subcomponent
// ─────────────────────────────────────────────────

interface GroupFormProps {
  group: PermissionGroup | null;
  clinicId: string;
  onClose: () => void;
}

const GroupForm = memo(function GroupForm({ group, clinicId, onClose }: GroupFormProps) {
  const isNew = group === null;

  const [name, setName] = useState(() => group?.name ?? "");
  const [description, setDescription] = useState(() => group?.description ?? "");
  const [color, setColor] = useState(() => group?.color ?? "#6B7280");
  const [rules, setRules] = useState<RuleInput[]>(() =>
    groupRulesToRuleInputs(group?.rules),
  );

  const createMutation = useCreatePermissionGroup();
  const updateMutation = useUpdatePermissionGroup(clinicId);
  const setRulesMutation = useSetPermissionGroupRules(clinicId);

  const [isSavePending, startSaveTransition] = useTransition();

  const handleNameChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setName(e.target.value);
  }, []);

  const handleDescriptionChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setDescription(e.target.value);
  }, []);

  const handleColorChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setColor(e.target.value);
  }, []);

  const handleRulesChange = useCallback((updated: RuleInput[]) => {
    setRules(updated);
  }, []);

  const handleSave = useCallback(() => {
    if (!name.trim()) {
      toast.error("グループ名を入力してください");
      return;
    }
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

  return (
    <div className="border rounded-lg p-4 mb-4 bg-muted/30">
      <div className="grid grid-cols-1 gap-4 mb-4">
        <div className="grid grid-cols-3 gap-4">
          <div className="col-span-2 space-y-1.5">
            <Label htmlFor="group-name">グループ名</Label>
            <Input
              id="group-name"
              value={name}
              onChange={handleNameChange}
              placeholder="例: 受付スタッフ"
              className="h-8 text-sm"
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="group-color">カラー</Label>
            <div className="flex items-center gap-2">
              <input
                id="group-color"
                type="color"
                value={color}
                onChange={handleColorChange}
                className="w-8 h-8 rounded border cursor-pointer"
              />
              <span className="text-xs text-muted-foreground">{color}</span>
            </div>
          </div>
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="group-description">説明（任意）</Label>
          <Textarea
            id="group-description"
            value={description}
            onChange={handleDescriptionChange}
            placeholder="このグループの役割・用途を入力"
            className="text-sm resize-none h-16"
          />
        </div>
      </div>

      <div className="border rounded-md p-3 bg-background mb-4">
        <p className="text-xs font-medium text-muted-foreground mb-3">ページ権限</p>
        <PermissionRuleTable rules={rules} onChange={handleRulesChange} />
      </div>

      <div className="flex justify-end gap-2">
        <Button variant="outline" size="sm" onClick={onClose} disabled={isSavePending}>
          キャンセル
        </Button>
        <Button size="sm" onClick={handleSave} disabled={isSavePending}>
          {isSavePending ? "保存中..." : "保存"}
        </Button>
      </div>
    </div>
  );
});

// ─────────────────────────────────────────────────
// Group card subcomponent
// ─────────────────────────────────────────────────

interface GroupCardProps {
  group: PermissionGroup;
  clinicId: string;
  isExpanded: boolean;
  onToggleExpand: (id: number) => void;
  onDelete: (group: PermissionGroup) => void;
}

const GroupCard = memo(function GroupCard({
  group,
  clinicId,
  isExpanded,
  onToggleExpand,
  onDelete,
}: GroupCardProps) {
  const handleToggle = useCallback(() => {
    onToggleExpand(group.id);
  }, [group.id, onToggleExpand]);

  const handleDelete = useCallback(() => {
    onDelete(group);
  }, [group, onDelete]);

  const handleClose = useCallback(() => {
    onToggleExpand(group.id);
  }, [group.id, onToggleExpand]);

  return (
    <div className="border rounded-lg overflow-hidden">
      <div className="flex items-center gap-3 p-3 bg-background">
        <div
          className="w-3 h-3 rounded-full flex-shrink-0"
          style={{ backgroundColor: group.color ?? "#6B7280" }}
        />
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium truncate">{group.name}</p>
          {group.description ? (
            <p className="text-xs text-muted-foreground truncate">{group.description}</p>
          ) : null}
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            onClick={handleToggle}
            aria-label={isExpanded ? "閉じる" : "編集"}
          >
            {isExpanded ? (
              <ChevronUp className="h-4 w-4" />
            ) : (
              <Pencil className="h-4 w-4" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-destructive hover:text-destructive"
            onClick={handleDelete}
            aria-label="削除"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>
      {isExpanded ? (
        <div className="border-t p-4">
          <GroupForm group={group} clinicId={clinicId} onClose={handleClose} />
        </div>
      ) : null}
    </div>
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

  const [isCreating, setIsCreating] = useState(false);
  const [expandedGroupId, setExpandedGroupId] = useState<number | null>(null);
  const [isDeletePending, startDeleteTransition] = useTransition();

  const handleCreateOpen = useCallback(() => {
    setIsCreating(true);
    setExpandedGroupId(null);
  }, []);

  const handleCreateClose = useCallback(() => {
    setIsCreating(false);
  }, []);

  const handleToggleExpand = useCallback((id: number) => {
    setExpandedGroupId((prev) => (prev === id ? null : id));
    setIsCreating(false);
  }, []);

  const handleDeleteRequest = useCallback(
    (group: PermissionGroup) => {
      if (!window.confirm(`「${group.name}」を削除しますか？\nこの操作は取り消せません。`)) return;
      startDeleteTransition(async () => {
        try {
          await deleteMutation.mutateAsync(group.id);
          toast.success("権限グループを削除しました");
        } catch {
          toast.error("削除に失敗しました");
        }
      });
    },
    [deleteMutation],
  );

  const groupCards = useMemo(
    () =>
      groups.map((g) => (
        <GroupCard
          key={g.id}
          group={g}
          clinicId={clinicId}
          isExpanded={expandedGroupId === g.id}
          onToggleExpand={handleToggleExpand}
          onDelete={handleDeleteRequest}
        />
      )),
    [groups, clinicId, expandedGroupId, handleToggleExpand, handleDeleteRequest],
  );

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-2">
          <Shield className="h-5 w-5 text-muted-foreground" />
          <h1 className="text-lg font-semibold">権限グループ管理</h1>
        </div>
        <Button size="sm" onClick={handleCreateOpen} disabled={isCreating || isDeletePending}>
          <Plus className="h-4 w-4 mr-1" />
          新規グループ作成
        </Button>
      </div>

      {isCreating ? (
        <GroupForm group={null} clinicId={clinicId} onClose={handleCreateClose} />
      ) : null}

      {isLoading ? (
        <div className="text-sm text-muted-foreground py-8 text-center">読み込み中...</div>
      ) : groups.length === 0 && !isCreating ? (
        <div className="text-sm text-muted-foreground py-8 text-center">
          権限グループがありません。「新規グループ作成」から追加してください。
        </div>
      ) : (
        <div className="space-y-3">{groupCards}</div>
      )}
    </div>
  );
}
