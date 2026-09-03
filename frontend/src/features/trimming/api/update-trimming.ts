import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { TrimmingUI } from "@/types";
import { transformTrimming } from "./transforms";
import type { BackendTrimming, UpdateTrimmingRequest } from "@/types/trimming";

const updateTrimming = async (id: string, req: UpdateTrimmingRequest): Promise<TrimmingUI> => {
  const { data } = await axios.patch<BackendTrimming>(`/v1/trimmings/${id}`, req);
  return transformTrimming(data);
};

export const useUpdateTrimming = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateTrimmingRequest }) =>
      updateTrimming(id, req),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.trimmings.all() });
      // 単一取得キャッシュ ["trimming", id] も無効化して詳細画面の古いデータを防ぐ
      queryClient.invalidateQueries({ queryKey: queryKeys.trimmings.detail(id) });
    },
    onError: (error) => {
      handleApiError(error, "更新");
    },
  });
};
