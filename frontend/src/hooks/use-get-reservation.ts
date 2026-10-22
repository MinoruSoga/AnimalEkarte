import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import {
  transformReservationWithIntent,
  type ReservationApiResponse,
} from "./reservation-intent-response";

const getReservation = async (id: string, clinicId?: string) =>
  transformReservationWithIntent(
    (
      await axios.get<ReservationApiResponse>(
        `/v1/reservations/${id}`,
        clinicId ? { headers: { "X-Clinic-ID": clinicId } } : undefined,
      )
    ).data,
  );

export function useGetReservation(id: string, clinicId?: string) {
  return useQuery({
    // Codex PR606 再指摘: 拠点横断フォームは行の clinicId を末尾に付けて分離
    queryKey: clinicId
      ? [...queryKeys.reservations.detail(id), clinicId]
      : queryKeys.reservations.detail(id),
    queryFn: () => getReservation(id, clinicId),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.SHORT,
  });
}
