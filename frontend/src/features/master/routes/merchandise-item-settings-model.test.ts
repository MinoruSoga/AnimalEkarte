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
    // EMR-530: tax_rate は exact decimal 文字列（10% は "0.1"）。
    taxRate: "0.1",
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

  it("persists exact decimal tax_rate string without number conversion", () => {
    const request = buildMerchandiseCreateRequest(makeFormData({ taxRate: "0.10" }));
    expect(request.tax_rate).toBe("0.10");
  });

  it("persists explicit zero tax_rate without collapsing to default", () => {
    const request = buildMerchandiseCreateRequest(makeFormData({ taxRate: "0" }));
    expect(request.tax_rate).toBe("0");
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

// EMR-537 AC4: GET の decimal 文字列が form 値 → save payload へ丸めずに通ること。
// （GET transform 自体は api/merchandise-items.ts の string 透過。form/builder はここで固定。）
describe("GET decimal string → form → save payload (EMR-537 AC4)", () => {
  it.each(["0", "0.10", "0.08333", "0.10000000000000001"])(
    "create payload keeps tax_rate %s unrounded",
    (taxRate) => {
      const request = buildMerchandiseCreateRequest(makeFormData({ taxRate }));
      expect(request.tax_rate).toBe(taxRate);
    },
  );

  it.each(["0", "0.10", "0.08333", "0.10000000000000001"])(
    "update payload keeps tax_rate %s unrounded",
    (taxRate) => {
      const request = buildMerchandiseUpdateRequest(makeFormData({ taxRate }));
      expect(request.tax_rate).toBe(taxRate);
    },
  );
});
