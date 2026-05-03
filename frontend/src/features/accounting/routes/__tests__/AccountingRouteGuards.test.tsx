import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { Suspense } from "react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/contexts/auth-context";
import { appRoutes } from "@/app/router";
import type { ResourceAction } from "@/types/auth";

vi.mock("@/features/accounting", () => ({
  AccountingList: () => null,
  AccountingPetSelection: () => null,
}));

vi.mock("@/app/pages/AccountingDetailPage", () => ({
  AccountingDetailPage: () => null,
}));

/**
 * /accounting 配下 route の権限ガード振る舞いテスト
 *
 * 権限分離の設計:
 * - 親 `/accounting`            → ResourceAccounting:view
 * - `/accounting/select-pet`    → ResourceAccounting:view(親) + :create(子)
 * - `/accounting/new`           → ResourceAccounting:view(親) + :create(子)
 * - `/accounting/close`         → ResourceCashRegisterClose:view (独立 route)
 * - `/accounting/close/history` → ResourceCashRegisterClose:view (独立 route)
 * - `/accounting/reports`       → ResourceAccountingReports:view (独立 route)
 *
 * 実際の appRoutes を使用し、lazy ロードを vi.mock でスタブする。
 * RequirePermission / resource 定数はプロダクションコードそのものを使用する。
 */

const CLINIC_ID = "clinic-test-1";
type PermGrant = [string, ResourceAction];

function makeHasPermission(grants: PermGrant[]) {
  return (resource: string, action: ResourceAction): boolean =>
    grants.some(([r, a]) => r === resource && a === action);
}

function makeAuthCtx(grants: PermGrant[]) {
  return {
    user: null,
    currentClinicId: CLINIC_ID,
    isAuthenticated: true,
    isLoading: false,
    isSwitchingClinic: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: makeHasPermission(grants),
    refreshPermissions: async () => {},
  };
}

function renderRoute(path: string, grants: PermGrant[]) {
  const router = createMemoryRouter(appRoutes, { initialEntries: [path] });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <AuthContext.Provider value={makeAuthCtx(grants)}>
      <QueryClientProvider client={qc}>
        <Suspense fallback={null}>
          <RouterProvider router={router} />
        </Suspense>
      </QueryClientProvider>
    </AuthContext.Provider>
  );
}

describe("/accounting — 権限ガード振る舞いテスト", () => {
  const deniedCases: [string, PermGrant[]][] = [
    // 親ガード: accounting:view なし → AccessDenied
    ["/accounting",               []                         ],
    // 子ガード: accounting:view あり・accounting:create なし → child AccessDenied
    ["/accounting/select-pet",    [["accounting", "view"]]   ],
    ["/accounting/new",           [["accounting", "view"]]   ],
    // 独立 route: 必要権限なし → AccessDenied
    ["/accounting/close",         []                         ],
    ["/accounting/close/history", []                         ],
    ["/accounting/reports",       []                         ],
  ];

  it.each(deniedCases)("権限不足: %s → AccessDenied", async (path, grants) => {
    renderRoute(path, grants);
    await screen.findByText(/アクセス権限がありません/);
  });

  const allowedCases: [string, PermGrant[]][] = [
    ["/accounting",         [["accounting",           "view"]]],
    ["/accounting/close",   [["cash-register-close",  "view"]]],
    ["/accounting/reports", [["accounting-reports",   "view"]]],
  ];

  it.each(allowedCases)("権限あり: %s → Layout 描画・アクセス拒否なし", async (path, grants) => {
    renderRoute(path, grants);
    await screen.findByRole("main");
    expect(screen.queryByText(/アクセス権限がありません/)).not.toBeInTheDocument();
  });
});
