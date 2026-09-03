import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { ReservationStatus } from "@/types";

interface UpdateStatusPayload {
  id: string;
  status: ReservationStatus;
}

/** 予約ステータスを更新する */
async function updateAppointmentStatus(id: string, status: ReservationStatus): Promise<void> {
  await axios.patch(`/v1/reservations/${id}`, { status });
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
