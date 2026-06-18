import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { CashRegisterClose as BackendCashRegisterClose } from "@/types/generated/models";
import type { CashRegisterPeriod } from "../constants";
import { transformCashRegisterClose } from "./transforms";
import type { CashRegisterClose } from "./transforms";

export interface CreateCashRegisterCloseRequest {
  date: string;
  period: CashRegisterPeriod;
  actual_cash: number;
  memo?: string;
}

export type { CashRegisterClose };

export const createCashRegisterClose = async (
  data: CreateCashRegisterCloseRequest,
): Promise<CashRegisterClose> => {
  const { data: res } = await axios.post<BackendCashRegisterClose>("/v1/cash-register/closes", data);
  return transformCashRegisterClose(res);
};

export const useCreateCashRegisterClose = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createCashRegisterClose,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["cash-register-closes"] });
      qc.invalidateQueries({ queryKey: ["cash-register-preview"] });
    },
    onError: (error) => handleApiError(error, "レジ締め作成"),
  });
};
