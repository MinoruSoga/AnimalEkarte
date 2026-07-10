import { useState, useRef, useTransition, useCallback } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { usePermission } from "@/hooks/use-permission";
import { useGetPet } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useCreateReservation } from "@/hooks/use-create-reservation";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { paths } from "@/config/paths";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useCreateMedicalRecord } from "../api/create-medical-record";
import { useUpdateMedicalRecord } from "../api/update-medical-record";
import { useUpdateInquiry } from "../api/inquiries";
import { useUpdateClinicalPlan } from "../api/clinical-plan";
import type { TreatmentItem } from "../components/TreatmentTable";
import { useApplyMedicalRecord } from "./use-apply-medical-record";
import { useMedicalRecordManualErrors } from "./use-medical-record-manual-errors";
import { useMedicalRecordOwnerChange } from "./use-medical-record-owner-change";
import { useMedicalRecordAutoCreate } from "./use-medical-record-auto-create";
import { useMedicalRecordDiagnosisState } from "./use-medical-record-diagnosis-state";
import { useMedicalRecordSaveAction } from "./use-medical-record-save-action";
import { useMedicalRecordQuickPatchActions } from "./use-medical-record-quick-patch-actions";
import {
  DEFAULT_CHIEF_COMPLAINT,
  DEFAULT_TREATMENT_POLICY,
  findGeneralReservationType,
  formatJSTDate,
  normalizeAppointmentId,
  normalizeVisitDate,
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

  const hasAutoCreatedRef = useRef(false);
  const [treatmentPlanItems, setTreatmentPlanItems] = useState<TreatmentItem[]>([]);
  const [treatmentCompletedItems, setTreatmentCompletedItems] = useState<TreatmentItem[]>([]);

  // 問診/SOAPS/診断マスタ/推奨理由の状態
  const {
    chiefComplaint,
    setChiefComplaint,
    chiefComplaintTypeId,
    setChiefComplaintTypeId,
    treatmentPolicy,
    setTreatmentPolicy,
    plan,
    setPlan,
    assessment,
    setAssessment,
    diagnosis1CategoryId,
    setDiagnosis1CategoryId,
    diagnosis1NameId,
    setDiagnosis1NameId,
    diagnosis2CategoryId,
    setDiagnosis2CategoryId,
    diagnosis2NameId,
    setDiagnosis2NameId,
    createRecommendationReason,
    setCreateRecommendationReason,
  } = useMedicalRecordDiagnosisState();

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

  const { formState, formAction, isSaving } = useMedicalRecordSaveAction({
    recordId,
    activeTab,
    canEdit,
    isNextVisitDateValid,
    diagnosis1CategoryId,
    diagnosis1NameId,
    diagnosis2CategoryId,
    diagnosis2NameId,
    plan,
    assessment,
    chiefComplaint,
    chiefComplaintDefault: DEFAULT_CHIEF_COMPLAINT,
    chiefComplaintTypeId,
    treatmentPolicy,
    treatmentPolicyDefault: DEFAULT_TREATMENT_POLICY,
    nextVisitDate,
    existingRecordVersion: existingRecord?.version,
    setManualErrors,
    queryClient,
    updateInquiryMutation,
    updateTreatmentPlanMutation,
    updateMutation,
  });

  const {
    isSavingTransition,
    startSaveTransition,
    handleChangeDoctor,
    handleVisitTypeChange,
    handleNextVisitDatePatch,
    handleChangeDate,
  } = useMedicalRecordQuickPatchActions({
    recordId,
    existingRecordVersion: existingRecord?.version,
    visitType,
    setVisitType,
    nextVisitDate,
    setNextVisitDate,
    queryClient,
    updateMutation,
  });

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
