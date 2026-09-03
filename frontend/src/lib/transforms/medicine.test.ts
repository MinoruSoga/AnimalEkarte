import { describe, it, expect } from "vitest";
import { transformBackendMedicineToFrontend } from "./medicine";
import type { Medicine as BackendMedicine } from "@/types/generated/models";

const minimalMedicine: BackendMedicine = {
  id: 1,
  clinic_id: 1,
  name: "テスト薬剤",
  is_active: true,
  is_non_insurance: false,
  description: "",
  default_quantity: 1,
  tax_type: "excluded",
  tax_rate: 0.1,
  sort_order: 0,
  created_at: "",
  updated_at: "",
};

describe("transformBackendMedicineToFrontend", () => {
  it("is_non_insurance=false をそのままマップする", () => {
    const result = transformBackendMedicineToFrontend(minimalMedicine);
    expect(result.isNonInsurance).toBe(false);
  });

  it("is_non_insurance=true をそのままマップする", () => {
    const result = transformBackendMedicineToFrontend({
      ...minimalMedicine,
      is_non_insurance: true,
    });
    expect(result.isNonInsurance).toBe(true);
  });

  it("id を文字列に変換する", () => {
    expect(transformBackendMedicineToFrontend(minimalMedicine).id).toBe("1");
  });

  it("name をそのまま返す", () => {
    expect(transformBackendMedicineToFrontend(minimalMedicine).name).toBe("テスト薬剤");
  });

  it("calculation_type 未設定時は none にデフォルトする(#201 default-deny)", () => {
    const result = transformBackendMedicineToFrontend(minimalMedicine);
    expect(result.calculationType).toBe("none");
  });

  it("calculation_type=per_weight をそのままマップする", () => {
    const result = transformBackendMedicineToFrontend({
      ...minimalMedicine,
      calculation_type: "per_weight",
      strength: 50,
      frequency_per_day: 2,
      default_duration_days: 7,
    });
    expect(result.calculationType).toBe("per_weight");
    expect(result.strength).toBe(50);
    expect(result.frequencyPerDay).toBe(2);
    expect(result.defaultDurationDays).toBe(7);
  });

  it("dose_params 未設定時は空配列にする", () => {
    expect(transformBackendMedicineToFrontend(minimalMedicine).doseParams).toEqual([]);
  });

  it("dose_params をそのまま透過する", () => {
    const doseParams = [
      {
        id: 1,
        clinic_id: 1,
        medicine_id: 1,
        species: "dog" as const,
        dose_basis: "per_administration" as const,
        dose_per_kg: 10,
        notes: "",
        created_at: "",
        updated_at: "",
      },
    ];
    const result = transformBackendMedicineToFrontend({
      ...minimalMedicine,
      dose_params: doseParams,
    });
    expect(result.doseParams).toEqual(doseParams);
  });
});
