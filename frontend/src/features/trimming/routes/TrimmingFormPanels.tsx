import { Suspense, useCallback, type ChangeEvent } from "react";
import { Scissors, Trash2 } from "lucide-react";

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
import type { MasterSelectItem } from "@/components/shared/MasterSelectModal";
import { ConfirmDialog, MasterSelectModal } from "./TrimmingLazyModals";
import { TRIMMING_FORM_ID, type TrimmingFormGate, type TrimmingSelectableItem } from "./trimming-form-model";
import { TrimmingFormColumns, type TrimmingPatient } from "./TrimmingFormColumns";

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
  // rerender-memo: MasterSelectModal は memo。onSelect を inline arrow で渡すと
  // 毎レンダー新規参照になり memo が無効化されるため useCallback で安定化する。
  const handleSelectCourse = useCallback(
    (item: MasterSelectItem) => onFormChange({ courseId: String(item.id) }),
    [onFormChange],
  );
  const handleSelectStaff = useCallback(
    (item: MasterSelectItem) => onFormChange({ staffName: item.name, staffId: String(item.id) }),
    [onFormChange],
  );

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
      {selectedPet?.status === "死亡" ? (
        <div
          role="status"
          aria-label="死亡ペットのため保存不可"
          className={`flex items-center gap-2 px-4 py-2.5 rounded-md border mt-4 ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
        >
          <span className="text-sm font-medium">
            死亡したペットのトリミング記録は保存できません
          </span>
        </div>
      ) : null}
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
          onSelect={handleSelectCourse}
        />
        <MasterSelectModal
          open={staffModalOpen}
          onOpenChange={onStaffModalOpenChange}
          title="担当スタッフ選択"
          items={activeStaffItems}
          selectedValue={formData.staffName}
          matchBy="name"
          onSelect={handleSelectStaff}
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
