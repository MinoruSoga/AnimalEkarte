import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { CashRegisterClose } from "@/types/generated/models";

export interface GetCashRegisterClosesParams {
  year?: number;
  month?: number;
  page?: number;
  limit?: number;
}

export interface CashRegisterClosesResponse {
  data: CashRegisterClose[];
  total: number;
}

export const getCashRegisterCloses = async (
  params?: GetCashRegisterClosesParams,
): Promise<CashRegisterClosesResponse> => {
  const { data } = await axios.get<CashRegisterClosesResponse>("/v1/cash-register/closes", {
    params,
  });
  return data;
};

export const useGetCashRegisterCloses = (params?: GetCashRegisterClosesParams) =>
  useQuery({
    queryKey: ["cash-register-closes", params],
    queryFn: () => getCashRegisterCloses(params),
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
