import { Outlet, type RouteObject } from "react-router";

import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { RequirePermission } from "@/components/shared/RequirePermission";
import {
  ResourceHospitalSettings,
  ResourceLstepAnalytics,
  ResourceReservations,
} from "@/types/generated/models";

export const operationsRoutes: RouteObject[] = [
  {
    path: "/lstep",
    element: (
      <RequirePermission resource={ResourceHospitalSettings}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        path: "checkup-sync",
        lazy: async () => {
          const { CheckupSyncPage } = await import("@/features/lstep");
          return { Component: CheckupSyncPage };
        },
      },
      {
        path: "delivery-monitor",
        lazy: async () => {
          const { LstepDeliveryMonitorPage } = await import("@/features/lstep");
          return { Component: LstepDeliveryMonitorPage };
        },
      },
      {
        path: "analytics",
        element: (
          <RequirePermission resource={ResourceLstepAnalytics}>
            <Outlet />
          </RequirePermission>
        ),
        children: [{
          index: true,
          lazy: async () => {
            const { LstepAnalyticsPage } = await import("@/features/lstep");
            return { Component: LstepAnalyticsPage };
          },
        }],
      },
    ],
  },
  {
    path: "/line-reservation",
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
          const { LineReservationSettings } = await import("@/features/line-reservation");
          return { Component: LineReservationSettings };
        },
      },
      {
        path: "settings",
        lazy: async () => {
          const { LineReservationSettings } = await import("@/features/line-reservation");
          return { Component: LineReservationSettings };
        },
      },
      {
        path: "page-editor",
        lazy: async () => {
          const { LineReservationPageEditor } = await import("@/features/line-reservation");
          return { Component: LineReservationPageEditor };
        },
      },
    ],
  },
  {
    path: "/settings/clinic",
    element: (
      <RequirePermission resource={ResourceHospitalSettings}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { ClinicMasterSettings } = await import("@/features/clinic-settings");
          return { Component: ClinicMasterSettings };
        },
      },
    ],
  },
  {
    path: "/manual",
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { ManualPage } = await import("@/features/manual");
          return { Component: ManualPage };
        },
      },
      {
        path: ":category/:slug",
        lazy: async () => {
          const { ManualPage } = await import("@/features/manual");
          return { Component: ManualPage };
        },
      },
    ],
  },
];
