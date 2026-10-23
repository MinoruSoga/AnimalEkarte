import { describe, expect, it } from "vitest";

import {
  formatMerchandiseTaxRate,
  isMerchandiseTaxRate,
  merchandiseToFormData,
} from "./merchandise-side-panel-model";

// EMR-530: preset 判定と表示は十進文字列の exact 処理（float の桁落ち・誤一致をしない）。
describe("isMerchandiseTaxRate", () => {
  it("matches numerically-equal notations across trailing zeros", () => {
    expect(isMerchandiseTaxRate("0.10", "0.1")).toBe(true);
    expect(isMerchandiseTaxRate("0.1", "0.10")).toBe(true);
    expect(isMerchandiseTaxRate("0.08", "0.08")).toBe(true);
  });

  it("does not match a value that only looks equal after float rounding", () => {
    expect(isMerchandiseTaxRate("0.10000000000000001", "0.1")).toBe(false);
    expect(isMerchandiseTaxRate("0.080", "0.08")).toBe(true);
    expect(isMerchandiseTaxRate("0.09", "0.1")).toBe(false);
  });

  it("returns false for out-of-contract tokens", () => {
    expect(isMerchandiseTaxRate("abc", "0.1")).toBe(false);
    expect(isMerchandiseTaxRate("-0.1", "0.1")).toBe(false);
  });
});

describe("formatMerchandiseTaxRate", () => {
  it("renders presets from exact string comparison", () => {
    expect(formatMerchandiseTaxRate("0.1")).toBe("10%");
    expect(formatMerchandiseTaxRate("0.10")).toBe("10%");
    expect(formatMerchandiseTaxRate("0.08")).toBe("8%");
  });

  it("renders arbitrary precision by exact decimal shift, not float", () => {
    expect(formatMerchandiseTaxRate("0.10000000000000001")).not.toBe("10%");
    expect(formatMerchandiseTaxRate("0.08333")).toBe("8.333%");
    expect(formatMerchandiseTaxRate("1")).toBe("100%");
    expect(formatMerchandiseTaxRate("0")).toBe("0%");
    expect(formatMerchandiseTaxRate("0.0050")).toBe("0.5%");
  });

  it("falls back to the raw token for out-of-contract input", () => {
    expect(formatMerchandiseTaxRate("invalid")).toBe("invalid");
  });
});

// EMR-537 AC4: GET の decimal 文字列が form 値へ丸めずに通ること（master api の
// transform は tax_rate を string のまま透過するため、form 変換の透過をここで固定）。
describe("merchandiseToFormData (EMR-537 AC4)", () => {
  const item = {
    id: "merch-1",
    name: "既存品目",
    category: "goods",
    unitPrice: 1200,
    taxType: "excluded" as const,
    taxRate: "0.1",
    isActive: true,
    sortOrder: 1,
  };

  it.each(["0", "0.10", "0.08333", "0.10000000000000001"])(
    "keeps GET tax_rate %s unrounded in the form value",
    (taxRate) => {
      expect(merchandiseToFormData({ ...item, taxRate }).taxRate).toBe(taxRate);
    },
  );

  it("new panel defaults to the exact string 0.1", () => {
    expect(merchandiseToFormData(null).taxRate).toBe("0.1");
  });
});
