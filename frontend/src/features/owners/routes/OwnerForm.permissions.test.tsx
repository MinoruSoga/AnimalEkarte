import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { Suspense } from "react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/hooks/auth-context";
import { OwnerForm } from "./OwnerForm";
import type { ResourceAction } from "@/types/auth";
import type { Owner } from "@/types/owner";

/**
 * OwnerForm — accounting:view 権限による会計履歴セクション表示制御
 *
 * OwnerForm は編集モード (ownerId あり) + accounting:view 権限がある場合のみ
 * 会計履歴セクション (<h2>会計履歴</h2> + OwnerAccountingHistory) を表示する。
 * この振る舞いを AuthContext.Provider でモックして検証する。
 */

const CLINIC_ID = "clinic-test-1";
const OWNER_ID = "123";
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

const mockOwner: Owner = {
  id: OWNER_ID,
  ownerName: "テスト飼主",
  ownerNameKana: "テストカイヌシ",
  company: "",
  postalCode: "",
  address1: "",
  address2: "",
  homePostalCode: "",
  homeAddress1: "",
  homeAddress2: "",
  phone: "",
  companyPhone: "",
  email: "",
  remarks: "",
  isDangerous: false,
  discountRate: 0,
  membershipType: "non_member",
  deliveryExcluded: false,
  isTransferred: false,
  lstepOptOut: false,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
  pets: [],
};

const ownerBaseGrants: PermGrant[] = [
  ["owners", "view"],
  ["owners", "edit"],
  ["owners", "create"],
  ["discount", "edit"],
];

function renderOwnerForm(options: { canViewAccounting: boolean; path?: string }) {
  const { canViewAccounting, path = `/owners/${OWNER_ID}` } = options;
  const grants: PermGrant[] = [
    ...ownerBaseGrants,
    ...(canViewAccounting ? [["accounting", "view"] as PermGrant] : []),
  ];
  // accountingSection は app 層 (OwnerFormPage) が注入する DI 契約のスタブ。
  // 表示可否の検証対象はあくまで OwnerForm 内の isEdit/canViewAccounting/accountingSection 有無。
  const accountingSection = canViewAccounting ? <div data-testid="accounting-stub" /> : undefined;
  // /owners/new を /owners/:id より先に置く (static match 優先)
  const router = createMemoryRouter(
    [
      {
        path: "/owners/new",
        element: (
          <Suspense fallback={null}>
            <OwnerForm accountingSection={accountingSection} />
          </Suspense>
        ),
        loader: () => ({ owner: undefined }),
      },
      {
        path: "/owners/:id",
        element: (
          <Suspense fallback={null}>
            <OwnerForm accountingSection={accountingSection} />
          </Suspense>
        ),
        loader: () => ({ owner: mockOwner }),
      },
    ],
    { initialEntries: [path] },
  );
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <AuthContext.Provider value={makeAuthCtx(grants)}>
      <QueryClientProvider client={qc}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </AuthContext.Provider>,
  );
}

describe("OwnerForm — accounting:view 権限による会計履歴セクション表示制御", () => {
  it("編集モード + accounting:view なし → 会計履歴セクション非表示", async () => {
    renderOwnerForm({ canViewAccounting: false });
    // フォームが描画されたことを確認してから会計履歴の有無を検証する
    await screen.findByRole("heading", { level: 1, name: /飼主・ペット/ });
    expect(screen.queryByRole("heading", { name: /会計履歴/ })).not.toBeInTheDocument();
  });

  it("編集モード + accounting:view あり → 会計履歴セクション表示", async () => {
    renderOwnerForm({ canViewAccounting: true });
    // <h2>会計履歴</h2> は Suspense の外側で即座に描画される
    await screen.findByRole("heading", { name: /会計履歴/ });
  });

  it("新規モード (ownerId なし) + accounting:view あり → isEdit=false → 会計履歴セクション非表示", async () => {
    renderOwnerForm({ canViewAccounting: true, path: "/owners/new" });
    await screen.findByRole("heading", { level: 1, name: /飼主・ペット/ });
    expect(screen.queryByRole("heading", { name: /会計履歴/ })).not.toBeInTheDocument();
  });
});
