import { useState, useRef, useTransition, useCallback, useMemo } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { usePermission } from "@/hooks/use-permission";
import { useGetPet, useGetPets } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useCreateReservation } from "@/hooks/use-create-reservation";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { paths } from "@/config/paths";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useCreateMedicalRecord } from "../api/create-medical-record";
import { useUpdateMedicalRecord } from "../api/update-medical-record";
import { useUpdateInquiry } from "../api/inquiries";
import { useUpdateClinicalPlan, useGetClinicalPlan } from "../api/clinical-plan";
import type { TreatmentItem } from "../components/TreatmentTable";
import { useApplyMedicalRecord } from "./use-apply-medical-record";
import { useApplyClinicalPlan } from "./use-apply-clinical-plan";
import { useMedicalRecordManualErrors } from "./use-medical-record-manual-errors";
import { useMedicalRecordOwnerChange } from "./use-medical-record-owner-change";
import { useMedicalRecordAutoCreate } from "./use-medical-record-auto-create";
import { useMedicalRecordDiagnosisState } from "./use-medical-record-diagnosis-state";
import { useMedicalRecordSaveAction } from "./use-medical-record-save-action";
import { useMedicalRecordQuickPatchActions } from "./use-medical-record-quick-patch-actions";
import { initialMedicalRecordTab } from "../routes/medical-record-form-model";
import {
  DEFAULT_CHIEF_COMPLAINT,
  DEFAULT_TREATMENT_POLICY,
  findGeneralReservationType,
  formatJSTDate,
  normalizeAppointmentId,
  normalizeVisitDate,
} from "./use-medical-record-form-model";
import {
  isNonDisclosureReadStatus,
  resolveEntityReadResult,
  type EntityReadResult,
} from "@/lib/entity-read-result";
import type { MedicalRecord } from "../api/transforms";
import type { Pet } from "@/types";
import { isMedicalRecordFinalizedStatus } from "../lib/medical-record-lock";

export function selectCohabitingPets(pets: Pet[], selectedPet: Pet): Pet[] {
  return pets.filter((pet) =>
    pet.ownerId === selectedPet.ownerId
    && pet.id !== selectedPet.id
    && pet.status !== "死亡"
  );
}

