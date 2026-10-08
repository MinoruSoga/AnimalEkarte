import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import {
  transformReservationWithIntent,
  type Reservation,
  type ReservationApiResponse,
} from "@/hooks/reservation-intent-response";
import type { ReservationStatus } from "@/types";

interface UpdateStatusPayload {
  id: string;
  status: ReservationStatus;
}

/**
 * 予約ステータスを更新する。
 * EMR-308: PATCH 応答を transform して返す — 呼び出し側が had_intended_date /
 * pet_id（tentative 由来 provenance）を見てキャンセル後の任意再設定を提示できるようにする。
 */
async function updateAppointmentStatus(
  id: string,
  status: ReservationStatus,
): Promise<Reservation> {
  const { data } = await axios.patch<ReservationApiResponse>(`/v1/reservations/${id}`, { status });
  return transformReservationWithIntent(data);
}

/** カンバン操作時のステータス更新 mutation hook */
export function useUpdateAppointmentStatus() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, status }: UpdateStatusPayload) => updateAppointmentStatus(id, status),
    onError: (error) => handleApiError(error, "受付ステータスの更新"),
    // 成否を問わず reception の server state へ exactly once 再同期する。
    // Promise を返し、再同期完了まで mutation の settled 処理を維持する。
    onSettled: () => queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() }),
  });
}
