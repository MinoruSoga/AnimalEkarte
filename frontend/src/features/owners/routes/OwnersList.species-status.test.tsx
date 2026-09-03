import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AuthContext } from "@/hooks/auth-context";
import type { useAnimalSpecies } from "@/hooks/use-animal-species";
import type { AuthUser, ResourceAction } from "@/types/auth";
import type { OwnersLoaderData } from "../loaders";
import { OwnersList } from "./OwnersList";

type AnimalSpeciesState = ReturnType<typeof useAnimalSpecies>;

const mocks = vi.hoisted(() => ({
  useAnimalSpecies: vi.fn<() => AnimalSpeciesState>(),
}));

vi.mock("@/hooks/use-animal-species", () => ({
  useAnimalSpecies: mocks.useAnimalSpecies,
}));

const animalSpecies = [
  {
    id: 1,
    name: "犬",
    is_active: true,
    sort_order: 1,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    label: "犬",
    isInactive: false,
  },
  {
    id: 2,
    name: "猫",
    is_active: true,
    sort_order: 2,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    label: "猫",
    isInactive: false,
  },
] satisfies AnimalSpeciesState["activeSpecies"];

function createSpeciesState(overrides: Partial<AnimalSpeciesState> = {}): AnimalSpeciesState {
  return {
    allSpecies: animalSpecies,
    activeSpecies: animalSpecies,
    isLoading: false,
    isError: false,
    error: null,
    ...overrides,
  };
}

const CLINIC_ID = "1";

function createAuthContext() {
  const user: AuthUser = {
    id: "10",
    email: "staff@example.com",
    displayName: "テストスタッフ",
    isSystemAdmin: false,
    mainClinicId: CLINIC_ID,
    clinic: null,
    clinics: [{ clinicId: CLINIC_ID, clinicName: "本院", isMain: true }],
    permissions: {},
  };

  return {
    user,
    currentClinicId: CLINIC_ID,
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: (_resource: string, _action: ResourceAction) => true,
    refreshPermissions: async () => {},
  };
}

function renderOwnersList() {
  const router = createMemoryRouter(
    [
      {
        path: "/owners",
        element: <OwnersList />,
        loader: (): OwnersLoaderData => ({
          pets: [],
          page: 1,
          limit: 20,
          total: 0,
        }),
      },
    ],
    { initialEntries: ["/owners"] },
  );

  render(
    <AuthContext.Provider value={createAuthContext()}>
      <RouterProvider router={router} />
    </AuthContext.Provider>,
  );
}

describe("OwnersList species status", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useAnimalSpecies.mockReturnValue(createSpeciesState());
  });

  it("取得失敗を最優先の accessible alert で示し、stale な動物種を隠して他操作を使える", async () => {
    const rawError = "GET /v1/masters/animal-species: database timeout";
    mocks.useAnimalSpecies.mockReturnValue(
      createSpeciesState({
        isLoading: true,
        isError: true,
        error: new Error(rawError),
      }),
    );
    const user = userEvent.setup();

    renderOwnersList();

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("動物種の取得に失敗しました。");
    expect(alert).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.queryByText(rawError)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新規登録" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "検索" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "フィルタを追加" })).toBeEnabled();

    await user.click(screen.getByRole("button", { name: "フィルタを追加" }));
    await user.click(screen.getByRole("option", { name: "種" }));
    await user.click(screen.getByRole("button", { name: "次と一致" }));
    expect(screen.getByText("選択肢がありません")).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "犬" })).not.toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "猫" })).not.toBeInTheDocument();
  });

  it("読み込み中を成功0件より優先して示し、stale な動物種を隠して他操作を使える", async () => {
    mocks.useAnimalSpecies.mockReturnValue(
      createSpeciesState({
        isLoading: true,
      }),
    );
    const user = userEvent.setup();

    renderOwnersList();

    const status = await screen.findByRole("status");
    expect(status).toHaveTextContent("動物種を読み込み中です。");
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(status).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByText("動物種マスタが登録されていません。")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新規登録" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "検索" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "フィルタを追加" })).toBeEnabled();

    await user.click(screen.getByRole("button", { name: "フィルタを追加" }));
    await user.click(screen.getByRole("option", { name: "種" }));
    await user.click(screen.getByRole("button", { name: "次と一致" }));
    expect(screen.getByText("選択肢がありません")).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "犬" })).not.toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "猫" })).not.toBeInTheDocument();
  });

  it("取得成功かつ0件を distinct accessible status で示し、種フィルタを理由なく消さない", async () => {
    mocks.useAnimalSpecies.mockReturnValue(
      createSpeciesState({
        allSpecies: [],
        activeSpecies: [],
      }),
    );
    const user = userEvent.setup();

    renderOwnersList();

    const status = await screen.findByRole("status");
    expect(status).toHaveTextContent("動物種マスタが登録されていません。");
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(status).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByText("動物種を読み込み中です。")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "フィルタを追加" }));
    expect(screen.getByRole("option", { name: "種" })).toBeInTheDocument();
  });

  it("取得成功かつ候補ありでは状態表示を消して動物種の選択肢を示す", async () => {
    const user = userEvent.setup();
    renderOwnersList();

    expect(await screen.findByText("データが見つかりません")).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "フィルタを追加" }));
    await user.click(screen.getByRole("option", { name: "種" }));
    await user.click(screen.getByRole("button", { name: "次と一致" }));
    expect(screen.getByRole("option", { name: "犬" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "猫" })).toBeInTheDocument();
  });
});
