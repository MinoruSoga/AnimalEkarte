import { describe, it, expect } from "vitest";
import { calcTaxBreakdown } from "./tax-breakdown";

const STANDARD_RATE = 0.1;
const REDUCED_RATE = 0.08;

describe("calcTaxBreakdown (FE6-14 特性テスト)", () => {
  it("標準税率のみの明細で内訳を計算する", () => {
    // 手計算: 1000*1-0=1000 / 500*2-100=900 → stdBase=1900
    // standardAmount = floor(1900*0.1) = floor(190) = 190
    const items = [
      { unitPrice: 1000, quantity: 1, discountAmount: 0, taxRate: STANDARD_RATE },
      { unitPrice: 500, quantity: 2, discountAmount: 100, taxRate: STANDARD_RATE },
    ];

    const result = calcTaxBreakdown(items, STANDARD_RATE, REDUCED_RATE);

    expect(result).toEqual({
      standardBase: 1900,
      reducedBase: 0,
      standardAmount: 190,
      reducedAmount: 0,
      standardRatePercent: 10,
      reducedRatePercent: 8,
    });
  });

  it("軽減税率のみ", () => {
    // 手計算: 800*3-200=2200 → redBase=2200
    // reducedAmount = floor(2200*0.08) = floor(176) = 176
    const items = [{ unitPrice: 800, quantity: 3, discountAmount: 200, taxRate: REDUCED_RATE }];

    const result = calcTaxBreakdown(items, STANDARD_RATE, REDUCED_RATE);

    expect(result).toEqual({
      standardBase: 0,
      reducedBase: 2200,
      standardAmount: 0,
      reducedAmount: 176,
      standardRatePercent: 10,
      reducedRatePercent: 8,
    });
  });

  it("標準・軽減混在", () => {
    // 手計算: 標準側 stdBase=1900 → floor(1900*0.1)=190
    //         軽減側 redBase=2200 → floor(2200*0.08)=176
    const items = [
      { unitPrice: 1000, quantity: 1, discountAmount: 0, taxRate: STANDARD_RATE },
      { unitPrice: 500, quantity: 2, discountAmount: 100, taxRate: STANDARD_RATE },
      { unitPrice: 800, quantity: 3, discountAmount: 200, taxRate: REDUCED_RATE },
    ];

    const result = calcTaxBreakdown(items, STANDARD_RATE, REDUCED_RATE);

    expect(result).toEqual({
      standardBase: 1900,
      reducedBase: 2200,
      standardAmount: 190,
      reducedAmount: 176,
      standardRatePercent: 10,
      reducedRatePercent: 8,
    });
  });

  it("端数はMath.floorで切り捨てる", () => {
    // 手計算: base = 999*1-0 = 999
    // 999*0.1 = 99.9 → Math.floor => 99（Math.round なら 100 になり区別できる）
    const items = [{ unitPrice: 999, quantity: 1, discountAmount: 0, taxRate: STANDARD_RATE }];

    const result = calcTaxBreakdown(items, STANDARD_RATE, REDUCED_RATE);

    expect(result.standardAmount).toBe(99);
    expect(result.standardAmount).not.toBe(Math.round(999 * STANDARD_RATE));
  });

  it("負の単価はクランプせず課税ベースに符号を残す", () => {
    const items = [{ unitPrice: -3000, quantity: 1, discountAmount: 0, taxRate: STANDARD_RATE }];
    const result = calcTaxBreakdown(items, STANDARD_RATE, REDUCED_RATE);
    expect(result.standardBase).toBe(-3000);
    expect(result.standardAmount).toBe(Math.floor(-3000 * STANDARD_RATE));
  });

  it("明細0件で全て0を返す", () => {
    // 手計算: base/amount は全て0。ただし *RatePercent は明細件数に依存せず
    // rate 引数から常に算出される（round(0.1*100)=10 / round(0.08*100)=8）。
    const result = calcTaxBreakdown([], STANDARD_RATE, REDUCED_RATE);

    expect(result).toEqual({
      standardBase: 0,
      reducedBase: 0,
      standardAmount: 0,
      reducedAmount: 0,
      standardRatePercent: 10,
      reducedRatePercent: 8,
    });
  });
});
