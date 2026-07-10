import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";

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
    queryKey: ["lstep-visit-conversion", yearMonth, days],
    queryFn: async () => {
      const clinicId = requireStoredClinicId();
      const { data } = await axios.get<VisitConversionSummaryResponse>(
        `/v1/clinics/${clinicId}/lstep/analytics/visit-conversion?year_month=${yearMonth}&days=${days}`
      );
      return data;
    },
    staleTime: 5 * 60 * 1000,
    enabled: !!yearMonth && days > 0,
  });
}
