import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformToAccounting } from "./transforms";
import type { Accounting } from "./transforms";
import type { BackendAccounting } from "./types";

// Accounting 詳細型（明細含む）を取得する hook
const getAccountingDetail = async (id: string): Promise<Accounting> => {
  const { data } = await axios.get<BackendAccounting>(`/v1/accountings/${id}`);
  return transformToAccounting(data);
};

export const useGetAccountingDetail = (id: string | undefined) => {
  return useQuery({
    queryKey: queryKeys.accountings.detail(id ?? ""),
    queryFn: () => getAccountingDetail(id!),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
