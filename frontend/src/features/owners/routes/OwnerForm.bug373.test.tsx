/**
 * BUG-373: OwnerForm.handlePetChangeOwner — 飼主変更 条件付き確認ダイアログ
 *
 * discount_rate / membership_type が旧/新 owner で異なる場合のみ ConfirmDialog を表示し、
 * 同じ場合は即時 updatePetMutate を呼ぶことを検証する。
 */
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor, act } from "@testing-library/react";
import { Suspense, createElement, useState, type ReactNode } from "react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/hooks/auth-context";
import { OwnerForm } from "./OwnerForm";
import { useOwnerForm } from "../hooks/use-owner-form";
import type { ResourceAction } from "@/types/auth";
import type { OwnerData, PetFormData } from "../types";
import type { PetMutations } from "@/types/pet";

vi.mock("../hooks/use-owner-form", () => ({
  useOwnerForm: vi.fn(),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({ isDirty: false, markDirty: vi.fn(), markClean: vi.fn() }),
}));

vi.mock("@/hooks/use-title", () => ({
  useTitle: vi.fn(),
}));

vi.mock("../hooks/use-postal-code-lookup", () => ({
  usePostalCodeLookup: () => ({ lookup: vi.fn() }),
}));

// named export (lazy import uses m.PetEditModal)
vi.mock("../components/PetEditModal", () => ({
  PetEditModal: (props: {
    open: boolean;
    onChangeOwner?: (o: {
      id: string;
      name: string;
      discountRate: number;
      membershipType: string;
    }) => void;
  }) => {
    if (!props.open) return null;
    return createElement(
      "div",
      { "data-testid": "pet-edit-modal" },
      // 旧 owner (discountRate=10, membershipType="会員") と同条件
      createElement(
        "button",
        {
          onClick: () =>
            props.onChangeOwner?.({
              id: "2",
              name: "同条件飼主",
              discountRate: 10,
              membershipType: "会員",
            }),
        },
        "btn-same",
      ),
      // discountRate が異なる
      createElement(
        "button",
        {
          onClick: () =>
            props.onChangeOwner?.({
              id: "3",
              name: "別飼主",
              discountRate: 20,
              membershipType: "会員",
            }),
        },
        "btn-diff-discount",
      ),
    );
  },
}));

// OwnerAccountingHistory は会計権限なしで非表示になるため mock 不要
// NavigationBlocker は react-router useBlocker を内部で使うが副作用なし

// ---------- helpers ----------

const mockPet: PetFormData = {
  id: "pet-1",
  petName: "ポチ",
  petNameKana: "",
  petNumber: "P001",
  status: "生存",
  species: "犬",
  gender: "オス",
  birthDate: "",
  color: "",
  weight: "",
  environment: "",
  remarks: "",
};

function makeOwnerFormReturn(overrides: Record<string, unknown> = {}) {
  return {
    isEdit: true,
    isLoading: false,
    ownerData: {
      ownerId: "123",
      membershipType: "会員",
      discountRate: 10,
      ownerName: "テスト飼主",
      ownerNameKana: "テストカイヌシ",
      postalCode: "",
      company: "",
      address1: "",
      address2: "",
      homeAddress1: "",
      homeAddress2: "",
      isDangerous: false,
      birthDate: "",
      email: "",
      phone: "090-0000-0000",
      companyPhone: "",
      remarks: "",
    } as OwnerData,
    setOwnerData: vi.fn(),
    pets: [],
    setPets: vi.fn(),
    petModalOpen: true, // モーダルは開いている状態
    setPetModalOpen: vi.fn(),
    editingPet: mockPet, // ペットは編集中
    handleAddPet: vi.fn(),
    handleEditPet: vi.fn(),
    handleDeletePet: vi.fn(),
    handleSavePet: vi.fn(),
    formAction: vi.fn() as never,
    formState: { success: false, fieldErrors: {}, timestamp: 0 } as never,
    fieldErrors: {} as Record<string, string>,
    clearFieldError: vi.fn(),
    ...overrides,
  };
}

