import { Suspense, useCallback, useEffect, useMemo, useState, type ChangeEvent } from "react";
import { useNavigate } from "react-router";
import { Scissors, Trash2 } from "lucide-react";

import { useGetMasterItems } from "@/hooks/use-master-items";
import { useGetTrimmingCourseTypes } from "@/hooks/use-trimming-course-types";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { paths } from "@/config/paths";
import { useTrimmingHistory } from "../hooks/use-trimming-history";
import {
  decorateTrimmingCourses,
  decorateTrimmingOptions,
  TRIMMING_PRIORITY_FIELDS,
} from "./trimming-form-model";

import { Button } from "@/components/ui/button";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { FormHeaderActions } from "@/components/shared/Form/FormHeaderActions";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { ICON, C, LAYOUT } from "@/lib/design-tokens";
import { ResourceTrimming } from "@/types/generated/models";
import type { MasterItem, SortOrder } from "@/types";
import type { TrimmingFormData } from "@/types/trimming";
import type { TrimmingHistoryItem } from "../components/trimming-form-column-types";
import { ConfirmDialog, MasterSelectModal } from "./TrimmingLazyModals";
import { TRIMMING_FORM_ID, type TrimmingFormGate, type TrimmingSelectableItem } from "./trimming-form-model";
import { TrimmingFormColumns, type TrimmingPatient } from "./trimming-form-body-columns";

