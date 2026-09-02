import type { Accounting } from "../api/transforms";

export type UnpaidGroupBy = "owner" | "billing" | "monthly";

export function parseUnpaidGroupBy(raw: string | null): UnpaidGroupBy {
  return raw === "billing" ? "billing" : raw === "monthly" ? "monthly" : "owner";
}

export function unpaidBillingAmount(billing: Accounting): number {
  // BUG-007: outstanding_amount を優先（クレジット訂正差額）。未設定時は明細合計へフォールバック。
  return (billing.outstandingAmount ?? 0) > 0
    ? billing.outstandingAmount
    : billing.items.reduce((sum, item) => sum + item.subtotal + item.taxAmount, 0);
}
