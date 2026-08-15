import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { transformExamResult, type ExamResult } from "./transforms";
import type { ExamItemsResponse, ReplaceExamItemsRequest } from "./types";

/**
 * PUT /api/v1/examinations/:id/items — 検査項目を一括置換する。
 * 既存項目を全削除して受け取った items を一括登録するセマンティクス。
 * status / is_abnormal と判定基準値は backend がマスタから導出するため送信不要。
 */
const updateExaminationItems = async (
  id: string,
  req: ReplaceExamItemsRequest,
): Promise<ExamResult[]> => {
  const { data } = await axios.put<ExamItemsResponse>(`/v1/examinations/${id}/items`, req);
  return (data.items ?? []).map(transformExamResult);
};

export const useUpdateExaminationItems = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: ReplaceExamItemsRequest }) =>
      updateExaminationItems(id, req),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.examinations.items(id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.examinations.detail(id) });
      // Nested under examinations.all() prefix is not enough here — items path does not call all().
      queryClient.invalidateQueries({
        queryKey: queryKeys.examinations.printSnapshot(id),
      });
    },
    onError: (error) => handleApiError(error, "検査項目の保存"),
  });
};
