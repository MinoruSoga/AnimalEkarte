import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { PaymentMethodMaster } from "@/types/generated/models";

export interface UpdatePaymentMethodRequest {
  name?: string;
  display_order?: number;
  is_active?: boolean;
}

export const updatePaymentMethod = async (
  id: number,
  data: UpdatePaymentMethodRequest,
): Promise<PaymentMethodMaster> => {
  const { data: res } = await axios.patch<PaymentMethodMaster>(`/v1/payment-methods/${id}`, data);
  return res;
};

export const useUpdatePaymentMethod = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdatePaymentMethodRequest }) =>
      updatePaymentMethod(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["payment-methods"] }),
  });
};
