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
import { useCreateReservation, useGetReservations } from "@/features/reservations";
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
import { useMedicalRecordManualErrors } from "./use-medical-record-manual-errors";
import { useMedicalRecordOwnerChange } from "./use-medical-record-owner-change";
import { useMedicalRecordAutoCreate } from "./use-medical-record-auto-create";
import {
  DEFAULT_ASSESSMENT,
  DEFAULT_CHIEF_COMPLAINT,
  DEFAULT_PLAN,
  DEFAULT_TREATMENT_POLICY,
  findGeneralReservationType,
  formatJSTDate,
  normalizeAppointmentId,
  normalizeVisitDate,
  toVisitTypeValue,
} from "./use-medical-record-form-model";

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

  // 編集モード: カルテからpetIdを取得
  const { data: existingRecord, isError: isRecordError, isLoading: isRecordLoading } = useGetMedicalRecord(recordId ?? "");

  useApplyMedicalRecord({
    existingRecord,
    setChiefComplaint,
    setTreatmentPolicy,
    setPlan,
    setAssessment,
    setVisitType,
    setNextVisitDate,
  });

  // petIdを決定: 新規作成時はURLパラメータ、編集時はカルテのpetId
  const resolvedPetId = isNewRecord ? (petId ?? "") : (existingRecord?.petId ?? "");

  // Petデータを取得
  const { data: selectedPet, isLoading: isPetLoading } = useGetPet(resolvedPetId);

  // Ownerデータを取得（飼主割引率用）
  const resolvedOwnerId = selectedPet?.ownerId ?? "";
  const { data: owner } = useGetOwner(resolvedOwnerId);
  const ownerDiscountRate = owner?.discountRate ?? 0;
  const appointmentIdFromState = normalizeAppointmentId(location.state?.appointmentId)
    ?? normalizeAppointmentId(searchParams.get("appointmentId"));
  const visitDateFromState = normalizeVisitDate(location.state?.visitDate)
    ?? normalizeVisitDate(searchParams.get("visitDate"));
  const { data: reservationTypeGroups } = useGetReservationTypesGrouped();
  const generalReservationType = findGeneralReservationType(reservationTypeGroups, visitType);
  const appointmentLookupDate = visitDateFromState ?? formatJSTDate(new Date());
  const { data: sameDayAppointments = [], isLoading: isSameDayAppointmentsLoading } = useGetReservations({
    date: appointmentLookupDate,
    petId: resolvedPetId,
    enabled: isNewRecord && !appointmentIdFromState && resolvedPetId !== "",
  });
  const reusableAppointment = sameDayAppointments.find((appointment) =>
    appointment.category === "general" &&
    !["completed", "cancelled", "no_show"].includes(String(appointment.status))
  );

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
            // 次回来院推奨日を更新（空欄 = クリア、値あり = 設定）
            await updateMutation.mutateAsync({
              id: recordId as string,
              req: {
                next_visit_recommended_date: nextVisitDate, // "" はBEでNULLクリア
                version: existingRecord?.version,
              } as UpdateMedicalRecordRequest,
            });
            break;
          }

          default:
            break;
        }

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

  // 来院種別変更ハンドラ（即時PATCH）
  // existingRecord?.version のみ参照するため object 全体を dep に含めない (OCC versioning)
  // eslint-disable-next-line react-hooks/preserve-manual-memoization
  const handleVisitTypeChange = useCallback((newVisitType: string) => {
    const prevVisitType = visitType;
    setVisitType(newVisitType);
    if (!recordId) return; // 新規作成時はローカルstateのみ
    startSaveTransition(async () => {
      try {
        await updateMutation.mutateAsync({
          id: recordId,
          req: {
            visit_type: toVisitTypeValue(newVisitType),
            version: existingRecord?.version,
          } as UpdateMedicalRecordRequest,
        });
        toast.success(`来院種別を ${newVisitType} に変更しました`);
      } catch (error) {
        setVisitType(prevVisitType); // H-1: rollback on PATCH failure
        handleApiError(error, "来院種別変更");
      }
    });
  }, [visitType, recordId, existingRecord?.version, updateMutation, startSaveTransition]);

  // 次回予定変更ハンドラ（ヘッダー NextVisitButton 用・即時PATCH）
  // eslint-disable-next-line react-hooks/preserve-manual-memoization
  const handleNextVisitDatePatch = useCallback((newDate: string) => {
    const prev = nextVisitDate;
    setNextVisitDate(newDate);
    if (!recordId) return;
    startSaveTransition(async () => {
      try {
        await updateMutation.mutateAsync({
          id: recordId,
          req: {
            next_visit_recommended_date: newDate, // "" = クリア
            version: existingRecord?.version,
          } as UpdateMedicalRecordRequest,
        });
        await queryClient.invalidateQueries({ queryKey: ["medical-record", recordId] });
        toast.success(newDate ? `次回予定を ${newDate} に設定しました` : "次回予定をクリアしました");
      } catch (error) {
        setNextVisitDate(prev); // rollback
        handleApiError(error, "次回予定変更");
      }
    });
  }, [nextVisitDate, recordId, existingRecord?.version, updateMutation, queryClient, startSaveTransition]);

  // 診察日変更ハンドラ
  // existingRecord?.version のみ参照するため object 全体を dep に含めない (OCC versioning)
  // eslint-disable-next-line react-hooks/preserve-manual-memoization
  const handleChangeDate = useCallback((newDate: string) => {
    if (!recordId) return;
    startSaveTransition(async () => {
      try {
        await updateMutation.mutateAsync({
          id: recordId,
          req: {
            date: `${newDate}T00:00:00+09:00`,
            version: existingRecord?.version,
          } as UpdateMedicalRecordRequest,
        });
        await queryClient.invalidateQueries({ queryKey: ["medical-record", recordId] });
        toast.success(`診察日を ${newDate} に変更しました`);
      } catch (error) {
        handleApiError(error, "診察日変更");
      }
    });
  }, [recordId, existingRecord?.version, updateMutation, queryClient, startSaveTransition]);

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
  useMedicalRecordAutoCreate({
    isNewRecord,
    selectedPet,
    hasAutoCreatedRef,
    appointmentIdFromState,
    reusableAppointment,
    isReusableAppointmentLoading: isSameDayAppointmentsLoading,
    visitDateFromState,
    generalReservationType,
    createReservationMutation,
    createMutation,
    startCreateTransition,
    visitType,
    createRecommendationReason,
    navigate,
  });

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
    // 来院種別変更
    handleVisitTypeChange,
    // 診察日
    recordDate: existingRecord?.date,
    handleChangeDate,
    // 飼主変更
    pendingOwnerChange,
    requestOwnerChange,
    confirmOwnerChange,
    cancelOwnerChange,
    fieldErrors: manualErrors,
    // 次回来院推奨日
    nextVisitDate,
    handleNextVisitDateChange: setNextVisitDate,
    handleNextVisitDatePatch,
    isNextVisitDateValid,
    handleNextVisitDateValidChange: setIsNextVisitDateValid,
    // 推奨理由: edit mode は existingRecord から、create mode は local state から
    recommendationReason: recordId
      ? (existingRecord?.recommendationReason ?? null)
      : createRecommendationReason,
    setRecommendationReason: setCreateRecommendationReason,
  };
}
