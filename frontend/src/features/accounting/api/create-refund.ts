import { axios } from "@/lib/axios";
import type { BillingRefund } from "@/types/generated/models";
import { transformToRefund } from "./transforms";
import type { Refund } from "./transforms";

export interface CreateRefundRequest {
  amount: number;
  reason?: string;
}

export const createRefund = async (
  billingId: string,
  data: CreateRefundRequest,
): Promise<Refund> => {
  const { data: res } = await axios.post<BillingRefund>(
    `/v1/accountings/${billingId}/refunds`,
    data,
  );
  return transformToRefund(res);
};
