import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ExaminationRecord } from "@/types";
import { transformExamination } from "./transforms";
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
  });
};

