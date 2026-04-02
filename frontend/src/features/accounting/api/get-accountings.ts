import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Accounting } from "../types";
import { transformToAccounting } from "./transforms";
import type { BackendAccounting } from "./types";

interface AccountingsListResponse {
  data: BackendAccounting[];
  total: number;
  page: number;
  limit: number;
}

export interface AccountingFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string;   // YYYY-MM-DD
}

export const getAccountings = async (
  filters?: AccountingFilters,
): Promise<Accounting[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  const { data } = await axios.get<AccountingsListResponse>("/v1/accountings", { params });
  return data.data.map(transformToAccounting);
};

export const useGetAccountings = (filters?: AccountingFilters) => {
  return useQuery({
    queryKey: ["accountings", filters],
    queryFn: () => getAccountings(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
