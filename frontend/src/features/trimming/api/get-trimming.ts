import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { TrimmingUI } from "@/types";
import { transformTrimming } from "./transforms";
import type { BackendTrimming } from "@/types/trimming";

const getTrimming = async (id: string): Promise<TrimmingUI> => {
  const { data } = await axios.get<BackendTrimming>(`/v1/trimmings/${id}`);
  return transformTrimming(data);
};

export const useGetTrimming = (id: string) => {
  return useQuery({
    queryKey: queryKeys.trimmings.detail(id),
    queryFn: () => getTrimming(id),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

// NO32: ペット単位の一覧 query（useGetTrimmingsByPetId）は cross-feature 共有のため
// @/hooks/use-pet-trimmings へ昇格。queryKey は trimmings.byPet のまま。
