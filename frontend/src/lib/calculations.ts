/**
 * Veterinary Billing Calculation Utilities
 * 
 * Centralizes rounding, tax calculation, and discount logic to ensure consistency
 * across Medical Records, Hospitalization, and Accounting.
 */

interface BillingItem {
  unitPrice: number;
  quantity: number;
  discountAmount?: number;
  isInsuranceApplicable?: boolean;
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

/**
 * Calculates standardized billing totals.
 * 
 * Flow:
 * 1. Sum (unitPrice * quantity - itemDiscount)
 * 2. Apply owner discount (percentage)
 * 3. Apply global discount (absolute amount)
 * 4. Calculate consumption tax (default 10%)
 * 5. Calculate insurance coverage (if applicable)
 */
export function calculateBillingTotals(
  items: BillingItem[],
  ownerDiscountRate: number = 0,
  globalDiscountAmount: number = 0,
  taxRate: number = 0.10,
  insuranceRatio: number = 0
): BillingTotals {
  // 1. Raw Subtotal (sum of all lines)
  const rawSubtotal = items.reduce((sum, item) => {
    const price = Number(item.unitPrice) || 0;
    const qty = Number(item.quantity) || 0;
    const itemDiscount = Number(item.discountAmount) || 0;
    return sum + (price * qty - itemDiscount);
  }, 0);

  // 2. Owner Discount (Percentage based)
  const ownerDiscountAmount = Math.floor(rawSubtotal * (ownerDiscountRate / 100));
  const afterOwnerDiscount = rawSubtotal - ownerDiscountAmount;

  // 3. Global Discount (Absolute amount)
  const actualGlobalDiscount = Math.min(afterOwnerDiscount, globalDiscountAmount);
  const taxableAmount = Math.max(0, afterOwnerDiscount - actualGlobalDiscount);

  // 4. Consumption Tax (Standard 10%)
  const tax = Math.floor(taxableAmount * taxRate);
  const total = taxableAmount + tax;

  // 5. Insurance Coverage (Applied to applicable items raw price)
  // Note: Usually insurance applies to the subtotal of specific items.
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
    billingAmount
  };
}
