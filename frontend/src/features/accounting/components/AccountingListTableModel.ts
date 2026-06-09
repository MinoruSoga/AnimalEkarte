// AccountingListTable の非コンポーネントロジック。コンポーネントファイルから分離して
// react-refresh/only-export-components 違反を解消する。
import type { Accounting as AccountingType } from "../types";

export function calculateAccountingTotal(accounting: AccountingType) {
  if (accounting.payment) return accounting.payment.totalAmount;

  return accounting.items.reduce((sum: number, item) => {
    // #85: 課税ベースは割引後（単価×数量 − 割引額）。負値は 0 クランプ。
    const base = Math.max(item.unitPrice * item.quantity - item.discountAmount, 0);
    const tax = Math.floor(base * item.taxRate);
    return sum + base + tax;
  }, 0);
}
