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
    path: "/owners",
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
          const [{ OwnersListPage }, { ownersLoader }] = await Promise.all([
            import("@/app/pages/OwnersListPage"),
            import("@/features/owners"),
          ]);
          return { Component: OwnersListPage, loader: ownersLoader };
        },
      },
      {
        // BUG-020: create 権限がないユーザーは新規作成フォームへアクセス不可
        path: "new",
        element: (
          <RequirePermission resource={ResourceOwners} action="create">
            <Outlet />
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
        path: ":id",
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
