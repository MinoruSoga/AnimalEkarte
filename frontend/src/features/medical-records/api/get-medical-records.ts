import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { MedicalRecord } from "@/types";
import { transformMedicalRecord } from "./transforms";
import type { BackendMedicalRecord } from "./types";

interface MedicalRecordsListResponse {
  data: BackendMedicalRecord[];
  total: number;
  page: number;
  limit: number;
}

export interface MedicalRecordFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string;   // YYYY-MM-DD
}

export const getMedicalRecords = async (
  filters?: MedicalRecordFilters,
): Promise<MedicalRecord[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  const { data } = await axios.get<MedicalRecordsListResponse>("/v1/medical-records", { params });
  return data.data.map(transformMedicalRecord);
};

export const useGetMedicalRecords = (filters?: MedicalRecordFilters) => {
  return useQuery({
    queryKey: ["medical-records", filters],
    queryFn: () => getMedicalRecords(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
