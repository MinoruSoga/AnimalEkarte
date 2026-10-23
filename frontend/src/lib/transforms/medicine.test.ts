import { describe, it, expect } from "vitest";
import { transformBackendMedicineToFrontend } from "./medicine";
import type { LegacyMedicineResponse } from "@/types/medicine";

const minimalMedicine: LegacyMedicineResponse = {
  id: 1,
  clinic_id: 1,
  name: "テスト薬剤",
  is_active: true,
  is_non_insurance: false,
  description: "",
  default_quantity: "1",
  tax_type: "excluded",
  tax_rate: "0.10",
  sort_order: 0,
  calculation_type: "none",
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
      strength: "50",
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

  it("EMR-447: exact decimal 文字列を number へ明示変換する", () => {
    const result = transformBackendMedicineToFrontend({
      ...minimalMedicine,
      default_quantity: "1.50",
      tax_rate: "0.08",
      strength: "0.0005",
    });
    expect(result.defaultQuantity).toBe(1.5);
    expect(result.taxRate).toBe(0.08);
    expect(result.strength).toBe(0.0005);
  });

  it("EMR-536: tax_rate の wire 文字列を丸めず taxRateExact に保持する", () => {
    const result = transformBackendMedicineToFrontend({
      ...minimalMedicine,
      tax_rate: "0.1000000000000000001",
    });
    expect(result.taxRate).toBe(0.1);
    expect(result.taxRateExact).toBe("0.1000000000000000001");
  });

  it("EMR-447: default_quantity の null と明示0を区別する", () => {
    const absent = { ...minimalMedicine, default_quantity: null };
    expect(transformBackendMedicineToFrontend(absent).defaultQuantity).toBeUndefined();
    const zero = { ...minimalMedicine, default_quantity: "0", tax_rate: "0" };
    expect(transformBackendMedicineToFrontend(zero)).toMatchObject({
      defaultQuantity: 0,
      taxRate: 0,
    });
  });

  it("EMR-447: wire 外の値は黙って NaN にせず例外にする", () => {
    const malformed = { ...minimalMedicine, tax_rate: "1e-1" };
    expect(() => transformBackendMedicineToFrontend(malformed)).toThrow("tax_rate");
  });

  it.each([undefined, null])(
    "EMR-447/527: tax_rate の欠落（%s）を 0.1 で補わず例外にする",
    (taxRate) => {
      // 実行時の JSON 応答を再現する（undefined は key 自体が落ちる）。
      const missing: LegacyMedicineResponse = JSON.parse(
        JSON.stringify({ ...minimalMedicine, tax_rate: taxRate }),
      );
      expect(() => transformBackendMedicineToFrontend(missing)).toThrow("tax_rate");
    },
  );
});
