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

// EMR-308/Codex: 拠点横断で開いたカルテ（record.clinicId）上の次回予定操作は、
// グローバル選択クリニックではなくレコード自身の clinicId を X-Clinic-ID で送る
// （P2-15 vitals の clinicHeaderConfig と同じ先例）。省略時はインターセプタが
// グローバル選択値にフォールバックする従来挙動を維持。
function intentClinicConfig(clinicId?: string) {
  return clinicId ? { headers: { "X-Clinic-ID": clinicId } } : undefined;
}

const setNextVisitIntent = async (
  variables: SetNextVisitIntentVariables,
  clinicId?: string,
): Promise<Reservation> => {
  const { data } = await axios.put<ReservationApiResponse>(
    `/v1/reservations/pets/${variables.petId}/next-visit-intent`,
    {
      intended_date: variables.intendedDate,
      ...(variables.medicalRecordId
        ? { source_medical_record_id: Number(variables.medicalRecordId) }
        : {}),
    },
    intentClinicConfig(clinicId),
  );
  return transformReservationWithIntent(data);
};

/**
 * EMR-308: DELETE /v1/reservations/pets/:petId/next-visit-intent
 * ペットのアクティブな次回予定を除去する（冪等）。BE 権限は reservations:edit。
 * active tentative はペットごとに1件（部分ユニーク索引）なので petId で一意に終わらせられる。
 */
const clearNextVisitIntent = async (petId: string, clinicId?: string): Promise<void> => {
  await axios.delete(
    `/v1/reservations/pets/${petId}/next-visit-intent`,
    intentClinicConfig(clinicId),
  );
};

export const useSetNextVisitIntent = (clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (variables: SetNextVisitIntentVariables) => setNextVisitIntent(variables, clinicId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.all() });
    },
    onError: (error) => handleApiError(error, "次回予定設定"),
  });
};

export const useClearNextVisitIntent = (clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (petId: string) => clearNextVisitIntent(petId, clinicId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.reservations.all() });
    },
    onError: (error) => handleApiError(error, "次回予定クリア"),
  });
};
