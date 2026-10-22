import { describe, expect, it } from "vitest";

import { formatMerchandiseTaxRate, isMerchandiseTaxRate } from "./merchandise-side-panel-model";

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
