export interface EstimateTotalsLine {
  readonly unitPrice: number;
  readonly quantity: number;
  readonly discountAmount?: number;
  readonly taxType?: string;
  readonly taxRate?: number;
}

export function calcEstimateLineAmount(line: EstimateTotalsLine): {
  subtotal: number;
  tax: number;
  total: number;
} {
  const discount = line.discountAmount ?? 0;
  const subtotal = Math.max(Math.round(line.unitPrice * line.quantity) - discount, 0);
  const rawTaxBase = Math.max(line.unitPrice * line.quantity - discount, 0);
  const rate = line.taxRate ?? 0;
  let tax = 0;
  if (line.taxType === "excluded") {
    tax = Math.round(rawTaxBase * rate);
  } else if (line.taxType === "included") {
    tax = Math.round((rawTaxBase * rate) / (1 + rate));
  }
  return { subtotal, tax, total: line.taxType === "excluded" ? subtotal + tax : subtotal };
}

export function calcEstimateTotals(lines: readonly EstimateTotalsLine[]): {
  subtotal: number;
  tax: number;
  total: number;
} {
  let subtotal = 0;
  let tax = 0;
  let total = 0;
  for (const line of lines) {
    const amount = calcEstimateLineAmount(line);
    subtotal += amount.subtotal;
    tax += amount.tax;
    total += amount.total;
  }
  return { subtotal, tax, total };
}
