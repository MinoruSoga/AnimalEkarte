import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ReservationAppointment, CreateReservationRequest } from "./types";

export async function createLineReservation(
  clinicId: string,
  payload: CreateReservationRequest
): Promise<ReservationAppointment> {
  const { data } = await axios.post<ReservationAppointment>(
    `/v1/clinics/${clinicId}/reservations`,
    payload
  );
  return data;
}

export function useCreateLineReservation(clinicId: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateReservationRequest) =>
      createLineReservation(clinicId!, payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["line-reservations", clinicId] });
    },
  });
}
