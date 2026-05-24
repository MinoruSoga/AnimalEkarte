import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { transformExamination, type ExaminationRecord } from "./transforms";
import type { BackendExamination, UpdateExaminationRequest } from "./types";

export const updateExamination = async (
  id: string,
  req: UpdateExaminationRequest
): Promise<ExaminationRecord> => {
  const { data } = await axios.patch<BackendExamination>(
    `/v1/examinations/${id}`,
    req
  );
  return transformExamination(data);
};

export const useUpdateExamination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      req,
    }: {
      id: string;
      req: UpdateExaminationRequest;
    }) => updateExamination(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["examinations"] });
    },
    onError: (error) => handleApiError(error, "検査更新"),
  });
};
