import { lazy, Suspense } from "react";
import { type RouteObject } from "react-router";

import { Layout } from "@/components/shared/Layout/Layout";
import { C } from "@/lib/design-tokens";

import { accountingRoutes } from "./accounting-routes";
import { clinicalRoutes } from "./clinical-routes";
import { operationsRoutes } from "./operations-routes";
import { settingsRoute } from "./settings-routes";

/* bundle-dynamic-imports: ログインページは未認証ユーザー専用。認証済みユーザーのバンドルに含めない */
const Login = lazy(() =>
  import("@/features/auth").then((m) => ({ default: m.Login })),
);

const authRoutes: RouteObject[] = [
  {
    path: "/login",
    element: (
      <Suspense fallback={null}>
        <Login />
      </Suspense>
    ),
  },
  {
    path: "/forgot-password",
    lazy: async () => {
      const { ForgotPasswordPage } = await import("@/features/auth");
      return { Component: ForgotPasswordPage };
    },
  },
  {
    path: "/reset-password",
    lazy: async () => {
      const { ResetPasswordPage } = await import("@/features/auth");
      return { Component: ResetPasswordPage };
    },
  },
];

const notFoundRoute: RouteObject = {
  path: "*",
  element: (
    <div className="flex-1 p-5 flex items-center justify-center">
      <p className={C.text50}>ページが見つかりません</p>
    </div>
  ),
};

// Exported for integration testing (AccountingRouteGuards etc.)
// createMemoryRouter(appRoutes, { initialEntries: [path] }) + AuthContext.Provider で権限ガードを検証できる。
export const appRoutes: RouteObject[] = [
  ...authRoutes,
  {
    element: <Layout />,
    children: [
      ...clinicalRoutes,
      ...accountingRoutes,
      settingsRoute,
      ...operationsRoutes,
      notFoundRoute,
    ],
  },
];
