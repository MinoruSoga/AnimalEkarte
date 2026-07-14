import { describe, it, expect } from "vitest";
import {
  CREATE_STATUS_OPTIONS,
  EDIT_STATUS_OPTIONS,
} from "./EstimateForm";

describe("EstimateForm status options", () => {
  it("Create 用選択肢は draft / sent のみ", () => {
    expect(CREATE_STATUS_OPTIONS.map((o) => o.value)).toEqual([
      "draft",
      "sent",
    ]);
  });

  it("Edit 用選択肢は draft / sent / approved / rejected の 4 値", () => {
    expect(EDIT_STATUS_OPTIONS.map((o) => o.value)).toEqual([
      "draft",
      "sent",
      "approved",
      "rejected",
    ]);
  });

  it("Create 用選択肢に approved / rejected は含まれない", () => {
    const values = CREATE_STATUS_OPTIONS.map((o) => o.value);
    expect(values).not.toContain("approved");
    expect(values).not.toContain("rejected");
  });
});
