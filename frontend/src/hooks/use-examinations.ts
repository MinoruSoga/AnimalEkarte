import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
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

function buildExaminationParams(filters?: ExaminationFilters): Record<string, string> {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.petId) params.pet_id = filters.petId;
  return params;
}

const getExaminations = async (
  filters?: ExaminationFilters,
): Promise<ExaminationRecord[]> => {
  const { data } = await axios.get<ExaminationsListResponse>("/v1/examinations", { params: buildExaminationParams(filters) });
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
    queryKey: queryKeys.examinations.list(filters),
    queryFn: () => getExaminations(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

export interface ExaminationsPage {
  data: ExaminationRecord[];
  total: number;
  page: number;
  limit: number;
}

export interface ExaminationPageFilters extends ExaminationFilters {
  page: number;
  limit: number;
}

const getExaminationsPage = async (
  filters: ExaminationPageFilters,
): Promise<ExaminationsPage> => {
  const params = buildExaminationParams(filters);
  params.page = String(filters.page);
  params.limit = String(filters.limit);
  const { data } = await axios.get<ExaminationsListResponse>("/v1/examinations", { params });
  return { data: data.data.map(transformExamination), total: data.total, page: data.page, limit: data.limit };
};

/**
 * BUG-411: サーバサイドページネーション版。ExaminationsList 専用。
 * useGetExaminations（medical-records/ExaminationForm が使う全件系）とは
 * queryKey の filters shape が異なる（page/limit を含む）ため自然に別キャッシュエントリになる。
 */
export const useGetExaminationsPage = (filters: ExaminationPageFilters) => {
  return useQuery({
    queryKey: queryKeys.examinations.list(filters),
    queryFn: () => getExaminationsPage(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
    placeholderData: keepPreviousData,
  });
};
