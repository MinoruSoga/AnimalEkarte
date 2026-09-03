import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { createTestWrapper } from "@/testing/TestUtils";
import { ResourceIdentityLinks } from "@/types/generated/models";
import type {
  OwnerGroupResponse,
  OwnerSearchItem,
  PetGroupResponse,
  PetSearchItem,
} from "@/types/generated/identitylink-responses";

const hasPermission = vi.fn();

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    hasPermission,
  }),
}));

const searchOwnersForLink = vi.fn();
const searchPetsForLink = vi.fn();
const createOwnerIdentityGroup = vi.fn();
const unlinkOwnerIdentityMember = vi.fn();
const createPetIdentityGroup = vi.fn();
const unlinkPetIdentityMember = vi.fn();
const getLinkedTreatmentHistory = vi.fn();
const findOwnerIdentityGroupByMember = vi.fn();
const findPetIdentityGroupByMember = vi.fn();

vi.mock("../api/identity-links-api", () => ({
  searchOwnersForLink: (...args: unknown[]) => searchOwnersForLink(...args),
  searchPetsForLink: (...args: unknown[]) => searchPetsForLink(...args),
  createOwnerIdentityGroup: (...args: unknown[]) => createOwnerIdentityGroup(...args),
  unlinkOwnerIdentityMember: (...args: unknown[]) => unlinkOwnerIdentityMember(...args),
  createPetIdentityGroup: (...args: unknown[]) => createPetIdentityGroup(...args),
  unlinkPetIdentityMember: (...args: unknown[]) => unlinkPetIdentityMember(...args),
  getLinkedTreatmentHistory: (...args: unknown[]) => getLinkedTreatmentHistory(...args),
  findOwnerIdentityGroupByMember: (...args: unknown[]) => findOwnerIdentityGroupByMember(...args),
  findPetIdentityGroupByMember: (...args: unknown[]) => findPetIdentityGroupByMember(...args),
}));

import { IdentityLinksPage } from "./IdentityLinksPage";

const ownerHit: OwnerSearchItem = {
  clinic_id: 1,
  owner_id: 10,
  name: "佐藤太郎",
  name_kana: "サトウタロウ",
  phone: "09011112222",
};

const ownerHitUnlinked: OwnerSearchItem = {
  clinic_id: 1,
  owner_id: 99,
  name: "未所属花子",
  name_kana: "ミショゾクハナコ",
  phone: "09099998888",
};

const petHit: PetSearchItem = {
  clinic_id: 1,
  pet_id: 5,
  owner_id: 10,
  name: "ポチ",
  name_kana: "ポチ",
  pet_number: "P-001",
};

const ownerGroup: OwnerGroupResponse = {
  id: 42,
  created_clinic_id: 1,
  version: 1,
  members: [{ clinic_id: 1, owner_id: 10 }],
};

const petGroup: PetGroupResponse = {
  id: 77,
  created_clinic_id: 1,
  owner_group_created_clinic_id: 1,
  owner_group_id: 42,
  version: 1,
  members: [{ clinic_id: 1, pet_id: 5 }],
};

function grantViewEdit() {
  hasPermission.mockImplementation((resource: string, action: string) => {
    return resource === ResourceIdentityLinks && (action === "view" || action === "edit");
  });
}

describe("IdentityLinksPage permission gates", () => {
  beforeEach(() => {
    hasPermission.mockReset();
    searchOwnersForLink.mockReset().mockResolvedValue([]);
    searchPetsForLink.mockReset().mockResolvedValue([]);
    createOwnerIdentityGroup.mockReset();
    unlinkOwnerIdentityMember.mockReset();
    createPetIdentityGroup.mockReset();
    unlinkPetIdentityMember.mockReset();
    getLinkedTreatmentHistory.mockReset();
    findOwnerIdentityGroupByMember.mockReset().mockResolvedValue(null);
    findPetIdentityGroupByMember.mockReset().mockResolvedValue(null);
  });

  it("view が無い場合はホームへリダイレクトする", () => {
    hasPermission.mockReturnValue(false);
    render(<IdentityLinksPage />, { wrapper: createTestWrapper({ router: true }) });
    // Navigate replaces content; FormHeader title must not appear
    expect(screen.queryByRole("heading", { name: "同一飼主・ペット連携" })).not.toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(ResourceIdentityLinks, "view");
  });

  it("view のみの場合は閲覧UIを出し link ボタンを出さない", () => {
    hasPermission.mockImplementation((resource: string, action: string) => {
      return resource === ResourceIdentityLinks && action === "view";
    });
    render(<IdentityLinksPage />, { wrapper: createTestWrapper({ router: true }) });
    // PageLayout / FormHeader owns the page title
    expect(screen.getByRole("heading", { name: "同一飼主・ペット連携" })).toBeInTheDocument();
    // children status note (+ PermissionBadges may also show 閲覧のみ)
    expect(screen.getByRole("status")).toHaveTextContent(/閲覧のみ/);
    expect(screen.queryByRole("button", { name: "飼主をリンク" })).not.toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(ResourceIdentityLinks, "edit");
  });

  it("edit がある場合は link ボタンを表示する", () => {
    grantViewEdit();
    render(<IdentityLinksPage />, { wrapper: createTestWrapper({ router: true }) });
    expect(screen.getByRole("heading", { name: "同一飼主・ペット連携" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "飼主をリンク" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "ペットをリンク" })).toBeInTheDocument();
  });
});

