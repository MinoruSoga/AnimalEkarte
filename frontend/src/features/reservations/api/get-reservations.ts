import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ReservationAppointment } from "@/types";
import { transformReservation } from "./transforms";
import type { ReservationAppointment as BackendReservation } from "@/types/generated/models";

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
  });
};
