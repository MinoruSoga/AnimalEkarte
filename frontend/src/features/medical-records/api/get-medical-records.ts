import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { MedicalRecord } from "@/types";
import { transformMedicalRecord, transformToHistoryItem } from "./transforms";
import type { BackendMedicalRecord } from "./types";
import type { InterviewHistoryItem } from "../types";

interface MedicalRecordsListResponse {
  data: BackendMedicalRecord[];
  total: number;
  page: number;
  limit: number;
}

export interface MedicalRecordFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string;   // YYYY-MM-DD
  petId?: string;
}

export const getMedicalRecords = async (
  filters?: MedicalRecordFilters,
): Promise<MedicalRecord[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.petId) params.pet_id = filters.petId;
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

/** FEAT-003: ペットの問診履歴を取得して InterviewHistoryItem[] に変換する */
export const useGetPetMedicalHistory = (
  petId: string | undefined,
  excludeRecordId?: string,
): { historyItems: InterviewHistoryItem[]; isLoading: boolean } => {
  const { data, isLoading } = useQuery({
    queryKey: ["medical-records", "history", petId],
    queryFn: async () => {
      const params: Record<string, string> = { limit: "50", page: "1" };
      if (petId) params.pet_id = petId;
      const { data: res } = await axios.get<MedicalRecordsListResponse>("/v1/medical-records", { params });
      return res.data;
    },
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });

  const historyItems: InterviewHistoryItem[] = (data ?? [])
    .filter((r) => !excludeRecordId || String(r.id ?? 0) !== excludeRecordId)
    .map(transformToHistoryItem);

  return { historyItems, isLoading };
};
