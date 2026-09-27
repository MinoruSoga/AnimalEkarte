import { lazy, Suspense } from "react";
import { type RouteObject } from "react-router";

import { Layout } from "@/components/shared/Layout/Layout";
import { SessionPending } from "@/components/shared/auth/SessionPending";
import { useAuth } from "@/hooks/use-auth";
import { C, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";

import { accountingRoutes } from "./accounting-routes";
import { clinicalRoutes } from "./clinical-routes";
import { operationsRoutes } from "./operations-routes";
import { settingsRoute } from "./settings-routes";

/* bundle-dynamic-imports: ログインページは未認証ユーザー専用。認証済みユーザーのバンドルに含めない */
const Login = lazy(() => import("@/features/auth").then((m) => ({ default: m.Login })));

/* サポートウィジェットはマニュアル MD バンドル（glob eager）を抱えるため、
 * 初期チャンクに含めず認証後に遅延ロードする。 */
const SupportWidget = lazy(() =>
  import("@/features/support").then((m) => ({ default: m.SupportWidget })),
);

const authRoutes: RouteObject[] = [
  {
    path: paths.auth.login.path,
    element: (
      <Suspense fallback={<SessionPending message="画面を読み込んでいます" />}>
        <Login />
      </Suspense>
    ),
  },
  {
    path: paths.auth.forgotPassword.path,
    lazy: async () => {
      const { ForgotPasswordPage } = await import("@/features/auth");
      return { Component: ForgotPasswordPage };
    },
  },
  {
    path: paths.auth.resetPassword.path,
    lazy: async () => {
      const { ResetPasswordPage } = await import("@/features/auth");
      return { Component: ResetPasswordPage };
    },
  },
];

// #158: 飼主単位カルテレポート。別ウィンドウ用に Layout（サイドバー）外のスタンドアロンで登録する。
// 認証ガードと medical-records:view ゲートは OwnerReport 自身が持つ。
const ownerReportRoute: RouteObject = {
  path: paths.owners.detail.report.path,
  lazy: async () => {
    const { OwnerReport } = await import("@/features/owner-report");
    return { Component: OwnerReport };
  },
};

const notFoundRoute: RouteObject = {
  path: "*",
  element: (
    <div className={`flex-1 p-6 flex items-center justify-center ${STYLE.page}`}>
      <p className={C.text50}>ページが見つかりません</p>
    </div>
  ),
};

/**
 * サポートウィジェット（全認証画面共通のフローティングボタン）。
 * Layout 側は feature 非依存（components/shared ルール）を維持するため、
 * 認証ゲートつきでルートシェルと兄弟にマウントする。
 */
function AuthenticatedSupportWidget() {
  const { isAuthenticated, isLoading } = useAuth();
  if (isLoading || !isAuthenticated) return null;
  return (
    <Suspense fallback={null}>
      <SupportWidget />
    </Suspense>
  );
}

// Exported for integration testing (AccountingRouteGuards etc.)
// createMemoryRouter(appRoutes, { initialEntries: [path] }) + AuthContext.Provider で権限ガードを検証できる。
export const appRoutes: RouteObject[] = [
  ...authRoutes,
  ownerReportRoute,
  {
    element: (
      <>
        <Layout />
        <AuthenticatedSupportWidget />
      </>
    ),
    children: [
      ...clinicalRoutes,
      ...accountingRoutes,
      settingsRoute,
      ...operationsRoutes,
      notFoundRoute,
    ],
  },
];
