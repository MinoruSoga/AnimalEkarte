import type { ReactNode } from "react";

import { RequirePermission } from "@/components/shared/RequirePermission";
import { ResourceOwners } from "@/types/generated/models";

import { OWNERS_CLINIC_SCOPE_ACTIVE } from "../lib/owners-clinic-scope-activation";

interface OwnersListRouteGuardProps {
  children: ReactNode;
}

/**
 * EMR-463: /owners 一覧 index の入口判定。
 * 活性時（OWNERS_CLINIC_SCOPE_ACTIVE=true）は ownersLoader が同一実行で取得した
 * owners 専有 scope／listState が fail-closed に入口判定するため、追加の view guard
 * を掛けない（guard 専用の別 GET /me、古い profile cache による二重判定は行わない）。
 * 非活性時（テスト専用の退避経路）は従来どおり owners:view を要求する。
 */
export function OwnersListRouteGuard({ children }: OwnersListRouteGuardProps) {
  if (!OWNERS_CLINIC_SCOPE_ACTIVE) {
    return <RequirePermission resource={ResourceOwners}>{children}</RequirePermission>;
  }
  return <>{children}</>;
}
