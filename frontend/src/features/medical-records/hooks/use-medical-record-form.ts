import { useState, useEffect, useRef, useTransition, useCallback, useActionState } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { useGetPet } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useCreateReservation } from "@/features/reservations/api/create-reservation";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useCreateMedicalRecord } from "../api/create-medical-record";
import { useUpdateMedicalRecord } from "../api/update-medical-record";
import { useUpdateInquiry } from "../api/inquiries";
import { useUpdateClinicalPlan } from "../api/clinical-plan";
import type { UpdateMedicalRecordRequest } from "../api/types";
import type { RecommendationReason } from "../constants/recommendation-reason";
import type { TreatmentItem } from "../components/TreatmentTable";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
import { useApplyMedicalRecord } from "./use-apply-medical-record";
import { useMedicalRecordDraft } from "./use-medical-record-draft";
import { useMedicalRecordManualErrors } from "./use-medical-record-manual-errors";
import { useMedicalRecordOwnerChange } from "./use-medical-record-owner-change";

const DEFAULT_CHIEF_COMPLAINT = "# どんな症状\n\n# どこが\n\n# いつから\n\n# その他・備考\n\n# フリースペース";
const DEFAULT_TREATMENT_POLICY = "# 治療方針";
const DEFAULT_PLAN = "# 治療方針";
const DEFAULT_ASSESSMENT = "# 診断詳細";
const DEFAULT_RECEPTION_APPOINTMENT_MINUTES = 15;
const JST_OFFSET_MS = 9 * 60 * 60 * 1000;

function toVisitTypeValue(visitTypeLabel: string): "first" | "revisit" {
  return visitTypeLabel === "初診" || visitTypeLabel === "first" ? "first" : "revisit";
}

function padDatePart(value: number): string {
  return String(value).padStart(2, "0");
}

function formatJSTDate(date: Date): string {
  const jstDate = new Date(date.getTime() + JST_OFFSET_MS);
  return `${jstDate.getUTCFullYear()}-${padDatePart(jstDate.getUTCMonth() + 1)}-${padDatePart(jstDate.getUTCDate())}`;
}

function formatJSTDateTime(date: Date): string {
  const jstDate = new Date(date.getTime() + JST_OFFSET_MS);
  return `${formatJSTDate(date)}T${padDatePart(jstDate.getUTCHours())}:${padDatePart(jstDate.getUTCMinutes())}:00+09:00`;
}

function createReceptionAppointmentTimeRange(durationMinutes: number) {
  const start = new Date();
  start.setSeconds(0, 0);
  const end = new Date(start);
  end.setMinutes(end.getMinutes() + durationMinutes);
  return {
    startTime: formatJSTDateTime(start),
    endTime: formatJSTDateTime(end),
  };
}

function normalizeAppointmentId(value: unknown): string | undefined {
  if (typeof value === "string" && value !== "") return value;
  if (typeof value === "number" && Number.isFinite(value)) return String(value);
  return undefined;
}

