import { Suspense, act } from "react";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { Owner } from "@/types/owner";
import type { Pet } from "@/types";
import type { PetMutations } from "@/types/pet";
import { OwnerForm } from "@/features/owners";

// EMR-174: /owners/:id?pet=:petId の deep link で既存ペットモーダルを直接開く。
// mutation 権限（canEdit）に関わらず詳細を開けることが契約。

const { mockCanEdit } = vi.hoisted(() => ({
  mockCanEdit: vi.fn(() => true),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canEdit: mockCanEdit(),
    canCreate: true,
    canDelete: true,
    canView: true,
  }),
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { mainClinicId: "clinic-1", clinics: [{ id: "clinic-1", name: "本院" }] },
    currentClinicId: "clinic-1",
  }),
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => ({ data: undefined, isLoading: false, isError: false }),
}));

vi.mock("@/features/owners/hooks/use-animal-species", () => ({
  useAnimalSpecies: () => ({
    allSpecies: [{ id: 1, name: "犬" }],
    activeSpecies: [{ id: 1, name: "犬" }],
    isLoading: false,
  }),
}));

vi.mock("@/features/owners/api/get-insurances", () => ({
  useGetInsurances: () => ({ data: [], isLoading: false }),
}));

vi.mock("@/features/owners/api/get-owner-shared-pets", () => ({
  useGetOwnerSharedPets: () => ({
    data: { shared_pets: [] },
    isError: false,
    isLoading: false,
  }),
}));

vi.mock("@/features/owners/components/PetSubOwnersSection", () => ({
  PetSubOwnersSection: () => null,
}));

const PET_ROW = {
  id: "7",
  ownerId: "owner-1",
  ownerName: "山田太郎",
  petNumber: "P-7",
  name: "ポチ",
  species: "犬",
  status: "生存",
  gender: "雄",
} as unknown as Pet;

const OWNER: Owner = {
  id: "owner-1",
  ownerName: "山田太郎",
  ownerNameKana: "ヤマダタロウ",
  company: "",
  postalCode: "",
  address1: "",
  address2: "",
  homePostalCode: "",
  homeAddress1: "",
  homeAddress2: "",
  phone: "090-0000-0000",
  companyPhone: "",
  email: "",
  remarks: "",
  isDangerous: false,
  discountRate: 0,
  membershipType: "member",
  deliveryExcluded: false,
  deliveryCaution: false,
  isTransferred: false,
  lstepOptOut: false,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
  pets: [PET_ROW],
} as Owner;

const petMutations: PetMutations = {
  createPetFn: vi.fn() as never,
  createPetMutate: vi.fn(),
  updatePetMutate: vi.fn(),
  deletePetMutate: vi.fn(),
  revokePetDeathMutate: vi.fn(),
};

function renderOwnerFormAt(entry: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const router = createMemoryRouter(
    [
      {
        path: "/owners/:id",
        loader: () => ({ owner: OWNER }),
        HydrateFallback: () => null,
        element: (
          <QueryClientProvider client={queryClient}>
            <Suspense fallback={null}>
              <OwnerForm petMutations={petMutations} />
            </Suspense>
          </QueryClientProvider>
        ),
      },
    ],
    { initialEntries: [entry] },
  );
  render(<RouterProvider router={router} />);
  return router;
}

describe("OwnerForm ?pet= deep link (EMR-174)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCanEdit.mockReturnValue(true);
  });

  it("?pet=7 で飼主詳細を開くと対象ペットのモーダルが自動で開く", async () => {
    renderOwnerFormAt("/owners/owner-1?pet=7");

    const dialog = await screen.findByRole("dialog", {}, { timeout: 5000 });
    expect(dialog).toBeInTheDocument();
    expect(screen.getByLabelText(/^ペット名\s*\*?$/)).toHaveValue("ポチ");
  });

  it("編集権限がなくても deep link で詳細モーダルが開き、フィールドは読み取り専用になる", async () => {
    mockCanEdit.mockReturnValue(false);
    renderOwnerFormAt("/owners/owner-1?pet=7");

    const dialog = await screen.findByRole("dialog", {}, { timeout: 5000 });
    expect(dialog).toBeInTheDocument();
    const nameInput = screen.getByLabelText(/^ペット名\s*\*?$/);
    expect(nameInput).toHaveValue("ポチ");
    await waitFor(() => {
      expect(nameInput).toBeDisabled();
    });
    // mutation ボタンは出さない（閲覧のみ）
    expect(screen.queryByRole("button", { name: "更新" })).not.toBeInTheDocument();
  });

  it("存在しないペットIDではモーダルを開かない", async () => {
    renderOwnerFormAt("/owners/owner-1?pet=999");

    await waitFor(() => {
      expect(screen.getByText("ペット情報")).toBeInTheDocument();
    });
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("?pet= なしの通常遷移ではモーダルを開かない", async () => {
    renderOwnerFormAt("/owners/owner-1");

    await waitFor(() => {
      expect(screen.getByText("ペット情報")).toBeInTheDocument();
    });
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("同じペットの deep link はモーダルを閉じた後に再度開ける", async () => {
    const user = userEvent.setup();
    const router = renderOwnerFormAt("/owners/owner-1?pet=7");

    expect(await screen.findByRole("dialog", {}, { timeout: 5000 })).toBeInTheDocument();
    await waitFor(() => {
      expect(router.state.location.search).not.toContain("pet=");
    });

    await user.keyboard("{Escape}");
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    // 同じペット名リンク（/owners/:id?pet=7）への再遷移で再度モーダルが開く
    await act(async () => {
      await router.navigate("/owners/owner-1?pet=7");
    });

    expect(await screen.findByRole("dialog", {}, { timeout: 5000 })).toBeInTheDocument();
    expect(screen.getByLabelText(/^ペット名\s*\*?$/)).toHaveValue("ポチ");
    await waitFor(() => {
      expect(router.state.location.search).not.toContain("pet=");
    });
  });
});
