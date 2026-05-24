import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformExamination, type ExaminationRecord } from "./transforms";
import type { BackendExamination } from "./types";

export const getExamination = async (id: string): Promise<ExaminationRecord> => {
  const { data } = await axios.get<BackendExamination>(`/v1/examinations/${id}`);
  return transformExamination(data);
};

export const useGetExamination = (id: string) => {
  return useQuery({
    queryKey: ["examination", id],
    queryFn: () => getExamination(id),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

