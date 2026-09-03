import { Outlet, type RouteObject } from "react-router";

import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { RequirePermission } from "@/components/shared/RequirePermission";
import { ResourceEstimates, ResourceInventory, ResourceShifts } from "@/types/generated/models";

export const clinicalBusinessRoutes: RouteObject[] = [
  // ── Inventory ────────────────────────────────────────────────
  {
    path: "/inventory",
    element: (
      <RequirePermission resource={ResourceInventory}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { InventoryList } = await import("@/features/inventory");
          return { Component: InventoryList };
        },
      },
      {
        // BUG-020: create 権限ガード
        path: "new",
        element: (
          <RequirePermission resource={ResourceInventory} action="create">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { InventoryForm } = await import("@/features/inventory");
              return { Component: InventoryForm };
            },
          },
        ],
      },
      {
        path: ":id",
        lazy: async () => {
          const { InventoryForm } = await import("@/features/inventory");
          return { Component: InventoryForm };
        },
      },
    ],
  },

  // ── Estimates ────────────────────────────────────────────────
  {
    path: "/estimates",
    element: (
      <RequirePermission resource={ResourceEstimates}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { EstimateList } = await import("@/features/estimates");
          return { Component: EstimateList };
        },
      },
      {
        // BUG-020: create 権限ガード
        path: "new",
        element: (
          <RequirePermission resource={ResourceEstimates} action="create">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { EstimateForm } = await import("@/features/estimates");
              return { Component: EstimateForm };
            },
          },
        ],
      },
      {
        path: ":id",
        lazy: async () => {
          const { EstimateDetail } = await import("@/features/estimates");
          return { Component: EstimateDetail };
        },
      },
      {
        // BUG-020: edit 権限ガード
        path: ":id/edit",
        element: (
          <RequirePermission resource={ResourceEstimates} action="edit">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { EstimateForm } = await import("@/features/estimates");
              return { Component: EstimateForm };
            },
          },
        ],
      },
    ],
  },

  // ── Shifts ───────────────────────────────────────────────────
  {
    path: "/shifts",
    element: (
      <RequirePermission resource={ResourceShifts}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { ShiftCalendarPage } = await import("@/features/shifts");
          return { Component: ShiftCalendarPage };
        },
      },
    ],
  },
];
