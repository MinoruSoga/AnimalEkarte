import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { transformReservation } from "@/lib/transforms/reservation";
import type { ReservationRoute } from "@/lib/transforms/reservation";
import type { Reservation as BackendReservation } from "@/types/generated/models";

interface UpdateReservationRouteBody {
  route: ReservationRoute | "";
}

const updateReservationRoute = async (id: string, body: UpdateReservationRouteBody) =>
  transformReservation(
    (await axios.patch<BackendReservation>(`/v1/reservations/${id}/reservation-route`, body)).data,
  );

export function useUpdateReservationRoute(reservationId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateReservationRouteBody) => updateReservationRoute(reservationId, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.detail(reservationId) });
    },
    onError: (error: unknown) => handleApiError(error, "予約経路更新"),
  });
}
