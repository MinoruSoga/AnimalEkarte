import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { CashRegisterClose as BackendCashRegisterClose } from "@/types/generated/models";
import { transformCashRegisterClose } from "./transforms";
import type { CashRegisterClose } from "./transforms";

export type { CashRegisterClose };

export interface GetCashRegisterClosesParams {
  year?: number;
  month?: number;
  page?: number;
  limit?: number;
}

interface BackendCashRegisterClosesResponse {
  data: BackendCashRegisterClose[];
  total: number;
}

export interface CashRegisterClosesResponse {
  data: CashRegisterClose[];
  total: number;
}

export const getCashRegisterCloses = async (
  params?: GetCashRegisterClosesParams,
): Promise<CashRegisterClosesResponse> => {
  const { data } = await axios.get<BackendCashRegisterClosesResponse>("/v1/cash-register/closes", {
    params,
  });
  return {
    data: data.data.map(transformCashRegisterClose),
    total: data.total,
  };
};

export const useGetCashRegisterCloses = (params?: GetCashRegisterClosesParams, enabled = true) =>
  useQuery({
    queryKey: ["cash-register-closes", params],
    queryFn: () => getCashRegisterCloses(params),
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
    enabled,
  });
