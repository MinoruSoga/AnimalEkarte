import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthUser, ResourceAction } from "@/types/auth";
import type { Pet } from "@/types";

// #266: PropertyFilter が返す ActiveFilter[] が OwnersList の URL 反映（species/status パラメータ
// + page リセット）へ正しく配線されていることを検証する。PropertyFilter 自体の内部実装（ポップオーバー
// 操作等）は別レイヤーの責務のためモックし、OwnersList 側の onFilterChange 配線のみに焦点を当てる。
vi.mock("@/components/shared/PropertyFilter/PropertyFilter", () => ({
  PropertyFilter: ({ onFilterChange }: { onFilterChange: (f: unknown[]) => void }) => (
    <button
      type="button"
      onClick={() =>
        onFilterChange([{ key: "species", condition: "is", value: "1", displayValue: "犬" }])
      }
    >
      種を犬に絞り込む
    </button>
  ),
}));

vi.mock("@/features/owners/components/PetEditModal", () => ({
  PetEditModal: () => null,
}));

vi.mock("@/hooks/use-animal-species", () => ({
  useAnimalSpecies: vi.fn(() => ({
    activeSpecies: [
      { id: 1, name: "犬" },
      { id: 2, name: "猫" },
    ],
    allSpecies: [
      { id: 1, name: "犬" },
      { id: 2, name: "猫" },
    ],
  })),
}));

import { OwnersList } from "./OwnersList";
import type { OwnersLoaderData } from "../loaders";

const CLINIC_ID = "1";

function makeUser(): AuthUser {
  return {
    id: "10",
    email: "staff@example.com",
    displayName: "テストスタッフ",
    isSystemAdmin: false,
    mainClinicId: CLINIC_ID,
    clinic: null,
    clinics: [{ clinicId: CLINIC_ID, clinicName: "本院", isMain: true }],
    permissions: {},
  };
}

const GRANTS: [string, ResourceAction][] = [
  ["owners", "view"],
  ["owners", "edit"],
  ["owners", "create"],
  ["owners", "delete"],
];

function makeAuthCtx() {
  return {
    user: makeUser(),
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

function makePet(): Pet {
  return {
    id: "1",
    ownerId: "1",
    ownerName: "山田太郎",
    ownerNumber: 1,
    name: "ポチ",
    species: "犬",
  } as unknown as Pet;
}

describe("OwnersList — #266 フィルタ選択→URL反映の配線", () => {
  it("フィルタ選択で species パラメータへ反映され、page がリセットされる", async () => {
    const loaderFn = vi.fn((): OwnersLoaderData => ({
      pets: [makePet()],
      page: 1,
      limit: 20,
      total: 1,
    }));

    const router = createMemoryRouter(
      [
        {
          path: "/owners",
          element: <OwnersList />,
          loader: () => loaderFn(),
        },
      ],
      { initialEntries: ["/owners?page=2"] },
    );
    render(
      <AuthContext.Provider value={makeAuthCtx()}>
        <RouterProvider router={router} />
      </AuthContext.Provider>,
    );
    await screen.findByText("山田太郎");

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "種を犬に絞り込む" }));

    await waitFor(() => {
      expect(router.state.location.search).toContain("species=1");
    });
    expect(router.state.location.search).not.toContain("page=2");
  });
});
