// React/Framework
import { memo, useMemo } from "react";

// Internal
import { CharCountTextarea } from "@/components/shared/CharCountTextarea";
import { MasterLink } from "@/components/shared/MasterLink";
import { C, STYLE } from "@/lib/design-tokens";
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";

// Relative
import { useGetDiagnosisTypes, useGetDiagnosisNames } from "../../api/get-diagnosis-options";

// ── Types ─────────────────────────────────────────────────────────────

interface ClinicalPlanSectionProps {
  /** 互換用（親が record id を渡す）。controlled 化後はデータ取得に使わない。 */
  medicalRecordId: string;
  canEdit?: boolean;
  /** P2-15 互換。controlled 化後はセクション内で子 API を呼ばない。 */
  recordClinicId?: string;
  // BUG-010: controlled — 親 form が state owner。独自 save / local state は持たない。
  physicalExam: string;
  onPhysicalExamChange: (value: string) => void;
  diagnosisDetails: string;
  onDiagnosisDetailsChange: (value: string) => void;
  treatmentPolicy: string;
  onTreatmentPolicyChange: (value: string) => void;
  diagnosisTypeId: number | null;
  onDiagnosisTypeIdChange: (id: number | null) => void;
  diagnosisNameId: number | null;
  onDiagnosisNameIdChange: (id: number | null) => void;
}

// ── Component ─────────────────────────────────────────────────────────

export const ClinicalPlanSection = memo(function ClinicalPlanSection({
  medicalRecordId: _medicalRecordId,
  recordClinicId: _recordClinicId,
  canEdit = false,
  physicalExam,
  onPhysicalExamChange,
  diagnosisDetails,
  onDiagnosisDetailsChange,
  treatmentPolicy,
  onTreatmentPolicyChange,
  diagnosisTypeId,
  onDiagnosisTypeIdChange,
  diagnosisNameId,
  onDiagnosisNameIdChange,
}: ClinicalPlanSectionProps) {
  void _medicalRecordId;
  void _recordClinicId;
  const { data: diagnosisTypes = [], isLoading: isTypesLoading } = useGetDiagnosisTypes();
  const { data: diagnosisNames = [], isLoading: isNamesLoading } =
    useGetDiagnosisNames(diagnosisTypeId);

  const typeOptions = useMemo<SearchableSelectOption[]>(
    () => diagnosisTypes.map((t) => ({ value: String(t.id), label: t.name })),
    [diagnosisTypes],
  );
  const nameOptions = useMemo<SearchableSelectOption[]>(
    () => diagnosisNames.map((n) => ({ value: String(n.id), label: n.name })),
    [diagnosisNames],
  );

  return (
    <div className={`${C.bgWhite} border ${C.borderMedium} rounded-xs p-4`}>
      <h2 className={`text-sm font-bold ${C.text} mb-4`}>診察所見・診断・治療方針</h2>

      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel} htmlFor="clinical-plan-physical-exam">
            身体検査所見
          </label>
          <CharCountTextarea
            id="clinical-plan-physical-exam"
            value={physicalExam}
            onChange={onPhysicalExamChange}
            placeholder="身体検査所見を入力してください"
            textareaClassName={`min-h-[100px] ${C.text} text-sm`}
            disabled={!canEdit}
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <div className="flex items-center justify-between">
            <label className={STYLE.formLabel}>診断カテゴリ</label>
            <MasterLink category="diagnosis_type" label="編集" className="text-2xs" />
          </div>
          <SearchableSelect
            value={diagnosisTypeId ? String(diagnosisTypeId) : ""}
            onValueChange={(value) => {
              onDiagnosisTypeIdChange(value ? Number(value) : null);
              onDiagnosisNameIdChange(null);
            }}
            options={typeOptions}
            disabled={isTypesLoading || !canEdit}
            placeholder={isTypesLoading ? "読み込み中..." : "カテゴリを選択"}
            searchPlaceholder="カテゴリを検索..."
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <div className="flex items-center justify-between">
            <label className={STYLE.formLabel}>診断病名</label>
            <MasterLink category="diagnosis_name" label="編集" className="text-2xs" />
          </div>
          <SearchableSelect
            value={diagnosisNameId ? String(diagnosisNameId) : ""}
            onValueChange={(value) => onDiagnosisNameIdChange(value ? Number(value) : null)}
            options={nameOptions}
            disabled={isNamesLoading || !diagnosisTypeId || !canEdit}
            placeholder={
              isNamesLoading
                ? "読み込み中..."
                : !diagnosisTypeId
                  ? "先にカテゴリを選択"
                  : "病名を選択"
            }
            searchPlaceholder="病名を検索..."
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel} htmlFor="clinical-plan-diagnosis-details">
            診断詳細
          </label>
          <CharCountTextarea
            id="clinical-plan-diagnosis-details"
            value={diagnosisDetails}
            onChange={onDiagnosisDetailsChange}
            placeholder="診断詳細を入力してください"
            textareaClassName={`min-h-[100px] ${C.text} text-sm`}
            disabled={!canEdit}
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel} htmlFor="clinical-plan-treatment-policy">
            治療方針
          </label>
          <CharCountTextarea
            id="clinical-plan-treatment-policy"
            value={treatmentPolicy}
            onChange={onTreatmentPolicyChange}
            placeholder="治療方針を入力してください"
            textareaClassName={`min-h-[100px] ${C.text} text-sm`}
            disabled={!canEdit}
          />
        </div>
      </div>
    </div>
  );
});
