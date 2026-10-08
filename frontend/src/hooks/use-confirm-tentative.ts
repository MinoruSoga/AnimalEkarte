import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import {
  transformReservationWithIntent,
  type Reservation,
  type ReservationApiResponse,
} from "./reservation-intent-response";
import {
  buildConfirmTentativeBody,
  type ConfirmTentativeRequestBody,
} from "@/components/shared/ReservationFormModal/confirm-tentative";

/**
 * EMR-308: POST /v1/reservations/:id/confirm-tentative
 * tentative（次回予定・未予約）行を同一行のまま実予約へ昇格する。
 * BE 権限は reservations:edit（routes.go:162）。
 * 新規作成ではなく対象 tentative 行を更新するため、別行の予約は作らない。
 *
 * src/hooks 配置の理由: medical-records ヘッダー（NextVisitButton）と
 * reservations（予約管理カレンダー/期限超過パネル）の双方が使う cross-feature mutation。
 */
export interface ConfirmTentativeVariables {
  /** 対象 tentative 予約 id */
  reservationId: string;
  /** 確定フォームの入力値（Partial<Reservation>）。buildConfirmTentativeBody で body 化。 */
  formData: Partial<Reservation>;
}

/**
 * formData.start/end が欠けた場合はモーダル側バリデーションが先に弾く想定だが、
 * fail-closed で body 未構築を明示エラーにする。
 */
class ConfirmTentativeBodyError extends Error {
  constructor() {
    super("confirm-tentative: start/end is required");
    this.name = "ConfirmTentativeBodyError";
  }
}

const confirmTentative = async (variables: ConfirmTentativeVariables): Promise<Reservation> => {
  const body: ConfirmTentativeRequestBody | null = buildConfirmTentativeBody(variables.formData);
  if (body === null) throw new ConfirmTentativeBodyError();
  const { data } = await axios.post<ReservationApiResponse>(
    `/v1/reservations/${variables.reservationId}/confirm-tentative`,
    body,
  );
  return transformReservationWithIntent(data);
};

export const useConfirmTentative = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: confirmTentative,
    onSuccess: (reservation) => {
      // useUpdateReservation と同等の invalidation（一覧 + 受付 + detail）
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.detail(reservation.id) });
    },
    onError: (error) => handleApiError(error, "次回予定の確定"),
  });
};
