import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { Estimate } from "../types";
import { transformEstimate } from "./transforms";
import type { BackendEstimate, UpdateEstimateRequest } from "./types";

interface UpdateEstimateParams {
  id: string;
  data: UpdateEstimateRequest;
}

async function updateEstimate({ id, data }: UpdateEstimateParams): Promise<Estimate> {
  const { data: responseData } = await axios.patch<BackendEstimate>(`/v1/estimates/${id}`, data);
  return transformEstimate(responseData);
}

export function useUpdateEstimate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updateEstimate,
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.estimates.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.estimates.detail(id) });
      toast.success("見積書を更新しました");
    },
    onError: (error) => {
      handleApiError(error, "更新");
    },
  });
}
