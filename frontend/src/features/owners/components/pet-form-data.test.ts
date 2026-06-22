import { describe, expect, it } from "vitest";

import { createPetFormData } from "./pet-form-data";

describe("createPetFormData", () => {
  it("血液型とマイクロチップ番号を初期値として保持する", () => {
    const form = createPetFormData({
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
      food: "療法食",
      environment: "室内",
      remarks: "",
    });

    expect(form.bloodType).toBe("DEA1.1陽性");
    expect(form.microchipNumber).toBe("392140000123456");
  });
});
