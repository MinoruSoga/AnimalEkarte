import { describe, it, expect } from "vitest";
import { summarizeCategoryTotals } from "./category-breakdown";

describe("summarizeCategoryTotals", () => {
  it("sums payment-method amounts per category and groups them by display category", () => {
    const breakdown = {
      categories: {
        // 同じ表示区分「診療」にまとまる複数の生カテゴリ
        examination: { 現金: 3000, クレジットカード: 2000 },
        medicine: { 現金: 1000 },
        // 単独区分
        surgery: { 現金: 5000 },
        food: { クレジットカード: 1500 },
      },
    };

    const result = summarizeCategoryTotals(breakdown);

    // 診療 = examination(5000) + medicine(1000) = 6000 に集約される
    expect(result).toContainEqual({ label: "診療", total: 6000 });
    expect(result).toContainEqual({ label: "外科", total: 5000 });
    expect(result).toContainEqual({ label: "フード", total: 1500 });
    // 表示区分順を維持（診療 → 外科 → ... → フード）
    expect(result.map((r) => r.label)).toEqual(["診療", "外科", "フード"]);
  });

  it("preserves unknown categories not present in the display grouping", () => {
    const breakdown = { categories: { mystery: { 現金: 800 } } };
    const result = summarizeCategoryTotals(breakdown);
    expect(result).toContainEqual({ label: "mystery", total: 800 });
  });

  it("omits zero-total categories", () => {
    const breakdown = { categories: { surgery: { 現金: 0 } } };
    expect(summarizeCategoryTotals(breakdown)).toEqual([]);
  });

  it("returns an empty array for null / malformed / missing input instead of throwing", () => {
    expect(summarizeCategoryTotals(null)).toEqual([]);
    expect(summarizeCategoryTotals(undefined)).toEqual([]);
    expect(summarizeCategoryTotals("not json")).toEqual([]);
    expect(summarizeCategoryTotals({})).toEqual([]);
    expect(summarizeCategoryTotals({ categories: null })).toEqual([]);
    expect(summarizeCategoryTotals({ categories: { food: "bad" } })).toEqual([]);
  });

  it("ignores non-numeric method amounts without corrupting the total", () => {
    const breakdown = {
      categories: { food: { 現金: 1200, broken: "x", missing: null } },
    };
    expect(summarizeCategoryTotals(breakdown)).toEqual([{ label: "フード", total: 1200 }]);
  });
});
