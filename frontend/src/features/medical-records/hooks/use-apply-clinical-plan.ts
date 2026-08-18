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

function toOptionalNumber(value: unknown): number | null {
  if (value == null || value === "") return null;
  if (typeof value === "object" && "id" in value) {
    return toOptionalNumber((value as { id: unknown }).id);
  }
  const n = typeof value === "number" ? value : Number(value);
  return Number.isFinite(n) ? n : null;
}

function readDiagnosisIds(clinicalPlan: ClinicalPlan) {
  return {
    type1: toOptionalNumber(clinicalPlan.diagnosis_type_id) ?? toOptionalNumber(clinicalPlan.diagnosis_type),
    name1: toOptionalNumber(clinicalPlan.diagnosis_name_id) ?? toOptionalNumber(clinicalPlan.diagnosis_name),
    type2: toOptionalNumber(clinicalPlan.diagnosis_2_type_id) ?? toOptionalNumber(clinicalPlan.diagnosis_2_type),
    name2: toOptionalNumber(clinicalPlan.diagnosis_2_name_id) ?? toOptionalNumber(clinicalPlan.diagnosis_2_name),
  };
}

/**
 * BUG-010 / BUG-013: clinical-plan GET が 3欄 + 診断マスタの正本。
 * 初回 GET で診断 ID が空なら、後続 GET で ID が入ったとき診断だけ再適用する。
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
  const diagnosisHydratedRef = useRef(false);

  useEffect(() => {
    if (!clinicalPlan) return;
    const recordKey = clinicalPlan.medical_record_id;
    if (!recordKey) return;

    const ids = readDiagnosisIds(clinicalPlan);
    const hasDiagnosis = ids.type1 != null || ids.name1 != null || ids.type2 != null || ids.name2 != null;
    const isFirst = hydratedRecordIdRef.current !== recordKey;

    if (isFirst) {
      hydratedRecordIdRef.current = recordKey;
      diagnosisHydratedRef.current = false;
      setPhysicalExam(clinicalPlan.physical_exam ?? "");
      setPlan(clinicalPlan.treatment_policy ?? "");
      setAssessment(clinicalPlan.diagnosis_details ?? "");
    }

    if (isFirst || !diagnosisHydratedRef.current) {
      setDiagnosis1CategoryId(ids.type1);
      setDiagnosis1NameId(ids.name1);
      setDiagnosis2CategoryId(ids.type2);
      setDiagnosis2NameId(ids.name2);
      if (hasDiagnosis) {
        diagnosisHydratedRef.current = true;
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- recordKey 単位の hydrate
  }, [clinicalPlan]);
}
