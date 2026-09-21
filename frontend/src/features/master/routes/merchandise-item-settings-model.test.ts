import { describe, expect, it } from "vitest";

import type { MerchandiseFormData } from "../lib/merchandise-side-panel-model";
import {
  buildMerchandiseCreateRequest,
  buildMerchandiseUpdateRequest,
} from "./merchandise-item-settings-model";

function makeFormData(overrides: Partial<MerchandiseFormData> = {}): MerchandiseFormData {
  return {
    name: "療法食",
    category: "goods",
    unitPrice: 1200,
    taxType: "excluded",
    taxRate: 0.1,
    isActive: true,
    ...overrides,
  };
}

describe("buildMerchandiseCreateRequest", () => {
  it("persists unit_price 0 distinctly from missing price", () => {
    const request = buildMerchandiseCreateRequest(makeFormData({ unitPrice: 0 }));
    expect(request.unit_price).toBe(0);
    expect(request).toHaveProperty("unit_price");
  });

  it("persists unit_price 1200 for accounting merchandise selection", () => {
    const request = buildMerchandiseCreateRequest(makeFormData({ unitPrice: 1200 }));
    expect(request.unit_price).toBe(1200);
    expect(request.category).toBe("goods");
  });
});

describe("buildMerchandiseUpdateRequest", () => {
  it("persists edited unit_price 2300 on update", () => {
    const request = buildMerchandiseUpdateRequest(makeFormData({ unitPrice: 2300 }));
    expect(request.unit_price).toBe(2300);
  });
});
