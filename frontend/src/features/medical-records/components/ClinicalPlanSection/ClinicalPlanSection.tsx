// React/Framework
import { useState, useEffect, useCallback } from "react";

// Internal
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { C, STYLE } from "@/lib/design-tokens";

// Relative
import { useClinicalPlan, useUpdateClinicalPlan } from "../../api/clinical-plan";
import type { UpdateClinicalPlanInput } from "../../api/clinical-plan";

// ── Types ─────────────────────────────────────────────────────────────

interface ClinicalPlanSectionProps {
  medicalRecordId: string;
}

// ── Component ─────────────────────────────────────────────────────────

export function ClinicalPlanSection({ medicalRecordId }: ClinicalPlanSectionProps) {
  const { data, isLoading } = useClinicalPlan(medicalRecordId);
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

  const handleSave = useCallback(() => {
    const input: UpdateClinicalPlanInput = {
      physical_exam: physicalExam,
      diagnosis_category_id: diagnosisCategoryId ? Number(diagnosisCategoryId) : null,
      diagnosis_name_id: diagnosisNameId ? Number(diagnosisNameId) : null,
      diagnosis_details: diagnosisDetails,
      treatment_policy: treatmentPolicy,
    };
    updateMutation.mutate(input);
  }, [physicalExam, diagnosisCategoryId, diagnosisNameId, diagnosisDetails, treatmentPolicy, updateMutation]);

  if (isLoading) {
    return (
      <div className={`flex items-center justify-center h-24 text-sm ${C.text40}`}>
        読み込み中...
      </div>
    );
  }

  return (
    <div className={`mt-4 bg-white border ${C.borderMedium} rounded-[4px] p-5`}>
      <h2 className={`text-sm font-bold ${C.text} mb-4`}>診察所見・診断・治療方針</h2>

      <div className="flex flex-col gap-4">
        {/* 身体検査所見 */}
        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel}>身体検査所見</label>
          <Textarea
            value={physicalExam}
            onChange={(e) => setPhysicalExam(e.target.value)}
            placeholder="身体検査所見を入力してください"
            className={`min-h-[100px] ${C.text} text-sm`}
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
            placeholder="診断カテゴリID"
            className={`${STYLE.formInput} border rounded-[4px] px-3 outline-none focus:ring-0`}
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
            placeholder="診断病名ID"
            className={`${STYLE.formInput} border rounded-[4px] px-3 outline-none focus:ring-0`}
          />
        </div>

        {/* 診断詳細 */}
        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel}>診断詳細</label>
          <Textarea
            value={diagnosisDetails}
            onChange={(e) => setDiagnosisDetails(e.target.value)}
            placeholder="診断詳細を入力してください"
            className={`min-h-[100px] ${C.text} text-sm`}
          />
        </div>

        {/* 治療方針 */}
        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel}>治療方針</label>
          <Textarea
            value={treatmentPolicy}
            onChange={(e) => setTreatmentPolicy(e.target.value)}
            placeholder="治療方針を入力してください"
            className={`min-h-[100px] ${C.text} text-sm`}
          />
        </div>

        {/* Save Button */}
        <div className="flex justify-end pt-1">
          <Button
            onClick={handleSave}
            disabled={updateMutation.isPending}
            className={`${STYLE.btnPrimary} px-5`}
          >
            {updateMutation.isPending ? "保存中..." : "保存"}
          </Button>
        </div>
      </div>
    </div>
  );
}
