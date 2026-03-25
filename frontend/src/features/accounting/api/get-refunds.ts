import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { BillingRefund } from "@/types/generated/models";
import type { Refund } from "../types";
import { transformToRefund } from "./transforms";

export const getRefunds = async (billingId: string): Promise<Refund[]> => {
  const { data } = await axios.get<BillingRefund[]>(
    `/v1/accountings/${billingId}/refunds`,
  );
  return data.map(transformToRefund);
};

export const useGetRefunds = (billingId: string | undefined) => {
  return useQuery({
    queryKey: ["accounting-refunds", billingId],
    queryFn: () => getRefunds(billingId!),
    enabled: !!billingId,
  });
};
