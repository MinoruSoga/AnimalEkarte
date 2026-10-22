import { describe, expect, it } from "vitest";

import { exactDecimalToNumber, numberToExactDecimal } from "./exact-decimal";

describe("exactDecimalToNumber", () => {
  it("decimal 文字列を number へ変換する", () => {
    expect(exactDecimalToNumber("0.1", "tax_rate")).toBe(0.1);
    expect(exactDecimalToNumber("0.10", "tax_rate")).toBe(0.1);
    expect(exactDecimalToNumber("0", "tax_rate")).toBe(0);
    expect(exactDecimalToNumber("12.5", "strength")).toBe(12.5);
  });

  it.each(["", "-1", "1e2", "+1", "01", "NaN", "1."])("wire 外の %s は例外", (value) => {
    expect(() => exactDecimalToNumber(value, "strength")).toThrow("strength");
  });

  it("旧 wire の JSON number は文字列化で素通りさせず例外", () => {
    const legacy: unknown = 0.1;
    expect(() => exactDecimalToNumber(legacy as string, "tax_rate")).toThrow("tax_rate");
  });
});

describe("numberToExactDecimal", () => {
  it("非負の有限 number を decimal 文字列へ変換する", () => {
    expect(numberToExactDecimal(0.1, "tax_rate")).toBe("0.1");
    expect(numberToExactDecimal(0.08, "tax_rate")).toBe("0.08");
    expect(numberToExactDecimal(0, "tax_rate")).toBe("0");
  });

  it.each([-0.1, Number.NaN, Number.POSITIVE_INFINITY, 1e-7])("%s は例外", (value) => {
    expect(() => numberToExactDecimal(value, "tax_rate")).toThrow("tax_rate");
  });
});
