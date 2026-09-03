import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { Examination } from "@/types/generated/models";
import { transformExamination, type ExaminationRecord } from "./transforms";
import type { UpdateExaminationRequest } from "../types";

const updateExamination = async (
  id: string,
  req: UpdateExaminationRequest,
): Promise<ExaminationRecord> => {
  const { data } = await axios.patch<Examination>(`/v1/examinations/${id}`, req);
  return transformExamination(data);
};

/**
 * Feature-local update supports the atomic parent + items PATCH contract.
 * The shared hook remains parent-only for consumers outside examinations.
 */
export const useUpdateExamination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateExaminationRequest }) =>
      updateExamination(id, req),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.examinations.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.examinations.detail(id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.examinations.items(id) });
    },
    onError: (error) => handleApiError(error, "検査更新"),
  });
};
