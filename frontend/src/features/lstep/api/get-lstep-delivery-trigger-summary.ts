import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

export interface DeliveryTriggerSummaryResponse {
  scheduled: number;
  fired: number;
  excluded: number;
  failed: number;
  suppressed_by_priority: number;
  excluded_reason_breakdown: Record<string, number>;
}

// GET /api/clinics/:clinic_id/lstep/delivery-monitor/summary
export function useGetLstepDeliveryTriggerSummary(from: string, to: string, triggerType?: string) {
  return useQuery({
    queryKey: queryKeys.lstepDeliveryTriggerSummary.summary(from, to, triggerType),
    queryFn: async () => {
      const clinicId = requireStoredClinicId();
      const params = new URLSearchParams({ from, to });
      if (triggerType) params.set("trigger_type", triggerType);
      const { data } = await axios.get<DeliveryTriggerSummaryResponse>(
        `/v1/clinics/${clinicId}/lstep/delivery-monitor/summary?${params}`,
      );
      return data;
    },
    staleTime: QUERY_STALE_TIMES.MINUTE, // 1分キャッシュ
  });
}
