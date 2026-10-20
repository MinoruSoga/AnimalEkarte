import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";
import { AuthContext } from "@/hooks/auth-context";
import { OwnersList } from "./OwnersList";
import type { OwnersLoaderData } from "../loaders";
import type { AuthUser } from "@/types/auth";
import type { Pet } from "@/types";

// EMR-463: 非活性時（loader が clinicScope=null・ready を返す）は PR 前と同じ #86 useClinicScope 表示。
// 値は合成のみ。

vi.mock("@/features/owners/components/PetEditModal", () => ({
  PetEditModal: () => null,
}));

vi.mock("@/hooks/use-animal-species", () => ({
  useAnimalSpecies: vi.fn(() => ({
    activeSpecies: [{ id: 1, name: "犬" }],
    allSpecies: [{ id: 1, name: "犬" }],
  })),
}));

const user: AuthUser = {
  id: "10",
  email: "staff@example.com",
  displayName: "合成スタッフ",
  isSystemAdmin: false,
  mainClinicId: "1",
  clinic: null,
  clinics: [
    { clinicId: "1", clinicName: "合成本院", isMain: true },
    { clinicId: "2", clinicName: "合成分院", isMain: false },
  ],
  permissions: {},
};

const authCtx = {
  user,
  currentClinicId: "1",
  isAuthenticated: true,
  isLoading: false,
  login: async () => {},
  logout: async () => {},
  switchClinic: () => {},
  hasPermission: () => true,
  refreshPermissions: async () => {},
};

const legacyData: OwnersLoaderData = {
  pets: [
    {
      id: "1",
      clinicId: "1",
      ownerId: "1",
      ownerName: "合成飼主",
      ownerNumber: 1,
      name: "合成ペット",
      species: "犬",
    } as unknown as Pet,
  ],
  page: 1,
  limit: 20,
  total: 1,
  listState: { kind: "ready", scopeRechecked: false },
  clinicScope: null,
};

describe("OwnersList — EMR-463 非活性時の既存医院フィルタ", () => {
  it("所属医院のフィルタを従来どおり表示し、未指定は現在の医院だけを選択状態にする", async () => {
    const router = createMemoryRouter(
      [{ path: "/owners", element: <OwnersList />, loader: () => legacyData }],
      { initialEntries: ["/owners"] },
    );
    render(
      <AuthContext.Provider value={authCtx}>
        <RouterProvider router={router} />
      </AuthContext.Provider>,
    );

    expect(await screen.findByRole("button", { name: "合成本院" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    expect(screen.getByRole("button", { name: "合成分院" })).toHaveAttribute(
      "aria-pressed",
      "false",
    );

    await userEvent.setup().click(screen.getByRole("button", { name: "合成分院" }));

    await waitFor(() => {
      expect(new URLSearchParams(router.state.location.search).get("clinics")).toBe("1,2");
    });
    // 非活性時は保存先へ書かない
    expect(window.localStorage.getItem("owners-list.clinics.v1.10")).toBeNull();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
