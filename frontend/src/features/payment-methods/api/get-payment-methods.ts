import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { PaymentMethodMaster } from "@/types/generated/models";

export const getPaymentMethods = async (): Promise<PaymentMethodMaster[]> => {
  const { data } = await axios.get<PaymentMethodMaster[]>("/v1/payment-methods");
  return data;
};

export const useGetPaymentMethods = () =>
  useQuery({
    queryKey: ["payment-methods"],
    queryFn: getPaymentMethods,
  });
