import { Outlet, type RouteObject } from "react-router";

import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { RequirePermission } from "@/components/shared/RequirePermission";
import {
  ResourceAccounting,
  ResourceAccountingReports,
  ResourceCashRegisterClose,
} from "@/types/generated/models";

export const accountingRoutes: RouteObject[] = [
  // ── Accounting ───────────────────────────────────────────────
  {
    path: "/accounting",
    element: (
      <RequirePermission resource={ResourceAccounting}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { AccountingList } = await import("@/features/accounting");
          return { Component: AccountingList };
        },
      },
      {
        path: "select-pet",
        element: (
          <RequirePermission resource={ResourceAccounting} action="create">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { AccountingPetSelection } = await import("@/features/accounting");
              return { Component: AccountingPetSelection };
            },
          },
        ],
      },
      {
        // BUG-020: create 権限ガード
        path: "new",
        element: (
          <RequirePermission resource={ResourceAccounting} action="create">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { AccountingDetailPage } = await import("@/app/pages/AccountingDetailPage");
              return { Component: AccountingDetailPage };
            },
          },
        ],
      },
      {
        path: ":id",
        lazy: async () => {
          const { AccountingDetailPage } = await import("@/app/pages/AccountingDetailPage");
          return { Component: AccountingDetailPage };
        },
      },
    ],
  },

  // ── CashRegisterClose（レジ締め / 締め履歴） ────────────────
  {
    path: "/accounting/close",
    element: (
      <RequirePermission resource={ResourceCashRegisterClose}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { CashRegisterClosePage } = await import("@/features/cash-register");
          return { Component: CashRegisterClosePage };
        },
      },
    ],
  },
  {
    path: "/accounting/close/history",
    element: (
      <RequirePermission resource={ResourceCashRegisterClose}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { CashRegisterHistoryPage } = await import("@/features/cash-register");
          return { Component: CashRegisterHistoryPage };
        },
      },
    ],
  },

  // ── AccountingReports（月次集計レポート） ────────────────────
  {
    path: "/accounting/reports",
    element: (
      <RequirePermission resource={ResourceAccountingReports}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { AccountingReportsPage } = await import("@/features/accounting-reports");
          return { Component: AccountingReportsPage };
        },
      },
    ],
  },
];
