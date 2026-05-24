import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformExamResult, type ExamResult } from "./transforms";
import type { ExamItemsResponse } from "./types";

export const getExaminationItems = async (id: string): Promise<ExamResult[]> => {
  const { data } = await axios.get<ExamItemsResponse>(`/v1/examinations/${id}/items`);
  return (data.items ?? []).map(transformExamResult);
};

export const useGetExaminationItems = (id: string) => {
  return useQuery({
    queryKey: ["examination-items", id],
    queryFn: () => getExaminationItems(id),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
