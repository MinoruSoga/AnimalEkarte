// React/Framework
import { memo, useState, useEffect, useCallback, useMemo } from "react";

// Internal
import { CharCountTextarea } from "@/components/shared/CharCountTextarea";
import { MasterLink } from "@/components/shared/MasterLink";
import { C, STYLE } from "@/lib/design-tokens";
import { LoadingFallback } from "@/components/shared/DataStates";
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";

// Relative
import { useGetClinicalPlan, useUpdateClinicalPlan } from "../../api/clinical-plan";
import { useGetDiagnosisTypes, useGetDiagnosisNames } from "../../api/get-diagnosis-options";
import type { UpdateClinicalPlanInput } from "../../api/clinical-plan";

// ── Types ─────────────────────────────────────────────────────────────

interface ClinicalPlanSectionProps {
  medicalRecordId: string;
  onRegisterSave?: (fn: () => Promise<void>) => void;
  canEdit?: boolean;
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId */
  recordClinicId?: string;
}

// ── Component ─────────────────────────────────────────────────────────

export const ClinicalPlanSection = memo(function ClinicalPlanSection({ medicalRecordId, onRegisterSave, canEdit = false, recordClinicId }: ClinicalPlanSectionProps) {
  const { data, isLoading } = useGetClinicalPlan(medicalRecordId, recordClinicId);
  const updateMutation = useUpdateClinicalPlan(medicalRecordId, recordClinicId);

  const [physicalExam, setPhysicalExam] = useState("");
  // 数値IDで管理（文字列を Number() 変換する旧実装の NaN バグを排除）
  const [diagnosisTypeId, setDiagnosisTypeId] = useState<number | null>(null);
  const [diagnosisNameId, setDiagnosisNameId] = useState<number | null>(null);
  const [diagnosisDetails, setDiagnosisDetails] = useState("");
  const [treatmentPolicy, setTreatmentPolicy] = useState("");

  const { data: diagnosisTypes = [], isLoading: isTypesLoading } = useGetDiagnosisTypes();
  const { data: diagnosisNames = [], isLoading: isNamesLoading } = useGetDiagnosisNames(diagnosisTypeId);

  // SearchableSelect 用に選択肢を {value,label} 形へ変換(参照安定のため memo 化)
  const typeOptions = useMemo<SearchableSelectOption[]>(
    () => diagnosisTypes.map((t) => ({ value: String(t.id), label: t.name })),
    [diagnosisTypes]
  );
  const nameOptions = useMemo<SearchableSelectOption[]>(
    () => diagnosisNames.map((n) => ({ value: String(n.id), label: n.name })),
    [diagnosisNames]
  );

  // Sync form state when data loads
  useEffect(() => {
    if (data) {
      setPhysicalExam(data.physical_exam ?? "");
      setDiagnosisTypeId(data.diagnosis_type_id ? Number(data.diagnosis_type_id) : null);
      setDiagnosisNameId(data.diagnosis_name_id ? Number(data.diagnosis_name_id) : null);
      setDiagnosisDetails(data.diagnosis_details ?? "");
      setTreatmentPolicy(data.treatment_policy ?? "");
    }
  }, [data]);

  const { mutateAsync: updateClinicalPlanAsync } = updateMutation;
  const handleSave = useCallback(async (): Promise<void> => {
    if (!canEdit) return;
    const input: UpdateClinicalPlanInput = {
      physical_exam: physicalExam,
      diagnosis_type_id: diagnosisTypeId,
      diagnosis_name_id: diagnosisNameId,
      diagnosis_details: diagnosisDetails,
      treatment_policy: treatmentPolicy,
    };
    await updateClinicalPlanAsync(input);
  }, [canEdit, physicalExam, diagnosisTypeId, diagnosisNameId, diagnosisDetails, treatmentPolicy, updateClinicalPlanAsync]);

  // Register save function with parent
  useEffect(() => {
    if (!onRegisterSave) return;
    onRegisterSave(handleSave);
  }, [onRegisterSave, handleSave]);

  if (isLoading) {
    return <LoadingFallback />;
  }

  return (
    <div className={`${C.bgWhite} border ${C.borderMedium} rounded-xs p-4`}>
      <h2 className={`text-sm font-bold ${C.text} mb-4`}>診察所見・診断・治療方針</h2>

      <div className="flex flex-col gap-4">
        {/* 身体検査所見 */}
        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel}>身体検査所見</label>
          <CharCountTextarea
            value={physicalExam}
            onChange={setPhysicalExam}
            placeholder="身体検査所見を入力してください"
            textareaClassName={`min-h-[100px] ${C.text} text-sm`}
            disabled={!canEdit}
          />
        </div>

        {/* 診断カテゴリ — セレクトで ID を安全に管理（旧: 自由入力 input で NaN 破損バグ） */}
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center justify-between">
            <label className={STYLE.formLabel}>診断カテゴリ</label>
            <MasterLink category="diagnosis_type" label="編集" className="text-[11px]" />
          </div>
          <SearchableSelect
            value={diagnosisTypeId ? String(diagnosisTypeId) : ""}
            onValueChange={(value) => {
              setDiagnosisTypeId(value ? Number(value) : null);
              setDiagnosisNameId(null); // カテゴリ変更時は病名をリセット
            }}
            options={typeOptions}
            disabled={isTypesLoading || !canEdit}
            placeholder={isTypesLoading ? "読み込み中..." : "カテゴリを選択"}
            searchPlaceholder="カテゴリを検索..."
          />
        </div>

        {/* 診断病名 — セレクトで ID を安全に管理（旧: 自由入力 input で NaN 破損バグ） */}
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center justify-between">
            <label className={STYLE.formLabel}>診断病名</label>
            <MasterLink category="diagnosis_name" label="編集" className="text-[11px]" />
          </div>
          <SearchableSelect
            value={diagnosisNameId ? String(diagnosisNameId) : ""}
            onValueChange={(value) => setDiagnosisNameId(value ? Number(value) : null)}
            options={nameOptions}
            disabled={isNamesLoading || !diagnosisTypeId || !canEdit}
            placeholder={
              isNamesLoading ? "読み込み中..." :
              !diagnosisTypeId ? "先にカテゴリを選択" :
              "病名を選択"
            }
            searchPlaceholder="病名を検索..."
          />
        </div>

        {/* 診断詳細 */}
        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel}>診断詳細</label>
          <CharCountTextarea
            value={diagnosisDetails}
            onChange={setDiagnosisDetails}
            placeholder="診断詳細を入力してください"
            textareaClassName={`min-h-[100px] ${C.text} text-sm`}
            disabled={!canEdit}
          />
        </div>

        {/* 治療方針 */}
        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel}>治療方針</label>
          <CharCountTextarea
            value={treatmentPolicy}
            onChange={setTreatmentPolicy}
            placeholder="治療方針を入力してください"
            textareaClassName={`min-h-[100px] ${C.text} text-sm`}
            disabled={!canEdit}
          />
        </div>

      </div>
    </div>
  );
});
