import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { AccountingRecord } from "@/types";
import { transformAccounting } from "./transforms";
import type { BackendAccounting, CreateAccountingRequest } from "./types";

export const createAccounting = async (
  req: CreateAccountingRequest
): Promise<AccountingRecord> => {
  const { data } = await axios.post<BackendAccounting>(
    "/v1/accountings",
    req
  );
  return transformAccounting(data);
};

export const useCreateAccounting = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createAccounting,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["accountings"] });
    },
    onError: (error) => {
      handleApiError(error, "作成");
    },
  });
};
