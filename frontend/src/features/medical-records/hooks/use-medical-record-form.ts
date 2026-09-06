import { useCallback } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { usePermission } from "@/hooks/use-permission";
import { useCreateReservation } from "@/hooks/use-create-reservation";
import { useCreateMedicalRecord } from "../api/create-medical-record";
import { useUpdateMedicalRecord } from "../api/update-medical-record";
import { useUpdateInquiry } from "../api/inquiries";
import { useUpdateClinicalPlan } from "../api/clinical-plan";
import { useMedicalRecordOwnerChange } from "./use-medical-record-owner-change";
import { useMedicalRecordAutoCreate } from "./use-medical-record-auto-create";
import { useMedicalRecordDiagnosisState } from "./use-medical-record-diagnosis-state";
import { useMedicalRecordSaveAction } from "./use-medical-record-save-action";
import { useMedicalRecordQuickPatchActions } from "./use-medical-record-quick-patch-actions";
import { DEFAULT_CHIEF_COMPLAINT, DEFAULT_TREATMENT_POLICY } from "./use-medical-record-form-model";
import {
  createMedicalRecordBackHandler,
  selectCohabitingPets,
  toMedicalRecordFormResult,
  useMedicalRecordFormRead,
  useMedicalRecordFormUiState,
} from "./use-medical-record-form-helpers";

export { selectCohabitingPets };

export function useMedicalRecordForm(recordId?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isNewRecord = !recordId;
  const { canCreate, canEdit } = usePermission("medical-records");

  const ui = useMedicalRecordFormUiState(searchParams.get("tab"));
  const diagnosis = useMedicalRecordDiagnosisState();
  const read = useMedicalRecordFormRead({
    recordId,
    isNewRecord,
    petId,
    visitType: ui.visitType,
    location,
    searchParams,
    diagnosis,
    setVisitType: ui.setVisitType,
    setNextVisitDate: ui.setNextVisitDate,
  });

  const queryClient = useQueryClient();
  const createMutation = useCreateMedicalRecord();
  const createReservationMutation = useCreateReservation();
  const updateMutation = useUpdateMedicalRecord(read.recordClinicId);
  const updateInquiryMutation = useUpdateInquiry(recordId ?? "");
  const updateTreatmentPlanMutation = useUpdateClinicalPlan(recordId ?? "", read.recordClinicId);

  const { formState, formAction, isSaving } = useMedicalRecordSaveAction({
    recordId,
    activeTab: ui.activeTab,
    canEdit,
    isSelectedPetDeceased: read.selectedPet?.status === "死亡",
    isFinalized: read.isFinalized,
    isNextVisitDateValid: ui.isNextVisitDateValid,
    diagnosis1CategoryId: diagnosis.diagnosis1CategoryId,
    diagnosis1NameId: diagnosis.diagnosis1NameId,
    diagnosis2CategoryId: diagnosis.diagnosis2CategoryId,
    diagnosis2NameId: diagnosis.diagnosis2NameId,
    physicalExam: diagnosis.physicalExam,
    plan: diagnosis.plan,
    assessment: diagnosis.assessment,
    chiefComplaint: diagnosis.chiefComplaint,
    chiefComplaintDefault: DEFAULT_CHIEF_COMPLAINT,
    chiefComplaintTypeId: diagnosis.chiefComplaintTypeId,
    treatmentPolicy: diagnosis.treatmentPolicy,
    treatmentPolicyDefault: DEFAULT_TREATMENT_POLICY,
    nextVisitDate: ui.nextVisitDate,
    existingRecordVersion: read.existingRecord?.version,
    existingClinicalPlanVersion: read.clinicalPlanVersion,
    onClinicalPlanSaved: read.onClinicalPlanSaved,
    setManualErrors: ui.setManualErrors,
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
    existingRecordVersion: read.existingRecord?.version,
    visitType: ui.visitType,
    setVisitType: ui.setVisitType,
    nextVisitDate: ui.nextVisitDate,
    setNextVisitDate: ui.setNextVisitDate,
    queryClient,
    updateMutation,
    isSelectedPetDeceased: read.selectedPet?.status === "死亡",
  });

  const handleBack = useCallback(() => {
    createMedicalRecordBackHandler({
      from: location.state?.from,
      recordId,
      navigate,
    })();
  }, [location.state, navigate, recordId]);

  const ownerChange = useMedicalRecordOwnerChange({
    owner: read.owner,
    recordId,
    existingRecord: read.existingRecord,
    updateMutation,
    startSaveTransition,
    isSelectedPetDeceased: read.selectedPet?.status === "死亡",
  });

  const autoCreate = useMedicalRecordAutoCreate({
    isNewRecord,
    canCreate,
    selectedPet: read.selectedPet ?? undefined,
    hasAutoCreatedRef: ui.hasAutoCreatedRef,
    appointmentIdFromState: read.appointmentIdFromState,
    reusableAppointment: read.reusableAppointment,
    isReusableAppointmentLoading: read.isSameDayAppointmentsLoading,
    isReservationTypesLoading: read.isReservationTypesLoading,
    isReservationTypesError: read.isReservationTypesError,
    visitDateFromState: read.visitDateFromState,
    generalReservationType: read.generalReservationType,
    createReservationMutation,
    createMutation,
    startCreateTransition: ui.startCreateTransition,
    visitType: ui.visitType,
    createRecommendationReason: diagnosis.createRecommendationReason,
    navigate,
    tab: searchParams.get("tab"),
  });

  const { createRecommendationReason, setCreateRecommendationReason, ...diagnosisFields } =
    diagnosis;

  return toMedicalRecordFormResult({
    isNewRecord,
    petId,
    recordId,
    ui,
    read,
    diagnosisFields,
    createRecommendationReason,
    setCreateRecommendationReason,
    handleBack,
    formAction,
    formState,
    isSaving,
    isSavingTransition,
    handleChangeDoctor,
    handleVisitTypeChange,
    handleNextVisitDatePatch,
    handleChangeDate,
    handleFinalize,
    ownerChange,
    autoCreate,
  });
}
