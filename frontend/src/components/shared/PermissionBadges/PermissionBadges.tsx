import { memo } from "react";
import { usePermission } from "@/features/auth/hooks/use-permission";
import { C } from "@/lib/design-tokens";
import type { Resource } from "@/types/generated/models";

interface PermissionBadgesProps {
  resource: Resource;
}

interface BadgeDef {
  key: string;
  label: string;
  active: boolean;
}

/**
 * 現在のユーザーが持つ権限をバッジで表示する。
 * 全権限（CRUD全て）を持っている場合は何も表示しない。
 * 一部制限がある場合のみ、持っている権限をバッジ表示する。
 * 閲覧のみの場合は「閲覧のみ」と1バッジで表示。
 */
export const PermissionBadges = memo(function PermissionBadges({
  resource,
}: PermissionBadgesProps) {
  const { canView, canCreate, canEdit, canDelete } = usePermission(resource);

  // 全権限ありなら非表示
  if (canView && canCreate && canEdit && canDelete) return null;

  // 閲覧のみの場合は専用表示
  if (canView && !canCreate && !canEdit && !canDelete) {
    return (
      <div className="flex items-center gap-1">
        <span
          className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${C.text50} bg-neutral-100 border border-neutral-200`}
        >
          閲覧のみ
        </span>
      </div>
    );
  }

  const badges: BadgeDef[] = [
    { key: "view", label: "閲覧", active: canView },
    { key: "create", label: "作成", active: canCreate },
    { key: "edit", label: "編集", active: canEdit },
    { key: "delete", label: "削除", active: canDelete },
  ];

  const activeBadges = badges.filter((b) => b.active);

  // 権限が全くない場合（通常は AccessDenied で弾かれるが念のため）
  if (activeBadges.length === 0) return null;

  return (
    <div className="flex items-center gap-1">
      {activeBadges.map((badge) => (
        <span
          key={badge.key}
          className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${C.text50} bg-neutral-100 border border-neutral-200`}
        >
          {badge.label}
        </span>
      ))}
    </div>
  );
});
