import type { RouteObject } from "react-router";
import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";

export const inventoryRoutes: RouteObject[] = [
  {
    path: "/inventory",
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { InventoryList } = await import(
            "@/features/inventory/routes/InventoryList"
          );
          return { Component: InventoryList };
        },
      },
      {
        path: "new",
        lazy: async () => {
          const { InventoryForm } = await import(
            "@/features/inventory/routes/InventoryForm"
          );
          return { Component: InventoryForm };
        },
      },
      {
        path: ":id",
        lazy: async () => {
          const { InventoryForm } = await import(
            "@/features/inventory/routes/InventoryForm"
          );
          return { Component: InventoryForm };
        },
      },
    ],
  },
];
