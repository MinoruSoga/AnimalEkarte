import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { Estimate } from "../types";
import { transformEstimate } from "./transforms";
import type { BackendEstimate, CreateEstimateRequest } from "./types";

async function createEstimate(req: CreateEstimateRequest): Promise<Estimate> {
  const { data } = await axios.post<BackendEstimate>("/v1/estimates", req);
  return transformEstimate(data);
}

export function useCreateEstimate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createEstimate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.estimates.all() });
      toast.success("見積書を作成しました");
    },
    onError: (error) => {
      handleApiError(error, "作成");
    },
  });
}