export function useMedicalRecordForm(recordId?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isNewRecord = !recordId;
  const { canEdit } = usePermission("medical-records");

  const [activeTab, setActiveTab] = useState("問診");
  const [visitType, setVisitType] = useState("再診");
  const [isCreating, startCreateTransition] = useTransition();
  const { manualErrors, setManualErrors } = useMedicalRecordManualErrors({ setActiveTab });

  // 次回来院推奨日
  const [nextVisitDate, setNextVisitDate] = useState("");
  const [isNextVisitDateValid, setIsNextVisitDateValid] = useState(true);

  // 推奨理由 (create mode 専用 state; edit mode では existingRecord から取得)
  const [createRecommendationReason, setCreateRecommendationReason] =
    useState<RecommendationReason | null>(null);

  const hasAutoCreatedRef = useRef(false);
  const [treatmentPlanItems, setTreatmentPlanItems] = useState<TreatmentItem[]>([]);
  const [treatmentCompletedItems, setTreatmentCompletedItems] = useState<TreatmentItem[]>([]);

  // 問診タブの状態
  const [chiefComplaint, setChiefComplaint] = useState(DEFAULT_CHIEF_COMPLAINT);
  const [chiefComplaintTypeId, setChiefComplaintTypeId] = useState<number | null>(null);
  const [treatmentPolicy, setTreatmentPolicy] = useState(DEFAULT_TREATMENT_POLICY);

  // 診察/治療プランタブの状態（SOAPS）
  const [plan, setPlan] = useState(DEFAULT_PLAN);
  const [assessment, setAssessment] = useState(DEFAULT_ASSESSMENT);

  // 診断マスタの状態
  const [diagnosis1CategoryId, setDiagnosis1CategoryId] = useState<number | null>(null);
  const [diagnosis1NameId, setDiagnosis1NameId] = useState<number | null>(null);
  const [diagnosis2CategoryId, setDiagnosis2CategoryId] = useState<number | null>(null);
  const [diagnosis2NameId, setDiagnosis2NameId] = useState<number | null>(null);

  const { draftKey } = useMedicalRecordDraft({
    recordId,
    chiefComplaint,
    treatmentPolicy,
    plan,
    assessment,
    visitType,
    diagnosis1CategoryId,
    diagnosis1NameId,
    diagnosis2CategoryId,
    diagnosis2NameId,
    setChiefComplaint,
    setTreatmentPolicy,
    setPlan,
    setAssessment,
    setVisitType,
    setDiagnosis1CategoryId,
    setDiagnosis1NameId,
    setDiagnosis2CategoryId,
    setDiagnosis2NameId,
  });

  // 編集モード: カルテからpetIdを取得
  const { data: existingRecord, isError: isRecordError, isLoading: isRecordLoading } = useGetMedicalRecord(recordId ?? "");

  useApplyMedicalRecord({
    existingRecord,
    setChiefComplaint,
    setTreatmentPolicy,
    setPlan,
    setAssessment,
  });

  // petIdを決定: 新規作成時はURLパラメータ、編集時はカルテのpetId
  const resolvedPetId = isNewRecord ? (petId ?? "") : (existingRecord?.petId ?? "");

  // Petデータを取得
  const { data: selectedPet, isLoading: isPetLoading } = useGetPet(resolvedPetId);

  // Ownerデータを取得（飼主割引率用）
  const resolvedOwnerId = selectedPet?.ownerId ?? "";
  const { data: owner } = useGetOwner(resolvedOwnerId);
  const ownerDiscountRate = owner?.discountRate ?? 0;
  const appointmentIdFromState = normalizeAppointmentId(location.state?.appointmentId);
  const { data: reservationTypeGroups } = useGetReservationTypesGrouped();
  const generalReservationType = reservationTypeGroups
    ?.flatMap((group) => group.types)
    .filter((type) => type.category === "general" && !type.is_internal)
    .sort((a, b) => a.sort_order - b.sort_order)
    .find((type) =>
      toVisitTypeValue(visitType) === "revisit"
        ? type.name.includes("再診")
        : !type.name.includes("再診"),
    )
    ?? reservationTypeGroups
      ?.flatMap((group) => group.types)
      .filter((type) => type.category === "general" && !type.is_internal)
      .sort((a, b) => a.sort_order - b.sort_order)[0];

  const queryClient = useQueryClient();
  const createMutation = useCreateMedicalRecord();
  const createReservationMutation = useCreateReservation();
  const updateMutation = useUpdateMedicalRecord();
  const updateInquiryMutation = useUpdateInquiry(recordId ?? "");
  const updateTreatmentPlanMutation = useUpdateClinicalPlan(recordId ?? "");

  // useTransition: save の pending 管理 (rerender-transitions)
  const [isSavingTransition, startSaveTransition] = useTransition();

  // activeTab を保存時に正確に参照するための ref
  const activeTabRef = useRef(activeTab);
  useEffect(() => {
    activeTabRef.current = activeTab;
  }, [activeTab]);

  /**
   * React 19 useActionState を使用したタブ別保存アクション
   */
  const [formState, formAction, isSaving] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      if (!recordId) return { success: false, timestamp: Date.now() };

      try {
        setManualErrors({});
        const currentTab = activeTabRef.current;

        switch (currentTab) {
          case "問診":
            await updateInquiryMutation.mutateAsync({
              chief_complaint: chiefComplaint !== DEFAULT_CHIEF_COMPLAINT ? chiefComplaint : undefined,
              chief_complaint_type_id: chiefComplaintTypeId,
              notes: treatmentPolicy !== DEFAULT_TREATMENT_POLICY ? treatmentPolicy : undefined,
            });
            break;

          case "診察/治療プラン": {
            if (!canEdit) break;
            if (!isNextVisitDateValid) {
              return { success: false, timestamp: Date.now() };
            }
            if (diagnosis1CategoryId && !diagnosis1NameId) {
              const diagError = { diagnosis1_name_id: "診断名を選択してください" };
              setManualErrors(diagError);
              return { success: false, fieldErrors: diagError, timestamp: Date.now() };
            }
            // BUG-102: DEFAULT値でも常に送信する（undefined を送ると BE が 400 を返す）
            const treatmentPlanPayload = {
              treatment_policy: plan,
              diagnosis_details: assessment,
              diagnosis_type_id: diagnosis1CategoryId ?? undefined,
              diagnosis_name_id: diagnosis1NameId ?? undefined,
              diagnosis_2_type_id: diagnosis2CategoryId ?? undefined,
              diagnosis_2_name_id: diagnosis2NameId ?? undefined,
            };
            await updateTreatmentPlanMutation.mutateAsync(treatmentPlanPayload);
            // 次回来院推奨日を更新（空欄の場合は送信しない）
            if (nextVisitDate) {
              await updateMutation.mutateAsync({
                id: recordId as string,
                req: { next_recommended_visit_date: nextVisitDate } as UpdateMedicalRecordRequest,
              });
            }
            break;
          }

          default:
            break;
        }

        localStorage.removeItem(draftKey);
        toast.success("保存しました");
        queryClient.invalidateQueries({ queryKey: ["reception"] });
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE
  );

  const handleBack = useCallback(() => {
    if (location.state?.from) {
      navigate(location.state.from);
      return;
    }

    if (!recordId) {
      navigate(paths.medicalRecords.selectPet.getHref());
    } else {
      navigate(paths.medicalRecords.getHref());
    }
  }, [location.state, navigate, recordId]);

  // 担当医変更ハンドラ
  const handleChangeDoctor = (newDoctorId: string, newDoctorName: string) => {
    if (!recordId) return;
    startSaveTransition(async () => {
      try {
        await updateMutation.mutateAsync({
          id: recordId,
          req: {
            doctor_id: Number(newDoctorId),
            version: existingRecord?.version,
          } as UpdateMedicalRecordRequest,
        });
        toast.success(`担当医を ${newDoctorName} に変更しました`);
      } catch (error) {
        handleApiError(error, "担当医変更");
      }
    });
  };

  const {
    pendingOwnerChange,
    requestOwnerChange,
    confirmOwnerChange,
    cancelOwnerChange,
  } = useMedicalRecordOwnerChange({
    owner,
    recordId,
    existingRecord,
    updateMutation,
    startSaveTransition,
  });

  // 新規作成時: ページ表示と同時にカルテを自動作成
  useEffect(() => {
    if (!isNewRecord || !selectedPet || hasAutoCreatedRef.current) return;
    if (!appointmentIdFromState && !generalReservationType) return;
    hasAutoCreatedRef.current = true;

    startCreateTransition(async () => {
      try {
        let appointmentId = appointmentIdFromState;
        if (!appointmentId) {
          const duration = generalReservationType?.duration_minutes || DEFAULT_RECEPTION_APPOINTMENT_MINUTES;
          const { startTime, endTime } = createReceptionAppointmentTimeRange(duration);
          const appointment = await createReservationMutation.mutateAsync({
            pet_id: Number(selectedPet.id),
            owner_id: Number(selectedPet.ownerId),
            start_time: startTime,
            end_time: endTime,
            visit_type: toVisitTypeValue(visitType),
            reservation_type_id: generalReservationType?.id ?? 0,
            status: "in_consultation",
            source: "manual",
          });
          appointmentId = appointment.id;
        }

        const today = formatJSTDate(new Date());
        const record = await createMutation.mutateAsync({
          pet_id: selectedPet.id,
          owner_id: selectedPet.ownerId,
          visit_date: today,
          visit_type: visitType,
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
  }, [isNewRecord, selectedPet?.id, appointmentIdFromState, generalReservationType?.id]);

  const shouldRedirectToSelectPet = isNewRecord && !petId;
  const notFound = !isNewRecord && !!recordId && !isRecordLoading && isRecordError;

  return {
    isNewRecord,
    activeTab,
    setActiveTab,
    visitType,
    setVisitType,
    selectedPet: selectedPet ?? null,
    isPetLoading,
    shouldRedirectToSelectPet,
    notFound,
    handleBack,
    formAction,
    formState,
    isSaving: isSaving || isSavingTransition,
    isCreating,
    treatmentPlanItems,
    setTreatmentPlanItems,
    treatmentCompletedItems,
    setTreatmentCompletedItems,
    // 問診タブ
    chiefComplaint,
    setChiefComplaint,
    chiefComplaintTypeId,
    setChiefComplaintTypeId,
    treatmentPolicy,
    setTreatmentPolicy,
    // 診察/治療プランタブ（SOAPS）
    plan,
    setPlan,
    assessment,
    setAssessment,
    // 診断マスタ
    diagnosis1CategoryId,
    setDiagnosis1CategoryId,
    diagnosis1NameId,
    setDiagnosis1NameId,
    diagnosis2CategoryId,
    setDiagnosis2CategoryId,
    diagnosis2NameId,
    setDiagnosis2NameId,
    // 飼主割引率
    ownerDiscountRate,
    // 医療記録
    visitCount: existingRecord?.visitCount,
    // 担当医変更
    handleChangeDoctor,
    // 飼主変更
    pendingOwnerChange,
    requestOwnerChange,
    confirmOwnerChange,
    cancelOwnerChange,
    fieldErrors: manualErrors,
    // 次回来院推奨日
    nextVisitDate,
    handleNextVisitDateChange: setNextVisitDate,
    isNextVisitDateValid,
    handleNextVisitDateValidChange: setIsNextVisitDateValid,
    // 推奨理由: edit mode は existingRecord から、create mode は local state から
    recommendationReason: recordId
      ? (existingRecord?.recommendationReason ?? null)
      : createRecommendationReason,
    setRecommendationReason: setCreateRecommendationReason,
  };
}
