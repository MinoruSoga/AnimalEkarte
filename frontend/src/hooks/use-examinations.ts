import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformExamination, type ExaminationRecord } from "@/lib/transforms/examination";
import type { Examination } from "@/types/generated/models";

interface ExaminationsListResponse {
  data: Examination[];
  total: number;
  page: number;
  limit: number;
}

export interface ExaminationFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string; // YYYY-MM-DD
  petId?: string;
}

const getExaminations = async (
  filters?: ExaminationFilters,
): Promise<ExaminationRecord[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.petId) params.pet_id = filters.petId;
  const { data } = await axios.get<ExaminationsListResponse>("/v1/examinations", { params });
  return data.data.map(transformExamination);
};

/**
 * Shared hook for fetching the examinations list.
 * R-F2-S8: examinations feature から昇格。medical-records (ExaminationImportDialog) が
 * examinations feature を直接 import しないための cross-feature 共有。
 * queryKey は examinations feature 側と同一 prefix を維持し React Query cache を共有する。
 */
export const useGetExaminations = (filters?: ExaminationFilters) => {
  return useQuery({
    queryKey: ["examinations", filters],
    queryFn: () => getExaminations(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