type PermGrant = [string, ResourceAction];

function makeAuthCtx(
  grants: PermGrant[] = [
    ["owners", "view"],
    ["owners", "edit"],
    ["owners", "create"],
    ["discount", "edit"],
  ],
) {
  return {
    user: null,
    currentClinicId: "clinic-1",
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: (resource: string, action: ResourceAction) =>
      grants.some(([r, a]) => r === resource && a === action),
    refreshPermissions: async () => {},
  };
}

function makePetMutations() {
  const updatePetMutate = vi.fn();
  const mutations: PetMutations = {
    updatePetMutate,
    createPetMutate: vi.fn(),
    deletePetMutate: vi.fn(),
    createPetFn: vi.fn() as never,
    revokePetDeathMutate: vi.fn(),
  };
  return { mutations, updatePetMutate };
}

function renderOwnerForm(mutations: PetMutations) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const router = createMemoryRouter(
    [
      {
        path: "/owners/:id",
        element: (
          <Suspense fallback={null}>
            <OwnerForm petMutations={mutations} />
          </Suspense>
        ),
        loader: () => ({ owner: undefined }),
      },
    ],
    { initialEntries: ["/owners/123"] },
  );
  render(
    <AuthContext.Provider value={makeAuthCtx()}>
      <QueryClientProvider client={qc}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </AuthContext.Provider>,
  );
}

function RevocableAuthProvider({ children }: { children: ReactNode }) {
  const [canEdit, setCanEdit] = useState(true);
  const grants: PermGrant[] = [
    ["owners", "view"],
    ["owners", "create"],
    ["discount", "edit"],
    ...(canEdit ? [["owners", "edit"] as PermGrant] : []),
  ];

  return (
    <>
      <button type="button" onClick={() => setCanEdit(false)}>
        編集権限を失効
      </button>
      <AuthContext.Provider value={makeAuthCtx(grants)}>{children}</AuthContext.Provider>
    </>
  );
}

function renderOwnerFormWithRevocablePermission(mutations: PetMutations) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const router = createMemoryRouter(
    [
      {
        path: "/owners/:id",
        element: (
          <Suspense fallback={null}>
            <OwnerForm petMutations={mutations} />
          </Suspense>
        ),
        loader: () => ({ owner: undefined }),
      },
    ],
    { initialEntries: ["/owners/123"] },
  );
  render(
    <RevocableAuthProvider>
      <QueryClientProvider client={qc}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </RevocableAuthProvider>,
  );
}

// ---------- tests ----------

