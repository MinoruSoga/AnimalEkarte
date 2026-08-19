// React/Framework
import { memo, useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router";

// External
import { Trash2 } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { PatientInfoCard, formatPatientPetDetails } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { C, STYLE, ICON, LAYOUT } from "@/lib/design-tokens";
import { normalizeKana } from "@/lib/normalize-kana";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { useGetVaccinations } from "../api/get-vaccinations";

// Relative
import { VaccinationFieldsPanel, VaccinationHistoryPanel } from "../components/VaccinationFormPanels";
import { useVaccinationForm } from "../hooks/use-vaccination-form";
import { usePermission } from "@/hooks/use-permission";
import { ResourceVaccinations } from "@/types/generated/models";

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
  const canSubmit = id ? canEdit : canCreate;

  const {
    isEdit,
    isReadLoading,
    isReadNotFound,
    isReadError,
    retryRead,
    petSelection,
    form,
    formAction,
    formState,
    isSaving,
    fieldErrors,
    handleDelete,
    isDeleting,
    historyFilter,
  } = useVaccinationForm(id, { canCreate, canEdit, canDelete });

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
    doctorName,
    date, setDate,
    vaccineId, setVaccineId, vaccineOptions,
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

  // --- 履歴セクション (BUG-007) ---
  // Server-side pet_id filter: unscoped page1 + client filter missed 2026 rows
  // behind 2029 seed dates. Key includes petId so caches never cross pets.
  const historyPetId = selectedPet?.id;
  const { data: petVaccinations = [] } = useGetVaccinations({
    // Always pass petId key so query never falls back to unscoped page-window list.
    petId: historyPetId ?? "",
  });

  // rerender-dependencies: オブジェクト参照ではなく primitive を deps に渡す
  const { historySearchTerm, filterStartDate, filterEndDate, sortOrder } = historyFilter;

  const petHistory = useMemo(() => {
    if (!historyPetId) return [];

    // Server already scoped to pet; still exclude the open edit record.
    let result = petVaccinations.filter((v) => v.id !== id);

    // キーワード検索
    const term = normalizeKana(historySearchTerm).toLowerCase();
    if (term) {
      result = result.filter((v) =>
        normalizeKana(v.vaccineName).toLowerCase().includes(term),
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
  }, [petVaccinations, historyPetId, id, historySearchTerm, filterStartDate, filterEndDate, sortOrder]);

  if (!selectedPet && !isEdit) {
    return (
      <div className={`flex items-center justify-center p-8 text-base ${C.text50}`}>
        <p>ペットを選択してください</p>
      </div>
    );
  }

  // BUG-016: never render blank editable form for missing / other-clinic / forbidden IDs
  if (isEdit && isReadLoading) {
    return (
      <PageLayout
        title="予防接種"
        resource={ResourceVaccinations}
        onBack={handleBack}
        maxWidth={LAYOUT.pageContentMaxWidth.formMid}
      >
        <LoadingFallback />
      </PageLayout>
    );
  }
  if (isEdit && isReadNotFound) {
    return (
      <PageLayout
        title="予防接種"
        resource={ResourceVaccinations}
        onBack={handleBack}
        maxWidth={LAYOUT.pageContentMaxWidth.formMid}
      >
        <ErrorFallback message="予防接種が見つかりません" />
      </PageLayout>
    );
  }
  if (isEdit && isReadError) {
    return (
      <PageLayout
        title="予防接種"
        resource={ResourceVaccinations}
        onBack={handleBack}
        maxWidth={LAYOUT.pageContentMaxWidth.formMid}
      >
        <div className="space-y-3">
          <ErrorFallback message="予防接種の取得に失敗しました" />
          {retryRead ? (
            <Button type="button" variant="outline" size="sm" onClick={retryRead}>
              再試行
            </Button>
          ) : null}
        </div>
      </PageLayout>
    );
  }

  return (
    <form action={formAction}>
      <PageLayout
        title={isEdit ? "予防接種詳細・編集" : "新規予防接種登録"}
        resource={ResourceVaccinations}
        onBack={handleBack}
        maxWidth={LAYOUT.pageContentMaxWidth.formMid}
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
                className="px-6 h-10 text-sm"
              >
                保存
              </SubmitButton>
            ) : null}
          </div>
        }
      >
        {/* FE6-8: jsx-no-leaked-render は非型認識のため isDirty を boolean と静的に断定できず !! で明示する */}
        <NavigationBlocker when={!!isDirty && !isSaving} />

        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
        {selectedPet ? (
          <PatientInfoCard
            ownerName={selectedPet.ownerName}
            petName={selectedPet.name}
            petNumber={selectedPet.petNumber ?? ""}
            weight={selectedPet.weight ?? ""}
            // BUG-006: 対象ペットの年齢・性別・去勢避妊を渡し、固定デフォルトを使わない
            petDetails={formatPatientPetDetails({
              species: selectedPet.species,
              birthDate: selectedPet.birthDate,
              gender: selectedPet.gender,
              neuteredDate: selectedPet.neuteredDate,
            })}
            status={selectedPet.status === "死亡" ? "deceased" : "alive"}
            insuranceName={selectedPet.insuranceName}
            insuranceDetails={selectedPet.insuranceDetails}
          />
        ) : null}

        <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
          <VaccinationFieldsPanel
            doctorName={doctorName}
            date={date}
            vaccineId={vaccineId}
            vaccineOptions={vaccineOptions}
            supplemental={supplemental}
            lot1={lot1}
            lot2={lot2}
            lot3={lot3}
            lot4={lot4}
            nextScheduleType={nextScheduleType}
            nextDate={nextDate}
            remarks={remarks}
            fieldErrors={fieldErrors}
            onDateChange={setDate}
            onVaccineIdChange={setVaccineId}
            onSupplementalChange={setSupplemental}
            onLot1Change={setLot1}
            onLot2Change={setLot2}
            onLot3Change={setLot3}
            onLot4Change={setLot4}
            onNextScheduleTypeChange={setNextScheduleType}
            onNextDateChange={setNextDate}
            onRemarksChange={setRemarks}
            onMarkDirty={markDirty}
          />
          <VaccinationHistoryPanel
            petHistory={petHistory}
            filterStartDate={historyFilter.filterStartDate}
            filterEndDate={historyFilter.filterEndDate}
            historySearchTerm={historyFilter.historySearchTerm}
            sortOrder={historyFilter.sortOrder}
            onFilterStartDateChange={historyFilter.setFilterStartDate}
            onFilterEndDateChange={historyFilter.setFilterEndDate}
            onHistorySearchTermChange={historyFilter.setHistorySearchTerm}
            onSortOrderChange={historyFilter.setSortOrder}
            onClear={historyFilter.handleClearHistoryFilter}
          />
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
