import { describe, expect, it } from "vitest";

import type { HospitalizationFormData } from "../lib/hospitalization-side-panel-model";
import {
  buildHospitalizationCreateRequest,
  buildHospitalizationUpdateRequest,
} from "./hospitalization-settings-model";

function makeFormData(overrides: Partial<HospitalizationFormData> = {}): HospitalizationFormData {
  return {
    name: "一般入院プラン",
    price: 1200,
    description: "",
    isActive: true,
    bodySize: "medium",
    billingUnit: "per_day",
    taxType: "excluded",
    taxRate: 0.1,
    ...overrides,
  };
}

describe("buildHospitalizationCreateRequest", () => {
  it("sends explicit price 0 on create instead of omitting (EMR-528)", () => {
    const request = buildHospitalizationCreateRequest(makeFormData({ price: 0 }));
    expect(request.price).toBe(0);
    expect(request.name).toBe("一般入院プラン");
  });

  it("persists priced create payload for save→reread", () => {
    const request = buildHospitalizationCreateRequest(makeFormData({ price: 1200 }));
    expect(request.price).toBe(1200);
  });
});

describe("buildHospitalizationUpdateRequest", () => {
  it("sends explicit price 0 on update instead of omitting", () => {
    const request = buildHospitalizationUpdateRequest(makeFormData({ price: 0 }));
    expect(request.price).toBe(0);
  });

  it("persists edited price 2300 on update", () => {
    const request = buildHospitalizationUpdateRequest(makeFormData({ price: 2300 }));
    expect(request.price).toBe(2300);
  });

  it("sends empty description so a non-empty note can be cleared (EMR-528)", () => {
    const request = buildHospitalizationUpdateRequest(makeFormData({ description: "" }));
    expect(request).toHaveProperty("description", "");
  });

  it("sends explicit null for unselected body size and billing unit (EMR-528)", () => {
    const request = buildHospitalizationUpdateRequest(
      makeFormData({ bodySize: "", billingUnit: "" }),
    );
    expect(request.body_size).toBeNull();
    expect(request.billing_unit).toBeNull();
  });
});
