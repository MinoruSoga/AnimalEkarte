import { describe, expect, it } from "vitest";

import { calculateBillingTotals } from "./calculations";

/**
 * EMR-164: 保険計算分岐の回帰テスト。
 *
 * calculateBillingTotals は isInsuranceApplicable=true の明細だけを
 * raw 単価×数量（税・明細割引前）で保険対象額に合算し、
 * insuranceAmount=floor(対象額×ratio)・billingAmount=max(0,total−insuranceAmount)
 * を返す（BE と同式・正の magnitude）。
 * ratio の新規選択肢は 0.5/0.7。0.9/1.0 は既存会計データ向けに残る
 * （features/accounting/components/InsuranceCard.tsx の LEGACY_INSURANCE_RATIO_LABELS）。
 */
describe("calculateBillingTotals 保険分岐 (EMR-164)", () => {
  // taxType "exempt" で税を 0 に固定し、保険分岐だけを検証する
  const applicable = { isInsuranceApplicable: true, taxType: "exempt" as const };

  it.each([
    { ratio: 0.5, insurance: 5000, billing: 5000 },
    { ratio: 0.7, insurance: 7000, billing: 3000 },
    { ratio: 0.9, insurance: 9000, billing: 1000 },
    { ratio: 1.0, insurance: 10000, billing: 0 },
  ])(
    "保険適用 ¥10,000・ratio=$ratio → insuranceAmount=$insurance・billingAmount=$billing",
    ({ ratio, insurance, billing }) => {
      const result = calculateBillingTotals(
        [{ unitPrice: 10000, quantity: 1, ...applicable }],
        0,
        0,
        0.1,
        ratio,
      );
      expect(result.total).toBe(10000);
      expect(result.insuranceAmount).toBe(insurance);
      expect(result.billingAmount).toBe(billing);
    },
  );

  it("ratio=0 は全額自己負担（insuranceAmount=0・billingAmount=total）", () => {
    const result = calculateBillingTotals(
      [{ unitPrice: 10000, quantity: 1, ...applicable }],
      0,
      0,
      0.1,
      0,
    );
    expect(result.insuranceAmount).toBe(0);
    expect(result.total).toBe(10000);
    expect(result.billingAmount).toBe(result.total);
  });

  it("isInsuranceApplicable 非対象の明細は保険対象額に含めない", () => {
    const result = calculateBillingTotals(
      [
        { unitPrice: 10000, quantity: 1, ...applicable },
        { unitPrice: 5000, quantity: 1, taxType: "exempt" }, // flag 未指定
        {
          unitPrice: 3000,
          quantity: 1,
          isInsuranceApplicable: false,
          taxType: "exempt",
        },
      ],
      0,
      0,
      0.1,
      1.0,
    );
    expect(result.total).toBe(18000);
    // 適用明細（¥10,000）だけが対象 — 未指定・false の明細は全額自己負担
    expect(result.insuranceAmount).toBe(10000);
    expect(result.billingAmount).toBe(8000);
  });

  it("insuranceAmount は floor（¥333×0.5=166.5 → 166）", () => {
    const result = calculateBillingTotals(
      [{ unitPrice: 333, quantity: 1, ...applicable }],
      0,
      0,
      0.1,
      0.5,
    );
    expect(result.insuranceAmount).toBe(166);
    expect(result.billingAmount).toBe(167);
  });

  it("quantity>1 は raw 単価×数量で保険対象額に乗る", () => {
    const result = calculateBillingTotals(
      [{ unitPrice: 2000, quantity: 3, ...applicable }],
      0,
      0,
      0.1,
      0.5,
    );
    expect(result.total).toBe(6000);
    expect(result.insuranceAmount).toBe(3000); // 2000×3×0.5
    expect(result.billingAmount).toBe(3000);
  });

  it("保険対象額は明細割引を引かない raw 単価×数量 — billingAmount は 0 に clamp", () => {
    const result = calculateBillingTotals(
      [{ unitPrice: 10000, quantity: 1, discountAmount: 9000, ...applicable }],
      0,
      0,
      0.1,
      1.0,
    );
    // lineBase=1,000 → total=1,000。保険対象額は raw ¥10,000 → max(0,…) で 0
    expect(result.total).toBe(1000);
    expect(result.insuranceAmount).toBe(10000);
    expect(result.billingAmount).toBe(0);
  });

  it("保険は税抜本体に適用され、外税の消費税は飼主負担に残る", () => {
    const result = calculateBillingTotals(
      [
        {
          unitPrice: 10000,
          quantity: 1,
          isInsuranceApplicable: true,
          taxRate: 0.1,
        },
      ],
      0,
      0,
      0.1,
      1.0,
    );
    expect(result.tax).toBe(1000);
    expect(result.total).toBe(11000);
    expect(result.insuranceAmount).toBe(10000);
    expect(result.billingAmount).toBe(1000);
  });
});
