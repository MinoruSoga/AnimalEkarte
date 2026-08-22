import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type MutableRefObject,
  type TransitionStartFunction,
} from "react";
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
  canCreate: boolean;
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
  tab?: string | null;
}

export function medicalRecordAfterCreateHref(recordId: string, tab?: string | null): string {
  const path = paths.medicalRecords.detail.getHref(recordId);
  if (!tab) return path;
  return `${path}?${new URLSearchParams({ tab }).toString()}`;
}

export type MedicalRecordAutoCreateFailurePhase = "appointment" | "medical-record";

type MedicalRecordAutoCreateFailure =
  | { phase: "appointment"; appointmentId: null }
  | { phase: "medical-record"; appointmentId: string };

export function useMedicalRecordAutoCreate({
  isNewRecord,
  canCreate,
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
  tab,
}: UseMedicalRecordAutoCreateParams) {
  const canCreateRef = useRef(canCreate);
  const selectedPetStatusRef = useRef(selectedPet?.status);
  const isCreateInFlightRef = useRef(false);
  const [failure, setFailure] = useState<MedicalRecordAutoCreateFailure | null>(null);

  useLayoutEffect(() => {
    canCreateRef.current = canCreate;
  }, [canCreate]);
  useLayoutEffect(() => {
    selectedPetStatusRef.current = selectedPet?.status;
  }, [selectedPet?.status]);

  const createAppointmentAndRecord = useCallback((retainedAppointmentId?: string) => {
    if (
      isCreateInFlightRef.current
      || !isNewRecord
      || !selectedPet
      || canCreateRef.current !== true
      || selectedPetStatusRef.current === "死亡"
    ) return;

    const availableAppointmentId = retainedAppointmentId
      ?? appointmentIdFromState
      ?? reusableAppointment?.id;
    if (!availableAppointmentId && !generalReservationType) return;
    if (!availableAppointmentId && isReusableAppointmentLoading) return;

    isCreateInFlightRef.current = true;
    startCreateTransition(async () => {
      let appointmentId = availableAppointmentId;
      try {
        if (!appointmentId) {
          if (
            canCreateRef.current !== true
            || selectedPetStatusRef.current === "死亡"
          ) return;

          const duration = generalReservationType?.duration_minutes || DEFAULT_RECEPTION_APPOINTMENT_MINUTES;
          const { startTime, endTime } = createReceptionAppointmentTimeRange(duration, visitDateFromState);
          const appointment = await createReservationMutation.mutateAsync({
            pet_id: Number(selectedPet.id),
            owner_id: Number(selectedPet.ownerId),
            start_time: startTime,
            end_time: endTime,
            visit_type: toVisitTypeValue(visitType) ?? "revisit",
            reservation_type_id: generalReservationType?.id ?? 0,
            status: "in_consultation",
            source: "manual",
            reservation_route: "record_shortcut",
          });
          appointmentId = appointment.id;
        }

        if (
          canCreateRef.current !== true
          || selectedPetStatusRef.current === "死亡"
        ) return;

        const today = visitDateFromState ?? formatJSTDate(new Date());
        const record = await createMutation.mutateAsync({
          pet_id: selectedPet.id,
          owner_id: selectedPet.ownerId,
          visit_date: today,
          visit_type: toVisitTypeValue(visitType) ?? "revisit",
          appointment_id: appointmentId,
          status: "draft",
          recommendation_reason: createRecommendationReason ?? "",
        });
        setFailure(null);
        navigate(medicalRecordAfterCreateHref(record.id, tab), { replace: true });
      } catch (error) {
        setFailure(
          appointmentId
            ? { phase: "medical-record", appointmentId }
            : { phase: "appointment", appointmentId: null },
        );
        handleApiError(error, "カルテ作成");
      } finally {
        isCreateInFlightRef.current = false;
      }
    });
  }, [
    appointmentIdFromState,
    createMutation,
    createRecommendationReason,
    createReservationMutation,
    generalReservationType,
    isNewRecord,
    isReusableAppointmentLoading,
    navigate,
    reusableAppointment?.id,
    selectedPet,
    startCreateTransition,
    visitDateFromState,
    visitType,
    tab,
  ]);

  useEffect(() => {
    if (!isNewRecord || !selectedPet || hasAutoCreatedRef.current) return;
    if (selectedPet.status === "死亡" || canCreate !== true) return;
    if (!appointmentIdFromState && !generalReservationType) return;
    if (!appointmentIdFromState && isReusableAppointmentLoading) return;

    hasAutoCreatedRef.current = true;
    createAppointmentAndRecord();
  }, [
    appointmentIdFromState,
    canCreate,
    createAppointmentAndRecord,
    generalReservationType,
    hasAutoCreatedRef,
    isNewRecord,
    isReusableAppointmentLoading,
    selectedPet,
  ]);

  const retry = useCallback(() => {
    if (failure === null) return;
    createAppointmentAndRecord(failure.appointmentId ?? undefined);
  }, [createAppointmentAndRecord, failure]);

  return {
    failurePhase: failure?.phase ?? null,
    retry,
  };
}
