import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PatientSelectionTable } from "./PatientSelectionTable";
import type { Pet } from "@/types";

vi.mock("@/hooks/use-pet", () => ({
  useSearchPets: vi.fn(),
}));

import { useSearchPets } from "@/hooks/use-pet";

const basePet = {
  id: "",
  ownerId: "100",
  ownerNumber: 100,
  ownerName: "",
  ownerNameKana: undefined,
  phone: "090-0000-0000",
  petNumber: undefined,
  name: "",
  petNameKana: undefined,
  species: "犬",
  gender: "雄",
  status: undefined,
  birthDate: undefined,
  weight: undefined,
  clinicId: undefined,
  address: undefined,
  animalSpeciesId: undefined,
  breed: undefined,
  color: undefined,
  bloodType: undefined,
  microchipNumber: undefined,
  neuteredDate: undefined,
  food: undefined,
  environment: undefined,
  acquisitionType: undefined,
  dangerLevel: undefined,
  lastVisit: undefined,
  insuranceId: undefined,
  insuranceName: undefined,
  insuranceDetails: undefined,
  remarks: undefined,
  deceasedAt: undefined,
} satisfies Pet;

const MOCK_PETS: Pet[] = [
  { ...basePet, id: "1", ownerName: "ヤマダ タロウ", ownerNameKana: "ヤマダ タロウ", name: "ポチ", petNameKana: "ポチ" },
  { ...basePet, id: "2", ownerName: "タナカ ハナコ", ownerNameKana: "タナカ ハナコ", name: "ミケ", petNameKana: "ミケ" },
];

beforeEach(() => {
  vi.mocked(useSearchPets).mockReturnValue({ pets: MOCK_PETS, isLoading: false } as ReturnType<typeof useSearchPets>);
});

async function searchByField(placeholder: string, value: string) {
  const user = userEvent.setup();
  render(<PatientSelectionTable onSelect={vi.fn()} selectedPets={[]} />);
  await user.type(screen.getByPlaceholderText(placeholder), value);
  await user.click(screen.getByRole("button", { name: /検索/ }));
}

describe("PatientSelectionTable — カナ混同検索", () => {
  it("検索前は結果を表示しない", () => {
    render(<PatientSelectionTable onSelect={vi.fn()} selectedPets={[]} />);
    expect(screen.queryByText("ヤマダ タロウ")).not.toBeInTheDocument();
  });

  it("ownerName ひらがなで検索するとカタカナ飼主名がヒットする", async () => {
    await searchByField("例: 林 文明", "やまだ");
    expect(screen.getByText("ヤマダ タロウ")).toBeInTheDocument();
    expect(screen.queryByText("タナカ ハナコ")).not.toBeInTheDocument();
  });

  it("petName ひらがなで検索するとカタカナペット名がヒットする", async () => {
    await searchByField("例: Iris", "ぽち");
    expect(screen.getByText("ポチ")).toBeInTheDocument();
    expect(screen.queryByText("ミケ")).not.toBeInTheDocument();
  });

  it("ownerNameKana フィールドで混同検索が効く", async () => {
    await searchByField("例: はやし", "やまだ");
    expect(screen.getByText("ヤマダ タロウ")).toBeInTheDocument();
    expect(screen.queryByText("タナカ ハナコ")).not.toBeInTheDocument();
  });

  it("petNameKana フィールドで混同検索が効く", async () => {
    await searchByField("例: いりす", "ぽち");
    expect(screen.getByText("ポチ")).toBeInTheDocument();
    expect(screen.queryByText("ミケ")).not.toBeInTheDocument();
  });

  it("検索語が空のまま検索すると全件表示する", async () => {
    const user = userEvent.setup();
    render(<PatientSelectionTable onSelect={vi.fn()} selectedPets={[]} />);
    await user.click(screen.getByRole("button", { name: /検索/ }));
    expect(screen.getByText("ヤマダ タロウ")).toBeInTheDocument();
    expect(screen.getByText("タナカ ハナコ")).toBeInTheDocument();
  });

  it("行は選択せず、固有名の44px buttonだけが患者を選択する", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(<PatientSelectionTable onSelect={onSelect} selectedPets={[]} />);
    await user.click(screen.getByRole("button", { name: /検索/ }));

    await user.click(screen.getByText("ヤマダ タロウ"));
    expect(onSelect).not.toHaveBeenCalled();

    const selectButton = screen.getByRole("button", { name: "選択: ポチ (ID 1)" });
    expect(selectButton.tagName).toBe("BUTTON");
    expect(selectButton).toHaveClass("min-h-11", "min-w-11");
    await user.click(selectButton);
    expect(onSelect).toHaveBeenCalledWith(MOCK_PETS[0]);
  });
});
