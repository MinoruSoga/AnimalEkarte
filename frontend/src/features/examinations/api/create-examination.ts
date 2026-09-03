import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { transformExamination, type ExaminationRecord } from "./transforms";
import type { BackendExamination, CreateExaminationRequest } from "../types";

const createExamination = async (req: CreateExaminationRequest): Promise<ExaminationRecord> => {
  const { data } = await axios.post<BackendExamination>("/v1/examinations", req);
  return transformExamination(data);
};

export const useCreateExamination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createExamination,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.examinations.all() });
    },
    onError: (error) => handleApiError(error, "検査作成"),
  });
};
