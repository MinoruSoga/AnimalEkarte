import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationAppointment } from "@/types";
import { transformReservation } from "./transforms";
import type { Appointment as BackendReservation } from "@/types/generated/models";

export const getReservation = async (
  id: string
): Promise<ReservationAppointment> => {
  const { data } = await axios.get<BackendReservation>(
    `/v1/reservations/${id}`
  );
  return transformReservation(data);
};

export const useGetReservation = (id: string) => {
  return useQuery({
    queryKey: ["reservation", id],
    queryFn: () => getReservation(id),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.REALTIME, // 予約は高頻度変更
    gcTime: QUERY_GC_TIMES.SHORT,
  });
};

