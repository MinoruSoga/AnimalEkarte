import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

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
    queryKey: ["lstep-delivery-stats", yearMonth],
    queryFn: async () => {
      const clinicId = localStorage.getItem("auth_current_clinic:v1");
      if (!clinicId) {
        throw new Error("クリニックが選択されていません。ページをリロードしてください。");
      }
      const { data } = await axios.get<MonthlyDeliveryStatsResponse>(
        `/v1/clinics/${clinicId}/lstep/analytics/delivery-stats?year_month=${yearMonth}`
      );
      return data;
    },
    staleTime: 5 * 60 * 1000, // 5分キャッシュ
    enabled: !!yearMonth,
  });
}
