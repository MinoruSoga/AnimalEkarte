import { describe, it, expect } from "vitest";
import { summarizeCategoryTotals } from "./category-breakdown";
import { transformCashRegisterClose } from "@/lib/transforms/cash-register";
import type { CashRegisterClose as BackendCashRegisterClose } from "@/types/generated/models";

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

  it("accepts transformCashRegisterClose's unknown-typed categoryBreakdown output as-is (R-F6-S3)", () => {
    const backendClose: BackendCashRegisterClose = {
      id: 1,
      clinic_id: 1,
      close_date: "2026-07-10",
      period: "am",
      theoretical_cash: 10000,
      actual_cash: 10000,
      cash_difference: 0,
      category_breakdown: { categories: { examination: { 現金: 3000 } } },
      memo: "",
      closed_at: "2026-07-10T09:00:00Z",
      created_at: "2026-07-10T09:00:00Z",
      updated_at: "2026-07-10T09:00:00Z",
    };

    const result = transformCashRegisterClose(backendClose);
    const totals = summarizeCategoryTotals(result.categoryBreakdown);

    expect(totals).toContainEqual({ label: "診療", total: 3000 });
  });
});

describe("summarizeCategoryTotals DEC-40 未分類・要確認件数", () => {
  it("新 snapshot の unclassified_other_count を件数として載せる", () => {
    const breakdown = {
      categories: { other: { 現金: 800 } },
      unclassified_other_count: 2,
    };
    expect(summarizeCategoryTotals(breakdown)).toContainEqual({
      label: "未分類・要確認",
      total: 800,
      count: 2,
    });
  });

  it("旧 snapshot（フィールド欠落）は count=null（記録なし）。0 扱いにしない", () => {
    const breakdown = {
      categories: { other: { 現金: 500 } },
    };
    const row = summarizeCategoryTotals(breakdown).find((r) => r.label === "未分類・要確認");
    expect(row).toEqual({ label: "未分類・要確認", total: 500, count: null });
    expect(row?.count).not.toBe(0);
  });

  it("件数 0・金額 0 の新 snapshot は行を出さない", () => {
    expect(
      summarizeCategoryTotals({ categories: {}, unclassified_other_count: 0 }),
    ).toEqual([]);
  });

  it("件数のみ（金額 0）の新 snapshot は行を出す", () => {
    expect(
      summarizeCategoryTotals({ categories: {}, unclassified_other_count: 1 }),
    ).toEqual([{ label: "未分類・要確認", total: 0, count: 1 }]);
  });
});

