import { useParams, useLocation, useSearchParams } from "react-router";

import { usePermission } from "@/hooks/use-permission";
import { useTrimmingForm } from "../hooks/use-trimming-form";
import { useTrimmingFormChrome } from "../hooks/use-trimming-form-chrome";
import { resolveTrimmingFormGate } from "./trimming-form-model";
import {
  TrimmingFormBody,
  TrimmingFormStatusView,
} from "./TrimmingFormPanels";

export function TrimmingForm() {
  const location = useLocation();
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");

  const { canEdit, canCreate, canDelete } = usePermission("trimming");
  const {
    mode,
    formData,
    setFormData,
    styleImagePreview,
    completedImagePreview,
    petSelection,
    handleStyleImageChange,
    handleCompletedImageChange,
    removeStyleImage,
    removeCompletedImage,
    formAction,
    formState,
    handleDelete,
    isSaving,
    isDeleting,
    fieldErrors,
    isLoading,
    isEditPetReady,
    notFound,
    hasExistingAppointment,
  } = useTrimmingForm(id, { canCreate, canEdit, canDelete });

  const selectedPet = petSelection.selectedPets[0];
  const isPetDeceased = selectedPet?.status === "死亡";
  const canSubmit = (mode === "edit" ? canEdit && isEditPetReady : canCreate) && !isPetDeceased;
  const redirectPath = typeof location.state?.from === "string" ? location.state.from : "/trimming";
  const fromPath = location.state?.from as string | undefined;
  const chrome = useTrimmingFormChrome({
    formData,
    setFormData,
    formState,
    selectedPetId: selectedPet?.id,
    redirectPath,
    fromPath,
    handleDelete,
  });

  const gate = resolveTrimmingFormGate({
    hasSelectedPet: Boolean(selectedPet),
    mode,
    petId,
    isLoading,
    notFound,
  });
  if (gate) {
    return <TrimmingFormStatusView gate={gate} onBack={chrome.handleBack} />;
  }

  return (
    <TrimmingFormBody
      mode={mode}
      canSubmit={canSubmit === true}
      canDelete={canDelete === true && isEditPetReady && !isPetDeceased}
      isSaving={isSaving}
      isDeleting={isDeleting}
      isDirty={chrome.isDirty}
      hasExistingAppointment={hasExistingAppointment}
      selectedPet={selectedPet}
      formData={formData}
      fieldErrors={fieldErrors}
      courses={chrome.courses}
      options={chrome.options}
      styleImagePreview={styleImagePreview}
      completedImagePreview={completedImagePreview}
      sortedHistory={chrome.history.sortedHistory}
      isHistoryLoading={chrome.history.isHistoryLoading}
      historySearchTerm={chrome.history.historySearchTerm}
      historySortOrder={chrome.history.historySortOrder}
      historyDateRange={chrome.history.historyDateRange}
      courseModalOpen={chrome.courseModalOpen}
      staffModalOpen={chrome.staffModalOpen}
      deleteConfirmOpen={chrome.deleteConfirmOpen}
      activeStaffItems={chrome.activeStaffItems}
      formAction={formAction}
      onBack={chrome.handleBack}
      onFormChange={chrome.handleFormChange}
      onOpenCourseModal={() => chrome.setCourseModalOpen(true)}
      onOpenStaffModal={() => chrome.setStaffModalOpen(true)}
      onOpenDeleteConfirm={() => chrome.setDeleteConfirmOpen(true)}
      onCourseModalOpenChange={chrome.setCourseModalOpen}
      onStaffModalOpenChange={chrome.setStaffModalOpen}
      onCloseDeleteConfirm={() => chrome.setDeleteConfirmOpen(false)}
      onConfirmDelete={chrome.handleDeleteClick}
      onStyleImageChange={handleStyleImageChange}
      onCompletedImageChange={handleCompletedImageChange}
      onRemoveStyleImage={removeStyleImage}
      onRemoveCompletedImage={removeCompletedImage}
      onHistorySearchTermChange={chrome.history.setHistorySearchTerm}
      onHistorySortOrderChange={chrome.history.setHistorySortOrder}
      onHistoryClear={chrome.history.handleHistoryClear}
      onHistoryStartDateChange={chrome.history.setHistoryDateRangeFrom}
      onHistoryEndDateChange={chrome.history.setHistoryDateRangeTo}
      onHistoryClick={chrome.handleHistoryClick}
    />
  );
}
