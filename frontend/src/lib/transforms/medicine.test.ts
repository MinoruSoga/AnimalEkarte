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
    const result = transformBackendMedicineToFrontend({ ...minimalMedicine, is_non_insurance: true });
    expect(result.isNonInsurance).toBe(true);
  });

  it("id を文字列に変換する", () => {
    expect(transformBackendMedicineToFrontend(minimalMedicine).id).toBe("1");
  });

  it("name をそのまま返す", () => {
    expect(transformBackendMedicineToFrontend(minimalMedicine).name).toBe("テスト薬剤");
  });
});
