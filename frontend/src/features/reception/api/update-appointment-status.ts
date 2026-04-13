import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { ReservationStatus } from "@/types";

interface UpdateStatusPayload {
  id: string;
  status: ReservationStatus;
}

/** 予約ステータスを更新する */
async function updateAppointmentStatus(
  id: string,
  status: ReservationStatus
): Promise<void> {
  await axios.patch(`/v1/reservations/${id}`, { status });
}

/** カンバン操作時のステータス更新 mutation hook */
export function useUpdateAppointmentStatus() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, status }: UpdateStatusPayload) =>
      updateAppointmentStatus(id, status),
    onSuccess: () => {
      // reception クエリキーを持つ全クエリを無効化して再取得
      queryClient.invalidateQueries({ queryKey: ["reception"] });
    },
    onError: (error) => handleApiError(error, "受付ステータスの更新"),
  });
}
