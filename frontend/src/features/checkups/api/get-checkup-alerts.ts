import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type {
  BackendCheckupAlertItem,
  BackendCheckupAlertsResponse,
  CheckupAlertItem,
  CheckupAlertsResponse,
} from "./types";

function transformAlertItem(b: BackendCheckupAlertItem): CheckupAlertItem {
  return {
    checkupId: b.checkup_id,
    petId: b.pet_id,
    petName: b.pet_name,
    ownerName: b.owner_name,
    checkupTypeName: b.checkup_type_name,
    lastCheckupDate: b.last_checkup_date,
    nextDate: b.next_date,
    days: b.days,
  };
}

function transformAlerts(b: BackendCheckupAlertsResponse): CheckupAlertsResponse {
  return {
    overdue: { count: b.overdue.count, items: b.overdue.items.map(transformAlertItem) },
    upcoming: { count: b.upcoming.count, items: b.upcoming.items.map(transformAlertItem) },
  };
}

export const useGetCheckupAlerts = (withinDays = 30) => {
  return useQuery({
    queryKey: ["checkup-alerts", withinDays],
    queryFn: async () => {
      const { data } = await axios.get<BackendCheckupAlertsResponse>("/v1/checkups/alerts", {
        params: { within_days: withinDays },
      });
      return transformAlerts(data);
    },
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
