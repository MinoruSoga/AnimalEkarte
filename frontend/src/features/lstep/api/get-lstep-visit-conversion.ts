import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

interface VisitConversionRow {
  trigger_type: string;
  delivered_count: number;
  visited_count: number;
  visit_rate: number;
}

export interface VisitConversionSummaryResponse {
  year_month: string;
  days: number;
  delivered_count: number;
  visited_count: number;
  visit_rate: number;
  rows: VisitConversionRow[];
}

// GET /api/v1/clinics/:clinic_id/lstep/analytics/visit-conversion?year_month=YYYY-MM&days=30
export function useGetLstepVisitConversion(yearMonth: string, days = 30) {
  return useQuery({
    queryKey: queryKeys.lstepVisitConversion(yearMonth, days),
    queryFn: async () => {
      const clinicId = requireStoredClinicId();
      const { data } = await axios.get<VisitConversionSummaryResponse>(
        `/v1/clinics/${clinicId}/lstep/analytics/visit-conversion?year_month=${yearMonth}&days=${days}`,
      );
      return data;
    },
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    enabled: !!yearMonth && days > 0,
  });
}
