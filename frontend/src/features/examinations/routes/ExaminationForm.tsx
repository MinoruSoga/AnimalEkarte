// React/Framework
import { memo, useCallback, useDeferredValue, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router";

// External
import { Trash2 } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { HistoryFilterPanel } from "@/components/shared/HistoryFilterPanel";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import type { SortOrder } from "@/types";

// Relative
import { useExaminationForm } from "../hooks/use-examination-form";
import { useGetExaminations } from "../api/get-examinations";
import { ExaminationCard } from "../components/ExaminationCard";
import { useMasterItems } from "@/hooks/use-master-items";
import { paths } from "@/config/paths";
import { usePermission } from "@/features/auth";
import type { ExaminationRecord } from "@/types";
import { ResourceExaminations } from "@/types/generated/models";

// rendering-hoist-jsx: ステータス選択肢は静的なのでモジュール定数に巻き上げ
const EXAM_STATUS_ITEMS = (
  <>
    <SelectItem value="依頼中">依頼中</SelectItem>
    <SelectItem value="検査中">検査中</SelectItem>
    <SelectItem value="結果入力済み">結果入力済み</SelectItem>
    <SelectItem value="完了">完了</SelectItem>
    <SelectItem value="確定">確定</SelectItem>
  </>
);

// rerender-memo: フォームフィールドセクションを独立 memo 化
interface FormFieldsSectionProps {
  formData: Partial<ExaminationRecord>;
  examTypes: { id: string; name: string }[];
  staffList: { id: string; name: string }[];
  isEdit: boolean;
  isSaving: boolean;
  isDeleting: boolean;
  isConfirmed: boolean;
  canEdit: boolean;
  canCreate: boolean;
  canDelete: boolean;
  onSetFormData: (next: Partial<ExaminationRecord>) => void;
  onBack: () => void;
  onDeleteClick: () => void;
}

const FormFieldsSection = memo(function FormFieldsSection({
  formData,
  examTypes,
  staffList,
  isEdit,
  isDeleting,
  isConfirmed,
  canEdit,
  canCreate,
  canDelete,
  onSetFormData,
  onBack,
  onDeleteClick,
}: FormFieldsSectionProps) {
  const canSubmit = isEdit ? canEdit : canCreate;
  return (
    <div className={`bg-white p-4 rounded-lg border ${C.borderMedium} space-y-4 shadow-sm`}>
      {isConfirmed ? (
        <p className={`text-sm font-medium ${C.text60}`}>確定済みのため編集できません。</p>
      ) : null}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label className={`text-sm ${C.text60}`}>検査種別</Label>
          <Select
            value={formData.testTypeId ?? ""}
            disabled={isConfirmed}
            onValueChange={(v) => {
              const item = examTypes.find((e) => e.id === v);
              onSetFormData({ testTypeId: v, testType: item?.name ?? v });
            }}
          >
            <SelectTrigger id="testTypeId" className={`h-10 text-sm ${C.text} bg-white ${C.borderMedium}`}>
              <SelectValue placeholder="選択してください" />
            </SelectTrigger>
            <SelectContent>
              {examTypes.map((item) => (
                <SelectItem key={item.id} value={String(item.id)}>
                  {item.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-1.5">
          <Label className={`text-sm ${C.text60}`}>担当医</Label>
          <Select
            value={formData.doctorId ?? ""}
            disabled={isConfirmed}
            onValueChange={(v) => {
              const staff = staffList.find((s) => String(s.id) === v);
              onSetFormData({ doctorId: v, doctor: staff?.name ?? v });
            }}
          >
            <SelectTrigger id="doctorId" className={`h-10 text-sm ${C.text} bg-white ${C.borderMedium}`}>
              <SelectValue placeholder="選択してください" />
            </SelectTrigger>
            <SelectContent>
              {staffList.map((staff) => (
                <SelectItem key={staff.id} value={String(staff.id)}>
                  {staff.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="space-y-1.5">
        <Label className={`text-sm ${C.text60}`}>検査日</Label>
        <NotionDatePicker
          value={formData.date ? formData.date.split("T")[0] : ""}
          onChange={(v) => onSetFormData({ date: v ? `${v}T00:00:00Z` : new Date().toISOString() })}
          disabledDays={{ after: new Date() }}
        />
      </div>

      <div className="space-y-1.5">
        <Label className={`text-sm ${C.text60}`}>ステータス</Label>
        <Select
          value={formData.status}
          disabled={isConfirmed}
          onValueChange={(v: "依頼中" | "検査中" | "結果入力済み" | "完了" | "確定") => onSetFormData({ status: v })}
        >
          <SelectTrigger className={`h-10 text-sm ${C.text} bg-white ${C.borderMedium}`}>
            <SelectValue placeholder="選択してください" />
          </SelectTrigger>
          <SelectContent>
            {EXAM_STATUS_ITEMS}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-1.5">
        <Label className={`text-sm ${C.text60}`}>備考・所見</Label>
        <Textarea
          className={`h-24 text-sm ${C.text} bg-white ${C.borderMedium} resize-none`}
          placeholder="検査結果や備考を入力"
          value={formData.resultSummary || ""}
          disabled={isConfirmed}
          onChange={(e) => onSetFormData({ resultSummary: e.target.value })}
        />
      </div>

      {isConfirmed ? null : (
        <div className="flex justify-end gap-2 pt-2">
          {canDelete && isEdit ? (
            <Button
              variant="ghost"
              type="button"
              className={`h-10 text-sm ${STYLE.btnDangerGhost} mr-auto`}
              onClick={onDeleteClick}
              disabled={isDeleting}
            >
              <Trash2 className={`mr-1.5 ${ICON.action}`} />
              {isDeleting ? "削除中..." : "削除"}
            </Button>
          ) : null}
          <Button variant="outline" type="button" onClick={onBack} className="h-10 text-sm">キャンセル</Button>
          {canSubmit ? (
            <SubmitButton
              className={`${C.bgAccent} ${C.bgAccentHover} text-white h-10 text-sm`}
            >
              保存
            </SubmitButton>
          ) : null}
        </div>
      )}
    </div>
  );
});

export function ExaminationForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const medicalRecordId = searchParams.get("medicalRecordId");
  const { canEdit, canCreate, canDelete } = usePermission("examinations");

  const { data: examTypesRaw } = useMasterItems("examination");
  const { data: staffListRaw } = useMasterItems("staff");
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

    const PRIORITY_FIELDS = ["testTypeId", "doctorId"];
    const firstError = PRIORITY_FIELDS.find((f) => errorFields.includes(f)) || errorFields[0];

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
      const lower = deferredHistorySearch.toLowerCase();
      result = result.filter(
        (e) =>
          e.testType.toLowerCase().includes(lower) ||
          (e.resultSummary ?? "").toLowerCase().includes(lower),
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
        <fieldset disabled={!canSubmit} className="contents">
        <div className="flex flex-col gap-4">
          {/* rerender-memo: PatientInfoCard — フォームフィールド変更では再レンダーしない */}
          {selectedPet ? (
            <PatientInfoCard
              ownerName={selectedPet.ownerName}
              petName={`${selectedPet.name}${selectedPet.species ? `(${selectedPet.species})` : ""}`}
              petNumber={selectedPet.petNumber || selectedPet.id}
              weight={selectedPet.weight || "-"}
              staffName="医師A"
              serviceType="検査"
              petDetails={`${selectedPet.birthDate ? `${selectedPet.birthDate}生` : ""} / ${selectedPet.species}`}
              insuranceName={selectedPet.insuranceName || "保険情報未登録"}
              insuranceDetails={selectedPet.insuranceDetails || "-"}
              nextVisitDate="-"
              nextVisitContent="-"
            />
          ) : null}

          {/* 2カラムレイアウト: 左 3/5（フォーム）・右 2/5（履歴） */}
          <div className="grid grid-cols-1 lg:grid-cols-5 gap-4 items-start">
            {/* 左カラム: フォームフィールド */}
            <div className="lg:col-span-3">
              <FormFieldsSection
                formData={formData}
                examTypes={examTypes}
                staffList={staffList}
                isEdit={isEdit}
                isSaving={isSaving}
                isDeleting={isDeleting}
                isConfirmed={isConfirmed}
                canEdit={canEdit}
                canCreate={canCreate}
                canDelete={canDelete}
                onSetFormData={handleSetFormData}
                onBack={handleBack}
                onDeleteClick={handleDeleteClick}
              />
            </div>

            {/* 右カラム: 過去の検査履歴 */}
            <div className="lg:col-span-2 space-y-3">
              <h3 className={`text-sm font-medium ${C.text60} px-1`}>過去の検査履歴</h3>
              <HistoryFilterPanel
                showDateRange={true}
                filterStartDate={historyStartDate}
                onFilterStartDateChange={setHistoryStartDate}
                filterEndDate={historyEndDate}
                onFilterEndDateChange={setHistoryEndDate}
                searchTerm={historySearchTerm}
                onSearchTermChange={setHistorySearchTerm}
                searchPlaceholder="検査種別・所見で検索..."
                sortOrder={historySortOrder}
                onSortOrderChange={setHistorySortOrder}
                onClear={handleHistoryClear}
              />
              <div className="space-y-2 max-h-[600px] overflow-y-auto">
                {filteredHistory.length > 0 ? (
                  filteredHistory.map((exam) => (
                    <ExaminationCard key={exam.id} examination={exam} />
                  ))
                ) : (
                  <p className={`text-sm ${C.text45} text-center py-6`}>
                    {currentPetId ? "検査記録がありません" : "ペットを選択してください"}
                  </p>
                )}
              </div>
            </div>
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
