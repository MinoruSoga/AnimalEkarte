import { axios } from "@/lib/axios";
import type { BillingRefund } from "@/types/generated/models";
import { transformToRefund } from "./transforms";
import type { Refund } from "./transforms";

export interface CreateRefundRequest {
  amount: number;
  reason?: string;
  /** 返金先の支払方法ID（payment_methods マスタ）。未指定可（#60）。 */
  paymentMethodId?: number;
}

export const createRefund = async (
  billingId: string,
  data: CreateRefundRequest,
): Promise<Refund> => {
  const { data: res } = await axios.post<BillingRefund>(
    `/v1/accountings/${billingId}/refunds`,
    {
      amount: data.amount,
      reason: data.reason,
      payment_method_id: data.paymentMethodId,
    },
  );
  return transformToRefund(res);
};
