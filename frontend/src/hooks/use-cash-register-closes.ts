import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { CashRegisterClose as BackendCashRegisterClose } from "@/types/generated/models";
import { transformCashRegisterClose } from "@/lib/transforms/cash-register";
import type { CashRegisterClose } from "@/lib/transforms/cash-register";

export type { CashRegisterClose };

export interface GetCashRegisterClosesParams {
  start_date?: string;
  end_date?: string;
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

/**
 * Shared hook for fetching cash register closes.
 * R-F2-S11: cash-register feature から昇格。accounting (AccountingDetail) が
 * cash-register feature を直接 import しないための cross-feature 共有。
 * queryKey は cash-register feature 側と同一 prefix を維持し React Query cache を共有する。
 */
export const useGetCashRegisterCloses = (params?: GetCashRegisterClosesParams, enabled = true) =>
  useQuery({
    queryKey: ["cash-register-closes", params],
    queryFn: () => getCashRegisterCloses(params),
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
    enabled,
  });
