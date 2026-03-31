import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { Accounting } from "../types";
import { transformToAccounting } from "./transforms";
import type { BackendAccounting, CreateAccountingRequest } from "./types";

export const createAccounting = async (
  req: CreateAccountingRequest,
): Promise<Accounting> => {
  const { data } = await axios.post<BackendAccounting>(
    "/v1/accountings",
    req,
  );
  return transformToAccounting(data);
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
