import { useCallback, useEffect, useMemo } from "react";
import {
  useNavigate,
  useParams,
  useLocation,
  useSearchParams,
} from "react-router";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { Button } from "@/components/ui/button";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { C, LAYOUT } from "@/lib/design-tokens";
import { useGetMasterItems } from "@/hooks/use-master-items";
import { useGetStaffs } from "@/hooks/use-staffs";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import {
  ResourceExaminations,
  ResourceExaminationUnconfirm,
} from "@/types/generated/models";

import { useExaminationForm } from "../hooks/use-examination-form";
import { useExaminationHistoryFilters } from "../hooks/use-examination-history-filters";
import { useExaminationFormPageActions } from "../hooks/use-examination-form-page-actions";
import { ExamItemsTable } from "../components/ExamItemsTable";
import { ExaminationFormFields } from "../components/ExaminationFormFields";
import { ExaminationFormHeader } from "../components/ExaminationFormHeader";
import { ExaminationFormStatusPage } from "../components/ExaminationFormStatusPage";
import { ExaminationHistoryPanel } from "../components/ExaminationHistoryPanel";
import { useGetExaminationPrintSnapshot } from "../api/get-examination-print-snapshot";
import { buildExaminationPrintModel } from "../lib/examination-print-model";

// rendering-hoist-jsx: アクセシビリティ用定数をモジュールレベルに巻き上げ（毎レンダー再生成を回避）
const EXAMINATION_PRIORITY_FIELDS = ["testTypeId", "doctorId"] as const;

export function ExaminationForm() {
  const { id } = useParams();
  return <ExaminationFormContent key={id ?? "new"} id={id} />;
}

