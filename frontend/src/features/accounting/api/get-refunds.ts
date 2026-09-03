import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { BillingRefund } from "@/types/generated/models";
import { transformToRefund } from "./transforms";
import type { Refund } from "./transforms";

const getRefunds = async (billingId: string): Promise<Refund[]> => {
  const { data } = await axios.get<BillingRefund[]>(`/v1/accountings/${billingId}/refunds`);
  return data.map(transformToRefund);
};

export const useGetRefunds = (billingId: string | undefined) => {
  return useQuery({
    queryKey: queryKeys.accountingRefunds(billingId!),
    queryFn: () => getRefunds(billingId!),
    enabled: !!billingId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
