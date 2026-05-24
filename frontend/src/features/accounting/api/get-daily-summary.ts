// BUG-368: レジ締め日次集計 API フック
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

export interface DailySummaryPaymentTotal {
  method: string;
  total: number;
}

export interface DailySummaryCategoryTotal {
  category: string;
  total: number;
}

export interface DailySummary {
  payment_totals: DailySummaryPaymentTotal[];
  category_totals: DailySummaryCategoryTotal[];
  billing_count: number;
  grand_total: number;
}

export const useGetDailySummary = (date: string) => {
  return useQuery({
    queryKey: ["accounting", "daily-summary", date] as const,
    queryFn: async (): Promise<DailySummary> => {
      const { data } = await axios.get<DailySummary>("/v1/accountings/daily-summary", {
        params: { date },
      });
      return data;
    },
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
    enabled: !!date,
  });
};
