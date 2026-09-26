import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { Pet } from "@/types";
import { MedicalRecordPetSelection } from "./MedicalRecordPetSelection";

const navigate = vi.fn();
const selectForNewRecord = vi.fn();
const locationState = { from: "/medical-records" };

vi.mock("react-router", () => ({
  useNavigate: () => navigate,
  useLocation: () => ({ search: "", state: locationState }),
}));

const LIVING_PET = {
  id: "10",
  ownerId: "20",
  ownerName: "山田 太郎",
  name: "ポチ",
  species: "犬",
  status: "生存",
} as unknown as Pet;

const DECEASED_PET = {
  id: "11",
  ownerId: "20",
  ownerName: "山田 太郎",
  name: "タマ",
  species: "猫",
  status: "死亡",
} as unknown as Pet;

vi.mock("@/hooks/use-pet-selection-page", () => ({
  usePetSelectionPage: () => ({
    searchParams: { search: "", ownerId: "", species: "" },
    setSearchParams: vi.fn(),
    petPage: {
      items: [LIVING_PET, DECEASED_PET],
      totalCount: 2,
      currentPage: 1,
      totalPages: 1,
      startIndex: 1,
      endIndex: 2,
      onPageChange: vi.fn(),
    },
    error: undefined,
    isLoading: false,
    handleClear: vi.fn(),
    handleSelect: selectForNewRecord,
    handleBack: vi.fn(),
  }),
}));

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/components/shared/PetSelection/PetSelectionSearchForm", () => ({
  PetSelectionSearchForm: () => <div />,
}));

// EMR-177: 死亡ペットは新規カルテ作成へ進めないため、選択時はカルテ一覧へ送る。
// 一覧から既存カルテの閲覧・確定済カルテへの追記（連絡記録）ができる。
describe("MedicalRecordPetSelection", () => {
  beforeEach(() => {
    navigate.mockClear();
    selectForNewRecord.mockClear();
  });

  it("死亡ペットの選択はそのペットのカルテ一覧へ遷移する", async () => {
    const user = userEvent.setup();
    render(<MedicalRecordPetSelection />);

    await user.click(screen.getByRole("button", { name: "死亡・選択: タマ (ID 11)" }));

    expect(navigate).toHaveBeenCalledWith("/medical-records?pet_id=11", {
      state: locationState,
    });
    expect(selectForNewRecord).not.toHaveBeenCalled();
  });

  it("生存ペットの選択は従来どおり新規作成フローへ委譲する", async () => {
    const user = userEvent.setup();
    render(<MedicalRecordPetSelection />);

    await user.click(screen.getByRole("button", { name: "選択: ポチ (ID 10)" }));

    expect(selectForNewRecord).toHaveBeenCalledWith(LIVING_PET);
    expect(navigate).not.toHaveBeenCalled();
  });
});
