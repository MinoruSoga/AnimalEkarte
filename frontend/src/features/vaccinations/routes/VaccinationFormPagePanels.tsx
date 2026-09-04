import { Trash2 } from "lucide-react";

import { ErrorFallback, LoadingFallback } from "@/components/shared/DataStates";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PatientInfoCard, formatPatientPetDetails } from "@/components/shared/PatientInfoCard";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Button } from "@/components/ui/button";
import { C, STYLE, ICON, LAYOUT } from "@/lib/design-tokens";
import { ResourceVaccinations } from "@/types/generated/models";
import type { SortOrder } from "@/types";
import type { VaccinationRecord } from "../api/transforms";
import {
  VaccinationFieldsPanel,
  VaccinationHistoryPanel,
} from "../components/VaccinationFormPanels";
import type { VaccinationFormGate } from "./vaccination-form-model";

export function VaccinationFormStatusView({
  gate,
  onBack,
}: {
  gate: VaccinationFormGate;
  onBack: () => void;
}) {
  if (gate.kind === "new-no-pet") {
    return (
      <div className={`flex items-center justify-center p-8 text-base ${C.text50}`}>
        <p>ペットを選択してください</p>
      </div>
    );
  }

  if (gate.kind === "edit-loading") {
    return (
      <PageLayout
        title="予防接種"
        resource={ResourceVaccinations}
        onBack={onBack}
        maxWidth={LAYOUT.pageContentMaxWidth.formMid}
      >
        <LoadingFallback />
      </PageLayout>
    );
  }

  if (gate.kind === "edit-not-found") {
    return (
      <PageLayout
        title="予防接種"
        resource={ResourceVaccinations}
        onBack={onBack}
        maxWidth={LAYOUT.pageContentMaxWidth.formMid}
      >
        <ErrorFallback message="予防接種が見つかりません" />
      </PageLayout>
    );
  }

  return (
    <PageLayout
      title="予防接種"
      resource={ResourceVaccinations}
      onBack={onBack}
      maxWidth={LAYOUT.pageContentMaxWidth.formMid}
    >
      <div className="space-y-3">
        <ErrorFallback message="予防接種の取得に失敗しました" />
        {gate.retryRead ? (
          <Button type="button" variant="outline" size="sm" onClick={gate.retryRead}>
            再試行
          </Button>
        ) : null}
      </div>
    </PageLayout>
  );
}

interface VaccinationFormFieldsState {
  doctorName: string;
  date: string;
  vaccineId: string;
  vaccineOptions: { value: string; label: string }[];
  supplemental: string;
  lot1: string;
  lot2: string;
  lot3: string;
  lot4: string;
  nextScheduleType: string;
  nextDate: string;
  remarks: string;
  setDate: (value: string) => void;
  setVaccineId: (value: string) => void;
  setSupplemental: (value: string) => void;
  setLot1: (value: string) => void;
  setLot2: (value: string) => void;
  setLot3: (value: string) => void;
  setLot4: (value: string) => void;
  setNextScheduleType: (value: string) => void;
  setNextDate: (value: string) => void;
  setRemarks: (value: string) => void;
}

interface VaccinationHistoryFilterState {
  filterStartDate: string;
  filterEndDate: string;
  historySearchTerm: string;
  sortOrder: SortOrder;
  setFilterStartDate: (value: string) => void;
  setFilterEndDate: (value: string) => void;
  setHistorySearchTerm: (value: string) => void;
  setSortOrder: (value: SortOrder) => void;
  handleClearHistoryFilter: () => void;
}

interface VaccinationPatient {
  ownerName: string;
  name: string;
  petNumber?: string;
  weight?: string;
  species?: string;
  birthDate?: string;
  gender?: string;
  neuteredDate?: string;
  status?: string;
  insuranceName?: string;
  insuranceDetails?: string;
}

