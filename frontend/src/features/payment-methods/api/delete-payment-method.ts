import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export const deletePaymentMethod = async (id: number): Promise<void> => {
  await axios.delete(`/v1/payment-methods/${id}`);
};

export const useDeletePaymentMethod = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deletePaymentMethod(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["payment-methods"] }),
  });
};
