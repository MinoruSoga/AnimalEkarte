import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { transformReservation } from "@/lib/transforms/reservation";
import type { Reservation } from "@/lib/transforms/reservation";
import type { Reservation as BackendReservation } from "@/types/generated/models";

/**
 * 予約更新リクエスト
 * models.ts の Appointment から導出
 */
export interface UpdateReservationRequest {
  start_time?: string;
  end_time?: string;
  pet_id?: number;
  owner_id?: number;
  visit_type?: string;
  reservation_type_id?: number;
  doctor_id?: number;
  is_designated?: boolean;
  status?: string;
  notes?: string;
}

const updateReservation = async (
  id: string,
  req: UpdateReservationRequest,
): Promise<Reservation> => {
  const { data } = await axios.patch<BackendReservation>(`/v1/reservations/${id}`, req);
  return transformReservation(data);
};

/**
 * Shared hook for updating a reservation.
 * R-F2-S12: reservations feature から昇格。reception (use-reception-modal-handlers)
 * が reservations feature を直接 import しないための cross-feature 共有。
 * invalidate queryKey prefix は reservations feature 側と同一を維持する。
 */
export const useUpdateReservation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateReservationRequest }) =>
      updateReservation(id, req),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() });
      // FE4-6 fix: detail クエリの実キーは queryKeys.reservations.detail(id)
      // （正本: src/hooks/use-get-reservation.ts）。list prefix invalidation はこれを包含しない。
      // 先例: ReservationRouteSelect/hooks/use-update-reservation-route.ts は list + detail の両方を invalidate している。
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.detail(id) });
    },
    onError: (error) => handleApiError(error, "予約更新"),
  });
};
