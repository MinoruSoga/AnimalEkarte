import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { CashRegisterClose as BackendCashRegisterClose } from "@/types/generated/models";
import { transformCashRegisterClose } from "./transforms";
import type { CashRegisterClose } from "./transforms";

export type { CashRegisterClose };

export const getCashRegisterClose = async (id: number): Promise<CashRegisterClose> => {
  const { data } = await axios.get<BackendCashRegisterClose>(`/v1/cash-register/closes/${id}`);
  return transformCashRegisterClose(data);
};

export const useGetCashRegisterClose = (id: number | null) =>
  useQuery({
    queryKey: ["cash-register-close", id],
    queryFn: () => getCashRegisterClose(id!),
    enabled: id !== null,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
