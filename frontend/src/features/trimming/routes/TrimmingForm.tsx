// React/Framework
import { ICON, C, LAYOUT } from "@/lib/design-tokens";
import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router";

// External
import { Scissors, Trash2 } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { PatientInfoCard, formatPatientPetDetails } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { FormHeaderActions } from "@/components/shared/Form/FormHeaderActions";
import { formatDate } from "@/lib/format/date";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { useMasterItems } from "@/hooks/use-master-items";
import { useGetTrimmingCourseTypes } from "@/hooks/use-trimming-course-types";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";

// Relative (direct file import, no barrel — bundle-barrel-imports)
import {
  TrimmingLeftColumn,
  TrimmingMiddleColumn,
  TrimmingRightColumn,
} from "../components/trimming-form-columns";
import { useTrimmingForm } from "../hooks/use-trimming-form";
import { filterActiveOrSelectedMasterItems } from "../hooks/trimming-form-utils";
import type { TrimmingFormData } from "@/types/trimming";
import { ResourceTrimming } from "@/types/generated/models";
import { ConfirmDialog, MasterSelectModal } from "./TrimmingLazyModals";
import { TRIMMING_FORM_ID, TRIMMING_PRIORITY_FIELDS } from "./trimming-form-model";
import { useTrimmingHistory } from "../hooks/use-trimming-history";

// ─── メインコンポーネント ────────────────────────────────────────────────────

export function TrimmingForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");

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
    hasExistingAppointment,
  } = useTrimmingForm(id);

  const { data: coursesRaw = [] } = useMasterItems("trimmingCourse");
  const { data: optionsRaw = [] } = useMasterItems("trimmingOption");
  const { data: staffItems = [] } = useMasterItems("staff");
  const { data: courseTypes = [] } = useGetTrimmingCourseTypes();
  // #73: コース選択で種別が分かるよう、コース名に種別名を併記する（例「[シャンプー] フルコース」）
  const courseTypeNameById = useMemo(
    () => new Map(courseTypes.map((t) => [t.id, t.name])),
    [courseTypes],
  );
  // #228: 無効化(is_active=false)されたコース/オプションは選択肢から除外する。
  // ただし編集中カルテに既に紐づく無効項目のみ「（無効）」表記で維持する（データを消さない）。
  const courses = useMemo(() => {
    const named = coursesRaw.map((c) => {
      const typeName = c.courseTypeId ? courseTypeNameById.get(c.courseTypeId) : undefined;
      return { ...c, id: String(c.id), name: typeName ? `[${typeName}] ${c.name}` : c.name };
    });
    return filterActiveOrSelectedMasterItems(named, formData.courseId ? [formData.courseId] : []);
  }, [coursesRaw, courseTypeNameById, formData.courseId]);
  const options = useMemo(() => {
    const named = optionsRaw.map((o) => ({ ...o, id: String(o.id) }));
    return filterActiveOrSelectedMasterItems(named, formData.optionIds);
  }, [optionsRaw, formData.optionIds]);

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
  }, [setHistoryDateRangeFrom]);

  const handleHistoryEndDateChange = useCallback((val: string) => {
    setHistoryDateRangeTo(val);
  }, [setHistoryDateRangeTo]);

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
        <ErrorFallback message="トリミング記録が見つかりません" />
      </PageLayout>
    );
  }

  return (
    <PageLayout
      title={mode === "new" ? "トリミング登録" : "トリミング編集"}
      onBack={handleBack}
      icon={<Scissors className={`${ICON.page} ${C.text}`} />}
      resource={ResourceTrimming}
      maxWidth={LAYOUT.pageContentMaxWidth.form}
      headerAction={
        <FormHeaderActions
          onCancel={handleBack}
          submitLabel={canSubmit ? (isSaving ? "保存中..." : "保存") : undefined}
          submitDisabled={isSaving}
          submitFormId={TRIMMING_FORM_ID}
          extra={mode === "edit" && canDelete ? (
            <Button
              type="button"
              onClick={() => setDeleteConfirmOpen(true)}
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
      {/* NavigationBlocker: isSaving 中はブロック無効化 */}
      {/* FE6-8: jsx-no-leaked-render は非型認識のため isDirty を boolean と静的に断定できず !! で明示する */}
      <NavigationBlocker when={!!isDirty && !isSaving} />
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
            staffLabel="担当医"
            staffButtonId="staffId"
            reservationType="トリミング"
            petDetails={formatPatientPetDetails({
              species: selectedPet.species,
              birthDate: selectedPet.birthDate,
              gender: selectedPet.gender,
              neuteredDate: selectedPet.neuteredDate,
            })}
            insuranceName={selectedPet.insuranceName}
            insuranceDetails={selectedPet.insuranceDetails}
            status={selectedPet.status === "死亡" ? "deceased" : "alive"}
            nextVisitDate={formData.nextDate ? formatDate(formData.nextDate) : undefined}
            onStaffClick={handleOpenStaffModal}
          />
          {/* BUG-027: inline staff validation error */}
          {fieldErrors.staffId ? (
            <FormFieldError message={fieldErrors.staffId} />
          ) : null}
          {fieldErrors.reservationTypeId ? (
            <FormFieldError message={fieldErrors.reservationTypeId} />
          ) : null}

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-5">
            <div className="space-y-6 lg:col-span-3">
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
              showInitialStatusSelector={mode === "new" && !hasExistingAppointment}
            />
            <TrimmingMiddleColumn
              formData={formData}
              completedImagePreview={completedImagePreview}
              onFormChange={handleFormChange}
              onCompletedImageChange={handleCompletedImageChange}
              onRemoveCompletedImage={removeCompletedImage}
            />
            </div>
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
