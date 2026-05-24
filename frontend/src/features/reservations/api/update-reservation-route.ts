import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { transformReservation } from "./transforms";
import type { BackendReservation } from "./types";
import type { ReservationRoute } from "../constants/reservation-route";

interface UpdateReservationRouteBody {
  route: ReservationRoute | "";
}

const updateReservationRoute = async (
  id: string,
  body: UpdateReservationRouteBody
) =>
  transformReservation(
    (await axios.patch<BackendReservation>(`/v1/reservations/${id}/reservation-route`, body))
      .data
  );

export function useUpdateReservationRoute(reservationId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateReservationRouteBody) =>
      updateReservationRoute(reservationId, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["reservations"] });
      queryClient.invalidateQueries({ queryKey: ["reservation", reservationId] });
    },
    onError: (error: unknown) => handleApiError(error, "予約経路更新"),
  });
}
