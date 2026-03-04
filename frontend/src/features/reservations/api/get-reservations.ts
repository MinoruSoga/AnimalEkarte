import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ReservationAppointment } from "@/types";
import { transformReservation } from "./transforms";
import type { BackendReservation } from "./types";

export const getReservations = async (): Promise<ReservationAppointment[]> => {
  const { data } = await axios.get<BackendReservation[]>("/v1/reservations");
  return data.map(transformReservation);
};

export const useGetReservations = () => {
  return useQuery({
    queryKey: ["reservations"],
    queryFn: getReservations,
  });
};
