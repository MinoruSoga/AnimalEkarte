import { useState, useLayoutEffect, useCallback, useMemo, useActionState, useRef } from "react";
import { useGetPet } from "@/hooks/use-pet";
import { useGetAllVaccinesMaster } from "@/hooks/use-treatment-master";
import { usePetSelection } from "@/hooks/use-pet-selection";
import {
  isNonDisclosureReadStatus,
  resolveEntityReadResult,
  type EntityReadResult,
} from "@/lib/entity-read-result";
import { useGetVaccination } from "../api/get-vaccination";
import { useCreateVaccination } from "../api/create-vaccination";
import { useUpdateVaccination } from "../api/update-vaccination";
import { useDeleteVaccination } from "../api/delete-vaccination";
import type { VaccinationRecord } from "@/types";
import {
  calculateNextDate,
  DENIED_MUTATION_PERMISSIONS,
  mergeVaccinationFormData,
  type VaccinationFormState,
  type VaccinationMutationPermissions,
} from "./use-vaccination-form-model";
import {
  createVaccinationDeleteHandler,
  runVaccinationSave,
  useVaccinationFormFields,
  useVaccinationFormPetSync,
  useVaccinationHistoryFilter,
  useVaccinationRoutePetId,
} from "./use-vaccination-form-helpers";

export { calculateNextDate };

const EMPTY_VACCINES_MASTER: NonNullable<ReturnType<typeof useGetAllVaccinesMaster>["data"]> = [];

export function useVaccinationForm(
  id?: string,
  permissions: Readonly<VaccinationMutationPermissions> = DENIED_MUTATION_PERMISSIONS,
) {
  const petId = useVaccinationRoutePetId();
  const isEdit = !!id;
  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  const { data: vaccinesMaster = EMPTY_VACCINES_MASTER } = useGetAllVaccinesMaster();
  const vaccineOptions = useMemo(
    () => vaccinesMaster.flatMap((v) => (v.isActive ? [{ value: v.id, label: v.name }] : [])),
    [vaccinesMaster],
  );

  const {
    data: vaccinationData,
    isLoading: isVaccinationLoading,
    isError: isVaccinationError,
    error: vaccinationError,
    refetch: refetchVaccination,
  } = useGetVaccination(id ?? "");
  const entityRead: EntityReadResult<VaccinationRecord> = resolveEntityReadResult({
    id: isEdit ? id : undefined,
    data: vaccinationData,
    isLoading: isVaccinationLoading,
    isError: isVaccinationError,
    error: vaccinationError,
    refetch: refetchVaccination,
  });
  const existingVaccination =
    entityRead.status === "found" ? entityRead.data : undefined;
  const entityReadRef = useRef(entityRead);
  useLayoutEffect(() => {
    entityReadRef.current = entityRead;
  }, [entityRead]);
  const editPetId = isEdit ? (existingVaccination?.petId ?? "") : "";
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");
  const { data: petFromEdit } = useGetPet(editPetId);
  const selectedPetRef = useRef(selectedPets[0]);
  const queryPetRef = useRef(petFromQuery);
  const editPetRef = useRef(petFromEdit);
  useLayoutEffect(() => {
    selectedPetRef.current = selectedPets[0];
    queryPetRef.current = petFromQuery;
  }, [petFromQuery, selectedPets]);
  useLayoutEffect(() => {
    editPetRef.current = petFromEdit;
  }, [petFromEdit]);
  const createMutation = useCreateVaccination();
  const updateMutation = useUpdateVaccination();
  const deleteMutation = useDeleteVaccination();
  const { canCreate, canEdit, canDelete } = permissions;
  const permissionsRef = useRef(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit, canDelete };
  }, [canCreate, canDelete, canEdit]);
  const isMutationAllowed = useCallback(
    (action: keyof VaccinationMutationPermissions) => permissionsRef.current[action] === true,
    [],
  );

  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [localOverrides, setLocalOverrides] = useState<Partial<VaccinationFormState>>({});
  const formData = mergeVaccinationFormData(isEdit, existingVaccination, localOverrides);
  const formDataRef = useRef(formData);
  useLayoutEffect(() => {
    formDataRef.current = formData;
  }, [formData]);

  const [formState, formAction, isPending] = useActionState(
    async () =>
      runVaccinationSave({
        isEdit,
        id,
        petId,
        formData,
        entityReadRef,
        selectedPetRef,
        queryPetRef,
        editPetRef,
        isMutationAllowed,
        setFieldErrors,
        updateMutation,
        createMutation,
      }),
    { success: false, timestamp: 0 },
  );

  useVaccinationFormPetSync({
    isEdit,
    petId,
    petFromQuery,
    petFromEdit,
    isPetLoading,
    setSelectedPets,
  });

  const historyFilter = useVaccinationHistoryFilter();
  const { form } = useVaccinationFormFields({
    formData,
    formDataRef,
    vaccinesMaster,
    vaccineOptions,
    doctorName: existingVaccination?.doctor ?? "",
    setLocalOverrides,
  });

  const { mutate: deleteVaccinationFn } = deleteMutation;
  const handleDelete = useCallback(
    createVaccinationDeleteHandler({
      isEdit,
      id,
      isMutationAllowed,
      isEditPetDeceased: () => editPetRef.current?.status === "死亡",
      deleteVaccination: deleteVaccinationFn,
    }),
    [isEdit, id, isMutationAllowed, deleteVaccinationFn],
  );

  return {
    isEdit,
    entityRead,
    isReadLoading: entityRead.status === "loading",
    isReadNotFound: isNonDisclosureReadStatus(entityRead.status),
    isReadError: entityRead.status === "error",
    retryRead: entityRead.status === "error" ? entityRead.retry : undefined,
    petSelection,
    form,
    historyFilter,
    formAction,
    formState,
    isSaving: isPending,
    fieldErrors,
    handleDelete,
    isDeleting: deleteMutation.isPending,
  };
}
