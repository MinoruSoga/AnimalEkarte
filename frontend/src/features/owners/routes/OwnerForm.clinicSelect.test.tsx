import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { Suspense } from "react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/hooks/auth-context";
import { OwnerForm } from "./OwnerForm";
import type { AuthUser, ClinicMembership, ResourceAction } from "@/types/auth";
import type { Owner } from "@/types/owner";

/**
 * OwnerForm — #84 登録先医院セレクトの表示制御
 *
 * Q11=A: 選択肢はユーザー所属医院 (user.clinics) のみ
 * Q12=A: 登録フォームのみ表示（編集モードでは非表示）
 * 単一所属ユーザーには表示しない（従来挙動の維持）
 */

const CLINIC_ID = "1";
const OWNER_ID = "123";

const TWO_CLINICS: ClinicMembership[] = [
  { clinicId: "1", clinicName: "本院", isMain: true },
  { clinicId: "2", clinicName: "分院", isMain: false },
];

const ONE_CLINIC: ClinicMembership[] = [{ clinicId: "1", clinicName: "本院", isMain: true }];

function makeUser(clinics: ClinicMembership[]): AuthUser {
  return {
    id: "10",
    email: "staff@example.com",
    displayName: "テストスタッフ",
    isSystemAdmin: false,
    mainClinicId: CLINIC_ID,
    clinic: null,
    clinics,
    permissions: {},
  };
}

const GRANTS: [string, ResourceAction][] = [
  ["owners", "view"],
  ["owners", "edit"],
  ["owners", "create"],
];

function makeAuthCtx(clinics: ClinicMembership[]) {
  return {
    user: makeUser(clinics),
    currentClinicId: CLINIC_ID,
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: (resource: string, action: ResourceAction) =>
      GRANTS.some(([r, a]) => r === resource && a === action),
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

function renderOwnerForm(options: { clinics: ClinicMembership[]; path?: string }) {
  const { clinics, path = "/owners/new" } = options;
  const router = createMemoryRouter(
    [
      {
        path: "/owners/new",
        element: (
          <Suspense fallback={null}>
            <OwnerForm />
          </Suspense>
        ),
        loader: () => ({ owner: undefined }),
      },
      {
        path: "/owners/:id",
        element: (
          <Suspense fallback={null}>
            <OwnerForm />
          </Suspense>
        ),
        loader: () => ({ owner: mockOwner }),
      },
    ],
    { initialEntries: [path] },
  );
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <AuthContext.Provider value={makeAuthCtx(clinics)}>
      <QueryClientProvider client={qc}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </AuthContext.Provider>,
  );
}

describe("OwnerForm — #84 登録先医院セレクトの表示制御", () => {
  it("新規モード + 複数所属 → 登録先医院セレクトを表示し現在の医院がデフォルト", async () => {
    renderOwnerForm({ clinics: TWO_CLINICS });
    await screen.findByRole("heading", { level: 1, name: /飼主・ペット/ });
    expect(screen.getByText("登録先医院")).toBeInTheDocument();
    // currentClinicId=1 の医院名がデフォルト表示される
    expect(screen.getByTestId("owner-clinic-select")).toHaveTextContent("本院");
  });

  it("新規モード + 単一所属 → 登録先医院セレクト非表示", async () => {
    renderOwnerForm({ clinics: ONE_CLINIC });
    await screen.findByRole("heading", { level: 1, name: /飼主・ペット/ });
    expect(screen.queryByText("登録先医院")).not.toBeInTheDocument();
  });

  it("編集モード + 複数所属 → 登録先医院セレクト非表示 (Q12=A 登録フォームのみ)", async () => {
    renderOwnerForm({ clinics: TWO_CLINICS, path: `/owners/${OWNER_ID}` });
    await screen.findByRole("heading", { level: 1, name: /飼主・ペット/ });
    expect(screen.queryByText("登録先医院")).not.toBeInTheDocument();
  });
});
