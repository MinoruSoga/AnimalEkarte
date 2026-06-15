import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";

export interface DeliveryTriggerSummaryResponse {
  scheduled: number;
  fired: number;
  excluded: number;
  failed: number;
  suppressed_by_priority: number;
  excluded_reason_breakdown: Record<string, number>;
}

// GET /api/clinics/:clinic_id/lstep/delivery-monitor/summary
export function useGetLstepDeliveryTriggerSummary(
  from: string,
  to: string,
  triggerType?: string
) {
  return useQuery({
    queryKey: ["lstep-delivery-trigger-summary", from, to, triggerType ?? ""],
    queryFn: async () => {
      const clinicId = requireStoredClinicId();
      const params = new URLSearchParams({ from, to });
      if (triggerType) params.set("trigger_type", triggerType);
      const { data } = await axios.get<DeliveryTriggerSummaryResponse>(
        `/v1/clinics/${clinicId}/lstep/delivery-monitor/summary?${params}`
      );
      return data;
    },
    staleTime: 60 * 1000, // 1分キャッシュ
  });
}
