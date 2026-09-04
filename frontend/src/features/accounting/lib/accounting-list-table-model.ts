// AccountingListTable の非コンポーネントロジック。コンポーネントファイルから分離して
// react-refresh/only-export-components 違反を解消する。
import type { Accounting as AccountingType } from "../types";
import { recordedLineNet } from "./tax-breakdown";

export function calculateAccountingTotal(accounting: AccountingType) {
  if (accounting.payment) return accounting.payment.totalAmount;

  return accounting.items.reduce((sum: number, item) => {
    const base = recordedLineNet(item);
    const tax = Math.floor(base * item.taxRate);
    return sum + base + tax;
  }, 0);
}