describe("IdentityLinksPage BUG-013 unlink for existing groups", () => {
  beforeEach(() => {
    hasPermission.mockReset();
    grantViewEdit();
    searchOwnersForLink.mockReset().mockResolvedValue([]);
    searchPetsForLink.mockReset().mockResolvedValue([]);
    createOwnerIdentityGroup.mockReset();
    unlinkOwnerIdentityMember.mockReset().mockResolvedValue(undefined);
    createPetIdentityGroup.mockReset();
    unlinkPetIdentityMember.mockReset().mockResolvedValue(undefined);
    getLinkedTreatmentHistory.mockReset();
    findOwnerIdentityGroupByMember.mockReset().mockResolvedValue(null);
    findPetIdentityGroupByMember.mockReset().mockResolvedValue(null);
  });

  it("リンク済み飼主を選択すると unlink が有効になり、クリックで group id 付き DELETE する", async () => {
    const user = userEvent.setup();
    searchOwnersForLink.mockResolvedValue([ownerHit]);
    findOwnerIdentityGroupByMember.mockResolvedValue(ownerGroup);

    render(<IdentityLinksPage />, { wrapper: createTestWrapper({ router: true }) });

    const ownerSearch = screen.getByPlaceholderText("氏名・カナ・電話");
    await user.type(ownerSearch, "佐藤");

    await waitFor(() => {
      expect(searchOwnersForLink).toHaveBeenCalledWith("佐藤");
    });

    await user.click(await screen.findByRole("button", { name: /\[医院 1\] 佐藤太郎/ }));

    await waitFor(() => {
      expect(findOwnerIdentityGroupByMember).toHaveBeenCalledWith(1, 10);
    });

    const unlinkBtn = await screen.findByRole("button", { name: "連携解除 1/10" });
    await waitFor(() => {
      expect(unlinkBtn).toBeEnabled();
    });

    await user.click(unlinkBtn);

    await waitFor(() => {
      expect(unlinkOwnerIdentityMember).toHaveBeenCalledWith(42, {
        clinic_id: 1,
        owner_id: 10,
      });
    });
  });

  it("lookup が null のとき unlink は disabled のまま、alert を出さない", async () => {
    const user = userEvent.setup();
    searchOwnersForLink.mockResolvedValue([ownerHit]);
    findOwnerIdentityGroupByMember.mockResolvedValue(null);

    render(<IdentityLinksPage />, { wrapper: createTestWrapper({ router: true }) });

    await user.type(screen.getByPlaceholderText("氏名・カナ・電話"), "佐藤");
    await waitFor(() => {
      expect(searchOwnersForLink).toHaveBeenCalledWith("佐藤");
    });

    await user.click(await screen.findByRole("button", { name: /\[医院 1\] 佐藤太郎/ }));

    await waitFor(() => {
      expect(findOwnerIdentityGroupByMember).toHaveBeenCalledWith(1, 10);
    });

    const unlinkBtn = await screen.findByRole("button", { name: "連携解除 1/10" });
    expect(unlinkBtn).toBeDisabled();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("先行 lookup で session group id があっても、404 メンバーの unlink は disabled のまま", async () => {
    const user = userEvent.setup();
    searchOwnersForLink.mockResolvedValue([ownerHit, ownerHitUnlinked]);
    findOwnerIdentityGroupByMember.mockImplementation(async (_clinicId: number, ownerId: number) => {
      if (ownerId === 10) return ownerGroup;
      return null;
    });

    render(<IdentityLinksPage />, { wrapper: createTestWrapper({ router: true }) });

    await user.type(screen.getByPlaceholderText("氏名・カナ・電話"), "佐藤");
    await waitFor(() => {
      expect(searchOwnersForLink).toHaveBeenCalledWith("佐藤");
    });

    await user.click(await screen.findByRole("button", { name: /\[医院 1\] 佐藤太郎/ }));
    await waitFor(() => {
      expect(findOwnerIdentityGroupByMember).toHaveBeenCalledWith(1, 10);
    });
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "連携解除 1/10" })).toBeEnabled();
    });

    // Deselect linked member (clears per-member id; session display id may remain).
    await user.click(screen.getByRole("button", { name: /\[医院 1\] 佐藤太郎/ }));
    await waitFor(() => {
      expect(screen.queryByRole("button", { name: "連携解除 1/10" })).not.toBeInTheDocument();
    });

    await user.click(await screen.findByRole("button", { name: /\[医院 1\] 未所属花子/ }));
    await waitFor(() => {
      expect(findOwnerIdentityGroupByMember).toHaveBeenCalledWith(1, 99);
    });

    const unlinkUnlinked = await screen.findByRole("button", { name: "連携解除 1/99" });
    expect(unlinkUnlinked).toBeDisabled();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(unlinkOwnerIdentityMember).not.toHaveBeenCalled();
  });

  it("リンク済みペットを選択すると unlink が有効になり、pet group id で DELETE する", async () => {
    const user = userEvent.setup();
    searchPetsForLink.mockResolvedValue([petHit]);
    findPetIdentityGroupByMember.mockResolvedValue(petGroup);

    render(<IdentityLinksPage />, { wrapper: createTestWrapper({ router: true }) });

    await user.type(screen.getByPlaceholderText("ペット名・番号"), "ポチ");
    await waitFor(() => {
      expect(searchPetsForLink).toHaveBeenCalledWith("ポチ");
    });

    await user.click(await screen.findByRole("button", { name: /\[医院 1\] ポチ/ }));

    await waitFor(() => {
      expect(findPetIdentityGroupByMember).toHaveBeenCalledWith(1, 5);
    });

    const unlinkBtn = await screen.findByRole("button", { name: "連携解除 1/5" });
    await waitFor(() => {
      expect(unlinkBtn).toBeEnabled();
    });

    await user.click(unlinkBtn);

    await waitFor(() => {
      expect(unlinkPetIdentityMember).toHaveBeenCalledWith(77, {
        clinic_id: 1,
        pet_id: 5,
      });
    });
  });
});

