import { memo, useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router";

import { paths } from "@/config/paths";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { useGetVaccinations } from "../api/get-vaccinations";
import { useVaccinationForm } from "../hooks/use-vaccination-form";
import { usePermission } from "@/hooks/use-permission";
import {
  VACCINATION_FIELD_ID_MAP,
  VACCINATION_PRIORITY_FIELDS,
  filterVaccinationHistory,
  resolveVaccinationFormGate,
} from "./vaccination-form-model";
import { VaccinationFormBody, VaccinationFormStatusView } from "./VaccinationFormPagePanels";

export const VaccinationForm = memo(function VaccinationForm() {
  const navigate = useNavigate();
  const { id } = useParams();
  const { canEdit, canCreate, canDelete } = usePermission("vaccinations");

  const {
    isEdit,
    isReadLoading,
    isEditPetReady,
    isReadNotFound,
    isReadError,
    retryRead,
    petSelection,
    form,
    formAction,
    formState,
    isSaving,
    fieldErrors,
    handleDelete,
    isDeleting,
    historyFilter,
  } = useVaccinationForm(id, { canCreate, canEdit, canDelete });

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  useEffect(() => {
    const errorFields = Object.keys(formState.fieldErrors || {});
    if (errorFields.length === 0) return;

    const firstError =
      VACCINATION_PRIORITY_FIELDS.find((field) => errorFields.includes(field)) || errorFields[0];
    const targetId = VACCINATION_FIELD_ID_MAP[firstError] || firstError;

    const element = document.getElementById(targetId);
    if (element) {
      element.focus();
      element.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }, [formState.fieldErrors, formState.timestamp]);

  useEffect(() => {
    if (formState.success) {
      markClean();
      navigate(paths.vaccinations.getHref());
    }
  }, [formState.success, formState.timestamp, navigate, markClean]);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];
  const isPetDeceased = selectedPet?.status === "死亡";
  const canSubmit = (id ? canEdit && isEditPetReady : canCreate) && !isPetDeceased;
  const allowDelete = canDelete === true && isEditPetReady;
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  const handleBack = useCallback(() => {
    navigate(paths.vaccinations.getHref());
  }, [navigate]);

  const historyPetId = selectedPet?.id;
  const { data: petVaccinations = [] } = useGetVaccinations({
    petId: historyPetId ?? "",
  });
  const { historySearchTerm, filterStartDate, filterEndDate, sortOrder } = historyFilter;
  const petHistory = useMemo(
    () =>
      filterVaccinationHistory({
        records: petVaccinations,
        historyPetId,
        excludeId: id,
        historySearchTerm,
        filterStartDate,
        filterEndDate,
        sortOrder,
      }),
    [
      petVaccinations,
      historyPetId,
      id,
      historySearchTerm,
      filterStartDate,
      filterEndDate,
      sortOrder,
    ],
  );

  const gate = resolveVaccinationFormGate({
    hasSelectedPet: Boolean(selectedPet),
    isEdit,
    isReadLoading,
    isReadNotFound,
    isReadError,
    retryRead,
  });
  if (gate) {
    return <VaccinationFormStatusView gate={gate} onBack={handleBack} />;
  }

  return (
    <VaccinationFormBody
      isEdit={isEdit}
      canSubmit={canSubmit === true}
      canDelete={allowDelete}
      isPetDeceased={isPetDeceased}
      isDeleting={isDeleting}
      isDirty={isDirty}
      isSaving={isSaving}
      selectedPet={selectedPet}
      form={form}
      fieldErrors={fieldErrors}
      petHistory={petHistory}
      historyFilter={historyFilter}
      deleteConfirmOpen={deleteConfirmOpen}
      formAction={formAction}
      onBack={handleBack}
      onMarkDirty={markDirty}
      onOpenDeleteConfirm={() => setDeleteConfirmOpen(true)}
      onCloseDeleteConfirm={() => setDeleteConfirmOpen(false)}
      onConfirmDelete={() => {
        handleDelete(() => {
          markClean();
          navigate(paths.vaccinations.getHref());
        });
      }}
    />
  );
});
