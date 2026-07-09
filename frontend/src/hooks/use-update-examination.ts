import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { transformExamination, type ExaminationRecord } from "@/lib/transforms/examination";
import type { Examination } from "@/types/generated/models";

// PATCH /v1/examinations/:id の BE リクエストボディ。
// 正本はここ（shared hook）。features/examinations/api/types.ts は
// `export type { UpdateExaminationRequest } from "@/hooks/use-update-examination"`
// で re-export するのみ — 契約は型レベルで単一ソース化されている。
export interface UpdateExaminationRequest {
  medical_record_id?: number | null;
  status?: "pending" | "in_progress" | "result_entered" | "completed" | "confirmed";
  result_summary?: string;
  machine?: string;
  date?: string;
}

export const updateExamination = async (
  id: string,
  req: UpdateExaminationRequest,
): Promise<ExaminationRecord> => {
  const { data } = await axios.patch<Examination>(`/v1/examinations/${id}`, req);
  return transformExamination(data);
};

/**
 * Shared hook for updating an examination (e.g. medical_record_id link on import).
 * R-F2-S8: examinations feature から昇格。invalidate prefix は移設前と同一 ["examinations"] を維持。
 */
export const useUpdateExamination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateExaminationRequest }) =>
      updateExamination(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["examinations"] });
    },
    onError: (error) => handleApiError(error, "検査更新"),
  });
};
