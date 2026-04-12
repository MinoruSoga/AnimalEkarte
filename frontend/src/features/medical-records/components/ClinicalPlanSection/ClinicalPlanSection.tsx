// React/Framework
import { memo, useState, useEffect, useCallback, useMemo } from "react";

// Internal
import { CharCountTextarea } from "@/components/shared/CharCountTextarea";
import { C, STYLE } from "@/lib/design-tokens";
import { LoadingFallback } from "@/components/shared/DataStates";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

// Relative
import { useGetClinicalPlan, useUpdateClinicalPlan } from "@/features/medical-records/api/clinical-plan";
import { useGetDiagnosisTypes, useGetDiagnosisNames } from "@/features/medical-records/api/get-diagnosis-options";
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
  // 数値IDで管理（文字列を Number() 変換する旧実装の NaN バグを排除）
  const [diagnosisTypeId, setDiagnosisTypeId] = useState<number | null>(null);
  const [diagnosisNameId, setDiagnosisNameId] = useState<number | null>(null);
  const [diagnosisDetails, setDiagnosisDetails] = useState("");
  const [treatmentPolicy, setTreatmentPolicy] = useState("");

  const { data: diagnosisTypes = [], isLoading: isTypesLoading } = useGetDiagnosisTypes();
  const { data: diagnosisNames = [], isLoading: isNamesLoading } = useGetDiagnosisNames(diagnosisTypeId);

  // js-cache-function-results: API データから生成する JSX リストを useMemo でキャッシュ
  const typeSelectItems = useMemo(
    () => diagnosisTypes.map((t) => (
      <SelectItem key={t.id} value={String(t.id)}>{t.name}</SelectItem>
    )),
    [diagnosisTypes]
  );
  const nameSelectItems = useMemo(
    () => diagnosisNames.map((n) => (
      <SelectItem key={n.id} value={String(n.id)}>{n.name}</SelectItem>
    )),
    [diagnosisNames]
  );

  // Sync form state when data loads
  useEffect(() => {
    if (data) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- 非同期サーバーデータでフォームを初期化するパターン。React 18 が自動バッチするため実害なし
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

        {/* 診断カテゴリ — Select で ID を安全に管理（旧: 自由入力 input で NaN 破損バグ） */}
        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel}>診断カテゴリ</label>
          <Select
            value={diagnosisTypeId ? String(diagnosisTypeId) : ""}
            onValueChange={(value) => {
              setDiagnosisTypeId(value ? Number(value) : null);
              setDiagnosisNameId(null); // カテゴリ変更時は病名をリセット
            }}
            disabled={isTypesLoading || !canEdit}
          >
            <SelectTrigger className={`bg-white ${C.borderMedium} h-10 text-sm`}>
              <SelectValue placeholder={isTypesLoading ? "読み込み中..." : "カテゴリを選択"} />
            </SelectTrigger>
            <SelectContent className="z-[9999]">
              {typeSelectItems}
            </SelectContent>
          </Select>
        </div>

        {/* 診断病名 — Select で ID を安全に管理（旧: 自由入力 input で NaN 破損バグ） */}
        <div className="flex flex-col gap-1.5">
          <label className={STYLE.formLabel}>診断病名</label>
          <Select
            value={diagnosisNameId ? String(diagnosisNameId) : ""}
            onValueChange={(value) => setDiagnosisNameId(value ? Number(value) : null)}
            disabled={isNamesLoading || !diagnosisTypeId || !canEdit}
          >
            <SelectTrigger className={`bg-white ${C.borderMedium} h-10 text-sm`}>
              <SelectValue placeholder={
                isNamesLoading ? "読み込み中..." :
                !diagnosisTypeId ? "先にカテゴリを選択" :
                "病名を選択"
              } />
            </SelectTrigger>
            <SelectContent className="z-[9999]">
              {nameSelectItems}
            </SelectContent>
          </Select>
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
