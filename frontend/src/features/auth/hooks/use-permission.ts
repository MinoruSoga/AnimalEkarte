import type { Resource } from "@/types/generated/models";
import { useAuth } from "@/features/auth/hooks/use-auth";

export interface UsePermissionResult {
  canView: boolean;
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}

/**
 * 現在のユーザーが指定リソースに対して持つ権限を返す。
 * system_admin / clinic_admin は常に true。
 * staff はグループ UNION から計算された実効権限（バックエンド計算済み）を使用。
 *
 * @param resource - リソース識別子（models.ts の Resource 定数を使用）
 */
export function usePermission(resource: Resource): UsePermissionResult {
  const { hasPermission } = useAuth();
  return {
    canView: hasPermission(resource, "view"),
    canCreate: hasPermission(resource, "create"),
    canEdit: hasPermission(resource, "edit"),
    canDelete: hasPermission(resource, "delete"),
  };
}
