import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ExaminationRecord } from "@/types";
import { transformExamination } from "./transforms";
import type { BackendExamination } from "./types";

interface ExaminationsListResponse {
  data: BackendExamination[];
  total: number;
  page: number;
  limit: number;
}

export interface ExaminationFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string; // YYYY-MM-DD
  petId?: string;
}

export const getExaminations = async (
  filters?: ExaminationFilters,
): Promise<ExaminationRecord[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.petId) params.pet_id = filters.petId;
  const { data } = await axios.get<ExaminationsListResponse>("/v1/examinations", { params });
  return data.data.map(transformExamination);
};

export const useGetExaminations = (filters?: ExaminationFilters) => {
  return useQuery({
    queryKey: ["examinations", filters],
    queryFn: () => getExaminations(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
