import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { getStoredClinicId } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { LineSendType } from "./send-line-message";

export interface LineSendHistoryItem {
  id: number;
  message_type: LineSendType;
  content_summary: string | null;
  status: "sent" | "failed" | "pending";
  error_message: string | null;
  sent_at: string;
}

interface LineSendHistoryResponse {
  items: LineSendHistoryItem[];
}

async function getLineSendHistory(
  clinicId: string,
  ownerId: string,
): Promise<LineSendHistoryItem[]> {
  const { data } = await axios.get<LineSendHistoryResponse>(
    `/v1/clinics/${clinicId}/owners/${ownerId}/line/send-logs`,
  );
  return data.items;
}

export function useGetLineSendHistory(ownerId: string) {
  const clinicId = getStoredClinicId();
  return useQuery({
    queryKey: queryKeys.lineSendHistory(ownerId),
    queryFn: () => {
      if (clinicId === null) {
        return Promise.reject(new Error("clinic_id is not selected"));
      }
      return getLineSendHistory(clinicId, ownerId);
    },
    enabled: !!ownerId && !!clinicId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.SHORT,
    // pending (送信中) の行がある間は 5 秒間隔でポーリングして送信結果を反映する
    refetchInterval: (query) => {
      const items = query.state.data ?? [];
      return items.some((item) => item.status === "pending") ? 5000 : false;
    },
  });
}
