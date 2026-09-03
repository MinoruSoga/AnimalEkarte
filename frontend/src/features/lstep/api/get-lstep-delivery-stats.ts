import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

export interface DeliveryStatsRow {
  trigger_type: string;
  status: string;
  count: number;
}

export interface MonthlyDeliveryStatsResponse {
  year_month: string;
  rows: DeliveryStatsRow[];
}

// GET /api/v1/clinics/:clinic_id/lstep/analytics/delivery-stats?year_month=YYYY-MM
export function useGetLstepDeliveryStats(yearMonth: string) {
  return useQuery({
    queryKey: queryKeys.lstepDeliveryStats(yearMonth),
    queryFn: async () => {
      const clinicId = requireStoredClinicId();
      const { data } = await axios.get<MonthlyDeliveryStatsResponse>(
        `/v1/clinics/${clinicId}/lstep/analytics/delivery-stats?year_month=${yearMonth}`,
      );
      return data;
    },
    staleTime: QUERY_STALE_TIMES.MEDIUM, // 5分キャッシュ
    enabled: !!yearMonth,
  });
}
