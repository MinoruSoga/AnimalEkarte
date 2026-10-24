import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";

import type { ClinicalPlan, ClinicalPlanDiagnosisRow } from "../api/clinical-plan";

export interface ClinicalPlanDraft {
  physicalExam: string;
  plan: string;
  assessment: string;
  diagnosis1CategoryId: number | null;
  diagnosis1NameId: number | null;
  diagnosis2CategoryId: number | null;
  diagnosis2NameId: number | null;
}

interface ClinicalPlanSnapshot extends ClinicalPlanDraft {
  medicalRecordId: string;
  version: number;
  /** BE が返した診断行の全件（返却順のまま）。保存時に変えていない slot の行を送り返す。 */
  diagnosisRows: ClinicalPlanDiagnosisRow[];
}

type ClinicalPlanSnapshotSource = Pick<
  ClinicalPlan,
  | "medical_record_id"
  | "version"
  | "physical_exam"
  | "treatment_policy"
  | "diagnosis_details"
  | "diagnoses"
>;

interface UseApplyClinicalPlanArgs extends ClinicalPlanDraft {
  recordId?: string;
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
  const numberValue = typeof value === "number" ? value : Number(value);
  return Number.isFinite(numberValue) ? numberValue : null;
}

function toSnapshot(
  clinicalPlan: ClinicalPlanSnapshotSource,
  recordId?: string,
): ClinicalPlanSnapshot {
  const diagnosisRows = Array.isArray(clinicalPlan.diagnoses) ? clinicalPlan.diagnoses : [];
  // slot の代表行は返却順の先頭。BE は slot, sort_order, id 昇順で返す（ListDiagnoses）。
  const slot1Row = diagnosisRows.find((row) => row.diagnosis_slot === 1);
  const slot2Row = diagnosisRows.find((row) => row.diagnosis_slot === 2);
  return {
    medicalRecordId: clinicalPlan.medical_record_id || recordId || "",
    version: clinicalPlan.version,
    physicalExam: clinicalPlan.physical_exam ?? "",
    plan: clinicalPlan.treatment_policy ?? "",
    assessment: clinicalPlan.diagnosis_details ?? "",
    diagnosisRows,
    diagnosis1CategoryId: toOptionalNumber(slot1Row?.diagnosis_type_id),
    diagnosis1NameId: toOptionalNumber(slot1Row?.diagnosis_name_id),
    diagnosis2CategoryId: toOptionalNumber(slot2Row?.diagnosis_type_id),
    diagnosis2NameId: toOptionalNumber(slot2Row?.diagnosis_name_id),
  };
}

function hasSameDraft(left: ClinicalPlanDraft, right: ClinicalPlanDraft): boolean {
  return (
    left.physicalExam === right.physicalExam &&
    left.plan === right.plan &&
    left.assessment === right.assessment &&
    left.diagnosis1CategoryId === right.diagnosis1CategoryId &&
    left.diagnosis1NameId === right.diagnosis1NameId &&
    left.diagnosis2CategoryId === right.diagnosis2CategoryId &&
    left.diagnosis2NameId === right.diagnosis2NameId
  );
}

/**
 * 診断行の全件が同一かを返す。slot ごとの代表行（先頭行）の ID だけでは、行の追加・
 * スナップショット変更・sort_order 変更を検出できないため、全行を返却順のまま比較する。
 */
function hasSameDiagnosisRows(
  left: ClinicalPlanDiagnosisRow[],
  right: ClinicalPlanDiagnosisRow[],
): boolean {
  return (
    left.length === right.length &&
    left.every((row, index) => {
      const other = right[index];
      return (
        row.id === other.id &&
        row.version === other.version &&
        row.diagnosis_slot === other.diagnosis_slot &&
        row.diagnosis_type_id === other.diagnosis_type_id &&
        row.diagnosis_name_id === other.diagnosis_name_id &&
        row.diagnosis_name_snapshot === other.diagnosis_name_snapshot &&
        row.category_name_snapshot === other.category_name_snapshot &&
        row.sort_order === other.sort_order
      );
    })
  );
}

function hasSameSnapshot(left: ClinicalPlanSnapshot, right: ClinicalPlanSnapshot): boolean {
  return (
    left.medicalRecordId === right.medicalRecordId &&
    left.version === right.version &&
    hasSameDraft(left, right) &&
    hasSameDiagnosisRows(left.diagnosisRows, right.diagnosisRows)
  );
}

/**
 * Maintains the clinical-plan form fields and CAS version as one server snapshot.
 * A changed remote snapshot is accepted only while the form still matches its baseline.
 */
