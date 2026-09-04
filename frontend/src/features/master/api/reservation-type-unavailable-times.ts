import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { unavailableTimesKey } from "@/hooks/use-reservation-type-unavailable-times";

// FE6-16: useGetUnavailableTimes / transformUnavailableTime / ReservationTypeUnavailableTime 型は
// components/shared/ReservationFormModal からも参照される cross-feature データ取得のため
// @/hooks/use-reservation-type-unavailable-times.ts へ移設。
// このファイルには master 画面専用の CRUD mutation のみ残す。
export { useGetUnavailableTimes } from "@/hooks/use-reservation-type-unavailable-times";

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export type CreateUnavailableTimeRequest = {
  unavailable_type: string;
  start_time: string;
  end_time: string;
  day_of_week?: number;
  specific_date?: string;
};

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function createUnavailableTime(
  reservationTypeId: string,
  req: CreateUnavailableTimeRequest,
): Promise<void> {
  await axios.post(`/v1/masters/reservation-types/${reservationTypeId}/unavailable-times`, req);
}

async function deleteUnavailableTime(reservationTypeId: string, id: number): Promise<void> {
  await axios.delete(`/v1/masters/reservation-types/${reservationTypeId}/unavailable-times/${id}`);
}

// ─────────────────────────────────────────────────
// Mutation hooks
// ─────────────────────────────────────────────────

export function useCreateUnavailableTime(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateUnavailableTimeRequest) =>
      createUnavailableTime(reservationTypeId, req),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: unavailableTimesKey(clinicId, reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "予約不可時間の作成"),
  });
}

export function useDeleteUnavailableTime(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteUnavailableTime(reservationTypeId, id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: unavailableTimesKey(clinicId, reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "予約不可時間の削除"),
  });
}
