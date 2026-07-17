import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { Insurance as BackendInsurance } from "@/types/generated/models";

const getInsurances = async (): Promise<BackendInsurance[]> => {
  const { data } = await axios.get<BackendInsurance[]>("/v1/masters/insurances");
  return data;
};

export const useGetInsurances = () => {
  return useQuery({
    queryKey: queryKeys.masters.category("insurances"),
    queryFn: getInsurances,
    staleTime: QUERY_STALE_TIMES.STATIC, // 静的マスタデータは30分キャッシュ
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
