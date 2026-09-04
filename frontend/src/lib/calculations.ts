/**
 * Veterinary Billing Calculation Utilities
 *
 * Centralizes rounding, tax calculation, and discount logic to ensure consistency
 * across Medical Records, Hospitalization, and Accounting.
 *
 * Tax per line aligns with backend billing.CalculateBillingTotals /
 * BillingItem.CalculateTaxAmount (外税・内税は Math.round、割引後ベース)。
 */

type BillingTaxType = "excluded" | "included" | "exempt";

interface BillingItem {
  unitPrice: number;
  quantity: number;
  discountAmount?: number;
  isInsuranceApplicable?: boolean;
  /** 明細税率（未指定時は calculateBillingTotals の taxRate 引数） */
  taxRate?: number;
  /** 課税区分。未指定は外税 */
  taxType?: BillingTaxType;
}

interface BillingTotals {
  subtotal: number;
  ownerDiscountAmount: number;
  globalDiscountAmount: number;
  taxableAmount: number;
  tax: number;
  insuranceAmount: number;
  total: number;
  billingAmount: number; // Amount to be paid by owner (total + insuranceAmount)
}

function lineBase(item: BillingItem): number {
  const price = Number(item.unitPrice) || 0;
  const qty = Number(item.quantity) || 0;
  const itemDiscount = Number(item.discountAmount) || 0;
  // BE: max(round(unitPrice*qty) - discount, 0)
  return Math.max(Math.round(price * qty) - itemDiscount, 0);
}

function lineTaxAmount(base: number, rate: number, taxType: BillingTaxType): number {
  switch (taxType) {
    case "excluded":
      return Math.round(base * rate);
    case "included":
      return Math.round((base * rate) / (1 + rate));
    case "exempt":
    default:
      return 0;
  }
}

/**
 * Calculates standardized billing totals.
 *
 * Flow:
 * 1. Sum (unitPrice * quantity - itemDiscount) per line
 * 2. Apply owner discount (percentage) and global discount (absolute)
 *    — when either is non-zero, taxable bases are scaled proportionally
 *      so multi-rate lines still tax correctly
 * 3. Sum per-line consumption tax (item.taxRate or default taxRate)
 * 4. total = taxableAmount + excluded tax only (included tax stays in base)
 * 5. Insurance coverage (if applicable)
 */
export function calculateBillingTotals(
  items: BillingItem[],
  ownerDiscountRate: number = 0,
  globalDiscountAmount: number = 0,
  taxRate: number | undefined = 0.1,
  insuranceRatio: number = 0,
): BillingTotals {
  const defaultTaxRate = taxRate ?? 0.1;
  const bases = items.map(lineBase);
  const rawSubtotal = bases.reduce((sum, b) => sum + b, 0);

  // 2. Owner Discount (Percentage based)
  const ownerDiscountAmount = Math.floor(rawSubtotal * (ownerDiscountRate / 100));
  const afterOwnerDiscount = rawSubtotal - ownerDiscountAmount;

  // 3. Global Discount (Absolute amount)
  const actualGlobalDiscount = Math.min(afterOwnerDiscount, globalDiscountAmount);
  const taxableAmount = Math.max(0, afterOwnerDiscount - actualGlobalDiscount);

  // Discount scale across lines (1 when no document-level discount)
  const scale = rawSubtotal > 0 ? taxableAmount / rawSubtotal : 0;

  let tax = 0;
  let excludedTax = 0;
  for (let i = 0; i < items.length; i++) {
    const item = items[i]!;
    const rate = typeof item.taxRate === "number" ? item.taxRate : defaultTaxRate;
    const taxType: BillingTaxType = item.taxType ?? "excluded";
    // Keep integer yen base after proportional discount
    const scaledBase = scale === 1 ? bases[i]! : Math.max(0, Math.round(bases[i]! * scale));
    const itemTax = lineTaxAmount(scaledBase, rate, taxType);
    tax += itemTax;
    if (taxType === "excluded") {
      excludedTax += itemTax;
    }
  }

  // 4. total: 外税のみ加算（内税は subtotal に内包 — BE CalculateBillingTotals）
  const total = taxableAmount + excludedTax;

  // 5. Insurance Coverage (Applied to applicable items raw price)
  const insuranceTargetTotal = items
    .filter((item) => item.isInsuranceApplicable)
    .reduce((sum, item) => sum + (Number(item.unitPrice) || 0) * (Number(item.quantity) || 0), 0);

  const insuranceAmount = Math.floor(insuranceTargetTotal * insuranceRatio) * -1;
  const billingAmount = Math.max(0, total + insuranceAmount);

  return {
    subtotal: rawSubtotal,
    ownerDiscountAmount,
    globalDiscountAmount: actualGlobalDiscount,
    taxableAmount,
    tax,
    insuranceAmount,
    total,
    billingAmount,
  };
}
