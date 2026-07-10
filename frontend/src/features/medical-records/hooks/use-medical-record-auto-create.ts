import { useEffect, type MutableRefObject, type TransitionStartFunction } from "react";
import type { NavigateFunction } from "react-router";

import { paths } from "@/config/paths";
import { handleApiError } from "@/lib/handle-api-error";
import type { MedicalRecord, Pet } from "@/types";

import type { CreateMedicalRecordRequest } from "../api/types";
import type { RecommendationReason } from "../constants/recommendation-reason";
import {
  createReceptionAppointmentTimeRange,
  DEFAULT_RECEPTION_APPOINTMENT_MINUTES,
  formatJSTDate,
  toVisitTypeValue,
  type MedicalRecordReservationType,
} from "./use-medical-record-form-model";

interface AsyncMutation<TVariables, TData> {
  mutateAsync: (variables: TVariables) => Promise<TData>;
}

/** 呼び出し元が保持する予約の最小参照。id 以外は auto-create で不要。 */
interface ReusableAppointmentRef {
  id: string;
}

/**
 * 予約作成リクエスト（auto-create 経路が実際に送信するフィールドのみ）。
 * BE 契約は features/reservations/api/types.ts の CreateReservationRequest と同一。
 * medical-records は reservations feature に依存しないためローカル定義する（S5 教訓: DRY より依存方向優先）。
 */
interface MedicalRecordAppointmentCreateRequest {
  pet_id: number;
  owner_id: number;
  start_time: string;
  end_time: string;
  visit_type: "first" | "revisit";
  reservation_type_id: number;
  status: "in_consultation";
  source: "manual" | "line";
  reservation_route: "record_shortcut";
}

interface UseMedicalRecordAutoCreateParams {
  isNewRecord: boolean;
  selectedPet: Pet | undefined;
  hasAutoCreatedRef: MutableRefObject<boolean>;
  appointmentIdFromState: string | undefined;
  reusableAppointment: ReusableAppointmentRef | undefined;
  isReusableAppointmentLoading: boolean;
  visitDateFromState: string | undefined;
  generalReservationType: MedicalRecordReservationType | undefined;
  createReservationMutation: AsyncMutation<MedicalRecordAppointmentCreateRequest, { id: string }>;
  createMutation: AsyncMutation<CreateMedicalRecordRequest, MedicalRecord>;
  startCreateTransition: TransitionStartFunction;
  visitType: string;
  createRecommendationReason: RecommendationReason | null;
  navigate: NavigateFunction;
}

export function useMedicalRecordAutoCreate({
  isNewRecord,
  selectedPet,
  hasAutoCreatedRef,
  appointmentIdFromState,
  reusableAppointment,
  isReusableAppointmentLoading,
  visitDateFromState,
  generalReservationType,
  createReservationMutation,
  createMutation,
  startCreateTransition,
  visitType,
  createRecommendationReason,
  navigate,
}: UseMedicalRecordAutoCreateParams) {
  useEffect(() => {
    if (!isNewRecord || !selectedPet || hasAutoCreatedRef.current) return;
    if (!appointmentIdFromState && !generalReservationType) return;
    if (!appointmentIdFromState && isReusableAppointmentLoading) return;
    hasAutoCreatedRef.current = true;

    startCreateTransition(async () => {
      try {
        let appointmentId = appointmentIdFromState;
        if (!appointmentId && reusableAppointment) {
          appointmentId = reusableAppointment.id;
        }
        if (!appointmentId) {
          const duration = generalReservationType?.duration_minutes || DEFAULT_RECEPTION_APPOINTMENT_MINUTES;
          const { startTime, endTime } = createReceptionAppointmentTimeRange(duration, visitDateFromState);
          const appointment = await createReservationMutation.mutateAsync({
            pet_id: Number(selectedPet.id),
            owner_id: Number(selectedPet.ownerId),
            start_time: startTime,
            end_time: endTime,
            visit_type: toVisitTypeValue(visitType),
            reservation_type_id: generalReservationType?.id ?? 0,
            status: "in_consultation",
            source: "manual",
            reservation_route: "record_shortcut",
          });
          appointmentId = appointment.id;
        }

        const today = visitDateFromState ?? formatJSTDate(new Date());
        const record = await createMutation.mutateAsync({
          pet_id: selectedPet.id,
          owner_id: selectedPet.ownerId,
          visit_date: today,
          visit_type: toVisitTypeValue(visitType),
          appointment_id: appointmentId,
          status: "draft",
          recommendation_reason: createRecommendationReason ?? "",
        });
        navigate(paths.medicalRecords.detail.getHref(record.id), { replace: true });
      } catch (error) {
        handleApiError(error, "カルテ作成");
        hasAutoCreatedRef.current = false;
      }
    });
  // eslint-disable-next-line react-hooks/exhaustive-deps -- intentional: run only when isNewRecord or petId changes; createMutation/navigate/visitType are stable references
  }, [isNewRecord, selectedPet?.id, appointmentIdFromState, reusableAppointment?.id, isReusableAppointmentLoading, visitDateFromState, generalReservationType?.id]);
}
