import { useMutation, useQueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { transformExamination, type ExaminationRecord } from "./transforms";
import type {
  BackendExamination,
  UnconfirmExaminationRequest,
} from "./types";

async function unconfirmExamination(
  id: string,
  request: UnconfirmExaminationRequest,
): Promise<ExaminationRecord> {
  const { data } = await axios.post<BackendExamination>(
    `/v1/examinations/${id}/unconfirm`,
    request,
  );
  return transformExamination(data);
}

export function useUnconfirmExamination() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      unconfirmExamination(id, { reason }),
    onSuccess: async (_data, { id }) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: queryKeys.examinations.all(),
        }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.examinations.detail(id),
        }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.examinations.items(id),
        }),
      ]);
    },
    onError: (error) => handleApiError(error, "検査確定解除"),
  });
}
