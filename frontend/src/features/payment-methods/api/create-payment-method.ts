import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { PaymentMethodMaster } from "@/types/generated/models";

export interface CreatePaymentMethodRequest {
  name: string;
  display_order?: number;
  is_active?: boolean;
}

export const createPaymentMethod = async (
  data: CreatePaymentMethodRequest,
): Promise<PaymentMethodMaster> => {
  const { data: res } = await axios.post<PaymentMethodMaster>("/v1/payment-methods", data);
  return res;
};

export const useCreatePaymentMethod = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createPaymentMethod,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["payment-methods"] }),
  });
};
