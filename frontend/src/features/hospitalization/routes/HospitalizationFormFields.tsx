import { useCallback } from "react";
import { MessageSquare, AlertCircle } from "lucide-react";

import { formatDate } from "@/lib/format/date";
import { PatientInfoCard, formatPatientPetDetails } from "@/components/shared/PatientInfoCard";
import { PastRecordHistoryPanel } from "@/components/shared/PastRecordHistoryPanel";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { C } from "@/lib/design-tokens";
import type { MasterItem } from "@/types";
import type { HospitalizationTreatmentPlan } from "@/types";
import { HospitalizationBasicInfo } from "../components/HospitalizationBasicInfo";
import { HospitalizationNoteCard } from "../components/HospitalizationNoteCard";
import { HospitalizationTreatmentTable } from "../components/HospitalizationTreatmentTable";
import { HospitalizationCostSummary } from "../components/HospitalizationCostSummary";
import { H_STYLES } from "../lib/styles";
import type { HospitalizationFormData } from "../types";

interface HospitalizationPatient {
  id?: string;
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

export interface HospitalizationFormFieldsProps {
  selectedPet: HospitalizationPatient | undefined;
  formData: HospitalizationFormData;
  fieldErrors: Record<string, string> | undefined;
  cageItems: MasterItem[] | undefined;
  isEdit: boolean;
  canDelete: boolean | undefined;
  hasChildTreatmentPlans: boolean;
  treatmentPlans: HospitalizationTreatmentPlan[];
  totals: {
    subtotalBeforeDiscount: number;
    subtotalAfterDiscount: number;
    consumptionTax: number;
    total: number;
  };
  historyItems: { id: string; date: string; title: string; subtitle?: string }[];
  isHistoryLoading: boolean;
  canSubmit: boolean;
  onFormChange: (updates: Partial<HospitalizationFormData>) => void;
  onOpenStaffModal: () => void;
  onAddTreatmentPlan: () => void;
  onUpdateTreatmentPlan: (
    id: string,
    field: keyof HospitalizationTreatmentPlan,
    value: string | number | boolean,
  ) => void;
  onRemoveTreatmentPlan: (id: string) => void;
}

// FE-RC-085: ネスト三項を早期return関数へ分解。
function resolveNextVisitDate(formData: HospitalizationFormData): string | undefined {
  if (formData.nextVisit) return formatDate(formData.nextVisit);
  if (formData.endDate) return formatDate(formData.endDate);
  return undefined;
}

function resolveNextVisitContent(formData: HospitalizationFormData): string | undefined {
  if (formData.nextVisit) return "次回来院";
  if (formData.endDate) return "退院予定";
  return undefined;
}

export function HospitalizationFormFields({
  selectedPet,
  formData,
  fieldErrors,
  cageItems,
  isEdit,
  canDelete,
  hasChildTreatmentPlans,
  treatmentPlans,
  totals,
  historyItems,
  isHistoryLoading,
  canSubmit,
  onFormChange,
  onOpenStaffModal,
  onAddTreatmentPlan,
  onUpdateTreatmentPlan,
  onRemoveTreatmentPlan,
}: HospitalizationFormFieldsProps) {
  // rerender-memo: HospitalizationNoteCard は memo。onChange を inline arrow で
  // 渡すと毎レンダー新規参照になり memo が無効化されるため useCallback で安定化する。
  const handleOwnerRequestChange = useCallback(
    (val: string) => onFormChange({ ownerRequest: val }),
    [onFormChange],
  );
  const handleStaffNotesChange = useCallback(
    (val: string) => onFormChange({ staffNotes: val }),
    [onFormChange],
  );

  return (
    <>
      {selectedPet ? (
        <PatientInfoCard
          ownerName={selectedPet.ownerName}
          petName={selectedPet.name}
          petNumber={selectedPet.petNumber || selectedPet.id || ""}
          weight={selectedPet.weight || "-"}
          staffName={formData.doctorName || "未設定"}
          staffLabel="担当医"
          staffButtonId="doctor_id"
          reservationType={formData.hospitalizationType}
          petDetails={formatPatientPetDetails({
            species: selectedPet.species,
            birthDate: selectedPet.birthDate,
            gender: selectedPet.gender,
            neuteredDate: selectedPet.neuteredDate,
          })}
          insuranceName={selectedPet.insuranceName}
          insuranceDetails={selectedPet.insuranceDetails}
          status={selectedPet.status === "死亡" ? "deceased" : "alive"}
          nextVisitDate={resolveNextVisitDate(formData)}
          nextVisitContent={resolveNextVisitContent(formData)}
          onStaffClick={canSubmit ? onOpenStaffModal : undefined}
        />
      ) : null}
      <FormFieldError message={fieldErrors?.pet} />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-5 mb-3">
        <div className="space-y-3 lg:col-span-3">
          <HospitalizationBasicInfo
            formData={formData}
            onChange={onFormChange}
            cageItems={cageItems ?? []}
            cageIdError={fieldErrors?.cage_id}
          />

          <HospitalizationNoteCard
            id="owner_request"
            title="主訴"
            icon={MessageSquare}
            value={formData.ownerRequest}
            onChange={handleOwnerRequestChange}
            placeholder="主訴を入力..."
          />

          <HospitalizationNoteCard
            id="staff_notes"
            title="スタッフへの連絡事項"
            icon={AlertCircle}
            value={formData.staffNotes}
            onChange={handleStaffNotesChange}
            placeholder="連絡事項を入力..."
          />

          {isEdit ? (
            <p className={`mb-2 ${H_STYLES.text.sm} ${C.text60}`}>
              登録時の治療プランはスナップショットとして参照のみです。この画面では変更・削除できません。入院中の投薬・給餌などは入院詳細のケアプランで管理します。
            </p>
          ) : (
            <p className={`mb-2 ${H_STYLES.text.sm} ${C.text60}`}>
              治療内容・メモが入力された行のみ、入院登録時に治療プラン（登録時スナップショット）として保存されます。空行は保存されません。
            </p>
          )}
          {hasChildTreatmentPlans ? (
            <p className={`mb-2 ${H_STYLES.text.sm} ${C.text60}`} role="status">
              治療プランが紐付いているため、この入院は削除できません。
            </p>
          ) : null}
          <HospitalizationTreatmentTable
            treatmentPlans={treatmentPlans}
            onAdd={onAddTreatmentPlan}
            onUpdate={onUpdateTreatmentPlan}
            onRemove={canDelete ? onRemoveTreatmentPlan : undefined}
            readOnly={isEdit}
          />

          <p className={`mb-2 ${H_STYLES.text.sm} ${C.text60}`}>
            一括割引（%／円）はこの画面では利用できません。表示金額は治療プラン明細に基づく概算です。
          </p>
          <HospitalizationCostSummary totals={totals} />
        </div>
        <PastRecordHistoryPanel
          title="過去の入院履歴"
          searchPlaceholder="種別・ステータスで検索..."
          items={historyItems}
          isLoading={isHistoryLoading}
        />
      </div>
    </>
  );
}
