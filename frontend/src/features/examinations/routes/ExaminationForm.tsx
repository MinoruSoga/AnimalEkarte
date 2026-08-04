// React/Framework
import {
  useCallback,
  useDeferredValue,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  useNavigate,
  useParams,
  useLocation,
  useSearchParams,
} from "react-router";

// Internal
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { C, LAYOUT } from "@/lib/design-tokens";
import type { Pet, SortOrder } from "@/types";

// Relative
import { useExaminationForm } from "../hooks/use-examination-form";
import { useGetExaminations } from "../api/get-examinations";
import { ExamItemsTable } from "../components/ExamItemsTable";
import { ExaminationFormFields } from "../components/ExaminationFormFields";
import { ExaminationHistoryPanel } from "../components/ExaminationHistoryPanel";
import { ExaminationPatientChangeDialog } from "../components/ExaminationPatientChangeDialog";
import { ExaminationUnconfirmDialog } from "../components/ExaminationUnconfirmDialog";
import { ExaminationPrintArea } from "../components/ExaminationPrintArea";
import { useGetExaminationPrintSnapshot } from "../api/get-examination-print-snapshot";
import { buildExaminationPrintModel } from "../lib/examination-print-model";
import { useMasterItems } from "@/hooks/use-master-items";
import { useGetStaffs } from "@/features/master";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import {
  ResourceExaminations,
  ResourceExaminationUnconfirm,
} from "@/types/generated/models";
import { normalizeKana } from "@/lib/normalize-kana";

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
  const canSubmit = id ? canEdit : canCreate && canEdit;

  const { data: examTypesRaw, isLoading: examTypesLoading } =
    useMasterItems("examination");
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
    formAction,
    formState,
    fieldErrors,
    handleDelete,
    isEdit,
    isSaving,
    isDeleting,
    formItems,
    setInspectionValue,
    addManualItem,
    removeItem,
    setItemName,
    handleUnconfirm,
    isPersistedConfirmed,
    isPatientChangeLocked,
  } = useExaminationForm(id, medicalRecordId ?? undefined, {
    canCreate,
    canEdit,
    canDelete,
    canUnconfirm,
  });

  // Print uses saved revision snapshot only — never formItems / unsaved edits.
  const { data: printSnapshot } = useGetExaminationPrintSnapshot(
    isEdit ? id : undefined,
  );
  const printModel = useMemo(
    () => (printSnapshot ? buildExaminationPrintModel(printSnapshot) : null),
    [printSnapshot],
  );

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

  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);

  const { selectedPets, setSelectedPets } = petSelection;
  const selectedPet = selectedPets[0];
  // 編集ロックはサーバ保存済み確定のみ。ドラフトで「確定」を選んでも保存 UI を消さない（A-S02-01）。
  const isConfirmed = isPersistedConfirmed;

  // 現在のペットID（履歴フィルタ用）
  const currentPetId = formData.petId ?? selectedPet?.id ?? petId ?? undefined;

  // 検査履歴取得
  const { data: allExaminations = [] } = useGetExaminations({
    petId: currentPetId,
    startDate: historyStartDate || undefined,
    endDate: historyEndDate || undefined,
  });

  const deferredHistorySearch = useDeferredValue(historySearchTerm);

  const petHistory = useMemo(() => {
    if (!currentPetId) return [];
    return allExaminations.filter((e) => e.petId === currentPetId);
  }, [allExaminations, currentPetId]);

  const searchedPetHistory = useMemo(() => {
    if (!deferredHistorySearch) return petHistory;

    const searchValue = normalizeKana(deferredHistorySearch).toLowerCase();
    return petHistory.filter(
      (examination) =>
        normalizeKana(examination.testType)
          .toLowerCase()
          .includes(searchValue) ||
        normalizeKana(examination.resultSummary ?? "")
          .toLowerCase()
          .includes(searchValue),
    );
  }, [deferredHistorySearch, petHistory]);

  // js-cache-function-results: カード履歴フィルタ結果をメモ化
  const filteredHistory = useMemo(() => {
    let result = searchedPetHistory;
    // 編集中の記録自体は除外
    if (isEdit && id) {
      result = result.filter((e) => e.id !== id);
    }
    return [...result].sort((a, b) => {
      const cmp = a.date.localeCompare(b.date);
      return historySortOrder === "asc" ? cmp : -cmp;
    });
  }, [searchedPetHistory, isEdit, id, historySortOrder]);

  const handleBack = useCallback(() => {
    if (location.state?.from) {
      navigate(location.state.from);
    } else {
      navigate(paths.examinations.getHref());
    }
  }, [location.state, navigate]);

  // rerender-memo: memo'd セクションに渡すハンドラを useCallback で安定化
  const handleSetFormData = useCallback(
    (next: Parameters<typeof setFormData>[0]) => {
      markDirty();
      setFormData(next);
    },
    [markDirty, setFormData],
  );

  const handleInspectionValueChange = useCallback(
    (key: string, value: string) => {
      markDirty();
      setInspectionValue(key, value);
    },
    [markDirty, setInspectionValue],
  );

  const handleItemNameChange = useCallback(
    (key: string, value: string) => {
      markDirty();
      setItemName(key, value);
    },
    [markDirty, setItemName],
  );

  const handleAddItem = useCallback(() => {
    markDirty();
    addManualItem();
  }, [addManualItem, markDirty]);

  const handleRemoveItem = useCallback(
    (key: string) => {
      markDirty();
      removeItem(key);
    },
    [markDirty, removeItem],
  );

  const handlePatientSelect = useCallback(
    (pet: Pet) => {
      markDirty();
      setSelectedPets([pet]);
    },
    [markDirty, setSelectedPets],
  );

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
      maxWidth={LAYOUT.pageContentMaxWidth.formMid}
      align="left"
    >
      {/* FE6-8: jsx-no-leaked-render は非型認識のため isDirty を boolean と静的に断定できず !! で明示する */}
      <NavigationBlocker when={!!isDirty && !isSaving} />
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

        {isEdit && !isPatientChangeLocked ? (
          <div className="flex justify-end">
            <ExaminationPatientChangeDialog
              selectedPet={selectedPet}
              onSelect={handlePatientSelect}
            />
          </div>
        ) : null}

        {isPersistedConfirmed && canUnconfirm && id ? (
          <div className="flex justify-end">
            <ExaminationUnconfirmDialog onUnconfirm={handleUnconfirm} />
          </div>
        ) : null}

        {isEdit && id ? (
          <div className="flex justify-end print:hidden">
            <button
              type="button"
              data-testid="examination-print-button"
              className={`rounded-xs border px-3 py-1.5 text-sm ${C.border} ${C.text60} hover:bg-black/5 disabled:opacity-50`}
              disabled={!printModel}
              onClick={() => window.print()}
            >
              印刷 / PDF出力
            </button>
          </div>
        ) : null}

        {printModel ? <ExaminationPrintArea model={printModel} /> : null}

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
                    disabled={isConfirmed}
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
