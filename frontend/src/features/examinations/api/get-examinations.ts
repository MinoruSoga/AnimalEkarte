import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ExaminationRecord } from "@/types";
import { transformExamination } from "./transforms";
import type { BackendExamination } from "./types";

interface ExaminationsListResponse {
  data: BackendExamination[];
  total: number;
  page: number;
  limit: number;
}

export const getExaminations = async (): Promise<ExaminationRecord[]> => {
  const { data } = await axios.get<ExaminationsListResponse>("/v1/examinations");
  return data.data.map(transformExamination);
};

export const useGetExaminations = () => {
  return useQuery({
    queryKey: ["examinations"],
    queryFn: getExaminations,
  });
};
