// FE-RC-045: trimming-form-panels.tsx (491行) から患者情報+3カラム部分を分離。
import type { ChangeEvent } from "react";
import { PatientInfoCard, formatPatientPetDetails } from "@/components/shared/PatientInfoCard";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { formatDate } from "@/lib/format/date";
import type { SortOrder } from "@/types";
import type { TrimmingFormData } from "@/types/trimming";
import {
  TrimmingLeftColumn,
  TrimmingMiddleColumn,
  TrimmingRightColumn,
} from "../lib/trimming-form-columns";
import type { TrimmingHistoryItem } from "../lib/trimming-form-column-types";
import type { TrimmingSelectableItem } from "./trimming-form-model";

export interface TrimmingPatient {
  ownerName: string;
  name: string;
  petNumber?: string;
  weight?: string;
  species?: string;
  birthDate?: string;
  gender?: string;
  neuteredDate?: string;
  insuranceName?: string;
  insuranceDetails?: string;
  status?: string;
}

interface TrimmingFormColumnsProps {
  selectedPet: TrimmingPatient;
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
  showInitialStatusSelector: boolean;
  onFormChange: (updates: Partial<TrimmingFormData>) => void;
  onOpenCourseModal: () => void;
  onOpenStaffModal: () => void;
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

export function TrimmingFormColumns({
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
  showInitialStatusSelector,
  onFormChange,
  onOpenCourseModal,
  onOpenStaffModal,
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
}: TrimmingFormColumnsProps) {
  return (
    <div className="space-y-6">
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
        onStaffClick={onOpenStaffModal}
      />
      {fieldErrors.staffId ? <FormFieldError message={fieldErrors.staffId} /> : null}
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
            onFormChange={onFormChange}
            onCourseModalOpen={onOpenCourseModal}
            onStyleImageChange={onStyleImageChange}
            onRemoveStyleImage={onRemoveStyleImage}
            courseError={fieldErrors.courseId}
            showInitialStatusSelector={showInitialStatusSelector}
          />
          <TrimmingMiddleColumn
            formData={formData}
            completedImagePreview={completedImagePreview}
            onFormChange={onFormChange}
            onCompletedImageChange={onCompletedImageChange}
            onRemoveCompletedImage={onRemoveCompletedImage}
          />
        </div>
        <TrimmingRightColumn
          sortedHistory={sortedHistory}
          isHistoryLoading={isHistoryLoading}
          historySearchTerm={historySearchTerm}
          historySortOrder={historySortOrder}
          historyDateRange={historyDateRange}
          onSearchTermChange={onHistorySearchTermChange}
          onSortOrderChange={onHistorySortOrderChange}
          onClear={onHistoryClear}
          onFilterStartDateChange={onHistoryStartDateChange}
          onFilterEndDateChange={onHistoryEndDateChange}
          onHistoryClick={onHistoryClick}
        />
      </div>
    </div>
  );
}
