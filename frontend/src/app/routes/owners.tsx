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
          const { OwnerFormPage } = await import("@/app/pages/OwnerFormPage");
          return { Component: OwnerFormPage };
        },
      },
      {
        path: ":id",
        lazy: async () => {
          const [{ OwnerFormPage }, { ownerLoader }] = await Promise.all([
            import("@/app/pages/OwnerFormPage"),
            import("@/features/owners/loaders"),
          ]);
          return { Component: OwnerFormPage, loader: ownerLoader };
        },
      },
    ],
  },
];
