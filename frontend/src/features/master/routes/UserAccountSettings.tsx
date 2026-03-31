import { ICON } from "@/lib/design-tokens";
import { memo, useCallback, useState, useTransition, useMemo } from "react";
import { Users, ChevronRight, X } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/features/auth/hooks/use-auth";
import { PermissionGroupSelector } from "@/features/master/components/PermissionGroupSelector";
import {
  useGetUsers,
  useGetUser,
  useSetUserPermissionGroups,
} from "@/features/master/api/user-accounts";
import type { UserAccountItem } from "@/features/master/api/user-accounts";
import { useGetPermissionGroups } from "@/features/master/api/permission-groups/permission-groups";

const USER_TYPE_LABELS: Record<string, string> = {
  system_admin: "運営管理者",
  clinic_admin: "医院管理者",
  staff: "スタッフ",
};

const ADMIN_USER_TYPES = new Set(["system_admin", "clinic_admin"]);

// ─────────────────────────────────────────────────
// Permission group editor panel
// ─────────────────────────────────────────────────

interface UserPermissionPanelProps {
  user: UserAccountItem;
  clinicId: string;
  onClose: () => void;
}

const UserPermissionPanel = memo(function UserPermissionPanel({
  user,
  clinicId,
  onClose,
}: UserPermissionPanelProps) {
  const { user: currentUser, refreshPermissions } = useAuth();
  const currentUserId = currentUser?.id;
  const { data: detail, isLoading: isDetailLoading } = useGetUser(user.id);
  const { data: groups = [] } = useGetPermissionGroups(clinicId);
  const setGroupsMutation = useSetUserPermissionGroups(clinicId);

  const [selectedGroupIds, setSelectedGroupIds] = useState<number[]>(() =>
    detail?.permissionGroupIds ?? [],
  );

  // detail がロードされたら初期値を同期
  const [initialized, setInitialized] = useState(false);
  if (detail && !initialized) {
    setSelectedGroupIds(detail.permissionGroupIds);
    setInitialized(true);
  }

  const [isSavePending, startSaveTransition] = useTransition();

  const isAdmin = ADMIN_USER_TYPES.has(user.user_type);

  const handleSave = useCallback(() => {
    startSaveTransition(async () => {
      try {
        await setGroupsMutation.mutateAsync({ userId: user.id, groupIds: selectedGroupIds });
        // 自分自身の権限グループが変更された場合は即時キャッシュ更新
        if (currentUserId === user.id) {
          await refreshPermissions();
        }
        toast.success("権限グループを更新しました");
        onClose();
      } catch {
        toast.error("更新に失敗しました");
      }
    });
  }, [user.id, selectedGroupIds, setGroupsMutation, onClose, currentUserId, refreshPermissions]);

  const handleGroupsChange = useCallback((ids: number[]) => {
    setSelectedGroupIds(ids);
  }, []);

  return (
    <div className="border rounded-lg p-4 bg-background">
      <div className="flex items-center justify-between mb-4">
        <div>
          <p className="font-medium text-sm">{user.display_name}</p>
          <p className="text-xs text-muted-foreground">{user.email}</p>
        </div>
        <Button variant="ghost" size="icon" className="h-7 w-7" onClick={onClose}>
          <X className={ICON.action} />
        </Button>
      </div>

      <div className="mb-4">
        <p className="text-xs font-medium text-muted-foreground mb-2">権限グループ</p>
        {isAdmin ? (
          <p className="text-sm text-muted-foreground">
            管理者アカウントはすべての権限を持ちます（グループ設定不要）。
          </p>
        ) : isDetailLoading ? (
          <p className="text-sm text-muted-foreground">読み込み中...</p>
        ) : (
          <PermissionGroupSelector
            availableGroups={groups}
            selectedGroupIds={selectedGroupIds}
            onChange={handleGroupsChange}
          />
        )}
      </div>

      {!isAdmin ? (
        <div className="flex justify-end gap-2">
          <Button variant="outline" size="sm" onClick={onClose} disabled={isSavePending}>
            キャンセル
          </Button>
          <Button size="sm" onClick={handleSave} disabled={isSavePending || isDetailLoading}>
            {isSavePending ? "保存中..." : "保存"}
          </Button>
        </div>
      ) : null}
    </div>
  );
});

// ─────────────────────────────────────────────────
// User row
// ─────────────────────────────────────────────────

interface UserRowProps {
  user: UserAccountItem;
  isSelected: boolean;
  onSelect: (user: UserAccountItem) => void;
}

const UserRow = memo(function UserRow({ user, isSelected, onSelect }: UserRowProps) {
  const handleClick = useCallback(() => {
    onSelect(user);
  }, [user, onSelect]);

  return (
    <button
      type="button"
      className={`w-full flex items-center gap-3 p-3 rounded-lg text-left transition-colors ${
        isSelected ? "bg-muted" : "hover:bg-muted/50"
      }`}
      onClick={handleClick}
    >
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium truncate">{user.display_name}</p>
        <p className="text-xs text-muted-foreground truncate">{user.email}</p>
      </div>
      <span className="text-xs text-muted-foreground flex-shrink-0">
        {USER_TYPE_LABELS[user.user_type] ?? user.user_type}
      </span>
      <ChevronRight className={`${ICON.action} text-muted-foreground flex-shrink-0`} />
    </button>
  );
});

// ─────────────────────────────────────────────────
// Main page
// ─────────────────────────────────────────────────

export function UserAccountSettings() {
  const { currentClinicId } = useAuth();
  const clinicId = currentClinicId ?? "";

  const { data: users = [], isLoading } = useGetUsers(clinicId);
  const [selectedUser, setSelectedUser] = useState<UserAccountItem | null>(null);

  const handleSelectUser = useCallback((user: UserAccountItem) => {
    setSelectedUser((prev) => (prev?.id === user.id ? null : user));
  }, []);

  const handleClosePanel = useCallback(() => {
    setSelectedUser(null);
  }, []);

  const userRows = useMemo(
    () =>
      users.map((u) => (
        <UserRow
          key={u.id}
          user={u}
          isSelected={selectedUser?.id === u.id}
          onSelect={handleSelectUser}
        />
      )),
    [users, selectedUser, handleSelectUser],
  );

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <div className="flex items-center gap-2 mb-6">
        <Users className={`${ICON.page} text-muted-foreground`} />
        <h1 className="text-lg font-semibold">ユーザーアカウント管理</h1>
      </div>

      {isLoading ? (
        <div className="text-sm text-muted-foreground py-8 text-center">読み込み中...</div>
      ) : users.length === 0 ? (
        <div className="text-sm text-muted-foreground py-8 text-center">
          ユーザーが登録されていません。
        </div>
      ) : (
        <div className="space-y-1">
          {userRows}
        </div>
      )}

      {selectedUser !== null ? (
        <div className="mt-4">
          <UserPermissionPanel
            key={selectedUser.id}
            user={selectedUser}
            clinicId={clinicId}
            onClose={handleClosePanel}
          />
        </div>
      ) : null}
    </div>
  );
}
