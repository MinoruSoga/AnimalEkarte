import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Estimate } from "../types";
import { transformEstimate } from "./transforms";
import type { EstimateListResponse } from "./types";

interface GetEstimatesParams {
  page?: number;
  limit?: number;
  status?: string;
  owner_id?: number;
  medical_record_id?: number;
}

interface EstimatesResult {
  data: Estimate[];
  total: number;
  page: number;
  limit: number;
}

async function getEstimates(params?: GetEstimatesParams): Promise<EstimatesResult> {
  const { data } = await axios.get<EstimateListResponse>("/v1/estimates", { params });
  return {
    data: data.data.map(transformEstimate),
    total: data.total,
    page: data.page,
    limit: data.limit,
  };
}

export function useGetEstimates(params?: GetEstimatesParams) {
  return useQuery({
    queryKey: queryKeys.estimates.list(params),
    queryFn: () => getEstimates(params),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
