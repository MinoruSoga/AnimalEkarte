import { Outlet, type RouteObject } from "react-router";

import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { RequirePermission } from "@/components/shared/RequirePermission";
import {
  ResourceHospitalSettings,
  ResourceLstepAnalytics,
  ResourceMasterReservationType,
  ResourceReservations,
} from "@/types/generated/models";

export const operationsRoutes: RouteObject[] = [
  {
    // Child-specific guards only (LINE residual FINAL R-06): parent must not AND HospitalSettings
    // onto Analytics-only routes such as delivery-monitor / analytics.
    path: "/lstep",
    element: <Outlet />,
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        path: "checkup-sync",
        element: (
          <RequirePermission resource={ResourceHospitalSettings}>
            <Outlet />
          </RequirePermission>
        ),
        children: [{
          index: true,
          lazy: async () => {
            const { CheckupSyncPage } = await import("@/features/lstep");
            return { Component: CheckupSyncPage };
          },
        }],
      },
      {
        path: "delivery-monitor",
        element: (
          <RequirePermission resource={ResourceLstepAnalytics}>
            <Outlet />
          </RequirePermission>
        ),
        children: [{
          index: true,
          lazy: async () => {
            const { LstepDeliveryMonitorPage } = await import("@/features/lstep");
            return { Component: LstepDeliveryMonitorPage };
          },
        }],
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
    // LINE予約枠（日別の予約可能開始時刻）。
    // API が /v1/masters/reservation-types 配下のため、権限は BE と同じ
    // ResourceMasterReservationType でガードする（ResourceReservations ではない）。
    path: "/line-reservation/slots",
    element: (
      <RequirePermission resource={ResourceMasterReservationType}>
        <Outlet />
      </RequirePermission>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { LineReservationSlotsSettings } = await import("@/features/master");
          return { Component: LineReservationSlotsSettings };
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
  {
    path: "/identity-links",
    errorElement: <RouteErrorBoundary />,
    lazy: async () => {
      const { IdentityLinksPage } = await import("@/features/identity-links");
      return { Component: IdentityLinksPage };
    },
  },
];
