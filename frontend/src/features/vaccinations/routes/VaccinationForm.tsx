import { memo, useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate, useParams } from "react-router";

import { paths } from "@/config/paths";
import { performBackNavigation, useBackNavigation } from "@/hooks/use-back-navigation";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { useGetVaccinations } from "../api/get-vaccinations";
import type { VaccinationRecord } from "../api/transforms";
import { useVaccinationForm } from "../hooks/use-vaccination-form";
import { usePermission } from "@/hooks/use-permission";
import { isPetDeceasedForClinicalWrite } from "@/lib/transforms/pet";
import {
  VACCINATION_FIELD_ID_MAP,
  VACCINATION_PRIORITY_FIELDS,
  filterVaccinationHistory,
  resolveVaccinationFormGate,
} from "./vaccination-form-model";
import { VaccinationFormBody, VaccinationFormStatusView } from "./VaccinationFormPagePanels";
import { collectReplacedVaccinationIds } from "@/lib/vaccination-replacement";

export const VaccinationForm = memo(function VaccinationForm() {
  const navigate = useNavigate();
  const location = useLocation();
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
    isVaccinesMasterError,
    retryVaccinesMaster,
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
      performBackNavigation(navigate, paths.vaccinations.getHref(), location.state);
    }
  }, [formState.success, formState.timestamp, navigate, markClean, location.state]);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];
  const isPetDeceased = selectedPet != null && isPetDeceasedForClinicalWrite(selectedPet);
  const canSubmit = (id ? canEdit && isEditPetReady : canCreate) && !isPetDeceased;
  const allowDelete = canDelete === true && isEditPetReady;
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  const handleBack = useBackNavigation(paths.vaccinations.getHref());

  const historyPetId = selectedPet?.id;
  const { data: petVaccinations = [] } = useGetVaccinations({
    petId: historyPetId ?? "",
  });
  const { historySearchTerm, filterStartDate, filterEndDate, sortOrder } = historyFilter;
  // make ci blocker (EMR-435): VaccinationRecord.petId は transform が string | undefined を
  // 返し得るため、petId が確定した記録だけを置き換わり判定に渡す。petId 不明の記録は
  // 「同一ペット×同一ワクチン」判定に参加できない（EMR-428 の設計どおり）。
  const replacedVaccinationIds = useMemo(
    () =>
      collectReplacedVaccinationIds(
        petVaccinations.filter((r): r is VaccinationRecord & { petId: string } => r.petId != null),
      ),
    [petVaccinations],
  );
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
      isVaccinesMasterError={isVaccinesMasterError}
      onRetryVaccinesMaster={retryVaccinesMaster}
      fieldErrors={fieldErrors}
      petHistory={petHistory}
      replacedVaccinationIds={replacedVaccinationIds}
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
          performBackNavigation(navigate, paths.vaccinations.getHref(), location.state);
        });
      }}
    />
  );
});
