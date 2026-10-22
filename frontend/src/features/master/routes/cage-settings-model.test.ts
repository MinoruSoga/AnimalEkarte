import { describe, expect, it } from "vitest";

import type { CageFormData } from "../lib/cage-side-panel-model";
import { buildCageCreateRequest, buildCageUpdateRequest } from "./cage-settings-model";

function makeFormData(overrides: Partial<CageFormData> = {}): CageFormData {
  return {
    name: "中型ケージ",
    cageType: "general",
    cageSize: "medium",
    price: 1200,
    description: "",
    isActive: true,
    ...overrides,
  };
}

describe("buildCageCreateRequest", () => {
  it("persists price 0 distinctly and does not invent billing ids", () => {
    const request = buildCageCreateRequest(makeFormData({ price: 0 }));
    expect(request.price).toBe(0);
    expect(request).not.toHaveProperty("cage_id");
    expect(request).not.toHaveProperty("billing_id");
  });

  it("persists price 1200 for cage master save→reread", () => {
    const request = buildCageCreateRequest(makeFormData({ price: 1200 }));
    expect(request.price).toBe(1200);
  });

  it("keeps create omitting an empty description", () => {
    const request = buildCageCreateRequest(makeFormData({ description: "" }));
    expect(request.description).toBeUndefined();
  });
});

describe("buildCageUpdateRequest", () => {
  it("persists edited cage price 2300", () => {
    const request = buildCageUpdateRequest(makeFormData({ price: 2300 }));
    expect(request.price).toBe(2300);
  });

  it("sends empty description so a non-empty note can be cleared (EMR-448)", () => {
    const request = buildCageUpdateRequest(makeFormData({ description: "" }));
    expect(request).toHaveProperty("description", "");
  });

  it("sends a new description after it was empty (EMR-448)", () => {
    const request = buildCageUpdateRequest(makeFormData({ description: "新備考" }));
    expect(request.description).toBe("新備考");
  });
});
