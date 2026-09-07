import { Suspense } from "react";
import { createBrowserRouter, Outlet } from "react-router";

import { RootErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { AuthProvider } from "@/features/auth/provider";

import { appRoutes } from "./routes/app-routes";
import { RootHydrateFallback } from "./root-hydrate-fallback";

export const router = createBrowserRouter([
  {
    // AuthProvider をアプリ全体に配置。
    // これにより /login でも useAuth() が使用可能になり、
    // LoginForm で login() を直接呼び出してから navigate() できる。
    HydrateFallback: RootHydrateFallback,
    element: (
      <Suspense fallback={null}>
        <AuthProvider>
          <Outlet />
        </AuthProvider>
      </Suspense>
    ),
    errorElement: <RootErrorBoundary />,
    children: appRoutes,
  },
]);
