import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationAppointment } from "@/types";
import { transformReservation } from "./transforms";
import type { Appointment as BackendReservation } from "@/types/generated/models";

interface ReservationsListResponse {
  data: BackendReservation[];
  total: number;
  page: number;
  limit: number;
}

export const getReservations = async (): Promise<ReservationAppointment[]> => {
  const { data } = await axios.get<ReservationsListResponse>("/v1/reservations");
  return data.data.map(transformReservation);
};

export const useGetReservations = () => {
  return useQuery({
    queryKey: ["reservations"],
    queryFn: getReservations,
    staleTime: QUERY_STALE_TIMES.REALTIME, // 予約一覧は高頻度変更
    gcTime: QUERY_GC_TIMES.SHORT,
  });
};
