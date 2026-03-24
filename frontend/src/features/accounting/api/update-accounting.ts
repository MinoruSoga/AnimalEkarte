import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { handleApiError } from "@/lib/handle-api-error";
import type { Accounting } from "../types";
import { transformToAccounting } from "./transforms";
import type { BackendAccounting, UpdateAccountingRequest } from "./types";

export const updateAccounting = async (
  id: string,
  req: UpdateAccountingRequest,
): Promise<Accounting> => {
  const { data } = await axios.patch<BackendAccounting>(
    `/v1/accountings/${id}`,
    req,
  );
  return transformToAccounting(data);
};

export const useUpdateAccounting = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      req,
    }: {
      id: string;
      req: UpdateAccountingRequest;
    }) => updateAccounting(id, req),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.accountings.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(id) });
    },
    onError: (error) => {
      handleApiError(error, "更新");
    },
  });
};
