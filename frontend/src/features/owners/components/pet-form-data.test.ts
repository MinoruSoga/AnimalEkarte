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

  it("新規ペットだけを生存で初期化し、既存ペットの欠損statusは不明にする", () => {
    expect(createPetFormData().status).toBe("生存");

    const existing = createPetFormData({
      id: "7",
      petNumber: "P-7",
      petName: "合成ペット",
      status: "",
      species: "犬",
      gender: "不明",
      birthDate: "",
      color: "",
      weight: "",
      environment: "",
      remarks: "",
    });

    expect(existing.status).toBe("不明");
  });

  it("既存死亡ペットのstatusと死亡日時をモーダル初期値へ保持する", () => {
    const deceasedAt = "2026-07-10T12:00:00+09:00";
    const form = createPetFormData({
      id: "7",
      petNumber: "P-7",
      petName: "合成ペット",
      status: "死亡",
      species: "犬",
      gender: "不明",
      birthDate: "",
      color: "",
      weight: "",
      environment: "",
      remarks: "",
      deceasedAt,
    });

    expect(form).toEqual(expect.objectContaining({ status: "死亡", deceasedAt }));
  });

  it("BUG-022: pending フラグをモーダル formData に保持する", () => {
    const form = createPetFormData({
      id: "temp-99",
      isPending: true,
      petNumber: "",
      petName: "未保存ポチ",
      status: "生存",
      species: "犬",
      gender: "不明",
      birthDate: "",
      color: "",
      weight: "",
      environment: "",
      remarks: "",
    });

    expect(form.id).toBe("temp-99");
    expect(form.isPending).toBe(true);
  });
});
