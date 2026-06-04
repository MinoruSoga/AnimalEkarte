// AccountingListTable の非コンポーネントロジック。コンポーネントファイルから分離して
// react-refresh/only-export-components 違反を解消する。
import type { Accounting as AccountingType } from "../types";

export function calculateAccountingTotal(accounting: AccountingType) {
  if (accounting.payment) return accounting.payment.totalAmount;

  return accounting.items.reduce((sum: number, item) => {
    const price = item.unitPrice * item.quantity;
    const tax = Math.floor(price * item.taxRate);
    return sum + price + tax;
  }, 0);
}
