import { describe, expect, it } from "vitest";
import { formatCurrency, formatCurrencyOrDash } from "./number";

describe("formatCurrency", () => {
  it("0 は ¥0 と表示する", () => {
    expect(formatCurrency(0)).toBe("¥0");
  });

  it("null はダッシュにする", () => {
    expect(formatCurrency(null)).toBe("-");
  });
});

describe("formatCurrencyOrDash", () => {
  it("正の金額を ¥ 区切りで表示する", () => {
    expect(formatCurrencyOrDash(1234)).toBe("¥1,234");
  });

  it("0 はダッシュにする", () => {
    expect(formatCurrencyOrDash(0)).toBe("-");
  });

  it("負値はダッシュにする", () => {
    expect(formatCurrencyOrDash(-500)).toBe("-");
  });

  it("null/undefined はダッシュにする", () => {
    expect(formatCurrencyOrDash(null)).toBe("-");
    expect(formatCurrencyOrDash(undefined)).toBe("-");
  });

  it("7 桁区切り", () => {
    expect(formatCurrencyOrDash(1234567)).toBe("¥1,234,567");
  });
});
