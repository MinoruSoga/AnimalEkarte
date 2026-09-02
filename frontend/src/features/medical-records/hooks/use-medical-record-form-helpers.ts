import { useState, useTransition, useRef, useMemo } from "react";
import type { Location } from "react-router";
import { useGetPet, useGetPets } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useGetClinicalPlan } from "../api/clinical-plan";
import type { TreatmentItem } from "../components/TreatmentTable";
import { useApplyMedicalRecord } from "./use-apply-medical-record";
import { useApplyClinicalPlan } from "./use-apply-clinical-plan";
import { useMedicalRecordManualErrors } from "./use-medical-record-manual-errors";
import { initialMedicalRecordTab } from "../routes/medical-record-form-model";
import {
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
import type { Reservation } from "@/lib/transforms/reservation";
import { isMedicalRecordFinalizedStatus } from "../lib/medical-record-lock";
import type { useMedicalRecordDiagnosisState } from "./use-medical-record-diagnosis-state";
import { paths } from "@/config/paths";

export function selectCohabitingPets(pets: Pet[], selectedPet: Pet): Pet[] {
  return pets.filter((pet) =>
    pet.ownerId === selectedPet.ownerId
    && pet.id !== selectedPet.id
    && pet.status !== "死亡"
  );
}

export function selectReusableGeneralAppointment(
  appointments: readonly Reservation[],
): Reservation | undefined {
  return appointments.find((appointment) =>
    appointment.category === "general" &&
    !["completed", "cancelled", "no_show"].includes(String(appointment.status))
  );
}

export function useMedicalRecordFormUiState(tabParam: string | null) {
  const [activeTab, setActiveTab] = useState(() =>
    initialMedicalRecordTab(tabParam),
  );
  const [visitType, setVisitType] = useState("再診");
  const [isCreating, startCreateTransition] = useTransition();
  const { manualErrors, setManualErrors } = useMedicalRecordManualErrors({ setActiveTab });
  const [nextVisitDate, setNextVisitDate] = useState("");
  const [isNextVisitDateValid, setIsNextVisitDateValid] = useState(true);
  const hasAutoCreatedRef = useRef(false);
  const [treatmentPlanItems, setTreatmentPlanItems] = useState<TreatmentItem[]>([]);
  const [treatmentCompletedItems, setTreatmentCompletedItems] = useState<TreatmentItem[]>([]);

  return {
    activeTab,
    setActiveTab,
    visitType,
    setVisitType,
    isCreating,
    startCreateTransition,
    manualErrors,
    setManualErrors,
    nextVisitDate,
    setNextVisitDate,
    isNextVisitDateValid,
    setIsNextVisitDateValid,
    hasAutoCreatedRef,
    treatmentPlanItems,
    setTreatmentPlanItems,
    treatmentCompletedItems,
    setTreatmentCompletedItems,
  };
}

type DiagnosisState = ReturnType<typeof useMedicalRecordDiagnosisState>;

export function useMedicalRecordFormRead(input: {
  recordId?: string;
  isNewRecord: boolean;
  petId: string | null;
  visitType: string;
  location: Location;
  searchParams: URLSearchParams;
  diagnosis: DiagnosisState;
  setVisitType: (value: string) => void;
  setNextVisitDate: (value: string) => void;
}) {
  const {
    recordId,
    isNewRecord,
    petId,
    visitType,
    location,
    searchParams,
    diagnosis,
    setVisitType,
    setNextVisitDate,
  } = input;

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
    setChiefComplaint: diagnosis.setChiefComplaint,
    setChiefComplaintTypeId: diagnosis.setChiefComplaintTypeId,
    setTreatmentPolicy: diagnosis.setTreatmentPolicy,
    setPlan: diagnosis.setPlan,
    setAssessment: diagnosis.setAssessment,
    setVisitType,
    setNextVisitDate,
    setDiagnosis1CategoryId: diagnosis.setDiagnosis1CategoryId,
    setDiagnosis1NameId: diagnosis.setDiagnosis1NameId,
    setDiagnosis2CategoryId: diagnosis.setDiagnosis2CategoryId,
    setDiagnosis2NameId: diagnosis.setDiagnosis2NameId,
  });

  const resolvedPetId = isNewRecord ? (petId ?? "") : (existingRecord?.petId ?? "");
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

  const resolvedOwnerId = selectedPet?.ownerId ?? "";
  const { data: owner } = useGetOwner(resolvedOwnerId);
  const ownerDiscountRate = owner?.discountRate ?? 0;
  const appointmentIdFromState = normalizeAppointmentId(location.state?.appointmentId)
    ?? normalizeAppointmentId(searchParams.get("appointmentId"));
  const visitDateFromState = normalizeVisitDate(location.state?.visitDate)
    ?? normalizeVisitDate(searchParams.get("visitDate"));
  const {
    data: reservationTypeGroups,
    isLoading: isReservationTypesLoading,
    isError: isReservationTypesError,
  } = useGetReservationTypesGrouped();
  const generalReservationType = findGeneralReservationType(reservationTypeGroups, visitType);
  const appointmentLookupDate = visitDateFromState ?? formatJSTDate(new Date());
  const { data: sameDayAppointments = [], isLoading: isSameDayAppointmentsLoading } = useGetReservations({
    date: appointmentLookupDate,
    petId: resolvedPetId,
    enabled: isNewRecord && !appointmentIdFromState && resolvedPetId !== "",
  });
  const reusableAppointment = selectReusableGeneralAppointment(sameDayAppointments);
  const recordClinicId = existingRecord?.clinicId;
  const { data: clinicalPlan } = useGetClinicalPlan(recordId ?? "", recordClinicId);

  useApplyClinicalPlan({
    clinicalPlan,
    setPhysicalExam: diagnosis.setPhysicalExam,
    setPlan: diagnosis.setPlan,
    setAssessment: diagnosis.setAssessment,
    setDiagnosis1CategoryId: diagnosis.setDiagnosis1CategoryId,
    setDiagnosis1NameId: diagnosis.setDiagnosis1NameId,
    setDiagnosis2CategoryId: diagnosis.setDiagnosis2CategoryId,
    setDiagnosis2NameId: diagnosis.setDiagnosis2NameId,
  });

  return {
    existingRecord,
    isReadLoading,
    isReadNotFound,
    isReadError,
    retryRead,
    isFinalized,
    notFound: isReadNotFound,
    selectedPet: selectedPet ?? null,
    isPetLoading,
    cohabitingPets,
    owner,
    ownerDiscountRate,
    appointmentIdFromState,
    visitDateFromState,
    isReservationTypesLoading,
    isReservationTypesError,
    generalReservationType,
    isSameDayAppointmentsLoading,
    reusableAppointment,
    recordClinicId,
    clinicalPlan,
    resolvedPetId,
  };
}