// eslint-disable-next-line react-refresh/only-export-components -- 150行分割で page chrome hook を panels と同居
export function useTrimmingFormChrome(input: {
  formData: TrimmingFormData;
  setFormData: (updates: Partial<TrimmingFormData>) => void;
  formState: { success?: boolean; timestamp?: number; fieldErrors?: Record<string, string> };
  selectedPetId: string | undefined;
  redirectPath: string;
  fromPath: string | undefined;
  handleDelete: (onSuccess: () => void) => void;
}) {
  const navigate = useNavigate();
  const { data: coursesRaw = [] } = useGetMasterItems("trimmingCourse");
  const { data: optionsRaw = [] } = useGetMasterItems("trimmingOption");
  const { data: staffItems = [] } = useGetMasterItems("staff");
  const { data: courseTypes = [] } = useGetTrimmingCourseTypes();
  const courseTypeNameById = useMemo(
    () => new Map(courseTypes.map((type) => [type.id, type.name])),
    [courseTypes],
  );
  const courses = useMemo(
    () => decorateTrimmingCourses(coursesRaw, courseTypeNameById, input.formData.courseId),
    [coursesRaw, courseTypeNameById, input.formData.courseId],
  );
  const options = useMemo(
    () => decorateTrimmingOptions(optionsRaw, input.formData.optionIds),
    [optionsRaw, input.formData.optionIds],
  );
  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  useEffect(() => {
    const errorFields = Object.keys(input.formState.fieldErrors || {});
    if (errorFields.length === 0) return;
    const firstError = TRIMMING_PRIORITY_FIELDS.find((field) => errorFields.includes(field)) || errorFields[0];
    const element = document.getElementById(firstError);
    if (element) {
      element.focus();
      element.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- fieldErrors は timestamp と同期。timestamp のみで十分
  }, [input.formState.timestamp]);

  useEffect(() => {
    if (input.formState.success) {
      markClean();
      navigate(input.redirectPath);
    }
  }, [input.formState.success, input.formState.timestamp, navigate, markClean, input.redirectPath]);

  const [courseModalOpen, setCourseModalOpen] = useState(false);
  const [staffModalOpen, setStaffModalOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const history = useTrimmingHistory(input.selectedPetId ?? "");
  const setFormData = input.setFormData;
  const handleFormChange = useCallback((updates: Partial<TrimmingFormData>) => {
    markDirty();
    setFormData(updates);
  }, [markDirty, setFormData]);
  const handleDelete = input.handleDelete;
  const handleDeleteClick = useCallback(() => {
    handleDelete(() => {
      markClean();
      navigate(paths.trimming.getHref());
    });
  }, [handleDelete, markClean, navigate]);
  const handleBack = useCallback(() => {
    navigate(input.fromPath ?? paths.trimming.getHref());
  }, [input.fromPath, navigate]);
  const handleHistoryClick = useCallback((updates: Partial<TrimmingFormData>) => {
    handleFormChange(history.handleHistoryClick(updates));
  }, [history, handleFormChange]);
  const activeStaffItems = useMemo(
    () => staffItems.filter((staff) => staff.status === "active"),
    [staffItems],
  );

  return {
    courses,
    options,
    isDirty,
    courseModalOpen,
    setCourseModalOpen,
    staffModalOpen,
    setStaffModalOpen,
    deleteConfirmOpen,
    setDeleteConfirmOpen,
    history,
    handleFormChange,
    handleDeleteClick,
    handleBack,
    handleHistoryClick,
    activeStaffItems,
  };
}

export function TrimmingFormStatusView({
  gate,
  onBack,
}: {
  gate: TrimmingFormGate;
  onBack: () => void;
}) {
  if (gate.kind === "new-pet-loading") {
    return (
      <div className={`flex items-center justify-center p-8 text-base ${C.text50}`}>
        <p>読み込み中...</p>
      </div>
    );
  }
  if (gate.kind === "new-no-pet") {
    return (
      <div className={`flex items-center justify-center p-8 text-base ${C.text50}`}>
        <p>ペットを選択してください</p>
      </div>
    );
  }
  if (gate.kind === "loading") {
    return (
      <PageLayout title="トリミング" onBack={onBack} icon={<Scissors className={`${ICON.page} ${C.text}`} />}>
        <LoadingFallback />
      </PageLayout>
    );
  }
  return (
    <PageLayout title="トリミング" onBack={onBack} icon={<Scissors className={`${ICON.page} ${C.text}`} />}>
      <ErrorFallback message="トリミング記録が見つかりません" />
    </PageLayout>
  );
}

interface TrimmingFormBodyProps {
  mode: "new" | "edit";
  canSubmit: boolean;
  canDelete: boolean | undefined;
  isSaving: boolean;
  isDeleting: boolean;
  isDirty: boolean;
  hasExistingAppointment: boolean;
  selectedPet: TrimmingPatient | undefined;
  formData: TrimmingFormData;
  fieldErrors: Record<string, string>;
  courses: TrimmingSelectableItem[];
  options: TrimmingSelectableItem[];
  styleImagePreview: string | null;
  completedImagePreview: string | null;
  sortedHistory: TrimmingHistoryItem[];
  isHistoryLoading: boolean;
  historySearchTerm: string;
  historySortOrder: SortOrder;
  historyDateRange: { from: string; to: string };
  courseModalOpen: boolean;
  staffModalOpen: boolean;
  deleteConfirmOpen: boolean;
  activeStaffItems: MasterItem[];
  formAction: (payload: FormData) => void;
  onBack: () => void;
  onFormChange: (updates: Partial<TrimmingFormData>) => void;
  onOpenCourseModal: () => void;
  onOpenStaffModal: () => void;
  onOpenDeleteConfirm: () => void;
  onCourseModalOpenChange: (open: boolean) => void;
  onStaffModalOpenChange: (open: boolean) => void;
  onCloseDeleteConfirm: () => void;
  onConfirmDelete: () => void;
  onStyleImageChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onCompletedImageChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onRemoveStyleImage: () => void;
  onRemoveCompletedImage: () => void;
  onHistorySearchTermChange: (value: string) => void;
  onHistorySortOrderChange: (value: SortOrder) => void;
  onHistoryClear: () => void;
  onHistoryStartDateChange: (value: string) => void;
  onHistoryEndDateChange: (value: string) => void;
  onHistoryClick: (updates: Partial<TrimmingFormData>) => void;
}

export function TrimmingFormBody({
  mode,
  canSubmit,
  canDelete,
  isSaving,
  isDeleting,
  isDirty,
  hasExistingAppointment,
  selectedPet,
  formData,
  fieldErrors,
  courses,
  options,
  styleImagePreview,
  completedImagePreview,
  sortedHistory,
  isHistoryLoading,
  historySearchTerm,
  historySortOrder,
  historyDateRange,
  courseModalOpen,
  staffModalOpen,
  deleteConfirmOpen,
  activeStaffItems,
  formAction,
  onBack,
  onFormChange,
  onOpenCourseModal,
  onOpenStaffModal,
  onOpenDeleteConfirm,
  onCourseModalOpenChange,
  onStaffModalOpenChange,
  onCloseDeleteConfirm,
  onConfirmDelete,
  onStyleImageChange,
  onCompletedImageChange,
  onRemoveStyleImage,
  onRemoveCompletedImage,
  onHistorySearchTermChange,
  onHistorySortOrderChange,
  onHistoryClear,
  onHistoryStartDateChange,
  onHistoryEndDateChange,
  onHistoryClick,
}: TrimmingFormBodyProps) {
  return (
    <PageLayout
      title={mode === "new" ? "トリミング登録" : "トリミング編集"}
      onBack={onBack}
      icon={<Scissors className={`${ICON.page} ${C.text}`} />}
      resource={ResourceTrimming}
      maxWidth={LAYOUT.pageContentMaxWidth.form}
      headerAction={
        <FormHeaderActions
          onCancel={onBack}
          submitLabel={canSubmit ? (isSaving ? "保存中..." : "保存") : undefined}
          submitDisabled={isSaving}
          submitFormId={TRIMMING_FORM_ID}
          extra={mode === "edit" && canDelete ? (
            <Button
              type="button"
              onClick={onOpenDeleteConfirm}
              variant="ghost-danger"
              className="h-10 rounded-sm text-sm px-4"
              disabled={isDeleting}
            >
              <Trash2 className={`mr-1.5 ${ICON.action}`} />
              削除
            </Button>
          ) : null}
        />
      }
    >
      <NavigationBlocker when={isDirty ? !isSaving : false} />
      <form id={TRIMMING_FORM_ID} action={formAction}>
      <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
      {selectedPet ? (
        <TrimmingFormColumns
          selectedPet={selectedPet}
          formData={formData}
          fieldErrors={fieldErrors}
          courses={courses}
          options={options}
          styleImagePreview={styleImagePreview}
          completedImagePreview={completedImagePreview}
          sortedHistory={sortedHistory}
          isHistoryLoading={isHistoryLoading}
          historySearchTerm={historySearchTerm}
          historySortOrder={historySortOrder}
          historyDateRange={historyDateRange}
          showInitialStatusSelector={mode === "new" ? !hasExistingAppointment : false}
          onFormChange={onFormChange}
          onOpenCourseModal={onOpenCourseModal}
          onOpenStaffModal={onOpenStaffModal}
          onStyleImageChange={onStyleImageChange}
          onCompletedImageChange={onCompletedImageChange}
          onRemoveStyleImage={onRemoveStyleImage}
          onRemoveCompletedImage={onRemoveCompletedImage}
          onHistorySearchTermChange={onHistorySearchTermChange}
          onHistorySortOrderChange={onHistorySortOrderChange}
          onHistoryClear={onHistoryClear}
          onHistoryStartDateChange={onHistoryStartDateChange}
          onHistoryEndDateChange={onHistoryEndDateChange}
          onHistoryClick={onHistoryClick}
        />
      ) : null}
      </fieldset>

      <Suspense fallback={null}>
        <MasterSelectModal
          open={courseModalOpen}
          onOpenChange={onCourseModalOpenChange}
          title="コース選択"
          items={courses}
          selectedValue={formData.courseId}
          matchBy="id"
          onSelect={(item) => onFormChange({ courseId: String(item.id) })}
        />
        <MasterSelectModal
          open={staffModalOpen}
          onOpenChange={onStaffModalOpenChange}
          title="担当スタッフ選択"
          items={activeStaffItems}
          selectedValue={formData.staffName}
          matchBy="name"
          onSelect={(item) => onFormChange({ staffName: item.name, staffId: String(item.id) })}
        />
        <ConfirmDialog
          open={deleteConfirmOpen}
          onClose={onCloseDeleteConfirm}
          title="削除確認"
          description="このトリミング情報を削除してもよろしいですか？"
          confirmLabel="削除"
          variant="destructive"
          onConfirm={onConfirmDelete}
        />
      </Suspense>
      </form>
    </PageLayout>
  );
}
