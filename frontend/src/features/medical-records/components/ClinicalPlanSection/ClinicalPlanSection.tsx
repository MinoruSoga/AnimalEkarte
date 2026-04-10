// React/Framework
import { memo, useState, useEffect, useCallback } from "react";

// Internal
import { CharCountTextarea } from "@/components/shared/CharCountTextarea";
import { C, STYLE } from "@/lib/design-tokens";
import { LoadingFallback } from "@/components/shared/DataStates";

// Relative
import { useGetClinicalPlan, useUpdateClinicalPlan } from "@/features/medical-records/api/clinical-plan";
import type { UpdateClinicalPlanInput } from "@/features/medical-records/api/clinical-plan";

// ── Types ─────────────────────────────────────────────────────────────

interface ClinicalPlanSectionProps {
  medicalRecordId: string;
  onRegisterSave?: (fn: () => Promise<void>) => void;
  canEdit?: boolean;
}

// ── Component ─────────────────────────────────────────────────────────

export const ClinicalPlanSection = memo(function ClinicalPlanSection({ medicalRecordId, onRegisterSave, canEdit = false }: ClinicalPlanSectionProps) {
  const { data, isLoading } = useGetClinicalPlan(medicalRecordId);
  const updateMutation = useUpdateClinicalPlan(medicalRecordId);

  const [physicalExam, setPhysicalExam] = useState("");
  const [diagnosisCategoryId, setDiagnosisCategoryId] = useState("");
  const [diagnosisNameId, setDiagnosisNameId] = useState("");
  const [diagnosisDetails, setDiagnosisDetails] = useState("");
  const [treatmentPolicy, setTreatmentPolicy] = useState("");

  // Sync form state when data loads
  useEffect(() => {
    if (data) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- 非同期サーバーデータでフォームを初期化するパターン。React 18 が自動バッチするため実害なし
      setPhysicalExam(data.physical_exam ?? "");
      setDiagnosisCategoryId(data.diagnosis_category_id ?? "");
      setDiagnosisNameId(data.diagnosis_name_id ?? "");
      setDiagnosisDetails(data.diagnosis_details ?? "");
      setTreatmentPolicy(data.treatment_policy ?? "");
    }
  }, [data]);

  const handleSave = useCallback(async (): Promise<void> => {
    if (!canEdit) return;
    const input: UpdateClinicalPlanInput = {
      physical_exam: physicalExam,
      diagnosis_category_id: diagnosisCategoryId ? Number(diagnosisCategoryId) : null,
      diagnosis_name_id: diagnosisNameId ? Number(diagnosisNameId) : null,
      diagnosis_details: diagnosisDetails,
      treatment_policy: treatmentPolicy,
    };
    await updateMutation.mutateAsync(input);
  }, [canEdit, physicalExam, diagnosisCategoryId, diagnosisNameId, diagnosisDetails, treatmentPolicy, updateMutation]);

  // Register save function with parent
  useEffect(() => {
    if (!onRegisterSave) return;
    onRegisterSave(handleSave);
  }, [onRegisterSave, handleSave]);

  if (isLoading) {
    return <LoadingFallback />;
  }

  return (
    <div className={`mt-4 bg-white border ${C.borderMedium} rounded-[4px] p-5`}>
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

        {/* 診断カテゴリ */}
        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel}>診断カテゴリ</label>
          <input
            type="text"
            value={
              data?.diagnosis_category
                ? data.diagnosis_category.name
                : diagnosisCategoryId
            }
            onChange={(e) => setDiagnosisCategoryId(e.target.value)}
            placeholder="カテゴリを選択"
            className={`${STYLE.formInput} border rounded-[4px] px-3 outline-none focus:ring-0`}
            disabled={!canEdit}
          />
        </div>

        {/* 診断病名 */}
        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel}>診断病名</label>
          <input
            type="text"
            value={
              data?.diagnosis_name
                ? data.diagnosis_name.name
                : diagnosisNameId
            }
            onChange={(e) => setDiagnosisNameId(e.target.value)}
            placeholder="病名を選択"
            className={`${STYLE.formInput} border rounded-[4px] px-3 outline-none focus:ring-0`}
            disabled={!canEdit}
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
