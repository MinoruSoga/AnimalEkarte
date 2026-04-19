// BUG-368: レジ締め日次集計 API フック
import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

export interface PaymentMethodTotal {
  method: string;
  total: number;
}

export interface CategoryTotal {
  category: string;
  total: number;
}

export interface DailySummaryResponse {
  payment_totals: PaymentMethodTotal[];
  category_totals: CategoryTotal[];
  billing_count: number;
  grand_total: number;
}

export const useGetDailySummary = (date: string) => {
  return useQuery({
    queryKey: ["accounting", "daily-summary", date] as const,
    queryFn: async (): Promise<DailySummaryResponse> => {
      const { data } = await axios.get<DailySummaryResponse>("/v1/accountings/daily-summary", {
        params: { date },
      });
      return data;
    },
    staleTime: QUERY_STALE_TIMES.MEDIUM,
  });
};
