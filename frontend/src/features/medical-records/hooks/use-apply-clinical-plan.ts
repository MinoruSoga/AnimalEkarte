import { useEffect, useRef } from "react";

import type { ClinicalPlan } from "../api/clinical-plan";

interface UseApplyClinicalPlanArgs {
  clinicalPlan?: ClinicalPlan;
  setPhysicalExam: (value: string) => void;
  setPlan: (value: string) => void;
  setAssessment: (value: string) => void;
  setDiagnosis1CategoryId: (id: number | null) => void;
  setDiagnosis1NameId: (id: number | null) => void;
  setDiagnosis2CategoryId: (id: number | null) => void;
  setDiagnosis2NameId: (id: number | null) => void;
}

function toOptionalNumber(value: string | null | undefined): number | null {
  if (value == null || value === "") return null;
  const n = Number(value);
  return Number.isFinite(n) ? n : null;
}

/**
 * BUG-010: clinical-plan GET が 3欄 + 診断マスタの正本。
 * medical-record detail wire には clinical_plan が載らないため、専用 GET から hydrate する。
 *
 * レコードごとに初回だけ form へ流し込む。保存後の invalidate/refetch では text を再適用しない
 * （version は React Query cache / setQueryData が正本。dirty 入力の上書きを防ぐ）。
 */
export function useApplyClinicalPlan({
  clinicalPlan,
  setPhysicalExam,
  setPlan,
  setAssessment,
  setDiagnosis1CategoryId,
  setDiagnosis1NameId,
  setDiagnosis2CategoryId,
  setDiagnosis2NameId,
}: UseApplyClinicalPlanArgs) {
  const hydratedRecordIdRef = useRef<string | null>(null);

  useEffect(() => {
    if (!clinicalPlan) return;
    const recordKey = clinicalPlan.medical_record_id;
    if (!recordKey) return;
    // 同一 medical_record では初回 hydrate のみ。保存後の再 GET で入力中フィールドを潰さない。
    if (hydratedRecordIdRef.current === recordKey) return;
    hydratedRecordIdRef.current = recordKey;

    setPhysicalExam(clinicalPlan.physical_exam ?? "");
    setPlan(clinicalPlan.treatment_policy ?? "");
    setAssessment(clinicalPlan.diagnosis_details ?? "");
    setDiagnosis1CategoryId(toOptionalNumber(clinicalPlan.diagnosis_type_id));
    setDiagnosis1NameId(toOptionalNumber(clinicalPlan.diagnosis_name_id));
    setDiagnosis2CategoryId(toOptionalNumber(clinicalPlan.diagnosis_2_type_id));
    setDiagnosis2NameId(toOptionalNumber(clinicalPlan.diagnosis_2_name_id));
    // eslint-disable-next-line react-hooks/exhaustive-deps -- recordKey 単位の one-shot hydrate
  }, [clinicalPlan]);
}