describe("IdentityLinksPage BUG-025 business Japanese copy", () => {
  beforeEach(() => {
    hasPermission.mockReset();
    grantViewEdit();
    searchOwnersForLink.mockReset().mockResolvedValue([]);
    searchPetsForLink.mockReset().mockResolvedValue([]);
    createOwnerIdentityGroup.mockReset();
    unlinkOwnerIdentityMember.mockReset();
    createPetIdentityGroup.mockReset();
    unlinkPetIdentityMember.mockReset();
    getLinkedTreatmentHistory.mockReset();
    findOwnerIdentityGroupByMember.mockReset().mockResolvedValue(null);
    findPetIdentityGroupByMember.mockReset().mockResolvedValue(null);
  });

  it("検索ラベルに debounce / 300ms などの実装用語を出さない", () => {
    render(<IdentityLinksPage />, { wrapper: createTestWrapper({ router: true }) });

    const labels = screen.getAllByText("検索");
    expect(labels.length).toBeGreaterThanOrEqual(2);
    expect(screen.queryByText(/debounce/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/300\s*ms/i)).not.toBeInTheDocument();
  });

  it("ペット連携の前提メッセージを業務日本語で示す", () => {
    render(<IdentityLinksPage />, { wrapper: createTestWrapper({ router: true }) });

    expect(screen.getByText("親となる飼主の連携グループが必要です。")).toBeInTheDocument();
    expect(screen.queryByText(/identity group/i)).not.toBeInTheDocument();
  });

  it("閲覧のみ・解除ボタンも開発者英語を出さない", () => {
    hasPermission.mockImplementation((resource: string, action: string) => {
      return resource === ResourceIdentityLinks && action === "view";
    });
    render(<IdentityLinksPage />, { wrapper: createTestWrapper({ router: true }) });

    expect(screen.getByRole("status")).toHaveTextContent("閲覧のみ（連携の変更権限がありません）");
    expect(screen.getByRole("status")).not.toHaveTextContent(/link\/unlink/i);
  });
});
