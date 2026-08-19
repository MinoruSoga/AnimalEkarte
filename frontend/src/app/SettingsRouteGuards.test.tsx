import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { Suspense } from "react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/hooks/auth-context";
import { appRoutes } from "@/app/routes/app-routes";
import type { ResourceAction } from "@/types/auth";

// AccountingRouteGuards と同じ理由で全 lazy module を mock し、lazy resolve 時間が
// findByText の timeout を超えないようにする。
// see STG-BLOCKER-001 (AccountingRouteGuards.test.tsx)
vi.mock("@/features/closing-settings", () => ({
  ClosingSettingsPage: () => null,
}));
vi.mock("@/features/master", () => ({
  PaymentMethodSettings: () => null,
  // MasterSettingsIndex 等、他の lazy が同じ index.ts 経由で読み込まれるが
  // /settings/closing-time と /settings/payment-methods の denied は素通りで構わない。
  MasterSettingsIndex: () => null,
  StaffSettings: () => null,
  TreatmentPlanMaster: () => null,
  DiagnosisSettings: () => null,
  AnimalSpeciesSettings: () => null,
  TrimmingSettings: () => null,
  MedicineSettings: () => null,
  ReservationTypeSettings: () => null,
  HospitalizationSettings: () => null,
  CageSettings: () => null,
  MerchandiseItemSettings: () => null,
  InsuranceSettings: () => null,
  OccupationSettings: () => null,
  PermissionGroupSettings: () => null,
  InterviewTemplateSettings: () => null,
  ChiefComplaintSettings: () => null,
  LabDeviceItemMasterSettings: () => null,
}));

/**
 * /settings/closing-time と /settings/payment-methods の権限ガード振る舞いテスト
 *
 * STG-BLOCKER-002: 2 ルートが RequirePermission を欠落していたため、URL 直叩きで
 * 権限なしユーザーが画面骨格に到達できる問題を修正した。本テストで回帰防止する。
 *
 * 権限設計:
 * - `/settings/closing-time`     → ResourceClosingSettings:view
 * - `/settings/payment-methods`  → ResourcePaymentMethod:view
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

afterEach(() => {
  cleanup();
});

const FIND_TIMEOUT = { timeout: 3000 };

describe("/settings — 権限ガード振る舞いテスト (STG-BLOCKER-002)", () => {
  const deniedCases: [string, PermGrant[]][] = [
    // 権限なし → AccessDenied （URL 直叩き想定）
    ["/settings/closing-time",    []],
    ["/settings/payment-methods", []],
    ["/settings/lab-device-item-masters", []],
    // 別 resource の view のみ → 該当 resource の view なしで AccessDenied
    ["/settings/closing-time",    [["master-payment-method", "view"]]],
    ["/settings/payment-methods", [["closing-settings",      "view"]]],
    ["/settings/lab-device-item-masters", [["closing-settings", "view"]]],
  ];

  it.each(deniedCases)("権限不足: %s (%j) → AccessDenied", async (path, grants) => {
    renderRoute(path, grants);
    await screen.findByText(/アクセス権限がありません/, undefined, FIND_TIMEOUT);
  });

  const allowedCases: [string, PermGrant[]][] = [
    ["/settings/closing-time",    [["closing-settings",      "view"]]],
    ["/settings/payment-methods", [["master-payment-method", "view"]]],
    ["/settings/lab-device-item-masters", [["lab-import", "view"]]],
  ];

  it.each(allowedCases)("権限あり: %s → Layout 描画・アクセス拒否なし", async (path, grants) => {
    renderRoute(path, grants);
    await screen.findByRole("main", undefined, FIND_TIMEOUT);
    expect(screen.queryByText(/アクセス権限がありません/)).not.toBeInTheDocument();
  });
});
