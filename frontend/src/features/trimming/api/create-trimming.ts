import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { TrimmingUI } from "@/types";
import { transformTrimming } from "./transforms";
import type { BackendTrimming, CreateTrimmingRequest } from "@/types/trimming";

const createTrimming = async (req: CreateTrimmingRequest): Promise<TrimmingUI> => {
  const { data } = await axios.post<BackendTrimming>("/v1/trimmings", req);
  return transformTrimming(data);
};

export const useCreateTrimming = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createTrimming,
    onSuccess: () => {
      // ["trimmings"] 配下のすべてのキー（一覧・ペット別一覧）を無効化
      queryClient.invalidateQueries({ queryKey: queryKeys.trimmings.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() });
    },
    onError: (error) => {
      handleApiError(error, "作成");
    },
  });
};
