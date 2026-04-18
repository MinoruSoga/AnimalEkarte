// React/Framework
import { memo, useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router";

// External
import { Trash2 } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { HistoryFilterPanel } from "@/components/shared/HistoryFilterPanel";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { MasterLink } from "@/components/shared/MasterLink";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { VaccinationCard } from "../components/VaccinationCard";
import { useGetVaccinations } from "../api/get-vaccinations";
import type { SortOrder } from "@/types";

// Relative
import { useVaccinationForm } from "../hooks/use-vaccination-form";
import { usePermission } from "@/hooks/use-permission";
import { ResourceVaccinations } from "@/types/generated/models";

// rendering-hoist-jsx: 静的SelectItem定数をモジュールスコープに巻き上げ
const VACCINE_TYPE_ITEMS = (
  <>
    <SelectItem value="1">混合ワクチン</SelectItem>
    <SelectItem value="2">狂犬病ワクチン</SelectItem>
  </>
);

const NEXT_SCHEDULE_ITEMS = (
  <>
    <SelectItem value="3weeks">3週後</SelectItem>
    <SelectItem value="4weeks">4週後</SelectItem>
    <SelectItem value="1year">1年後</SelectItem>
    <SelectItem value="custom">以外（手動）</SelectItem>
  </>
);

// rendering-hoist-jsx: アクセシビリティ用定数をモジュールレベルに巻き上げ（毎レンダー再生成を回避）
const VACCINATION_PRIORITY_FIELDS = ["date", "vaccineId"] as const;
const VACCINATION_FIELD_ID_MAP: Record<string, string> = {
  date: "vaccination-date",
  vaccineId: "vaccine-select",
};

export const VaccinationForm = memo(function VaccinationForm() {
  const navigate = useNavigate();
  const { id } = useParams();
  const { canEdit, canCreate, canDelete } = usePermission("vaccinations");

  const {
    isEdit,
    petSelection,
    form,
    formAction,
    formState,
    isSaving,
    fieldErrors,
    handleDelete,
    isDeleting,
    historyFilter,
  } = useVaccinationForm(id);

  const canSubmit = isEdit ? canEdit : canCreate;

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  // --- Focus Management (Accessibility) ---
  useEffect(() => {
    const errorFields = Object.keys(formState.fieldErrors || {});
    if (errorFields.length === 0) return;

    const firstError = VACCINATION_PRIORITY_FIELDS.find((f) => errorFields.includes(f)) || errorFields[0];
    const targetId = VACCINATION_FIELD_ID_MAP[firstError] || firstError;

    const element = document.getElementById(targetId);
    if (element) {
      element.focus();
      element.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }, [formState.fieldErrors, formState.timestamp]);

  // React 19 Action の成功を検知して遷移
  useEffect(() => {
    if (formState.success) {
      markClean();
      navigate(paths.vaccinations.getHref());
    }
  }, [formState.success, formState.timestamp, navigate, markClean]);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];

  const {
    date, setDate,
    vaccineId, setVaccineId,
    nextScheduleType, setNextScheduleType,
    nextDate, setNextDate,
    supplemental, setSupplemental,
    lot1, setLot1,
    lot2, setLot2,
    lot3, setLot3,
    lot4, setLot4,
    remarks, setRemarks,
  } = form;

  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  const handleBack = useCallback(() => {
    navigate(paths.vaccinations.getHref());
  }, [navigate]);

  // --- 履歴セクション ---
  const { data: allVaccinations = [] } = useGetVaccinations();

  // rerender-dependencies: オブジェクト参照ではなく primitive を deps に渡す
  const { historySearchTerm, filterStartDate, filterEndDate, sortOrder } = historyFilter;

  const petHistory = useMemo(() => {
    if (!selectedPet) return [];

    let result = allVaccinations.filter(
      (v) => v.petId === selectedPet.id && v.id !== id,
    );

    // キーワード検索
    const term = historySearchTerm.toLowerCase();
    if (term) {
      result = result.filter((v) =>
        v.vaccineName.toLowerCase().includes(term),
      );
    }

    // 日付フィルタ
    if (filterStartDate) {
      result = result.filter((v) => v.date >= filterStartDate);
    }
    if (filterEndDate) {
      result = result.filter((v) => v.date <= filterEndDate);
    }

    // ソート
    result = [...result].sort((a, b) =>
      sortOrder === "asc"
        ? a.date.localeCompare(b.date)
        : b.date.localeCompare(a.date),
    );

    return result;
  }, [allVaccinations, selectedPet, id, historySearchTerm, filterStartDate, filterEndDate, sortOrder]);

  if (!selectedPet && !isEdit) {
    return (
      <div className={`flex items-center justify-center p-8 text-base ${C.text50}`}>
        <p>ペットを選択してください</p>
      </div>
    );
  }

  return (
    <form action={formAction}>
      <PageLayout
        title={isEdit ? "予防接種詳細・編集" : "新規予防接種登録"}
        resource={ResourceVaccinations}
        onBack={handleBack}
        maxWidth="max-w-[1200px]"
        headerAction={
          <div className="flex gap-2">
            {canDelete && isEdit ? (
              <Button
                variant="ghost"
                type="button"
                className={`${STYLE.btnDangerGhost} px-4 h-10 text-sm`}
                onClick={() => setDeleteConfirmOpen(true)}
                disabled={isDeleting}
              >
                <Trash2 className={`mr-1.5 ${ICON.action}`} />
                {isDeleting ? "削除中..." : "削除"}
              </Button>
            ) : null}
            {canSubmit ? (
              <SubmitButton
                className={`${C.bgAccent} ${C.bgAccentHover} ${C.textWhite} shadow-sm px-6 h-10 text-sm`}
              >
                保存
              </SubmitButton>
            ) : null}
          </div>
        }
      >
        <NavigationBlocker when={isDirty && !isSaving} />

        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
        {selectedPet ? (
          <PatientInfoCard
            ownerName={selectedPet.ownerName}
            petName={selectedPet.name}
            petNumber={selectedPet.petNumber ?? ""}
            weight={selectedPet.weight ?? ""}
          />
        ) : null}

        <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
          {/* ===== 左カラム: 入力フォーム ===== */}
          <div className="lg:col-span-3">
            <div className={`${C.bgWhite} p-6 rounded-lg border ${C.borderLight} space-y-6`}>
              {/* 接種日 / ワクチン */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-2">
                  <Label htmlFor="vaccination-date">
                    接種日<span className={`${C.textRequired} ml-1`}>*</span>
                  </Label>
                  <NotionDatePicker
                    id="vaccination-date"
                    value={date}
                    onChange={(v) => { markDirty(); setDate(v); }}
                    disabledDays={{ after: new Date() }}
                  />
                  <FormFieldError message={fieldErrors.date} />
                </div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="vaccine-select">
                      ワクチン<span className={`${C.textRequired} ml-1`}>*</span>
                    </Label>
                    <MasterLink category="vaccine" label="編集" className="text-[11px]" />
                  </div>
                  <Select value={vaccineId} onValueChange={(v) => { markDirty(); setVaccineId(v); }}>
                    <SelectTrigger id="vaccine-select">
                      <SelectValue placeholder="選択してください" />
                    </SelectTrigger>
                    <SelectContent>{VACCINE_TYPE_ITEMS}</SelectContent>
                  </Select>
                  <FormFieldError message={fieldErrors.vaccineId} />
                </div>
              </div>

              {/* 補助説明 */}
              <div className="space-y-2">
                <Label>補助説明</Label>
                <Input
                  value={supplemental}
                  onChange={(e) => { markDirty(); setSupplemental(e.target.value); }}
                  placeholder="補助説明を入力"
                />
              </div>

              {/* LOT番号 1〜4 */}
              <div className="space-y-2">
                <Label>LOT番号</Label>
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1">
                    <Label className={`text-xs ${C.text60}`}>LOT 1</Label>
                    <Input
                      value={lot1}
                      onChange={(e) => { markDirty(); setLot1(e.target.value); }}
                      placeholder="LOT番号 1"
                      maxLength={50}
                    />
                  </div>
                  <div className="space-y-1">
                    <Label className={`text-xs ${C.text60}`}>LOT 2</Label>
                    <Input
                      value={lot2}
                      onChange={(e) => { markDirty(); setLot2(e.target.value); }}
                      placeholder="LOT番号 2"
                      maxLength={50}
                    />
                  </div>
                  <div className="space-y-1">
                    <Label className={`text-xs ${C.text60}`}>LOT 3</Label>
                    <Input
                      value={lot3}
                      onChange={(e) => { markDirty(); setLot3(e.target.value); }}
                      placeholder="LOT番号 3"
                      maxLength={50}
                    />
                  </div>
                  <div className="space-y-1">
                    <Label className={`text-xs ${C.text60}`}>LOT 4</Label>
                    <Input
                      value={lot4}
                      onChange={(e) => { markDirty(); setLot4(e.target.value); }}
                      placeholder="LOT番号 4"
                      maxLength={50}
                    />
                  </div>
                </div>
              </div>

              {/* 次回予定 */}
              <div className="space-y-2">
                <Label>次回の予定</Label>
                <div className="flex gap-3 items-center flex-wrap">
                  <Select
                    value={nextScheduleType}
                    onValueChange={(v) => { markDirty(); setNextScheduleType(v); }}
                  >
                    <SelectTrigger className="w-[130px]">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>{NEXT_SCHEDULE_ITEMS}</SelectContent>
                  </Select>
                  <NotionDatePicker
                    value={nextDate}
                    onChange={(v) => { markDirty(); setNextDate(v); }}
                    className="flex-1 min-w-[160px]"
                  />
                </div>
                {/* BUG-096: 過去日付エラーメッセージ */}
                <FormFieldError message={fieldErrors.nextDate} />
              </div>

              {/* 備考 */}
              <div className="space-y-2">
                <Label>備考</Label>
                <Textarea
                  value={remarks}
                  onChange={(e) => { markDirty(); setRemarks(e.target.value); }}
                  placeholder="備考を入力"
                  className="min-h-[100px]"
                />
              </div>
            </div>
          </div>

          {/* ===== 右カラム: 接種履歴 ===== */}
          <div className="lg:col-span-2 flex flex-col gap-3">
            <h3 className={`text-base font-semibold ${C.text}`}>過去の接種履歴</h3>

            <HistoryFilterPanel
              showDateRange
              filterStartDate={historyFilter.filterStartDate}
              onFilterStartDateChange={historyFilter.setFilterStartDate}
              filterEndDate={historyFilter.filterEndDate}
              onFilterEndDateChange={historyFilter.setFilterEndDate}
              searchTerm={historyFilter.historySearchTerm}
              onSearchTermChange={historyFilter.setHistorySearchTerm}
              searchPlaceholder="ワクチン名で検索..."
              sortOrder={historyFilter.sortOrder as SortOrder}
              onSortOrderChange={historyFilter.setSortOrder}
              onClear={historyFilter.handleClearHistoryFilter}
            />

            <div className="flex flex-col gap-2 overflow-y-auto max-h-[600px]">
              {petHistory.length === 0 ? (
                <p className={`text-sm ${C.text45} py-4 text-center`}>履歴がありません</p>
              ) : (
                petHistory.map((v) => (
                  <VaccinationCard key={v.id} vaccination={v} />
                ))
              )}
            </div>
          </div>
        </div>

        </fieldset>
        <ConfirmDialog
          open={deleteConfirmOpen}
          onClose={() => setDeleteConfirmOpen(false)}
          title="削除確認"
          description="この予防接種情報を削除してもよろしいですか？"
          confirmLabel="削除"
          variant="destructive"
          onConfirm={() => {
            handleDelete(() => {
              markClean();
              navigate(paths.vaccinations.getHref());
            });
          }}
        />
      </PageLayout>
    </form>
  );
});