describe("BUG-373: OwnerForm.handlePetChangeOwner — 飼主変更 条件付き確認", () => {
  beforeEach(() => {
    vi.mocked(useOwnerForm).mockReturnValue(makeOwnerFormReturn() as never);
  });

  it("[ペット編集] 旧/新 owner で discount_rate も membership_type も同じ → ConfirmDialog 表示なしで即変更", async () => {
    const { mutations, updatePetMutate } = makePetMutations();
    renderOwnerForm(mutations);

    // PetEditModal が描画されるまで待つ (lazy Suspense 解決)
    await screen.findByTestId("pet-edit-modal");

    act(() => {
      fireEvent.click(screen.getByText("btn-same"));
    });

    // ConfirmDialog は表示されない
    expect(screen.queryByText("飼主変更の確認")).not.toBeInTheDocument();
    // 即時 updatePetMutate が呼ばれる
    expect(updatePetMutate).toHaveBeenCalledWith(
      { id: "pet-1", req: { owner_id: 2 } },
      expect.any(Object),
    );
  });

  it("[ペット編集] 旧/新 owner で discount_rate 異なる → ConfirmDialog 表示", async () => {
    const { mutations } = makePetMutations();
    renderOwnerForm(mutations);

    await screen.findByTestId("pet-edit-modal");

    act(() => {
      fireEvent.click(screen.getByText("btn-diff-discount"));
    });

    // ConfirmDialog が表示される
    await screen.findByText("飼主変更の確認");
  });

  it("[ペット編集] ConfirmDialog → 続行 で更新実行", async () => {
    const { mutations, updatePetMutate } = makePetMutations();
    renderOwnerForm(mutations);

    await screen.findByTestId("pet-edit-modal");

    act(() => {
      fireEvent.click(screen.getByText("btn-diff-discount"));
    });

    await screen.findByText("飼主変更の確認");

    act(() => {
      fireEvent.click(screen.getByRole("button", { name: "続行" }));
    });

    await waitFor(() => {
      expect(updatePetMutate).toHaveBeenCalledWith(
        { id: "pet-1", req: { owner_id: 3 } },
        expect.any(Object),
      );
    });
  });

  it("[ペット編集] ConfirmDialog → キャンセル で更新実行されない", async () => {
    const { mutations, updatePetMutate } = makePetMutations();
    renderOwnerForm(mutations);

    await screen.findByTestId("pet-edit-modal");

    act(() => {
      fireEvent.click(screen.getByText("btn-diff-discount"));
    });

    await screen.findByText("飼主変更の確認");

    act(() => {
      fireEvent.click(screen.getByRole("button", { name: "キャンセル" }));
    });

    await waitFor(() => {
      expect(screen.queryByText("飼主変更の確認")).not.toBeInTheDocument();
    });
    expect(updatePetMutate).not.toHaveBeenCalled();
  });

  it("明示的な死亡ペットは同条件飼主への即時変更mutationを発行しない", async () => {
    vi.mocked(useOwnerForm).mockReturnValue(
      makeOwnerFormReturn({ editingPet: { ...mockPet, status: "死亡" } }) as never,
    );
    const { mutations, updatePetMutate } = makePetMutations();
    renderOwnerForm(mutations);
    await screen.findByTestId("pet-edit-modal");

    fireEvent.click(screen.getByText("btn-same"));

    expect(updatePetMutate).not.toHaveBeenCalled();
  });

  it("明示的な死亡ペットは条件差のある飼主変更確認を開始しない", async () => {
    vi.mocked(useOwnerForm).mockReturnValue(
      makeOwnerFormReturn({ editingPet: { ...mockPet, status: "死亡" } }) as never,
    );
    const { mutations, updatePetMutate } = makePetMutations();
    renderOwnerForm(mutations);
    await screen.findByTestId("pet-edit-modal");

    fireEvent.click(screen.getByText("btn-diff-discount"));

    expect(screen.queryByText("飼主変更の確認")).not.toBeInTheDocument();
    expect(updatePetMutate).not.toHaveBeenCalled();
  });

  it("取得済み飼主変更callbackは最新の編集権限がfalseなら即時updateを発行しない", async () => {
    const { mutations, updatePetMutate } = makePetMutations();
    renderOwnerFormWithRevocablePermission(mutations);
    await screen.findByTestId("pet-edit-modal");

    fireEvent.click(screen.getByRole("button", { name: "編集権限を失効" }));
    fireEvent.click(screen.getByText("btn-same"));

    expect(updatePetMutate).not.toHaveBeenCalled();
  });

  it("確認dialogの取得済み続行callbackは最新の編集権限がfalseならupdateを発行しない", async () => {
    const { mutations, updatePetMutate } = makePetMutations();
    renderOwnerFormWithRevocablePermission(mutations);
    await screen.findByTestId("pet-edit-modal");
    const revokeButton = screen.getByRole("button", { name: "編集権限を失効" });

    fireEvent.click(screen.getByText("btn-diff-discount"));
    await screen.findByText("飼主変更の確認");
    fireEvent.click(revokeButton);
    fireEvent.click(screen.getByRole("button", { name: "続行" }));

    expect(updatePetMutate).not.toHaveBeenCalled();
  });
});
