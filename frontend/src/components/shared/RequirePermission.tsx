import { memo, type ReactNode } from "react";
import { C } from "@/lib/design-tokens";
import type { Resource } from "@/types/generated/models";
import { useAuth } from "@/hooks/use-auth";
import type { ResourceAction } from "@/hooks/use-permission";

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
 * isSystemAdmin=true の場合は常に children を表示する。
 */
export const RequirePermission = memo(function RequirePermission({
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
});

function AccessDenied() {
  return (
    <div className="flex flex-col items-center justify-center h-full gap-4 text-center p-8">
      <h1 className={`text-xl font-semibold ${C.text}`}>アクセス権限がありません</h1>
      <p className={`text-sm ${C.text50}`}>
        このページを表示するための権限が付与されていません。
      </p>
    </div>
  );
}