export function useApplyClinicalPlan({
  recordId,
  clinicalPlan,
  physicalExam,
  plan,
  assessment,
  diagnosis1CategoryId,
  diagnosis1NameId,
  diagnosis2CategoryId,
  diagnosis2NameId,
  setPhysicalExam,
  setPlan,
  setAssessment,
  setDiagnosis1CategoryId,
  setDiagnosis1NameId,
  setDiagnosis2CategoryId,
  setDiagnosis2NameId,
}: UseApplyClinicalPlanArgs) {
  const baselineRef = useRef<ClinicalPlanSnapshot | null>(null);
  const activeRecordIdRef = useRef(recordId);
  const resetRecordIdRef = useRef(recordId);
  const [clinicalPlanVersion, setClinicalPlanVersion] = useState<number>();
  const [clinicalPlanDiagnoses, setClinicalPlanDiagnoses] = useState<ClinicalPlanDiagnosisRow[]>();
  const currentDraftRef = useRef<ClinicalPlanDraft>({
    physicalExam,
    plan,
    assessment,
    diagnosis1CategoryId,
    diagnosis1NameId,
    diagnosis2CategoryId,
    diagnosis2NameId,
  });

  useLayoutEffect(() => {
    activeRecordIdRef.current = recordId;
    currentDraftRef.current = {
      physicalExam,
      plan,
      assessment,
      diagnosis1CategoryId,
      diagnosis1NameId,
      diagnosis2CategoryId,
      diagnosis2NameId,
    };
  });

  const applySnapshot = useCallback(
    (snapshot: ClinicalPlanSnapshot) => {
      baselineRef.current = snapshot;
      setClinicalPlanVersion(snapshot.version);
      setClinicalPlanDiagnoses(snapshot.diagnosisRows);
      setPhysicalExam(snapshot.physicalExam);
      setPlan(snapshot.plan);
      setAssessment(snapshot.assessment);
      setDiagnosis1CategoryId(snapshot.diagnosis1CategoryId);
      setDiagnosis1NameId(snapshot.diagnosis1NameId);
      setDiagnosis2CategoryId(snapshot.diagnosis2CategoryId);
      setDiagnosis2NameId(snapshot.diagnosis2NameId);
    },
    [
      setAssessment,
      setDiagnosis1CategoryId,
      setDiagnosis1NameId,
      setDiagnosis2CategoryId,
      setDiagnosis2NameId,
      setPhysicalExam,
      setPlan,
    ],
  );

  useEffect(() => {
    if (resetRecordIdRef.current === recordId) return;

    resetRecordIdRef.current = recordId;
    baselineRef.current = null;
    setClinicalPlanVersion(undefined);
    setClinicalPlanDiagnoses(undefined);
    setPhysicalExam("");
    setPlan("");
    setAssessment("");
    setDiagnosis1CategoryId(null);
    setDiagnosis1NameId(null);
    setDiagnosis2CategoryId(null);
    setDiagnosis2NameId(null);
  }, [
    recordId,
    setAssessment,
    setDiagnosis1CategoryId,
    setDiagnosis1NameId,
    setDiagnosis2CategoryId,
    setDiagnosis2NameId,
    setPhysicalExam,
    setPlan,
  ]);

  useEffect(() => {
    if (!clinicalPlan) return;
    if (recordId && clinicalPlan.medical_record_id && clinicalPlan.medical_record_id !== recordId) {
      return;
    }

    const remoteSnapshot = toSnapshot(clinicalPlan, recordId);
    const baseline = baselineRef.current;
    if (!baseline || baseline.medicalRecordId !== remoteSnapshot.medicalRecordId) {
      applySnapshot(remoteSnapshot);
      return;
    }
    // useUpdateClinicalPlan writes its response before invalidating. A delayed refetch of an
    // older version must not roll the just-confirmed snapshot backward.
    if (remoteSnapshot.version < baseline.version) return;
    if (hasSameSnapshot(baseline, remoteSnapshot)) return;

    const currentDraft: ClinicalPlanDraft = {
      physicalExam,
      plan,
      assessment,
      diagnosis1CategoryId,
      diagnosis1NameId,
      diagnosis2CategoryId,
      diagnosis2NameId,
    };
    if (!hasSameDraft(currentDraft, baseline)) return;

    applySnapshot(remoteSnapshot);
  }, [
    applySnapshot,
    assessment,
    clinicalPlan,
    diagnosis1CategoryId,
    diagnosis1NameId,
    diagnosis2CategoryId,
    diagnosis2NameId,
    physicalExam,
    plan,
    recordId,
  ]);

  const onClinicalPlanSaved = useCallback(
    (savedClinicalPlan: ClinicalPlanSnapshotSource, submittedDraft: ClinicalPlanDraft) => {
      const savedSnapshot = toSnapshot(savedClinicalPlan);
      if (activeRecordIdRef.current !== savedSnapshot.medicalRecordId) return;

      if (hasSameDraft(currentDraftRef.current, submittedDraft)) {
        applySnapshot(savedSnapshot);
        return;
      }

      baselineRef.current = savedSnapshot;
      setClinicalPlanVersion(savedSnapshot.version);
      setClinicalPlanDiagnoses(savedSnapshot.diagnosisRows);
    },
    [applySnapshot],
  );

  return { clinicalPlanVersion, clinicalPlanDiagnoses, onClinicalPlanSaved };
}
