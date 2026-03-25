import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { BillingRefund } from "@/types/generated/models";

export const getRefunds = async (billingId: string): Promise<BillingRefund[]> => {
  const { data } = await axios.get<BillingRefund[]>(
    `/v1/accountings/${billingId}/refunds`,
  );
  return data;
};

export const useGetRefunds = (billingId: string | undefined) => {
  return useQuery({
    queryKey: ["accounting-refunds", billingId],
    queryFn: () => getRefunds(billingId!),
    enabled: !!billingId,
  });
};
