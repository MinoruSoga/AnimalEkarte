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

export interface ClinicDailySummaryItem {
  clinic_id: number;
  summary: DailySummary;
}

export interface PerClinicDailySummaryResponse {
  per_clinic: ClinicDailySummaryItem[];
}

export const useGetDailySummary = (date: string, clinicIds?: string[]) => {
  const isMultiClinic = clinicIds !== undefined && clinicIds.length > 1;
  return useQuery({
    queryKey: ["accounting", "daily-summary", date, clinicIds] as const,
    queryFn: async (): Promise<DailySummary | PerClinicDailySummaryResponse> => {
      const params: Record<string, string> = { date };
      if (isMultiClinic) params.clinic_ids = clinicIds.join(",");
      const { data } = await axios.get<DailySummary | PerClinicDailySummaryResponse>(
        "/v1/accountings/daily-summary",
        { params },
      );
      return data;
    },
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
    enabled: !!date,
  });
};
