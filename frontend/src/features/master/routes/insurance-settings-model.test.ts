import { describe, expect, it } from "vitest";

import { buildInsuranceCreateRequest, validateInsuranceForm } from "./insurance-settings-model";

const base = {
  name: "V04保険",
  description: "",
  coverageRate: "50",
  contactPhone: "",
  isActive: true,
};

describe("validateInsuranceForm (BUG-026)", () => {
  it("名称未入力を拒否する", () => {
    expect(validateInsuranceForm({ ...base, name: "  " })).toBe("名称は必須です");
  });

  it("補償率 0 と 100 を許可する", () => {
    expect(validateInsuranceForm({ ...base, coverageRate: "0" })).toBeNull();
    expect(validateInsuranceForm({ ...base, coverageRate: "100" })).toBeNull();
  });

  it("補償率 101 を拒否する（成功トースト経路に入らせない）", () => {
    expect(validateInsuranceForm({ ...base, coverageRate: "101" })).toBe(
      "補償率は0〜100の範囲で入力してください",
    );
    expect(validateInsuranceForm({ ...base, coverageRate: "150" })).toBe(
      "補償率は0〜100の範囲で入力してください",
    );
  });

  it("負数・小数・非数を拒否する", () => {
    expect(validateInsuranceForm({ ...base, coverageRate: "-1" })).toBe(
      "補償率は0〜100の範囲で入力してください",
    );
    expect(validateInsuranceForm({ ...base, coverageRate: "10.5" })).toBe(
      "補償率は0〜100の整数で入力してください",
    );
    expect(validateInsuranceForm({ ...base, coverageRate: "abc" })).toBe(
      "補償率は0〜100の整数で入力してください",
    );
  });

  it("空の補償率は 0 としてリクエスト化し、validate は通す", () => {
    expect(validateInsuranceForm({ ...base, coverageRate: "" })).toBeNull();
    expect(buildInsuranceCreateRequest({ ...base, coverageRate: "" }).coverage_rate).toBe(0);
  });
});