export function useMedicalRecordForm(recordId?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isNewRecord = !recordId;
  const { canCreate, canEdit } = usePermission("medical-records");

  const [activeTab, setActiveTab] = useState(() =>
    initialMedicalRecordTab(searchParams.get("tab")),
  );
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
    physicalExam,
    setPhysicalExam,
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
  // BUG-017: classify read failures; never fold missing ID into blank page via selectedPet=null
  const {
    data: existingRecordData,
    isError: isRecordError,
    isLoading: isRecordLoading,
    error: recordError,
    refetch: refetchRecord,
  } = useGetMedicalRecord(recordId ?? "");
  const entityRead: EntityReadResult<MedicalRecord> = resolveEntityReadResult({
    id: isNewRecord ? undefined : recordId,
    data: existingRecordData,
    isLoading: isRecordLoading,
    isError: isRecordError,
    error: recordError,
    refetch: refetchRecord,
  });
  const existingRecord =
    entityRead.status === "found" ? entityRead.data : undefined;
  const isReadLoading = !isNewRecord && entityRead.status === "loading";
  const isReadNotFound =
    !isNewRecord && isNonDisclosureReadStatus(entityRead.status);
  const isReadError = !isNewRecord && entityRead.status === "error";
  const retryRead =
    entityRead.status === "error" ? entityRead.retry : undefined;
  const isFinalized = isMedicalRecordFinalizedStatus(existingRecord?.status);

  useApplyMedicalRecord({
    existingRecord,
    setChiefComplaint,
    setChiefComplaintTypeId,
    setTreatmentPolicy,
    setPlan,
    setAssessment,
    setVisitType,
    setNextVisitDate,
    setDiagnosis1CategoryId,
    setDiagnosis1NameId,
    setDiagnosis2CategoryId,
    setDiagnosis2NameId,
  });

  // petIdを決定: 新規作成時はURLパラメータ、編集時はカルテのpetId
  const resolvedPetId = isNewRecord ? (petId ?? "") : (existingRecord?.petId ?? "");

  // Petデータを取得
  const { data: selectedPet, isLoading: isPetLoading } = useGetPet(resolvedPetId);
  const cohabitingOwnerId = !isNewRecord ? selectedPet?.ownerId : undefined;
  const { data: ownerPets = [] } = useGetPets(
    cohabitingOwnerId,
    { includeDeceased: true },
    { enabled: Boolean(cohabitingOwnerId) },
  );
  const cohabitingPets = useMemo(
    () => !isNewRecord && selectedPet
      ? selectCohabitingPets(ownerPets, selectedPet)
      : [],
    [isNewRecord, ownerPets, selectedPet],
  );

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

  // P2-15: 拠点横断で開いたカルテ（record.clinicId）を保存する場合、グローバル選択クリニックではなく
  // レコード自身の clinicId を X-Clinic-ID として送る必要がある。
  const recordClinicId = existingRecord?.clinicId;

  const queryClient = useQueryClient();
  const createMutation = useCreateMedicalRecord();
  const createReservationMutation = useCreateReservation();
  const updateMutation = useUpdateMedicalRecord(recordClinicId);
  const updateInquiryMutation = useUpdateInquiry(recordId ?? "");
  const updateTreatmentPlanMutation = useUpdateClinicalPlan(recordId ?? "", recordClinicId);
  // BUG-416③: clinical_plan の楽観ロック用に現在バージョンを取得する。
  // medical_record GET レスポンスに clinical_plan は同梱されないため、
  // useGetClinicalPlan が version の唯一の取得元（TanStack Query が
  // クエリキーで dedupe するため、MedicalRecordForm.tsx 側の呼び出しと
  // 追加ネットワークリクエストにはならない）。
  // BUG-010: 同データから 3欄 + 診断マスタも hydrate する（detail wire には載らない）。
  const { data: clinicalPlan } = useGetClinicalPlan(recordId ?? "", recordClinicId);

  useApplyClinicalPlan({
    clinicalPlan,
    setPhysicalExam,
    setPlan,
    setAssessment,
    setDiagnosis1CategoryId,
    setDiagnosis1NameId,
    setDiagnosis2CategoryId,
    setDiagnosis2NameId,
  });

  const { formState, formAction, isSaving } = useMedicalRecordSaveAction({
    recordId,
    activeTab,
    canEdit,
    isSelectedPetDeceased: selectedPet?.status === "死亡",
    isFinalized,
    isNextVisitDateValid,
    diagnosis1CategoryId,
    diagnosis1NameId,
    diagnosis2CategoryId,
    diagnosis2NameId,
    physicalExam,
    plan,
    assessment,
    chiefComplaint,
    chiefComplaintDefault: DEFAULT_CHIEF_COMPLAINT,
    chiefComplaintTypeId,
    treatmentPolicy,
    treatmentPolicyDefault: DEFAULT_TREATMENT_POLICY,
    nextVisitDate,
    existingRecordVersion: existingRecord?.version,
    existingClinicalPlanVersion: clinicalPlan?.version,
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
    handleFinalize,
  } = useMedicalRecordQuickPatchActions({
    recordId,
    existingRecordVersion: existingRecord?.version,
    visitType,
    setVisitType,
    nextVisitDate,
    setNextVisitDate,
    queryClient,
    updateMutation,
    isSelectedPetDeceased: selectedPet?.status === "死亡",
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
    isSelectedPetDeceased: selectedPet?.status === "死亡",
  });

  // 新規作成時: ページ表示と同時にカルテを自動作成
  const {
    failurePhase: autoCreateFailurePhase,
    retry: retryAutoCreate,
  } = useMedicalRecordAutoCreate({
    isNewRecord,
    canCreate,
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
    tab: searchParams.get("tab"),
  });

  const shouldRedirectToSelectPet = isNewRecord && !petId;
  // backward-compat alias for MedicalRecordForm / tests (BUG-017 non-disclosure UI)
  const notFound = isReadNotFound;

  return {
    isNewRecord,
    activeTab,
    setActiveTab,
    visitType,
    setVisitType,
    selectedPet: selectedPet ?? null,
    cohabitingPets,
    isPetLoading,
    shouldRedirectToSelectPet,
    notFound,
    isReadLoading,
    isReadNotFound,
    isReadError,
    retryRead,
    handleBack,
    formAction,
    formState,
    isSaving: isSaving || isSavingTransition,
    isFinalized,
    isCreating,
    autoCreateFailurePhase,
    retryAutoCreate,
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
    // 診察/治療プランタブ（clinical_plan 3欄）
    physicalExam,
    setPhysicalExam,
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
    // カルテ確定（SPEC-GAP）
    handleFinalize,
    isFinalizeSaving: isSavingTransition,
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
