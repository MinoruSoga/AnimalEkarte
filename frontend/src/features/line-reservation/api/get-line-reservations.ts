import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationAppointment } from "./types";

export interface GetLineReservationsParams {
  view: "month" | "day";
  date: string;
}

export async function getLineReservations(
  clinicId: string,
  params: GetLineReservationsParams
): Promise<ReservationAppointment[]> {
  const { data } = await axios.get<ReservationAppointment[]>(
    `/v1/clinics/${clinicId}/reservations`,
    { params }
  );
  return data ?? [];
}

export function useGetLineReservations(
  clinicId: string | null,
  params: GetLineReservationsParams
) {
  return useQuery({
    queryKey: ["line-reservations", clinicId, params.view, params.date],
    queryFn: () => getLineReservations(clinicId!, params),
    enabled: clinicId !== null,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.SHORT,
  });
}
