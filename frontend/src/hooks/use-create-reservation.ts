import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { transformReservation } from "@/lib/transforms/reservation";
import type { Reservation, ReservationRoute } from "@/lib/transforms/reservation";
import type { Reservation as BackendReservation } from "@/types/generated/models";

/**
 * 予約作成リクエスト
 * models.ts の Appointment から導出
 */
export interface CreateReservationRequest {
  pet_id: number;
  owner_id: number;
  doctor_id?: number;
  start_time: string;
  end_time: string;
  visit_type: string;
  reservation_type_id: number;
  is_designated?: boolean;
  status?: string;
  notes?: string;
  source?: "manual" | "line";
  reservation_route?: ReservationRoute;
}

export interface CreateReservationBatchRequest extends Omit<CreateReservationRequest, "pet_id" | "owner_id"> {
  pets: { owner_id: number; pet_id: number }[];
}

const createReservation = async (
  req: CreateReservationRequest
): Promise<Reservation> => {
  const { data } = await axios.post<BackendReservation>(
    "/v1/reservations",
    req
  );
  return transformReservation(data);
};

const createReservationBatch = async (req: CreateReservationBatchRequest): Promise<Reservation[]> => {
  const { data } = await axios.post<BackendReservation[]>("/v1/reservations/batch", req);
  return data.map(transformReservation);
};

/**
 * Shared hook for creating a reservation.
 * R-F2-S13: reservations feature から昇格。medical-records (use-medical-record-auto-create 経由)
 * が reservations feature を直接 import しないための cross-feature 共有。
 * invalidate queryKey prefix は reservations feature 側と同一を維持する。
 */
export const useCreateReservation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createReservation,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() });
    },
    onError: (error) => handleApiError(error, "予約作成"),
  });
};

export const useCreateReservationBatch = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createReservationBatch,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() });
    },
    onError: (error) => handleApiError(error, "予約作成"),
  });
};
