import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformReservation } from "@/lib/transforms/reservation";
import type { Reservation as BackendReservation } from "@/types/generated/models";

const getReservation = async (id: string) =>
  transformReservation((await axios.get<BackendReservation>(`/v1/reservations/${id}`)).data);

export function useGetReservation(id: string) {
  return useQuery({
    queryKey: queryKeys.reservations.detail(id),
    queryFn: () => getReservation(id),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.SHORT,
  });
}
