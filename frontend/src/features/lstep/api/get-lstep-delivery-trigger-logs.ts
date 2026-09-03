import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

interface DeliveryTriggerLogItem {
  id: string;
  owner_id: string;
  owner_name: string;
  trigger_type: string;
  scheduled_at: string; // ISO datetime
  status: string;
  fired_at?: string; // ISO datetime
  excluded_reason?: string;
}

export interface DeliveryTriggerLogsPageResponse {
  items: DeliveryTriggerLogItem[];
  total: number;
  page: number;
  per_page: number;
}

// GET /api/clinics/:clinic_id/lstep/delivery-monitor/logs
export function useGetLstepDeliveryTriggerLogs(
  from: string,
  to: string,
  triggerType?: string,
  status?: string,
  page = 1,
) {
  return useQuery({
    queryKey: queryKeys.lstepDeliveryTriggerLogs.logs(from, to, triggerType, status, page),
    queryFn: async () => {
      const clinicId = requireStoredClinicId();
      const params = new URLSearchParams({ from, to, page: String(page) });
      if (triggerType) params.set("trigger_type", triggerType);
      if (status) params.set("status", status);
      const { data } = await axios.get<DeliveryTriggerLogsPageResponse>(
        `/v1/clinics/${clinicId}/lstep/delivery-monitor/logs?${params}`,
      );
      return data;
    },
    staleTime: QUERY_STALE_TIMES.MINUTE, // 1分キャッシュ
  });
}