interface VaccinationFormBodyProps {
  isEdit: boolean;
  canSubmit: boolean;
  canDelete: boolean | undefined;
  isPetDeceased: boolean;
  isDeleting: boolean;
  isDirty: boolean;
  isSaving: boolean;
  selectedPet: VaccinationPatient | undefined;
  form: VaccinationFormFieldsState;
  fieldErrors: Record<string, string>;
  petHistory: VaccinationRecord[];
  historyFilter: VaccinationHistoryFilterState;
  deleteConfirmOpen: boolean;
  formAction: (payload: FormData) => void;
  onBack: () => void;
  onMarkDirty: () => void;
  onOpenDeleteConfirm: () => void;
  onCloseDeleteConfirm: () => void;
  onConfirmDelete: () => void;
}

export function VaccinationFormBody({
  isEdit,
  canSubmit,
  canDelete,
  isPetDeceased,
  isDeleting,
  isDirty,
  isSaving,
  selectedPet,
  form,
  fieldErrors,
  petHistory,
  historyFilter,
  deleteConfirmOpen,
  formAction,
  onBack,
  onMarkDirty,
  onOpenDeleteConfirm,
  onCloseDeleteConfirm,
  onConfirmDelete,
}: VaccinationFormBodyProps) {
  return (
    <form action={formAction}>
      <PageLayout
        title={isEdit ? "予防接種詳細・編集" : "新規予防接種登録"}
        resource={ResourceVaccinations}
        onBack={onBack}
        maxWidth={LAYOUT.pageContentMaxWidth.formMid}
        headerAction={
          <div className="flex gap-2">
            {canDelete && isEdit && !isPetDeceased ? (
              <Button
                variant="ghost"
                type="button"
                className={`${STYLE.btnDangerGhost} px-4 h-10 text-sm`}
                onClick={onOpenDeleteConfirm}
                disabled={isDeleting}
              >
                <Trash2 className={`mr-1.5 ${ICON.action}`} />
                {isDeleting ? "削除中..." : "削除"}
              </Button>
            ) : null}
            {canSubmit ? <SubmitButton className="px-6 h-10 text-sm">保存</SubmitButton> : null}
          </div>
        }
      >
        <NavigationBlocker when={isDirty ? !isSaving : false} />

        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
          {selectedPet ? (
            <PatientInfoCard
              ownerName={selectedPet.ownerName}
              petName={selectedPet.name}
              petNumber={selectedPet.petNumber ?? ""}
              weight={selectedPet.weight ?? ""}
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
          {isPetDeceased ? (
            <div
              role="status"
              aria-label="死亡ペットのため保存不可"
              className={`flex items-center gap-2 px-4 py-2.5 rounded-md border mt-4 ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
            >
              <span className="text-sm font-medium">
                死亡したペットの予防接種記録は保存できません
              </span>
            </div>
          ) : null}

          <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
            <VaccinationFieldsPanel
              doctorName={form.doctorName}
              date={form.date}
              vaccineId={form.vaccineId}
              vaccineOptions={form.vaccineOptions}
              supplemental={form.supplemental}
              lot1={form.lot1}
              lot2={form.lot2}
              lot3={form.lot3}
              lot4={form.lot4}
              nextScheduleType={form.nextScheduleType}
              nextDate={form.nextDate}
              remarks={form.remarks}
              fieldErrors={fieldErrors}
              onDateChange={form.setDate}
              onVaccineIdChange={form.setVaccineId}
              onSupplementalChange={form.setSupplemental}
              onLot1Change={form.setLot1}
              onLot2Change={form.setLot2}
              onLot3Change={form.setLot3}
              onLot4Change={form.setLot4}
              onNextScheduleTypeChange={form.setNextScheduleType}
              onNextDateChange={form.setNextDate}
              onRemarksChange={form.setRemarks}
              onMarkDirty={onMarkDirty}
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
          onClose={onCloseDeleteConfirm}
          title="削除確認"
          description="この予防接種情報を削除してもよろしいですか？"
          confirmLabel="削除"
          variant="destructive"
          onConfirm={onConfirmDelete}
        />
      </PageLayout>
    </form>
  );
}