function ExaminationFormContent({ id }: { id: string | undefined }) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const medicalRecordId = searchParams.get("medicalRecordId");
  const historyView =
    searchParams.get("historyView") === "pivot" ? "pivot" : "cards";
  const { canEdit, canCreate, canDelete } = usePermission(ResourceExaminations);
  const { canEdit: canUnconfirm } = usePermission(ResourceExaminationUnconfirm);

  const { data: examTypesRaw, isLoading: examTypesLoading } =
    useGetMasterItems("examination");
  // BUG-005: typed staff source keeps staffType/isActive; generic master transform drops them.
  const { data: staffsRaw = [], isLoading: staffLoading } = useGetStaffs();
  const masterLoading = examTypesLoading || staffLoading;
  const examTypes = useMemo(
    () => examTypesRaw.map((t) => ({ id: String(t.id), name: t.name })),
    [examTypesRaw],
  );
  const staffList = useMemo(
    () =>
      staffsRaw
        .filter((s) => s.staffType === "doctor" && s.isActive)
        .map((s) => ({ id: String(s.id), name: s.name })),
    [staffsRaw],
  );

  const {
    formData,
    setFormData,
    petSelection,
    isPetDeceased,
    formAction,
    formState,
    fieldErrors,
    handleDelete,
    isEdit,
    isReadLoading,
    isReadNotFound,
    isReadError,
    retryRead,
    isSaving,
    isDeleting,
    formItems,
    setInspectionValue,
    addManualItem,
    removeItem,
    setItemName,
    handleUnconfirm,
    isPersistedConfirmed,
    isPersistedCompletedLocked,
    isPersistedResultsLocked,
    isPatientChangeLocked,
  } = useExaminationForm(id, medicalRecordId ?? undefined, {
    canCreate,
    canEdit,
    canDelete,
    canUnconfirm,
  });

  // FE-RC-002: 死亡ペットは render 側でも SubmitButton を非表示にする（callback 側の拒否と二重防壁）。
  const canSubmit = (id ? canEdit : canCreate && canEdit) && !isPetDeceased;

  // FE-RC-048: 固定文字列「医師A」ではなく選択中の担当医名を表示する
  const selectedDoctorName = useMemo(
    () => staffList.find((staff) => staff.id === formData.doctorId)?.name ?? "",
    [staffList, formData.doctorId],
  );

  // Print uses saved revision snapshot only — never formItems / unsaved edits.
  const { data: printSnapshot } = useGetExaminationPrintSnapshot(
    isEdit ? id : undefined,
  );
  const printModel = useMemo(
    () => (printSnapshot ? buildExaminationPrintModel(printSnapshot) : null),
    [printSnapshot],
  );

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  const handleHistoryViewChange = useCallback(
    (nextView: "cards" | "pivot") => {
      const nextParams = new URLSearchParams(searchParams);
      if (nextView === "pivot") {
        nextParams.set("historyView", "pivot");
      } else {
        nextParams.delete("historyView");
      }
      setSearchParams(nextParams, { replace: true });
    },
    [searchParams, setSearchParams],
  );

  // --- Focus Management (Accessibility) ---
  useEffect(() => {
    const errorFields = Object.keys(formState.fieldErrors || {});
    if (errorFields.length === 0) return;

    const firstError =
      EXAMINATION_PRIORITY_FIELDS.find((f) => errorFields.includes(f)) ||
      errorFields[0];

    const element = document.getElementById(firstError);
    if (element) {
      element.focus();
      element.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }, [formState.fieldErrors, formState.timestamp]);

  // React 19 Action の成功を検知して遷移
  useEffect(() => {
    if (formState.success) {
      markClean();
      navigate(paths.examinations.getHref());
    }
  }, [formState.success, formState.timestamp, navigate, markClean]);

  const { selectedPets, setSelectedPets } = petSelection;
  const selectedPet = selectedPets[0];
  const isConfirmed = isPersistedConfirmed;
  const isCompletedLocked = isPersistedCompletedLocked;
  const isResultsLocked = isPersistedResultsLocked;

  // 現在のペットID（履歴フィルタ用）
  const currentPetId = formData.petId ?? selectedPet?.id ?? petId ?? undefined;

  // FE-RC-045/046: 履歴検索/絞り込みの state・派生値は use-examination-history-filters.ts へ分離。
  const {
    historySearchTerm,
    setHistorySearchTerm,
    historySortOrder,
    setHistorySortOrder,
    historyStartDate,
    setHistoryStartDate,
    historyEndDate,
    setHistoryEndDate,
    handleHistoryClear,
    searchedPetHistory,
    filteredHistory,
  } = useExaminationHistoryFilters({
    currentPetId,
    isEdit,
    excludeId: id,
  });

  const {
    handleBack,
    handleSetFormData,
    handleInspectionValueChange,
    handleItemNameChange,
    handleAddItem,
    handleRemoveItem,
    handlePatientSelect,
    isDeleteConfirmOpen,
    handleDeleteClick,
    handleDeleteCancel,
    handleDeleteConfirm,
  } = useExaminationFormPageActions({
    navigate,
    fromPath: location.state?.from,
    markDirty,
    markClean,
    setFormData,
    setInspectionValue,
    setItemName,
    addManualItem,
    removeItem,
    setSelectedPets,
    handleDelete,
  });

  // FE-RC-033: ペット未選択時のリダイレクトは useExaminationFormPetSync（use-examination-form-helpers.ts）に一元化済み。
  // ここでは redirect 完了までの間、フォームを描画しないための render guard のみを持つ。
  if (!selectedPet && !isEdit) return null;

  // BUG-016: never render blank editable form for missing / other-clinic / forbidden IDs
  if (isEdit && isReadLoading) {
    return (
      <ExaminationFormStatusPage resource={ResourceExaminations} onBack={handleBack}>
        <LoadingFallback />
      </ExaminationFormStatusPage>
    );
  }
  if (isEdit && isReadNotFound) {
    return (
      <ExaminationFormStatusPage resource={ResourceExaminations} onBack={handleBack}>
        <ErrorFallback message="検査記録が見つかりません" />
      </ExaminationFormStatusPage>
    );
  }
  if (isEdit && isReadError) {
    return (
      <ExaminationFormStatusPage resource={ResourceExaminations} onBack={handleBack}>
        <div className="space-y-3">
          <ErrorFallback message="検査記録の取得に失敗しました" />
          {retryRead ? (
            <Button type="button" variant="outline" size="sm" onClick={retryRead}>
              再試行
            </Button>
          ) : null}
        </div>
      </ExaminationFormStatusPage>
    );
  }

  return (
    <PageLayout
      title={isEdit ? "検査詳細・編集" : "新規検査登録"}
      resource={ResourceExaminations}
      onBack={handleBack}
      maxWidth={LAYOUT.pageContentMaxWidth.formMid}
      align="left"
    >
      <NavigationBlocker when={isDirty ? !isSaving : false} />
      <div className="flex flex-col gap-4">
        <ExaminationFormHeader
          selectedPet={selectedPet}
          selectedDoctorName={selectedDoctorName}
          isPetDeceased={isPetDeceased}
          isEdit={isEdit}
          isPatientChangeLocked={isPatientChangeLocked}
          onPatientSelect={handlePatientSelect}
          isPersistedConfirmed={isPersistedConfirmed}
          canUnconfirm={canUnconfirm}
          examinationId={id}
          onUnconfirm={handleUnconfirm}
          printModel={printModel}
        />

        {/* 2カラムレイアウト: 左 3/5（フォーム）・右 2/5（履歴） */}
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-4 items-start">
          {/* 左カラム: フォームフィールド + 検査項目テーブル */}
          <form action={formAction} className="min-w-0 lg:col-span-3">
            <fieldset
              disabled={!canSubmit}
              className="border-0 p-0 m-0 min-w-0"
            >
              <div className="space-y-4">
                <ExaminationFormFields
                  formData={formData}
                  examTypes={examTypes}
                  staffList={staffList}
                  masterLoading={masterLoading}
                  isEdit={isEdit}
                  isDeleting={isDeleting}
                  isConfirmed={isConfirmed}
                  isCompletedLocked={isCompletedLocked}
                  isPetDeceased={isPetDeceased}
                  canEdit={canEdit}
                  canCreate={canCreate}
                  canDelete={canDelete}
                  fieldErrors={fieldErrors}
                  onSetFormData={handleSetFormData}
                  onBack={handleBack}
                  onDeleteClick={handleDeleteClick}
                />

                <div className="space-y-2">
                  <h3 className={`text-sm font-medium ${C.text60} px-1`}>
                    検査項目
                  </h3>
                  {formState.fieldErrors?.examItems ? (
                    <p
                      id="examItems"
                      role="alert"
                      tabIndex={-1}
                      className={`rounded-xs border px-3 py-2 text-sm ${C.danger} ${C.bgDanger8} ${C.borderDanger20}`}
                    >
                      {formState.fieldErrors.examItems}
                    </p>
                  ) : null}
                  <ExamItemsTable
                    items={formItems}
                    onChangeInspectionValue={handleInspectionValueChange}
                    onChangeName={handleItemNameChange}
                    onAddItem={handleAddItem}
                    onRemoveItem={handleRemoveItem}
                    disabled={isResultsLocked}
                  />
                </div>
              </div>
            </fieldset>
          </form>

          <ExaminationHistoryPanel
            filteredHistory={filteredHistory}
            pivotHistory={searchedPetHistory}
            currentPetId={currentPetId}
            historyStartDate={historyStartDate}
            historyEndDate={historyEndDate}
            historySearchTerm={historySearchTerm}
            historySortOrder={historySortOrder}
            historyView={historyView}
            onHistoryStartDateChange={setHistoryStartDate}
            onHistoryEndDateChange={setHistoryEndDate}
            onHistorySearchTermChange={setHistorySearchTerm}
            onHistorySortOrderChange={setHistorySortOrder}
            onHistoryViewChange={handleHistoryViewChange}
            onHistoryClear={handleHistoryClear}
          />
        </div>
      </div>

      <ConfirmDialog
        open={isDeleteConfirmOpen}
        onClose={handleDeleteCancel}
        onConfirm={handleDeleteConfirm}
        title="検査記録を削除しますか？"
        description="この操作は取り消せません。"
        confirmLabel="削除する"
        cancelLabel="キャンセル"
        variant="destructive"
      />
    </PageLayout>
  );
}
