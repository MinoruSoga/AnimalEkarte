import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ReservationAppointment } from "@/types";
import { transformReservation } from "./transforms";
import type { ReservationAppointment as BackendReservation } from "@/types/generated/models";

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
  });
};

