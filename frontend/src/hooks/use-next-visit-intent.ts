import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import {
  transformReservationWithIntent,
  type Reservation,
  type ReservationApiResponse,
} from "./reservation-intent-response";

/**
 * EMR-308: PUT /v1/reservations/pets/:petId/next-visit-intent
 * ペットの次回予定（tentative appointment）を作成または更新する。
 * BE 権限は reservations:create。body は intended_date（YYYY-MM-DD）必須、
 * 既存カルテからの設定時は source_medical_record_id も送る。
 *
 * src/hooks 配置の理由: medical-records ヘッダー（NextVisitButton）だけでなく
 * reception（キャンセル後の再設定）と reservations（期限超過パネルの再設定）も
 * 使う cross-feature mutation。feature 同士の相互 import は禁止のため共有層に置く
 * （use-get-reservations / use-update-reservation と同じ先例）。
 */
export interface SetNextVisitIntentVariables {
  petId: string;
  /** 意図日（YYYY-MM-DD、JST 暦日） */
  intendedDate: string;
  /** 保存済みカルテがあればその id（source_medical_record_id として送信） */
  medicalRecordId?: string;
}

const setNextVisitIntent = async (variables: SetNextVisitIntentVariables): Promise<Reservation> => {
  const { data } = await axios.put<ReservationApiResponse>(
    `/v1/reservations/pets/${variables.petId}/next-visit-intent`,
    {
      intended_date: variables.intendedDate,
      ...(variables.medicalRecordId
        ? { source_medical_record_id: Number(variables.medicalRecordId) }
        : {}),
    },
  );
  return transformReservationWithIntent(data);
};

/**
 * EMR-308: DELETE /v1/reservations/pets/:petId/next-visit-intent
 * ペットのアクティブな次回予定を除去する（冪等）。BE 権限は reservations:edit。
 * active tentative はペットごとに1件（部分ユニーク索引）なので petId で一意に終わらせられる。
 */
const clearNextVisitIntent = async (petId: string): Promise<void> => {
  await axios.delete(`/v1/reservations/pets/${petId}/next-visit-intent`);
};

export const useSetNextVisitIntent = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: setNextVisitIntent,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.all() });
    },
    onError: (error) => handleApiError(error, "次回予定設定"),
  });
};

export const useClearNextVisitIntent = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: clearNextVisitIntent,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.all() });
    },
    onError: (error) => handleApiError(error, "次回予定クリア"),
  });
};
