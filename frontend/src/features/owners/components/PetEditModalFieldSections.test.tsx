import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { MemoryRouter } from "react-router";
import { describe, expect, it } from "vitest";

import { PetEditModalFields } from "./PetEditModalFields";
import type { PetFormData } from "../types";

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
});
