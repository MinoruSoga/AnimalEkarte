import { describe, expect, it } from "vitest";

import type { TreatmentFormData } from "../components/TreatmentItemSidePanel";
import {
  buildProcedureCreateRequest,
  buildProcedureUpdateRequest,
} from "./treatment-plan-master-model";

function makeFormData(overrides: Partial<TreatmentFormData> = {}): TreatmentFormData {
  return {
    name: "V04処置テスト",
    price: 500,
    description: "",
    isActive: true,
    taxType: "excluded",
    taxRate: 0.1,
    isNonInsurance: false,
    anesthesia: "none",
    ...overrides,
  };
}

describe("buildProcedureCreateRequest", () => {
  it.each([
    ["none", "none"],
    ["local", "local"],
    ["sedation", "sedation"],
    ["general", "general"],
  ] as const)("sends anesthesia=%s when form selects %s", (selected) => {
    const request = buildProcedureCreateRequest(makeFormData({ anesthesia: selected }));
    expect(request.anesthesia).toBe(selected);
    expect(request.name).toBe("V04処置テスト");
    expect(request.price).toBe(500);
  });

  it("includes tax and parent fields alongside anesthesia", () => {
    const request = buildProcedureCreateRequest(
      makeFormData({
        anesthesia: "general",
        taxType: "included",
        taxRate: 0.08,
        parentId: "12",
        description: "備考",
      }),
    );
    expect(request).toMatchObject({
      name: "V04処置テスト",
      price: 500,
      anesthesia: "general",
      tax_type: "included",
      tax_rate: 0.08,
      parent_id: 12,
      description: "備考",
      is_active: true,
    });
  });
});

describe("buildProcedureUpdateRequest", () => {
  it("forwards anesthesia change and clears parent when empty string", () => {
    const request = buildProcedureUpdateRequest(
      makeFormData({ anesthesia: "local", parentId: "" }),
    );
    expect(request.anesthesia).toBe("local");
    expect(request.clear_parent_id).toBe(true);
  });
});