export function toMedicalRecordFormResult(input: {
  isNewRecord: boolean;
  petId: string | null;
  recordId?: string;
  ui: ReturnType<typeof useMedicalRecordFormUiState>;
  read: ReturnType<typeof useMedicalRecordFormRead>;
  diagnosisFields: Omit<
    ReturnType<typeof useMedicalRecordDiagnosisState>,
    "createRecommendationReason" | "setCreateRecommendationReason"
  >;
  createRecommendationReason: ReturnType<typeof useMedicalRecordDiagnosisState>["createRecommendationReason"];
  setCreateRecommendationReason: ReturnType<typeof useMedicalRecordDiagnosisState>["setCreateRecommendationReason"];
  handleBack: () => void;
  formAction: (payload: FormData) => void;
  formState: unknown;
  isSaving: boolean;
  isSavingTransition: boolean;
  handleChangeDoctor: unknown;
  handleVisitTypeChange: unknown;
  handleNextVisitDatePatch: unknown;
  handleChangeDate: unknown;
  handleFinalize: unknown;
  ownerChange: object;
  autoCreate: { failurePhase: unknown; retry: unknown };
}) {
  return {
    isNewRecord: input.isNewRecord,
    activeTab: input.ui.activeTab,
    setActiveTab: input.ui.setActiveTab,
    visitType: input.ui.visitType,
    setVisitType: input.ui.setVisitType,
    selectedPet: input.read.selectedPet,
    cohabitingPets: input.read.cohabitingPets,
    isPetLoading: input.read.isPetLoading,
    shouldRedirectToSelectPet: input.isNewRecord && !input.petId,
    notFound: input.read.notFound,
    isReadLoading: input.read.isReadLoading,
    isReadNotFound: input.read.isReadNotFound,
    isReadError: input.read.isReadError,
    retryRead: input.read.retryRead,
    handleBack: input.handleBack,
    formAction: input.formAction,
    formState: input.formState,
    isSaving: input.isSaving || input.isSavingTransition,
    isFinalized: input.read.isFinalized,
    isCreating: input.ui.isCreating,
    autoCreateFailurePhase: input.autoCreate.failurePhase,
    retryAutoCreate: input.autoCreate.retry,
    treatmentPlanItems: input.ui.treatmentPlanItems,
    setTreatmentPlanItems: input.ui.setTreatmentPlanItems,
    treatmentCompletedItems: input.ui.treatmentCompletedItems,
    setTreatmentCompletedItems: input.ui.setTreatmentCompletedItems,
    ...input.diagnosisFields,
    ownerDiscountRate: input.read.ownerDiscountRate,
    visitCount: input.read.existingRecord?.visitCount,
    handleChangeDoctor: input.handleChangeDoctor,
    handleVisitTypeChange: input.handleVisitTypeChange,
    recordDate: input.read.existingRecord?.date,
    handleChangeDate: input.handleChangeDate,
    handleFinalize: input.handleFinalize,
    isFinalizeSaving: input.isSavingTransition,
    ...input.ownerChange,
    fieldErrors: input.ui.manualErrors,
    nextVisitDate: input.ui.nextVisitDate,
    handleNextVisitDateChange: input.ui.setNextVisitDate,
    handleNextVisitDatePatch: input.handleNextVisitDatePatch,
    isNextVisitDateValid: input.ui.isNextVisitDateValid,
    handleNextVisitDateValidChange: input.ui.setIsNextVisitDateValid,
    recommendationReason: input.recordId
      ? (input.read.existingRecord?.recommendationReason ?? null)
      : input.createRecommendationReason,
    setRecommendationReason: input.setCreateRecommendationReason,
  };
}

export function createMedicalRecordBackHandler(input: {
  from: unknown;
  recordId?: string;
  navigate: (to: string) => void;
}): () => void {
  return () => {
    if (input.from) {
      input.navigate(input.from as string);
      return;
    }
    if (!input.recordId) {
      input.navigate(paths.medicalRecords.selectPet.getHref());
    } else {
      input.navigate(paths.medicalRecords.getHref());
    }
  };
}
