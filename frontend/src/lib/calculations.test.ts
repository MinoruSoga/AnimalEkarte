import { describe, expect, it } from "vitest";

import { calculateBillingTotals } from "./calculations";

/**
 * BUG-006 / BUG-013: 軽減税率明細のフッター合計が 10% 固定にならないこと。
 * BE CalculateBillingTotals と整合（明細ごと税額を合算、外税は Math.round）。
 */
describe("calculateBillingTotals", () => {
  it("BUG-006: 8% 単品（ヒルズ z/d 2kg ¥5,800）は税額 464・合計 6,264", () => {
    const result = calculateBillingTotals(
      [{ unitPrice: 5800, quantity: 1, discountAmount: 0, taxRate: 0.08 }],
      0,
      0,
    );
    expect(result.subtotal).toBe(5800);
    expect(result.tax).toBe(464);
    expect(result.total).toBe(6264);
    expect(result.billingAmount).toBe(6264);
  });

  it("10% のみ会計は従来どおり（回帰なし）", () => {
    const result = calculateBillingTotals(
      [{ unitPrice: 1000, quantity: 1, discountAmount: 0, taxRate: 0.1 }],
      0,
      0,
    );
    expect(result.subtotal).toBe(1000);
    expect(result.tax).toBe(100);
    expect(result.total).toBe(1100);
  });

  it("8% と 10% 混在は明細税率どおり合算する", () => {
    const result = calculateBillingTotals(
      [
        { unitPrice: 5800, quantity: 1, discountAmount: 0, taxRate: 0.08 },
        { unitPrice: 1000, quantity: 1, discountAmount: 0, taxRate: 0.1 },
      ],
      0,
      0,
    );
    // 5800+1000=6800; tax 464+100=564; total 7364
    expect(result.subtotal).toBe(6800);
    expect(result.tax).toBe(564);
    expect(result.total).toBe(7364);
    expect(result.billingAmount).toBe(7364);
  });

  it("taxRate 未指定の明細は引数 taxRate（既定 10%）を使う（後方互換）", () => {
    const result = calculateBillingTotals([{ unitPrice: 1000, quantity: 1 }], 0, 0);
    expect(result.tax).toBe(100);
    expect(result.total).toBe(1100);
  });

  it("明細割引後の課税ベースで税額を出す（#85 / BE 整合）", () => {
    const result = calculateBillingTotals(
      [{ unitPrice: 1000, quantity: 2, discountAmount: 500, taxRate: 0.1 }],
      0,
      0,
    );
    expect(result.subtotal).toBe(1500);
    expect(result.tax).toBe(150);
    expect(result.total).toBe(1650);
  });

  it("内税は total に税額を二重加算しない", () => {
    const result = calculateBillingTotals(
      [{ unitPrice: 1100, quantity: 1, taxRate: 0.1, taxType: "included" }],
      0,
      0,
    );
    expect(result.subtotal).toBe(1100);
    expect(result.tax).toBe(100);
    expect(result.total).toBe(1100);
  });

  it("非課税は税額 0", () => {
    const result = calculateBillingTotals(
      [{ unitPrice: 500, quantity: 3, taxRate: 0.1, taxType: "exempt" }],
      0,
      0,
    );
    expect(result.subtotal).toBe(1500);
    expect(result.tax).toBe(0);
    expect(result.total).toBe(1500);
  });
});
