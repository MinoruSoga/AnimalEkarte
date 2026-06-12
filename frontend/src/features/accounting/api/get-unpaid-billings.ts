// BUG-370: 月末未納者一覧 API フック
import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { BackendAccounting } from "./types";
import { transformToAccounting } from "./transforms";
import type { Accounting } from "./transforms";

export interface UnpaidOwner {
  owner_id: number;
  owner_name: string;
  count: number;
  total_amount: number;
  oldest_scheduled: string;
  latest_scheduled: string;
}

export interface UnpaidSummary {
  total_amount: number;
  billing_count: number;
  owner_count: number;
}

export interface UnpaidByOwnerResponse {
  data: UnpaidOwner[];
  total: number;
  page: number;
  limit: number;
  summary: UnpaidSummary;
}

// #120: start_date/end_date 必須。両方揃うまでクエリは発火しない
interface UnpaidQueryParams {
  startDate: string;
  endDate: string;
  groupBy: "owner" | "billing";
  page: number;
  limit: number;
}

export const useGetUnpaidByOwner = (params: UnpaidQueryParams) => {
  return useQuery({
    queryKey: ["accounting", "unpaid", "owner", params] as const,
    queryFn: async (): Promise<UnpaidByOwnerResponse> => {
      const { data } = await axios.get<UnpaidByOwnerResponse>("/v1/accountings/unpaid", {
        params: {
          start_date: params.startDate,
          end_date: params.endDate,
          group_by: "owner",
          page: params.page,
          limit: params.limit,
        },
      });
      return data;
    },
    enabled: params.groupBy === "owner" && !!params.startDate && !!params.endDate,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

export interface UnpaidByBillingResponse {
  data: Accounting[];
  total: number;
  page: number;
  limit: number;
}

interface BackendUnpaidByBillingResponse {
  data: BackendAccounting[];
  total: number;
  page: number;
  limit: number;
}

export const useGetUnpaidByBilling = (params: UnpaidQueryParams) => {
  return useQuery({
    queryKey: ["accounting", "unpaid", "billing", params] as const,
    queryFn: async (): Promise<UnpaidByBillingResponse> => {
      const { data } = await axios.get<BackendUnpaidByBillingResponse>("/v1/accountings/unpaid", {
        params: {
          start_date: params.startDate,
          end_date: params.endDate,
          group_by: "billing",
          page: params.page,
          limit: params.limit,
        },
      });
      return {
        data: data.data.map(transformToAccounting),
        total: data.total,
        page: data.page,
        limit: data.limit,
      };
    },
    enabled: params.groupBy === "billing" && !!params.startDate && !!params.endDate,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
