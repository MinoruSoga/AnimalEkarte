import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { Estimate } from "../types";
import { transformEstimate } from "./transforms";
import type { BackendEstimate } from "./types";

export type CreateEstimateSuccessorRequest = {
  reason: string;
  title?: string;
  comment?: string;
  notes?: string;
};

type CreateEstimateSuccessorVars = {
  id: string;
} & CreateEstimateSuccessorRequest;

async function createEstimateSuccessor(
  id: string,
  req: CreateEstimateSuccessorRequest,
): Promise<Estimate> {
  const { data } = await axios.post<BackendEstimate>(`/v1/estimates/${id}/successors`, req);
  return transformEstimate(data);
}

export function useCreateEstimateSuccessor() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, ...req }: CreateEstimateSuccessorVars) => createEstimateSuccessor(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.estimates.all() });
      toast.success("後継ドラフトを作成しました");
    },
    onError: (error) => {
      handleApiError(error, "後継ドラフト作成");
    },
  });
}
