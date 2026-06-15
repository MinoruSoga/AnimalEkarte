// React/Framework
import { useCallback, useDeferredValue, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router";

// Internal
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { C } from "@/lib/design-tokens";
import type { SortOrder } from "@/types";

// Relative
import { useExaminationForm } from "../hooks/use-examination-form";
import { useGetExaminations } from "../api/get-examinations";
import { ExamItemsTable } from "../components/ExamItemsTable";
import { ExaminationFormFields } from "../components/ExaminationFormFields";
import { ExaminationHistoryPanel } from "../components/ExaminationHistoryPanel";
import { useMasterItems } from "@/hooks/use-master-items";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { ResourceExaminations } from "@/types/generated/models";
import { normalizeKana } from "@/lib/normalize-kana";

// rendering-hoist-jsx: アクセシビリティ用定数をモジュールレベルに巻き上げ（毎レンダー再生成を回避）
const EXAMINATION_PRIORITY_FIELDS = ["testTypeId", "doctorId"] as const;

export function ExaminationForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const medicalRecordId = searchParams.get("medicalRecordId");
  const { canEdit, canCreate, canDelete } = usePermission("examinations");

  const { data: examTypesRaw, isLoading: examTypesLoading } = useMasterItems("examination");
  const { data: staffListRaw, isLoading: staffLoading } = useMasterItems("staff");
  const masterLoading = examTypesLoading || staffLoading;
  const examTypes = useMemo(
    () => examTypesRaw.map((t) => ({ id: String(t.id), name: t.name })),
    [examTypesRaw],
  );
  const staffList = useMemo(
    () => staffListRaw.map((s) => ({ id: String(s.id), name: s.name })),
    [staffListRaw],
  );

  const {
    formData,
    setFormData,
    petSelection,
    formAction,
    formState,
    handleDelete,
    isEdit,
    isSaving,
    isDeleting,
    formItems,
    setInspectionValue,
  } = useExaminationForm(id, medicalRecordId ?? undefined);

  const canSubmit = isEdit ? canEdit : canCreate;

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  // --- 履歴フィルタ状態 ---
  const [historySearchTerm, setHistorySearchTerm] = useState("");
  const [historySortOrder, setHistorySortOrder] = useState<SortOrder>("desc");
  const [historyStartDate, setHistoryStartDate] = useState("");
  const [historyEndDate, setHistoryEndDate] = useState("");

  const handleHistoryClear = useCallback(() => {
    setHistorySearchTerm("");
    setHistorySortOrder("desc");
    setHistoryStartDate("");
    setHistoryEndDate("");
  }, []);

  // --- Focus Management (Accessibility) ---
  useEffect(() => {
    const errorFields = Object.keys(formState.fieldErrors || {});
    if (errorFields.length === 0) return;

    const firstError = EXAMINATION_PRIORITY_FIELDS.find((f) => errorFields.includes(f)) || errorFields[0];

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

  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];
  const isConfirmed = formData.status === "確定";

  // 現在のペットID（履歴フィルタ用）
  const currentPetId = formData.petId ?? selectedPet?.id ?? petId ?? undefined;

  // 検査履歴取得
  const { data: allExaminations = [] } = useGetExaminations({
    startDate: historyStartDate || undefined,
    endDate: historyEndDate || undefined,
  });

  const deferredHistorySearch = useDeferredValue(historySearchTerm);

  // js-cache-function-results: 履歴フィルタ結果をメモ化
  const filteredHistory = useMemo(() => {
    if (!currentPetId) return [];
    let result = allExaminations.filter((e) => e.petId === currentPetId);
    // 編集中の記録自体は除外
    if (isEdit && id) {
      result = result.filter((e) => e.id !== id);
    }
    if (deferredHistorySearch) {
      const lower = normalizeKana(deferredHistorySearch).toLowerCase();
      result = result.filter(
        (e) =>
          normalizeKana(e.testType).toLowerCase().includes(lower) ||
          normalizeKana(e.resultSummary ?? "").toLowerCase().includes(lower),
      );
    }
    return [...result].sort((a, b) => {
      const cmp = a.date.localeCompare(b.date);
      return historySortOrder === "asc" ? cmp : -cmp;
    });
  }, [allExaminations, currentPetId, deferredHistorySearch, isEdit, id, historySortOrder]);

  const handleBack = useCallback(() => {
    if (location.state?.from) {
      navigate(location.state.from);
    } else {
      navigate(paths.examinations.getHref());
    }
  }, [location.state, navigate]);

  // rerender-memo: memo'd セクションに渡すハンドラを useCallback で安定化
  const handleSetFormData = useCallback((next: Parameters<typeof setFormData>[0]) => {
    markDirty();
    setFormData(next);
  }, [markDirty, setFormData]);

  const handleDeleteClick = useCallback(() => {
    setIsDeleteConfirmOpen(true);
  }, []);

  const handleDeleteCancel = useCallback(() => {
    setIsDeleteConfirmOpen(false);
  }, []);

  const handleDeleteConfirm = useCallback(() => {
    markClean();
    handleDelete(() => navigate(paths.examinations.getHref()));
    setIsDeleteConfirmOpen(false);
  }, [markClean, handleDelete, navigate]);

  useEffect(() => {
    if (!selectedPet && !isEdit && !petId) {
      navigate(paths.examinations.selectPet.getHref());
    }
  }, [selectedPet, isEdit, navigate, petId]);

  if (!selectedPet && !isEdit && petId) return null;
  if (!selectedPet && !isEdit) return null;

  return (
    <PageLayout
      title={isEdit ? "検査詳細・編集" : "新規検査登録"}
      resource={ResourceExaminations}
      onBack={handleBack}
      maxWidth="max-w-[1200px]"
      align="left"
    >
      <NavigationBlocker when={isDirty && !isSaving} />
      <form action={formAction}>
        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
        <div className="flex flex-col gap-4">
          {/* rerender-memo: PatientInfoCard — フォームフィールド変更では再レンダーしない */}
          {selectedPet ? (
            <PatientInfoCard
              ownerName={selectedPet.ownerName}
              petName={`${selectedPet.name}${selectedPet.species ? `(${selectedPet.species})` : ""}`}
              petNumber={selectedPet.petNumber || selectedPet.id}
              weight={selectedPet.weight || "-"}
              staffName="医師A"
              reservationType="検査"
              petDetails={`${selectedPet.birthDate ? `${selectedPet.birthDate}生` : ""} / ${selectedPet.species}`}
              insuranceName={selectedPet.insuranceName || "保険情報未登録"}
              insuranceDetails={selectedPet.insuranceDetails || "-"}
              nextVisitDate="-"
              nextVisitContent="-"
            />
          ) : null}

          {/* 2カラムレイアウト: 左 3/5（フォーム）・右 2/5（履歴） */}
          <div className="grid grid-cols-1 lg:grid-cols-5 gap-4 items-start">
            {/* 左カラム: フォームフィールド + 検査項目テーブル */}
            <div className="lg:col-span-3 space-y-4">
              <ExaminationFormFields
                formData={formData}
                examTypes={examTypes}
                staffList={staffList}
                masterLoading={masterLoading}
                isEdit={isEdit}
                isDeleting={isDeleting}
                isConfirmed={isConfirmed}
                canEdit={canEdit}
                canCreate={canCreate}
                canDelete={canDelete}
                onSetFormData={handleSetFormData}
                onBack={handleBack}
                onDeleteClick={handleDeleteClick}
              />

              <div className="space-y-2">
                <h3 className={`text-sm font-medium ${C.text60} px-1`}>検査項目</h3>
                <ExamItemsTable
                  items={formItems}
                  onChangeInspectionValue={setInspectionValue}
                  disabled={isConfirmed}
                />
              </div>
            </div>

            <ExaminationHistoryPanel
              filteredHistory={filteredHistory}
              currentPetId={currentPetId}
              historyStartDate={historyStartDate}
              historyEndDate={historyEndDate}
              historySearchTerm={historySearchTerm}
              historySortOrder={historySortOrder}
              onHistoryStartDateChange={setHistoryStartDate}
              onHistoryEndDateChange={setHistoryEndDate}
              onHistorySearchTermChange={setHistorySearchTerm}
              onHistorySortOrderChange={setHistorySortOrder}
              onHistoryClear={handleHistoryClear}
            />
          </div>
        </div>
        </fieldset>
      </form>

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
