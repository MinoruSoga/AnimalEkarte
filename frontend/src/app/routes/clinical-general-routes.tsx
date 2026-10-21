import { Outlet, type RouteObject } from "react-router";

import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { RequirePermission } from "@/components/shared/RequirePermission";
import { ResourceOwners, ResourceReception, ResourceReservations } from "@/types/generated/models";

export const clinicalGeneralRoutes: RouteObject[] = [
  // ── Reception（当日の受付）────────────────────────────────────
  {
    path: "/",
    element: (
      <RequirePermission resource={ResourceReception}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { Reception } = await import("@/features/reception");
          return { Component: Reception };
        },
      },
    ],
  },

  // ── Owners ───────────────────────────────────────────────────
  {
    // EMR-463: 親の選択医院 view wrapper は撤去する。一覧 index の入口判定は
    // ownersLoader が同一実行で取得する owners 専有 scope／listState が担う
    // （guard 専用の別 GET /me、古い profile cache の二重判定は行わない）。
    path: "/owners",
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        // 一覧 index: OWNERS_CLINIC_SCOPE_ACTIVE=true なら guard なしで通し、
        // 非活性時（テスト専用の退避経路）は従来どおり owners:view を要求する。
        // guard は lazy 内で読み込み、owners feature を初期 bundle へ静的 import しない。
        index: true,
        lazy: async () => {
          const [{ OwnersListPage }, { ownersLoader, OwnersListRouteGuard }] = await Promise.all([
            import("@/app/pages/OwnersListPage"),
            import("@/features/owners"),
          ]);
          function OwnersListEntry() {
            return (
              <OwnersListRouteGuard>
                <OwnersListPage />
              </OwnersListRouteGuard>
            );
          }
          return { Component: OwnersListEntry, loader: ownersLoader };
        },
      },
      {
        // BUG-020: create 権限がないユーザーは新規作成フォームへアクセス不可
        // EMR-463: 親の view wrapper 撤去に伴い、view＋create の両要求をこの枝で明示する。
        path: "new",
        element: (
          <RequirePermission resource={ResourceOwners}>
            <RequirePermission resource={ResourceOwners} action="create">
              <Outlet />
            </RequirePermission>
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { OwnerFormPage } = await import("@/app/pages/OwnerFormPage");
              return { Component: OwnerFormPage };
            },
          },
        ],
      },
      {
        // EMR-463: 詳細枝は従来どおり owners:view を要求する。
        // element と lazy Component は同一 route に置かない（BUG-20260906-003）。
        path: ":id",
        element: (
          <RequirePermission resource={ResourceOwners}>
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const [{ OwnerFormPage }, { ownerLoader }] = await Promise.all([
                import("@/app/pages/OwnerFormPage"),
                import("@/features/owners"),
              ]);
              return { Component: OwnerFormPage, loader: ownerLoader };
            },
          },
        ],
      },
    ],
  },

  // ── Aggregation ──────────────────────────────────────────────
  {
    path: "/aggregation",
    element: (
      <RequirePermission resource={ResourceOwners}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { AggregationDashboardPage } = await import("@/features/aggregation");
          return { Component: AggregationDashboardPage };
        },
      },
    ],
  },

  // ── Reservations ─────────────────────────────────────────────
  {
    path: "/reservations",
    element: (
      <RequirePermission resource={ResourceReservations}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { ReservationsPage } = await import("@/app/pages/ReservationsPage");
          return { Component: ReservationsPage };
        },
      },
    ],
  },
];
