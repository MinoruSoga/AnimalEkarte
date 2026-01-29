import type { RouteObject } from "react-router";
import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";

export const accountingRoutes: RouteObject[] = [
  {
    path: "/accounting",
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { Accounting } = await import("@/features/accounting/routes/Accounting");
          return { Component: Accounting };
        },
      },
      {
        path: "select-pet",
        lazy: async () => {
          const { AccountingPetSelection } = await import("@/features/accounting/routes/AccountingPetSelection");
          return { Component: AccountingPetSelection };
        },
      },
      {
        path: "new",
        lazy: async () => {
          const { AccountingDetail } = await import("@/features/accounting/routes/AccountingDetail");
          return { Component: AccountingDetail };
        },
      },
      {
        path: ":id",
        lazy: async () => {
          const { AccountingDetail } = await import("@/features/accounting/routes/AccountingDetail");
          return { Component: AccountingDetail };
        },
      },
    ],
  },
];
