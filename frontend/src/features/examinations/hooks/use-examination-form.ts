import {
  useLayoutEffect,
  useTransition,
  useCallback,
  useActionState,
  useRef,
  useEffect,
} from "react";
import { useSearchParams } from "react-router";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { INITIAL_ACTION_STATE } from "@/types/form";
import {
  isPersistedConfirmedStatus,
  isPersistedResultsLocked,
} from "../lib/examination-lock";
import {
  createBlankExaminationForm,
  DENIED_MUTATION_PERMISSIONS,
  deriveExaminationLockFlags,
  isExaminationPatientChangeLocked,
  type ExaminationMutationPermissions,
} from "./use-examination-form-model";
import { useCreateExamination } from "../api/create-examination";
import { useUpdateExamination } from "../api/update-examination";
import { useDeleteExamination } from "../api/delete-examination";
import { useUnconfirmExamination } from "../api/unconfirm-examination";
import {
  createExaminationDeleteHandler,
  createExaminationUnconfirmHandler,
  runExaminationSave,
  useExaminationFormItems,
  useExaminationFormLoad,
  useExaminationFormOverrides,
  useExaminationFormPetSync,
} from "./use-examination-form-helpers";
import { isNonDisclosureReadStatus } from "@/lib/entity-read-result";

export function useExaminationForm(
  id?: string,
  medicalRecordIdParam?: string,
  permissions: Readonly<ExaminationMutationPermissions> = DENIED_MUTATION_PERMISSIONS,
) {
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const doctorId = searchParams.get("doctorId");
  const medicalRecordId =
    medicalRecordIdParam ?? searchParams.get("medicalRecordId") ?? "";
  const isEdit = !!id;
  const activeExaminationIDRef = useRef(id);
  useLayoutEffect(() => {
    activeExaminationIDRef.current = id;
  }, [id]);

  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;
  const load = useExaminationFormLoad(id, isEdit, petId);
  const { entityRead, existingExam, mutationPet, isPetLoading } = load;
  const entityReadRef = useRef(entityRead);
  useLayoutEffect(() => {
    entityReadRef.current = entityRead;
  }, [entityRead]);

  const createMutation = useCreateExamination();
  const updateMutation = useUpdateExamination();
  const deleteMutation = useDeleteExamination();
  const unconfirmMutation = useUnconfirmExamination();
  const { canCreate, canEdit, canDelete, canUnconfirm } = permissions;
  const permissionsRef = useRef(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit, canDelete, canUnconfirm };
  }, [canCreate, canDelete, canEdit, canUnconfirm]);
  const hasExplicitlyDeceasedPet =
    mutationPet?.status === "死亡" || selectedPets[0]?.status === "死亡";
  const hasExplicitlyDeceasedPetRef = useRef(hasExplicitlyDeceasedPet);
  useLayoutEffect(() => {
    hasExplicitlyDeceasedPetRef.current = hasExplicitlyDeceasedPet;
  }, [hasExplicitlyDeceasedPet]);
  const isMutationAllowed = useCallback(
    (action: keyof ExaminationMutationPermissions) =>
      permissionsRef.current[action] === true,
    [],
  );
  const isPetExplicitlyDeceased = useCallback(
    () => hasExplicitlyDeceasedPetRef.current === true,
    [],
  );

  const isPersistedResultsLockedRef = useRef(false);
  const isPersistedConfirmedRef = useRef(false);
  useLayoutEffect(() => {
    isPersistedConfirmedRef.current =
      isEdit && isPersistedConfirmedStatus(existingExam?.status);
    isPersistedResultsLockedRef.current =
      isEdit && isPersistedResultsLocked(
        existingExam?.status,
        existingExam?.currentRevisionVersion,
      );
  }, [isEdit, existingExam?.status, existingExam?.currentRevisionVersion]);

  const isPatientChangeLocked = isExaminationPatientChangeLocked(isEdit, canEdit, existingExam);
  const isPatientChangeLockedRef = useRef(isPatientChangeLocked);
  const existingPetIdRef = useRef(existingExam?.petId);
  useLayoutEffect(() => {
    isPatientChangeLockedRef.current = isPatientChangeLocked;
    existingPetIdRef.current = existingExam?.petId;
  }, [existingExam?.petId, isPatientChangeLocked]);

  const [isDeleteTransitionPending, startDeleteTransition] = useTransition();
  const {
    localOverrides,
    setFormData,
    manualFieldErrors,
    setManualFieldErrors,
  } = useExaminationFormOverrides(id);
  const formData = isEdit && existingExam
    ? { ...existingExam, ...localOverrides }
    : createBlankExaminationForm(doctorId, localOverrides);
  const activePet = selectedPets[0] ?? mutationPet;
  const formDataWithPet = activePet
    ? { ...formData, ownerName: activePet.ownerName, petName: activePet.name, petId: activePet.id }
    : formData;
  const formDataWithPetRef = useRef(formDataWithPet);
  useLayoutEffect(() => {
    formDataWithPetRef.current = formDataWithPet;
  });
  const activePetRef = useRef(activePet);
  useLayoutEffect(() => {
    activePetRef.current = activePet;
  }, [activePet]);

  const items = useExaminationFormItems({
    id,
    isEdit,
    existingItems: load.existingItems,
    existingItemsQuerySucceeded: load.existingItemsQuerySucceeded,
    currentTestTypeId: formData.testTypeId ?? "",
  });

  const [formState, formAction, isPending] = useActionState(
    async () => runExaminationSave({
      id, isEdit, medicalRecordId, formDataWithPetRef,
      formItemsRef: items.formItemsRef,
      itemsReadyForIDRef: items.itemsReadyForIDRef,
      activeExaminationIDRef, isPersistedConfirmedRef, isPersistedResultsLockedRef,
      isPatientChangeLockedRef, existingPetIdRef, entityReadRef, activePetRef,
      isPetExplicitlyDeceased, isMutationAllowed, updateMutation, createMutation,
    }),
    INITIAL_ACTION_STATE,
  );

  useExaminationFormPetSync({ isEdit, petId, mutationPet, isPetLoading, setSelectedPets });

  const { mutateAsync } = unconfirmMutation;
  const handleUnconfirm = useCallback((rawReason: string) => {
    return createExaminationUnconfirmHandler({
      isEdit, id, isMutationAllowed,
      isPersistedConfirmed: () => isPersistedConfirmedRef.current,
      unconfirm: (vars) => mutateAsync(vars),
    })(rawReason);
  }, [id, isEdit, isMutationAllowed, mutateAsync]);
  const { mutate } = deleteMutation;
  const handleDelete = useCallback((onSuccess?: () => void) => {
    createExaminationDeleteHandler({
      isEdit, id, isMutationAllowed, isPetExplicitlyDeceased, startDeleteTransition,
      isResultsLocked: () => isPersistedResultsLockedRef.current,
      deleteExamination: (examinationId, opts) => mutate(examinationId, opts),
    })(onSuccess);
  }, [isEdit, id, isMutationAllowed, isPetExplicitlyDeceased, mutate, startDeleteTransition]);

  useEffect(() => {
    setManualFieldErrors(formState.fieldErrors || {});
  }, [formState.fieldErrors, formState.timestamp, setManualFieldErrors]);

  const lockFlags = deriveExaminationLockFlags(isEdit, existingExam);
  return {
    formData: formDataWithPet,
    setFormData,
    petSelection,
    // FE-RC-002: callback 側（runExaminationSave 等）の拒否と対になる render 側の防壁。
    // SubmitButton 非表示・PatientInfoCard の死亡表示に使う（二重防壁）。
    isPetDeceased: hasExplicitlyDeceasedPet,
    formAction,
    formState,
    fieldErrors: manualFieldErrors,
    handleDelete,
    isEdit,
    entityRead,
    isReadLoading: entityRead.status === "loading",
    isReadNotFound: isNonDisclosureReadStatus(entityRead.status),
    isReadError: entityRead.status === "error",
    retryRead: entityRead.status === "error" ? entityRead.retry : undefined,
    isSaving: isPending,
    isDeleting: deleteMutation.isPending || isDeleteTransitionPending,
    handleUnconfirm,
    isUnconfirming: unconfirmMutation.isPending,
    ...lockFlags,
    isPatientChangeLocked,
    formItems: items.visibleFormItems,
    setInspectionValue: items.setInspectionValue,
    addManualItem: items.addManualItem,
    removeItem: items.removeItem,
    setItemName: items.setItemName,
  };
}
