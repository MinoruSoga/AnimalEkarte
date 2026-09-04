import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";
import { AuthContext } from "@/hooks/auth-context";
import { OwnersList } from "./OwnersList";
import { activeFiltersToParams, paramsToActiveFilters } from "../lib/owners-list-filters";
import type { OwnersLoaderData } from "../loaders";
import type { AuthUser, ResourceAction } from "@/types/auth";
import type { Pet } from "@/types";

// #266: /owners のサーバサイドページネーション・検索・フィルタが URL 経由で
// loader に正しく転送されることを検証する（白画面バグの UI 層回帰防止）。

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

function makePet(overrides: Partial<Pet> = {}): Pet {
  return {
    id: "1",
    ownerId: "1",
    ownerName: "山田太郎",
    ownerNumber: 1,
    name: "ポチ",
    species: "犬",
    ...overrides,
  } as unknown as Pet;
}

function renderOwnersList(loader: (request: Request) => OwnersLoaderData, initialPath = "/owners") {
  const router = createMemoryRouter(
    [
      {
        path: "/owners",
        element: <OwnersList />,
        loader: ({ request }) => loader(request),
      },
    ],
    { initialEntries: [initialPath] },
  );
  render(
    <AuthContext.Provider value={makeAuthCtx()}>
      <RouterProvider router={router} />
    </AuthContext.Provider>,
  );
  return router;
}

describe("OwnersList — #266 サーバサイド検索・フィルタ・ページネーション", () => {
  it("検索入力はデバウンス後に URL の search パラメータへ反映され loader が再フェッチする", async () => {
    const requests: string[] = [];
    const loaderFn = vi.fn((request: Request) => {
      requests.push(new URL(request.url).search);
      return { pets: [makePet()], page: 1, limit: 20, total: 1 };
    });

    const router = renderOwnersList(loaderFn);
    await screen.findByText("山田太郎");

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "検索" }));
    const searchInput = screen.getByPlaceholderText(
      "飼主名、ペット名、電話番号、飼主No、ペット番号...",
    );
    await user.type(searchInput, "田中");

    await waitFor(
      () => {
        expect(router.state.location.search).toContain("search=%E7%94%B0%E4%B8%AD");
      },
      { timeout: 2000 },
    );

    expect(requests.some((s) => s.includes("search=%E7%94%B0%E4%B8%AD"))).toBe(true);
  });

  // フィルタ選択→URL反映の統合検証は OwnersList.filter-wiring.test.tsx（PropertyFilter を
  // 全体モックした別ファイル）で行う。本ファイルは実 PropertyFilter を使う検索テストと同居するため、
  // 同一ファイル内でのモジュールモック切替（vi.doMock 後の動的 import）はモジュールキャッシュの
  // 制約で機能しない。

  it("activeFiltersToParams: ActiveFilter[] から species/include_deceased のみ抽出する", () => {
    expect(
      activeFiltersToParams([
        { key: "species", condition: "is", value: "2", displayValue: "猫" },
        {
          key: "include_deceased",
          condition: "is",
          value: "true",
          displayValue: "死亡ペットも含める",
        },
      ]),
    ).toEqual({ species: "2", include_deceased: "true" });
    expect(activeFiltersToParams([])).toEqual({ species: undefined, include_deceased: undefined });
  });

  it("activeFiltersToParams: include_deceased=false（既定値）は URL に残さない", () => {
    expect(
      activeFiltersToParams([
        { key: "include_deceased", condition: "is", value: "false", displayValue: "生存のみ" },
      ]),
    ).toEqual({ species: undefined, include_deceased: undefined });
  });

  it("activeFiltersToParams: condition が is 以外（is_not/空/空でない）は転送しない（意味反転のサイレントバグ防止）", () => {
    // FilterAddPopover は conditions 上書きを無視して全条件を提示するため、is_not 等が選ばれうる。
    // is_not の value をそのまま「一致」として送ると絞り込みの意味が反転するため、is 以外は無視する。
    expect(
      activeFiltersToParams([
        { key: "species", condition: "is_not", value: "1", displayValue: "犬" },
      ]),
    ).toEqual({ species: undefined, include_deceased: undefined });
    expect(
      activeFiltersToParams([
        { key: "include_deceased", condition: "is_empty", value: "", displayValue: "空" },
      ]),
    ).toEqual({ species: undefined, include_deceased: undefined });
  });

  it("paramsToActiveFilters: URL パラメータから ActiveFilter[] を復元する（表示ラベル付き）", () => {
    const speciesOptions = [
      { value: "1", label: "犬" },
      { value: "2", label: "猫" },
    ];
    const params = new URLSearchParams("species=2&include_deceased=true");
    expect(paramsToActiveFilters(params, speciesOptions)).toEqual([
      { key: "species", condition: "is", value: "2", displayValue: "猫" },
      {
        key: "include_deceased",
        condition: "is",
        value: "true",
        displayValue: "死亡ペットも含める",
      },
    ]);
    expect(paramsToActiveFilters(new URLSearchParams(), speciesOptions)).toEqual([]);
  });

  it("total が limit を超える場合にページネーションが表示され、次へクリックで page=2 を loader に渡す", async () => {
    const requests: string[] = [];
    const loaderFn = vi.fn((request: Request) => {
      requests.push(new URL(request.url).search);
      return { pets: [makePet()], page: 1, limit: 20, total: 41 };
    });

    renderOwnersList(loaderFn);
    await screen.findByText("山田太郎");

    const user = userEvent.setup();
    const nextButton = await screen.findByRole("button", { name: /次へ|次のページ|Next/i });
    await user.click(nextButton);

    await waitFor(() => {
      expect(requests.some((s) => s.includes("page=2"))).toBe(true);
    });
  });
});
