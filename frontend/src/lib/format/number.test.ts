import { describe, expect, it } from "vitest";
import {
  formatCount,
  formatCurrency,
  formatCurrencyIfNonzero,
  formatCurrencyOrDash,
  formatDecimal,
  formatNumber,
  formatPercent,
} from "./number";

describe("formatNumber", () => {
  it("3 桁区切りで整形し単位を付けない", () => {
    expect(formatNumber(1234567)).toBe("1,234,567");
    expect(formatNumber(0)).toBe("0");
  });

  it("負数は符号を保つ", () => {
    expect(formatNumber(-1234)).toBe("-1,234");
  });
});

describe("formatCount", () => {
  it("既定の単位は件で、数字と単位の間に空白を入れない", () => {
    expect(formatCount(1234)).toBe("1,234件");
  });

  it("単位を差し替えられる", () => {
    expect(formatCount(3, "名")).toBe("3名");
  });
});

describe("formatDecimal", () => {
  it("小数桁を固定する", () => {
    expect(formatDecimal(38.5, 1)).toBe("38.5");
    expect(formatDecimal(38, 1)).toBe("38.0");
    expect(formatDecimal(2.345, 2)).toBe("2.35");
  });
});

describe("formatPercent", () => {
  it("比率を小数 1 桁の百分率にする", () => {
    expect(formatPercent(0.123)).toBe("12.3%");
    expect(formatPercent(1)).toBe("100.0%");
  });

  it("小数桁を指定できる", () => {
    expect(formatPercent(0.5, 0)).toBe("50%");
  });
});

describe("formatCurrency", () => {
  it("0 は ¥0 と表示する", () => {
    expect(formatCurrency(0)).toBe("¥0");
  });

  it("負額は符号のまま表示する", () => {
    expect(formatCurrency(-3000)).toBe("-¥3,000");
  });

  it("正の金額を ¥ 区切りで表示する", () => {
    expect(formatCurrency(1234)).toBe("¥1,234");
  });

  it("null / undefined はEM DASH にする", () => {
    expect(formatCurrency(null)).toBe("—");
    expect(formatCurrency(undefined)).toBe("—");
  });
});

describe("formatCurrencyIfNonzero", () => {
  it("正の金額を ¥ 区切りで表示する", () => {
    expect(formatCurrencyIfNonzero(1234)).toBe("¥1,234");
  });

  it("負額は符号のまま表示する", () => {
    expect(formatCurrencyIfNonzero(-500)).toBe("-¥500");
  });

  it("0 / null / undefined はEM DASH にする", () => {
    expect(formatCurrencyIfNonzero(0)).toBe("—");
    expect(formatCurrencyIfNonzero(null)).toBe("—");
    expect(formatCurrencyIfNonzero(undefined)).toBe("—");
  });
});

describe("formatCurrencyOrDash", () => {
  it("正の金額を ¥ 区切りで表示する", () => {
    expect(formatCurrencyOrDash(1234)).toBe("¥1,234");
  });

  it("0 はEM DASH にする", () => {
    expect(formatCurrencyOrDash(0)).toBe("—");
  });

  it("負値はEM DASH にする", () => {
    expect(formatCurrencyOrDash(-500)).toBe("—");
  });

  it("null/undefined はEM DASH にする", () => {
    expect(formatCurrencyOrDash(null)).toBe("—");
    expect(formatCurrencyOrDash(undefined)).toBe("—");
  });

  it("7 桁区切り", () => {
    expect(formatCurrencyOrDash(1234567)).toBe("¥1,234,567");
  });
});
