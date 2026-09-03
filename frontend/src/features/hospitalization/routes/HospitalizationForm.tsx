import { useCallback, useLayoutEffect, useRef, useTransition, useState } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router";

import { usePermission } from "@/hooks/use-permission";
import { useHospitalizationForm } from "../hooks/use-hospitalization-form";
import { useDeleteHospitalization } from "../api/delete-hospitalization";
import { paths } from "@/config/paths";
import { resolveHospitalizationFormGate } from "./hospitalization-form-model";
import {
  HospitalizationFormBody,
  HospitalizationFormStatusView,
  useHospitalizationFormChrome,
} from "./HospitalizationFormPanels";

export function HospitalizationForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const { id: hospitalizationId } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const { canEdit, canCreate, canDelete } = usePermission("hospitalization");
  const canSubmit = hospitalizationId ? canEdit : canCreate;
  const canDeleteRef = useRef(canDelete);
  const deleteMutation = useDeleteHospitalization();
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [isDeletePending, startDeleteTransition] = useTransition();
  const locationFrom = location.state?.from as string | undefined;

  const form = useHospitalizationForm(hospitalizationId, canSubmit === true);
  const selectedPet = form.petSelection.selectedPets[0];
  const petIsDeceased = selectedPet?.status === "死亡";
  const petIsDeceasedRef = useRef(petIsDeceased);
  useLayoutEffect(() => {
    canDeleteRef.current = canDelete;
    petIsDeceasedRef.current = petIsDeceased;
  }, [canDelete, petIsDeceased]);

  const handleDelete = useCallback(() => {
    if (
      !hospitalizationId ||
      canDeleteRef.current !== true ||
      petIsDeceasedRef.current === true
    ) return;
    startDeleteTransition(() => {
      // FE-RC-005: useDeleteHospitalization.onError が既に handleApiError でトースト表示済みのため、
      // ここでは再通知しない。
      deleteMutation.mutate(hospitalizationId, {
        onSuccess: () => {
          navigate(paths.hospitalization.getHref());
        },
      });
    });
  }, [hospitalizationId, deleteMutation, navigate]);

  const chrome = useHospitalizationFormChrome({
    hospitalizationId,
    petId,
    locationFrom,
    isEdit: form.isEdit,
    formData: form.formData,
    formState: form.formState,
    handleFormDataChangeRaw: form.handleFormDataChange,
    calculateTotals: form.calculateTotals,
    selectedPet,
    canDelete,
    treatmentPlans: form.treatmentPlans,
  });

  const gate = resolveHospitalizationFormGate({
    hasSelectedPet: Boolean(selectedPet),
    isEdit: form.isEdit,
    petId,
    isReadLoading: form.isReadLoading,
    isReadNotFound: form.isReadNotFound,
    isReadError: form.isReadError,
    retryRead: form.retryRead,
  });
  if (gate) {
    return <HospitalizationFormStatusView gate={gate} onBack={chrome.handleBack} />;
  }
  if (!form.isEdit && petIsDeceased) {
    return <HospitalizationFormStatusView gate={{ kind: "new-deceased" }} onBack={chrome.handleBack} />;
  }

  return (
    <HospitalizationFormBody
      hospitalizationId={hospitalizationId}
      canSubmit={canSubmit === true}
      canShowDelete={chrome.canShowDelete}
      isDirty={chrome.isDirty}
      isDeleteConfirmOpen={isDeleteConfirmOpen}
      isDeletePending={isDeletePending}
      staffModalOpen={chrome.staffModalOpen}
      doctorStaffItems={chrome.doctorStaffItems}
      formData={form.formData}
      formAction={form.formAction}
      onBack={chrome.handleBack}
      onOpenDetail={() => navigate(paths.hospitalization.detail.getHref(String(hospitalizationId)))}
      onOpenDeleteConfirm={() => setIsDeleteConfirmOpen(true)}
      onCloseDeleteConfirm={() => setIsDeleteConfirmOpen(false)}
      onConfirmDelete={handleDelete}
      onStaffModalOpenChange={chrome.setStaffModalOpen}
      onSelectDoctor={chrome.handleSelectDoctor}
      fields={{
        selectedPet,
        formData: form.formData,
        fieldErrors: form.formState.fieldErrors,
        cageItems: chrome.cageItems,
        isEdit: form.isEdit,
        canDelete,
        hasChildTreatmentPlans: chrome.hasChildTreatmentPlans,
        treatmentPlans: form.treatmentPlans,
        totals: chrome.totals,
        historyItems: chrome.historyItems,
        isHistoryLoading: chrome.isHistoryLoading,
        canSubmit: canSubmit === true,
        onFormChange: chrome.handleFormChange,
        onOpenStaffModal: () => chrome.setStaffModalOpen(true),
        onAddTreatmentPlan: form.addTreatmentPlan,
        onUpdateTreatmentPlan: form.updateTreatmentPlan,
        onRemoveTreatmentPlan: form.removeTreatmentPlan,
      }}
    />
  );
}
