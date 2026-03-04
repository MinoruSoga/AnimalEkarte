import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ExaminationRecord } from "@/types";
import { transformExamination } from "./transforms";
import type { BackendExamination, CreateExaminationRequest } from "./types";

export const createExamination = async (
  req: CreateExaminationRequest
): Promise<ExaminationRecord> => {
  const { data } = await axios.post<BackendExamination>(
    "/v1/examinations",
    req
  );
  return transformExamination(data);
};

export const useCreateExamination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createExamination,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["examinations"] });
    },
  });
};
