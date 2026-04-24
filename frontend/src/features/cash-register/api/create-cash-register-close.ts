import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { CashRegisterClose } from "@/types/generated/models";

export interface CreateCashRegisterCloseRequest {
  date: string;
  period: "am" | "pm";
  actual_cash: number;
  memo?: string;
}

export const createCashRegisterClose = async (
  data: CreateCashRegisterCloseRequest,
): Promise<CashRegisterClose> => {
  const { data: res } = await axios.post<CashRegisterClose>("/v1/cash-register/closes", data);
  return res;
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
