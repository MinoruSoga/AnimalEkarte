import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { transformExamination, type ExaminationRecord } from "@/lib/transforms/examination";
import type { Examination } from "@/types/generated/models";

// PATCH /v1/examinations/:id の BE リクエストボディ。
// 正本はここ（shared hook）。features/examinations/types がこの型を拡張する。
export interface UpdateExaminationRequest {
  medical_record_id?: number | null;
  pet_id?: number | null;
  doctor_id?: number | null;
  status?: "pending" | "in_progress" | "result_entered" | "completed" | "confirmed";
  result_summary?: string;
  machine?: string;
  date?: string;
}

const updateExamination = async (
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
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.examinations.all() });
      // FE4-6 fix: detail クエリの実キーは ["examination", id]（単数形。features/examinations/api/get-examination.ts）。
      // list prefix invalidation はこれを包含しないため、更新後も詳細画面が stale のまま残っていた
      // （先例: update-examination-items.ts:26-27 は既に両方を invalidate している）。
      queryClient.invalidateQueries({
        queryKey: queryKeys.examinations.detail(id),
      });
    },
    onError: (error) => handleApiError(error, "検査更新"),
  });
};
