import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { LineSendType } from "./send-line-message";

export interface LineSendHistoryItem {
  id: string;
  message_type: LineSendType;
  text: string | null;
  file_url: string | null;
  purpose_tag: string | null;
  sent_at: string;
  status: "sent" | "failed" | "pending";
}

export interface LineSendHistoryResponse {
  items: LineSendHistoryItem[];
}

async function getLineSendHistory(
  clinicId: string,
  ownerId: string
): Promise<LineSendHistoryResponse> {
  const { data } = await axios.get<LineSendHistoryResponse>(
    `/v1/clinics/${clinicId}/owners/${ownerId}/lstep/send-history`
  );
  return data;
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
