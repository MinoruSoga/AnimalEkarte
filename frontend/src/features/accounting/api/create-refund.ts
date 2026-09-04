import { axios } from "@/lib/axios";
import type { BillingRefund, PaymentMethod } from "@/types/generated/models";
import { transformToRefund } from "./transforms";
import type { Refund } from "./transforms";

export interface CreateRefundRequest {
  amount: number;
  reason?: string;
  /** 返金先の支払方法（ENUM）。会計 payment_splits.method と同体系。未指定可（#60）。 */
  paymentMethod?: PaymentMethod;
}

export const createRefund = async (
  billingId: string,
  data: CreateRefundRequest,
): Promise<Refund> => {
  const { data: res } = await axios.post<BillingRefund>(`/v1/accountings/${billingId}/refunds`, {
    amount: data.amount,
    reason: data.reason,
    payment_method: data.paymentMethod,
  });
  return transformToRefund(res);
};
