import type { AccountingItem } from "./types";

// BUG-367: 軽減税率判定の浮動小数誤差吸収
const TAX_RATE_EPSILON = 0.0001;
function approxEqual(a: number, b: number): boolean {
  return Math.abs(a - b) < TAX_RATE_EPSILON;
}

/** {@link calcTaxBreakdown} が参照する明細の最小形。 */
type TaxBreakdownItem = Pick<AccountingItem, "unitPrice" | "quantity" | "discountAmount" | "taxRate">;

/** 記録された明細ネット。移行の負額は 0 にクランプしない。 */
export function recordedLineNet(item: Pick<AccountingItem, "unitPrice" | "quantity" | "discountAmount">): number {
  return item.unitPrice * item.quantity - item.discountAmount;
}

export interface TaxBreakdown {
  standardBase: number;
  reducedBase: number;
  standardAmount: number;
  reducedAmount: number;
  standardRatePercent: number;
  reducedRatePercent: number;
}

/**
 * BUG-367: 明細兼領収書（AccountingDocument）の標準/軽減税率内訳を計算する。
 *
 * 領収書・明細書に印字される法定金額のため、丸め規則（消費税額は Math.floor で
 * 切り捨て、表示用パーセントは Math.round）を変更してはならない。
 *
 * FE6-14: AccountingDocument.tsx の useMemo から抽出。消費税は Math.floor。
 * 移行負額は line net を 0 クランプしない（AE-MIG-NEG-PRINT-1）。
 */
export function calcTaxBreakdown(
  items: readonly TaxBreakdownItem[],
  standardRate: number,
  reducedRate: number,
): TaxBreakdown {
  const stdItems = items.filter(i => approxEqual(i.taxRate, standardRate));
  const redItems = items.filter(i => approxEqual(i.taxRate, reducedRate));

  const stdBase = stdItems.reduce((sum, i) => sum + recordedLineNet(i), 0);
  const redBase = redItems.reduce((sum, i) => sum + recordedLineNet(i), 0);

  return {
    standardBase: stdBase,
    reducedBase: redBase,
    standardAmount: Math.floor(stdBase * standardRate),
    reducedAmount: Math.floor(redBase * reducedRate),
    standardRatePercent: Math.round(standardRate * 100),
    reducedRatePercent: Math.round(reducedRate * 100),
  };
}
