import type { PaymentSplitDraft } from "../components/PaymentCard";
import type { Accounting } from "../types";

export interface AccountingFormState {
  success: boolean;
  timestamp: number;
}

export function createInitialPaymentSplits(baseAccounting: Accounting | null): PaymentSplitDraft[] {
  const splits = baseAccounting?.paymentSplits;
  if (splits && splits.length > 0) {
    return splits.map((split) => ({
      method: split.method,
      amount: split.amount.toString(),
      receivedAmount: split.receivedAmount > 0 ? split.receivedAmount.toString() : "",
    }));
  }

  const payment = baseAccounting?.payment;
  if (payment) {
    return [{
      method: payment.method,
      amount: payment.billingAmount.toString(),
      receivedAmount: payment.receivedAmount > 0 ? payment.receivedAmount.toString() : "",
    }];
  }

  return [{ method: "cash", amount: "", receivedAmount: "" }];
}
