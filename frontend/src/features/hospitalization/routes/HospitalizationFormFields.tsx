import { useCallback } from "react";
import { MessageSquare, AlertCircle } from "lucide-react";

import { formatDate } from "@/lib/format/date";
import { paths } from "@/config/paths";
import { PatientInfoCard, formatPatientPetDetails } from "@/components/shared/PatientInfoCard";
import { PastRecordHistoryPanel } from "@/components/shared/PastRecordHistoryPanel";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { C } from "@/lib/design-tokens";
import { isPetDeceasedForClinicalWrite } from "@/lib/transforms/pet";
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
  /** EMR-174: 詳細 deep link 用（選択 Pet 由来） */
  ownerId?: string;
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
  /** 死亡OR契約の deceased_at 側。Pet transform が供給する。 */
  deceasedAt?: string | null;
  /** スタッフ向け飼主特記マーク (EMR-173/231)。Pet transform が供給する。 */
  ownerIsDangerous?: boolean;
  /** ペット特記レベル (表示値 "高"/"中"/"低")。高/中のみ Popover バッジ。 */
  dangerLevel?: string;
  dangerReason?: string;
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

// EMR-412 X4/R2: 補足文言は見出しと重複するため出さない。
// 実 nextVisit は「次回予定」、endDate フォールバックは「退院予定」の見出しで区別する。
function resolveNextVisitLabel(formData: HospitalizationFormData): string {
  return formData.nextVisit ? "次回予定" : "退院予定";
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
          ownerDetailHref={
            selectedPet.ownerId ? paths.owners.detail.getHref(selectedPet.ownerId) : undefined
          }
          petDetailHref={
            selectedPet.ownerId && selectedPet.id
              ? paths.owners.detail.pet.getHref(selectedPet.ownerId, selectedPet.id)
              : undefined
          }
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
          ownerIsDangerous={selectedPet.ownerIsDangerous}
          petDangerLevel={selectedPet.dangerLevel}
          petDangerReason={selectedPet.dangerReason}
          status={isPetDeceasedForClinicalWrite(selectedPet) ? "deceased" : "alive"}
          birthDate={selectedPet.birthDate}
          nextVisitDate={resolveNextVisitDate(formData)}
          nextVisitLabel={resolveNextVisitLabel(formData)}
          nextVisitShowAlert={false}
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
