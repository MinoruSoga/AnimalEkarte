import type { ReactNode } from "react";
import type { Resource } from "@/types/generated/models";
import { useAuth } from "@/features/auth";
import type { ResourceAction } from "@/features/auth/types";

interface RequirePermissionProps {
  resource: Resource;
  action?: ResourceAction;
  children: ReactNode;
  /** 権限がない場合に表示するフォールバック。省略時は AccessDenied を表示 */
  fallback?: ReactNode;
}

/**
 * 指定リソース×アクションの権限がない場合に fallback を表示する。
 * デフォルト action は "view"（ページアクセス制御に使用）。
 * system_admin / clinic_admin は常に children を表示する。
 */
export function RequirePermission({
  resource,
  action = "view",
  children,
  fallback,
}: RequirePermissionProps) {
  const { hasPermission } = useAuth();

  if (!hasPermission(resource, action)) {
    return fallback !== undefined ? <>{fallback}</> : <AccessDenied />;
  }

  return <>{children}</>;
}

function AccessDenied() {
  return (
    <div className="flex flex-col items-center justify-center h-full gap-4 text-center p-8">
      <p className="text-lg font-medium text-foreground">アクセス権限がありません</p>
      <p className="text-sm text-muted-foreground">
        このページを表示するための権限が付与されていません。
      </p>
    </div>
  );
}
