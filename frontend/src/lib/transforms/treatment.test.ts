import { describe, it, expect } from "vitest";
import { transformExaminationType } from "./treatment";
import type { ExaminationType } from "@/types/generated/models";

const minimalExamType: ExaminationType = {
  id: 1,
  name: "血液検査",
  is_active: true,
  is_non_insurance: false,
  created_at: "",
  updated_at: "",
};

describe("transformExaminationType", () => {
  it("is_non_insurance=false をそのままマップする", () => {
    const result = transformExaminationType(minimalExamType);
    expect(result.isNonInsurance).toBe(false);
  });

  it("is_non_insurance=true をそのままマップする", () => {
    const result = transformExaminationType({ ...minimalExamType, is_non_insurance: true });
    expect(result.isNonInsurance).toBe(true);
  });

  it("id を文字列に変換する", () => {
    expect(transformExaminationType(minimalExamType).id).toBe("1");
  });

  it("name をそのまま返す", () => {
    expect(transformExaminationType(minimalExamType).name).toBe("血液検査");
  });
});
