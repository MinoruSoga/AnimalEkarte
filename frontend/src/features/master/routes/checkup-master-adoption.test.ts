import { describe, expect, it } from "vitest";
import type { TreatmentFormData } from "../components/TreatmentItemSidePanel";
import { buildCheckupUpdateRequest } from "./treatment-plan-master-model";

const form: TreatmentFormData = {
  name: "受入項目",
  price: 500,
  description: "",
  isActive: true,
  taxType: "excluded",
  taxRate: 0.1,
  isNonInsurance: null,
  anesthesia: "none",
  version: 7,
};

describe("checkup consumer activation", () => {
  it("uses the captured version and preserves explicit parent clearing", () => {
    expect(buildCheckupUpdateRequest({ ...form, parentId: "" })).toMatchObject({
      version: 7,
      clear_parent_id: true,
    });
  });

  it("rejects legacy or invalid editing snapshots before transport", () => {
    for (const version of [undefined, null, 0, -1, 1.5, NaN, Infinity]) {
      expect(() => buildCheckupUpdateRequest({ ...form, version })).toThrow("編集開始時");
    }
  });
});
