import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
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
  ownerId: string
): Promise<LineSendHistoryItem[]> {
  const { data } = await axios.get<LineSendHistoryResponse>(
    `/v1/clinics/${clinicId}/owners/${ownerId}/line/send-logs`
  );
  return data.items;
}

export function useGetLineSendHistory(ownerId: string) {
  const clinicId = localStorage.getItem("auth_current_clinic:v1") ?? "";
  return useQuery({
    queryKey: ["line-send-history", ownerId],
    queryFn: () => getLineSendHistory(clinicId, ownerId),
    enabled: !!ownerId && !!clinicId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.SHORT,
  });
}
