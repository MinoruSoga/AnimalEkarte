import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { CashRegisterClose as BackendCashRegisterClose } from "@/types/generated/models";
import type { CashRegisterPeriod } from "../lib/constants";
import { transformCashRegisterClose } from "./transforms";
import type { CashRegisterClose } from "./transforms";

interface CreateCashRegisterCloseRequest {
  date: string;
  period: CashRegisterPeriod;
  actual_cash: number;
  memo?: string;
}

const createCashRegisterClose = async (
  data: CreateCashRegisterCloseRequest,
): Promise<CashRegisterClose> => {
  const { data: res } = await axios.post<BackendCashRegisterClose>(
    "/v1/cash-register/closes",
    data,
  );
  return transformCashRegisterClose(res);
};

export const useCreateCashRegisterClose = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createCashRegisterClose,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.cashRegister.closes.all() });
      qc.invalidateQueries({ queryKey: queryKeys.cashRegister.preview.all() });
    },
    onError: (error) => handleApiError(error, "レジ締め作成"),
  });
};
