import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Estimate } from "../types";
import { transformEstimate } from "./transforms";
import type { BackendEstimate } from "./types";

async function getEstimate(id: string): Promise<Estimate> {
  const { data } = await axios.get<BackendEstimate>(`/v1/estimates/${id}`);
  return transformEstimate(data);
}

export function useGetEstimate(id: string | undefined) {
  return useQuery({
    queryKey: queryKeys.estimates.detail(id!),
    queryFn: () => getEstimate(id!),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
