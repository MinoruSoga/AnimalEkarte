import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { ChiefComplaintType as ModelChiefComplaintType } from "@/types/generated/models";

function transformChiefComplaintType(item: ModelChiefComplaintType) {
  return {
    id: Number(item.id ?? 0),
    name: item.name,
  };
}
type ChiefComplaintType = ReturnType<typeof transformChiefComplaintType>;

const getChiefComplaintTypes = async (): Promise<ChiefComplaintType[]> => {
  const { data } = await axios.get<ModelChiefComplaintType[]>("/v1/masters/chief-complaint-types");
  return data.map(transformChiefComplaintType);
};

export const useGetChiefComplaintTypes = () =>
  useQuery({
    queryKey: queryKeys.masters.category("chief-complaint-types"),
    queryFn: getChiefComplaintTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
