import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { CashRegisterClose } from "@/types/generated/models";

export const getCashRegisterClose = async (id: number): Promise<CashRegisterClose> => {
  const { data } = await axios.get<CashRegisterClose>(`/v1/cash-register/closes/${id}`);
  return data;
};

export const useGetCashRegisterClose = (id: number | null) =>
  useQuery({
    queryKey: ["cash-register-close", id],
    queryFn: () => getCashRegisterClose(id!),
    enabled: id !== null,
  });
