import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformHospitalization } from "./transforms";
import type { Hospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

interface HospitalizationPaginatedResponse {
  data: BackendHospitalization[];
  total: number;
  page: number;
  limit: number;
}

export interface HospitalizationFilters {
  startDate?: string; // YYYY-MM-DD（入院開始日の範囲）
  endDate?: string;   // YYYY-MM-DD
}

const getHospitalizations = async (filters?: HospitalizationFilters): Promise<Hospitalization[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  const { data } = await axios.get<HospitalizationPaginatedResponse>(
    "/v1/hospitalizations",
    { params },
  );
  return data.data.map(transformHospitalization);
};

export const useGetHospitalizations = (filters?: HospitalizationFilters) => {
  return useQuery({
    queryKey: queryKeys.hospitalizations.list(filters),
    queryFn: () => getHospitalizations(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
