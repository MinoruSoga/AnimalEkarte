import type { RouteObject } from "react-router";
import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";

export const ownersRoutes: RouteObject[] = [
  {
    path: "/owners",
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const [{ OwnersList }, { ownersLoader }] = await Promise.all([
            import("@/features/owners/routes/OwnersList"),
            import("@/features/owners/loaders"),
          ]);
          return { Component: OwnersList, loader: ownersLoader };
        },
      },
      {
        path: "new",
        lazy: async () => {
          const { OwnerForm } = await import("@/features/owners/routes/OwnerForm");
          return { Component: OwnerForm };
        },
      },
      {
        path: ":id",
        lazy: async () => {
          const [{ OwnerForm }, { ownerLoader }] = await Promise.all([
            import("@/features/owners/routes/OwnerForm"),
            import("@/features/owners/loaders"),
          ]);
          return { Component: OwnerForm, loader: ownerLoader };
        },
      },
    ],
  },
];
