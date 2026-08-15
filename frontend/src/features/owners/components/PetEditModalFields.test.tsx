import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { PetEditModal } from "./PetEditModal";
import { PetEditModalFields } from "./PetEditModalFields";
import type { PetFormData } from "../types";

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => ({ data: undefined, isLoading: false, isError: false }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canEdit: true }),
}));

vi.mock("../hooks/use-animal-species", () => ({
  useAnimalSpecies: () => ({
    allSpecies: [{ id: 1, name: "犬" }],
    activeSpecies: [{ id: 1, name: "犬" }],
    isLoading: false,
  }),
}));

vi.mock("../api/get-insurances", () => ({
  useGetInsurances: () => ({ data: [], isLoading: false }),
}));

vi.mock("./PetSubOwnersSection", () => ({
  PetSubOwnersSection: ({ petId }: { petId: string }) => (
    <div data-testid="pet-sub-owners-section">{petId}</div>
  ),
}));

const basePet: PetFormData = {
  id: "7",
  petNumber: "42-1",
  petName: "ポチ",
  petNameKana: "ぽち",
  status: "生存",
  species: "犬",
  animalSpeciesId: "1",
  gender: "雄",
  birthDate: "2015-04-14",
  breed: "柴犬",
  color: "赤",
  bloodType: "DEA1.1陽性",
  microchipNumber: "392140000123456",
  weight: "7.35",
  neuteredDate: "2016-05-20",
  acquisitionType: "購入",
  dangerLevel: "中",
  dangerReason: "保定時に噛む",
  food: "療法食",
  environment: "室内",
  remarks: "咬傷注意",
  insuranceId: "",
};

function PetFieldsHarness() {
  const [formData, setFormData] = useState<PetFormData>(basePet);

  return (
    <MemoryRouter>
      <PetEditModalFields
        formData={formData}
        setFormData={setFormData}
        fieldErrors={{}}
        clearFieldError={() => {}}
        animalSpeciesList={[{ id: 1, name: "犬" }]}
        isLoadingSpecies={false}
        insuranceList={[]}
        isLoadingInsurances={false}
        canEdit
        isEdit
      />
    </MemoryRouter>
  );
}

describe("PetEditModalFields", () => {
  it("保存済みの危険理由を初期表示する", () => {
    render(<PetFieldsHarness />);

    expect(screen.getByLabelText("危険と判断した理由")).toHaveValue("保定時に噛む");
  });

  it("血液型とマイクロチップ番号を編集できる", async () => {
    const user = userEvent.setup();
    render(<PetFieldsHarness />);

    const bloodTypeInput = screen.getByLabelText("血液型");
    const microchipInput = screen.getByLabelText("マイクロチップ番号");

    expect(bloodTypeInput).toHaveValue("DEA1.1陽性");
    expect(microchipInput).toHaveValue("392140000123456");

    await user.clear(bloodTypeInput);
    await user.type(bloodTypeInput, "B");
    await user.clear(microchipInput);
    await user.type(microchipInput, "900000000000001");

    expect(screen.getByLabelText("血液型")).toHaveValue("B");
    expect(screen.getByLabelText("マイクロチップ番号")).toHaveValue("900000000000001");
  });

  it("編集時は PetEditModalFields から副飼主セクションを差し込む", () => {
    render(<PetFieldsHarness />);

    expect(screen.getByTestId("pet-sub-owners-section")).toHaveTextContent("7");
  });

  it("BUG-022: pending (temp-*) ペットでは副飼主セクションを出さない", () => {
    function PendingHarness() {
      const [formData, setFormData] = useState<PetFormData>({
        ...basePet,
        id: "temp-1710000000000",
        isPending: true,
      });
      return (
        <MemoryRouter>
          <PetEditModalFields
            formData={formData}
            setFormData={setFormData}
            fieldErrors={{}}
            clearFieldError={() => {}}
            animalSpeciesList={[{ id: 1, name: "犬" }]}
            isLoadingSpecies={false}
            insuranceList={[]}
            isLoadingInsurances={false}
            canEdit
            isEdit
          />
        </MemoryRouter>
      );
    }

    render(<PendingHarness />);
    expect(screen.queryByTestId("pet-sub-owners-section")).not.toBeInTheDocument();
  });
});

function renderPetEditModal(
  petData: PetFormData,
  onSave = vi.fn(),
) {
  render(
    <MemoryRouter>
      <PetEditModal
        open
        onOpenChange={() => {}}
        ownerName="山田太郎"
        petData={petData}
        onSave={onSave}
      />
    </MemoryRouter>,
  );
  return onSave;
}

async function chooseDangerLevel(user: ReturnType<typeof userEvent.setup>, level: string) {
  await user.click(screen.getByRole("combobox", { name: "ペットの危険度" }));
  await user.click(screen.getByRole("option", { name: level }));
}

describe("PetEditModal danger reason validation", () => {
  it("dangerLevel=高 で危険理由が空白のみなら保存をブロックする", async () => {
    const user = userEvent.setup();
    const onSave = renderPetEditModal({ ...basePet, dangerReason: " \t " });

    await chooseDangerLevel(user, "高");
    await user.click(screen.getByRole("button", { name: "更新" }));

    expect(onSave).not.toHaveBeenCalled();
    expect(screen.getByText("危険度が高の場合は理由を入力してください")).toBeInTheDocument();
  });

  it("危険理由が501 Unicode文字なら保存をブロックする", async () => {
    const user = userEvent.setup();
    const onSave = renderPetEditModal({
      ...basePet,
      dangerReason: "犬".repeat(501),
    });

    await user.click(screen.getByRole("button", { name: "更新" }));

    expect(onSave).not.toHaveBeenCalled();
    expect(screen.getByText("危険理由は500文字以内で入力してください")).toBeInTheDocument();
  });

  it("危険理由が500 Unicode文字なら保存できる", async () => {
    const user = userEvent.setup();
    const onSave = renderPetEditModal({
      ...basePet,
      dangerReason: "🐕".repeat(500),
    });

    await user.click(screen.getByRole("button", { name: "更新" }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ dangerReason: "🐕".repeat(500) }),
    );
  });

  it("危険理由のエラーは入力時と高以外への変更時に解消する", async () => {
    const user = userEvent.setup();
    renderPetEditModal({ ...basePet, dangerLevel: "高", dangerReason: "" });

    await user.click(screen.getByRole("button", { name: "更新" }));
    expect(screen.getByText("危険度が高の場合は理由を入力してください")).toBeInTheDocument();

    await user.type(screen.getByLabelText("危険と判断した理由"), "噛む");
    expect(screen.queryByText("危険度が高の場合は理由を入力してください")).not.toBeInTheDocument();

    await user.clear(screen.getByLabelText("危険と判断した理由"));
    await user.click(screen.getByRole("button", { name: "更新" }));
    expect(screen.getByText("危険度が高の場合は理由を入力してください")).toBeInTheDocument();

    await chooseDangerLevel(user, "低");
    expect(screen.queryByText("危険度が高の場合は理由を入力してください")).not.toBeInTheDocument();
  });
});
