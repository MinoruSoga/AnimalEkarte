import { describe, expect, it } from "vitest";

import type { MedicineDoseParam } from "@/types/generated/models";

import {
  buildUpsertDoseParamRequest,
  doseParamToFormData,
  findDoseParamBySpecies,
  INITIAL_DOSE_PARAM_FORM,
  isDoseParamFormEmpty,
  validateDoseParamForm,
  type DoseParamFormData,
} from "../lib/medicine-dose-params-editor-model";

function makeForm(overrides: Partial<DoseParamFormData> = {}): DoseParamFormData {
  return { ...INITIAL_DOSE_PARAM_FORM, dosePerKg: "10", maxMgPerKg: "20", ...overrides };
}

function makeParam(overrides: Partial<MedicineDoseParam> = {}): MedicineDoseParam {
  return {
    id: 1,
    clinic_id: 1,
    medicine_id: 2,
    species: "dog",
    dose_basis: "per_administration",
    dose_per_kg: 5,
    notes: "",
    created_at: "",
    updated_at: "",
    ...overrides,
  };
}

describe("validateDoseParamForm", () => {
  it("passes with dose_per_kg and max_mg_per_kg only", () => {
    const result = validateDoseParamForm(makeForm());
    expect(result).toEqual({ isValid: true, errors: [] });
  });

  it("rejects dose_per_kg <= 0", () => {
    const result = validateDoseParamForm(makeForm({ dosePerKg: "0" }));
    expect(result.isValid).toBe(false);
    expect(result.errors).toContain("投与量(mg/kg)は0より大きい値を入力してください");
  });

  it("rejects blank dose_per_kg", () => {
    const result = validateDoseParamForm(makeForm({ dosePerKg: "" }));
    expect(result.isValid).toBe(false);
  });

  it("rejects min > max", () => {
    const result = validateDoseParamForm(makeForm({ minMgPerKg: "30", maxMgPerKg: "20" }));
    expect(result.errors).toContain("下限は上限以下にしてください");
  });

  it("accepts min == max (BUG-013)", () => {
    const result = validateDoseParamForm(
      makeForm({
        dosePerKg: "10",
        minMgPerKg: "10",
        maxMgPerKg: "10",
      }),
    );
    expect(result.isValid).toBe(true);
  });

  it("rejects dose_per_kg below min", () => {
    const result = validateDoseParamForm(
      makeForm({ dosePerKg: "5", minMgPerKg: "10", maxMgPerKg: "20" }),
    );
    expect(result.errors).toContain("投与量は下限以上にしてください");
  });

  it("rejects dose_per_kg above max", () => {
    const result = validateDoseParamForm(makeForm({ dosePerKg: "30", maxMgPerKg: "20" }));
    expect(result.errors).toContain("投与量は上限以下にしてください");
  });

  it("rejects when neither max_mg_per_kg nor absolute_max_dose is set", () => {
    const result = validateDoseParamForm(makeForm({ maxMgPerKg: "" }));
    expect(result.errors).toContain("過量防止のため上限(mg/kgまたはmg)のいずれかは必須です");
  });

  it("passes when only absolute_max_dose is set (no max_mg_per_kg)", () => {
    const result = validateDoseParamForm(makeForm({ maxMgPerKg: "", absoluteMaxDose: "100" }));
    expect(result.isValid).toBe(true);
  });

  it("rejects rounding_step set without rounding_mode", () => {
    const result = validateDoseParamForm(makeForm({ roundingStep: "0.5", roundingMode: "" }));
    expect(result.errors).toContain(
      "丸め幅と丸め方向はどちらも設定するか、どちらも未設定にしてください",
    );
  });

  it("rejects rounding_mode set without rounding_step", () => {
    const result = validateDoseParamForm(makeForm({ roundingStep: "", roundingMode: "up" }));
    expect(result.errors).toContain(
      "丸め幅と丸め方向はどちらも設定するか、どちらも未設定にしてください",
    );
  });

  it("passes when both rounding_step and rounding_mode are set together", () => {
    const result = validateDoseParamForm(makeForm({ roundingStep: "0.5", roundingMode: "up" }));
    expect(result.isValid).toBe(true);
  });
});

describe("buildUpsertDoseParamRequest", () => {
  it("omits unset optional fields", () => {
    const request = buildUpsertDoseParamRequest(makeForm());
    expect(request).toEqual({
      dose_basis: "per_administration",
      dose_per_kg: 10,
      max_mg_per_kg: 20,
    });
  });

  it("includes all optional fields when set", () => {
    const request = buildUpsertDoseParamRequest(
      makeForm({
        minMgPerKg: "5",
        absoluteMaxDose: "50",
        roundingStep: "0.5",
        roundingMode: "up",
        notes: "備考",
      }),
    );
    expect(request).toEqual({
      dose_basis: "per_administration",
      dose_per_kg: 10,
      min_mg_per_kg: 5,
      max_mg_per_kg: 20,
      absolute_max_dose: 50,
      rounding_step: 0.5,
      rounding_mode: "up",
      notes: "備考",
    });
  });
});

describe("doseParamToFormData", () => {
  it("returns the initial form when param is undefined", () => {
    expect(doseParamToFormData(undefined)).toEqual(INITIAL_DOSE_PARAM_FORM);
  });

  it("maps an existing param to form data", () => {
    const param = makeParam({
      dose_basis: "per_day",
      dose_per_kg: 15,
      min_mg_per_kg: 5,
      max_mg_per_kg: 25,
      rounding_mode: "nearest",
      notes: "既存メモ",
    });

    expect(doseParamToFormData(param)).toEqual({
      doseBasis: "per_day",
      dosePerKg: "15",
      minMgPerKg: "5",
      maxMgPerKg: "25",
      absoluteMaxDose: "",
      roundingStep: "",
      roundingMode: "nearest",
      notes: "既存メモ",
    });
  });
});

describe("findDoseParamBySpecies", () => {
  it("finds the matching species entry", () => {
    const dog = makeParam({ species: "dog" });
    const cat = makeParam({ species: "cat", id: 2 });
    expect(findDoseParamBySpecies([dog, cat], "cat")).toBe(cat);
  });

  it("returns undefined when no match exists", () => {
    expect(findDoseParamBySpecies([makeParam({ species: "dog" })], "cat")).toBeUndefined();
  });

  it("returns undefined when params is undefined", () => {
    expect(findDoseParamBySpecies(undefined, "dog")).toBeUndefined();
  });
});

describe("isDoseParamFormEmpty", () => {
  it("treats the initial form as empty", () => {
    expect(isDoseParamFormEmpty(INITIAL_DOSE_PARAM_FORM)).toBe(true);
  });

  it("treats a filled dose as not empty", () => {
    expect(isDoseParamFormEmpty(makeForm())).toBe(false);
  });
});
