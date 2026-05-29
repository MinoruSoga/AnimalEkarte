// React/Framework
import { ICON, C, STYLE } from "@/lib/design-tokens";
import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router";

// External
import { Scissors, Trash2 } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { LoadingFallback } from "@/components/shared/DataStates";
import { useMasterItems } from "@/hooks/use-master-items";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";

// Relative (direct file import, no barrel — bundle-barrel-imports)
import {
  TrimmingLeftColumn,
  TrimmingMiddleColumn,
  TrimmingRightColumn,
} from "../components/TrimmingFormColumns.ts";
import { useTrimmingForm } from "../hooks/use-trimming-form";
import type { TrimmingFormData } from "@/types/trimming";
import { ResourceTrimming } from "@/types/generated/models";
import { ConfirmDialog, MasterSelectModal } from "./TrimmingLazyModals";
import { TRIMMING_FORM_ID, TRIMMING_PRIORITY_FIELDS } from "./TrimmingFormModel";
import { useTrimmingHistory } from "./useTrimmingHistory";

// ─── メインコンポーネント ────────────────────────────────────────────────────

export function TrimmingForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");

  const { data: coursesRaw = [] } = useMasterItems("trimmingCourse");
  const { data: optionsRaw = [] } = useMasterItems("trimmingOption");
  const { data: staffItems = [] } = useMasterItems("staff");
  const courses = useMemo(() => coursesRaw.map((c) => ({ ...c, id: String(c.id) })), [coursesRaw]);
  const options = useMemo(() => optionsRaw.map((o) => ({ ...o, id: String(o.id) })), [optionsRaw]);

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
    notFound,
  } = useTrimmingForm(id);

  const { canEdit, canCreate, canDelete } = usePermission("trimming");
  const canSubmit = mode === "edit" ? canEdit : canCreate;
  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  // --- Focus Management (Accessibility) ---
  // rerender-dependencies: formState.fieldErrors (object) は deps に入れない。
  // timestamp が変わるたびに fieldErrors も変わるため timestamp だけで十分。
  useEffect(() => {
    const errorFields = Object.keys(formState.fieldErrors || {});
    if (errorFields.length === 0) return;

    // 優先順位に基づいたエラーフィールドの特定
    const firstError = TRIMMING_PRIORITY_FIELDS.find((f) => errorFields.includes(f)) || errorFields[0];

    const element = document.getElementById(firstError);
    if (element) {
      element.focus();
      element.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- fieldErrors は timestamp と同期。timestamp のみで十分
  }, [formState.timestamp]);

  // rerender-dependencies: location.state (object) から primitive を抽出して deps を安定化
  const redirectPath = typeof location.state?.from === "string" ? location.state.from : "/trimming";

  // React 19 Action の成功を検知して遷移
  useEffect(() => {
    if (formState.success) {
      markClean();
      navigate(redirectPath);
    }
  }, [formState.success, formState.timestamp, navigate, markClean, redirectPath]);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];

  const [courseModalOpen, setCourseModalOpen] = useState(false);
  const [staffModalOpen, setStaffModalOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  const {
    handleHistoryClear,
    handleHistoryClick: buildHistoryUpdates,
    historyDateRange,
    historySearchTerm,
    historySortOrder,
    isHistoryLoading,
    setHistoryDateRangeFrom,
    setHistoryDateRangeTo,
    setHistorySearchTerm,
    setHistorySortOrder,
    sortedHistory,
  } = useTrimmingHistory(selectedPet?.id ?? "");

  // rerender-functional-setstate: useCallback でハンドラを安定化
  // ※ React hooks はコンポーネントのトップレベルで呼ぶ（ガード節の前に定義）
  const handleFormChange = useCallback((updates: Partial<TrimmingFormData>) => {
    markDirty();
    setFormData(updates);
  }, [markDirty, setFormData]);

  const handleDeleteClick = useCallback(() => {
    handleDelete(() => {
      markClean();
      navigate(paths.trimming.getHref());
    });
  }, [handleDelete, markClean, navigate]);

  // rerender-dependencies: location.state (object) から primitive を抽出して deps を安定化
  const fromPath = location.state?.from as string | undefined;
  const handleBack = useCallback(() => {
    navigate(fromPath ?? paths.trimming.getHref());
  }, [fromPath, navigate]);

  const handleOpenCourseModal = useCallback(() => setCourseModalOpen(true), []);
  const handleOpenStaffModal = useCallback(() => setStaffModalOpen(true), []);

  const handleHistoryClick = useCallback((updates: Partial<TrimmingFormData>) => {
    handleFormChange(buildHistoryUpdates(updates));
  }, [buildHistoryUpdates, handleFormChange]);

  const handleHistoryStartDateChange = useCallback((val: string) => {
    setHistoryDateRangeFrom(val);
  }, []);

  const handleHistoryEndDateChange = useCallback((val: string) => {
    setHistoryDateRangeTo(val);
  }, []);

  const activeStaffItems = useMemo(
    () => staffItems.filter((s) => s.status === "active"),
    [staffItems]
  );

  if (!selectedPet && mode === "new" && petId) {
    return (
      <div className={`flex items-center justify-center p-8 text-base ${C.text50}`}>
        <p>読み込み中...</p>
      </div>
    );
  }
  if (!selectedPet && mode === "new") {
    return (
      <div className={`flex items-center justify-center p-8 text-base ${C.text50}`}>
        <p>ペットを選択してください</p>
      </div>
    );
  }

  if (isLoading) {
    return (
      <PageLayout title="トリミング" onBack={handleBack} icon={<Scissors className={`${ICON.page} ${C.text}`} />}>
        <LoadingFallback />
      </PageLayout>
    );
  }

  if (notFound) {
    return (
      <PageLayout title="トリミング" onBack={handleBack} icon={<Scissors className={`${ICON.page} ${C.text}`} />}>
        <div className={`px-6 py-12 text-center text-base ${C.text50}`}>トリミング記録が見つかりません</div>
      </PageLayout>
    );
  }

  return (
    <PageLayout
      title={mode === "new" ? "トリミング登録" : "トリミング編集"}
      onBack={handleBack}
      icon={<Scissors className={`${ICON.page} ${C.text}`} />}
      resource={ResourceTrimming}
      maxWidth="max-w-[1400px]"
      headerAction={
        <div className="flex gap-2">
          {/* rendering-conditional-render: && → ? ... : null */}
          {mode === "edit" && canDelete ? (
            <Button
              type="button"
              onClick={() => setDeleteConfirmOpen(true)}
              variant="ghost-danger"
              className="h-10 rounded-[6px] text-sm px-4"
              disabled={isDeleting}
            >
              <Trash2 className={`mr-1.5 ${ICON.action}`} />
              削除
            </Button>
          ) : null}
          {canSubmit ? (
            <Button
              type="submit"
              form={TRIMMING_FORM_ID}
              className={`${STYLE.confirmPrimary} h-10`}
              disabled={isSaving}
            >
              {isSaving ? "保存中..." : "保存"}
            </Button>
          ) : null}
        </div>
      }
    >
      {/* NavigationBlocker: isSaving 中はブロック無効化 */}
      <NavigationBlocker when={isDirty && !isSaving} />
      <form id={TRIMMING_FORM_ID} action={formAction}>
      <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
      {/* rendering-conditional-render: && → ? ... : null */}
      {selectedPet ? (
        <div className="space-y-6">
          {/* Patient Info Card */}
          <PatientInfoCard
            ownerName={selectedPet.ownerName}
            petName={selectedPet.name}
            petNumber={selectedPet.petNumber || ""}
            weight={selectedPet.weight || ""}
            staffName={formData.staffName}
            staffButtonId="staffId"
            reservationType="トリミング"
            nextVisitDate="-"
            nextVisitContent="-"
            onStaffClick={handleOpenStaffModal}
          />
          {/* BUG-027: inline staff validation error */}
          {fieldErrors.staffId ? (
            <FormFieldError message={fieldErrors.staffId} />
          ) : null}

          {/* Main Content - 3 column layout */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* rerender-memo: 各カラムを独立したmemo化コンポーネントに分離 */}
            <TrimmingLeftColumn
              formData={formData}
              courses={courses}
              options={options}
              styleImagePreview={styleImagePreview}
              onFormChange={handleFormChange}
              onCourseModalOpen={handleOpenCourseModal}
              onStyleImageChange={handleStyleImageChange}
              onRemoveStyleImage={removeStyleImage}
              courseError={fieldErrors.courseId}
            />
            <TrimmingMiddleColumn
              formData={formData}
              completedImagePreview={completedImagePreview}
              onFormChange={handleFormChange}
              onCompletedImageChange={handleCompletedImageChange}
              onRemoveCompletedImage={removeCompletedImage}
            />
            <TrimmingRightColumn
              sortedHistory={sortedHistory}
              isHistoryLoading={isHistoryLoading}
              historySearchTerm={historySearchTerm}
              historySortOrder={historySortOrder}
              historyDateRange={historyDateRange}
              onSearchTermChange={setHistorySearchTerm}
              onSortOrderChange={setHistorySortOrder}
              onClear={handleHistoryClear}
              onFilterStartDateChange={handleHistoryStartDateChange}
              onFilterEndDateChange={handleHistoryEndDateChange}
              onHistoryClick={handleHistoryClick}
            />
          </div>
        </div>
      ) : null}
      </fieldset>

      <Suspense fallback={null}>
        {/* Course Modal */}
        <MasterSelectModal
          open={courseModalOpen}
          onOpenChange={setCourseModalOpen}
          title="コース選択"
          items={courses}
          selectedValue={formData.courseId}
          matchBy="id"
          onSelect={(item) => handleFormChange({ courseId: String(item.id) })}
        />

        {/* Staff Modal */}
        <MasterSelectModal
          open={staffModalOpen}
          onOpenChange={setStaffModalOpen}
          title="担当スタッフ選択"
          items={activeStaffItems}
          selectedValue={formData.staffName}
          matchBy="name"
          onSelect={(item) => handleFormChange({ staffName: item.name, staffId: String(item.id) })}
        />

        {/* Delete Confirmation Dialog */}
        <ConfirmDialog
          open={deleteConfirmOpen}
          onClose={() => setDeleteConfirmOpen(false)}
          title="削除確認"
          description="このトリミング情報を削除してもよろしいですか？"
          confirmLabel="削除"
          variant="destructive"
          onConfirm={handleDeleteClick}
        />
      </Suspense>
      </form>
    </PageLayout>
  );
}
