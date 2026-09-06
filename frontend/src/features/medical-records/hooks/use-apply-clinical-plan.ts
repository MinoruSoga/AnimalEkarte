import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";

import type { ClinicalPlan } from "../api/clinical-plan";

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
}

type ClinicalPlanSnapshotSource = Pick<
  ClinicalPlan,
  | "medical_record_id"
  | "version"
  | "physical_exam"
  | "treatment_policy"
  | "diagnosis_details"
  | "diagnosis_type_id"
  | "diagnosis_name_id"
  | "diagnosis_2_type_id"
  | "diagnosis_2_name_id"
  | "diagnosis_type"
  | "diagnosis_name"
  | "diagnosis_2_type"
  | "diagnosis_2_name"
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
  if (typeof value === "object" && "id" in value) {
    return toOptionalNumber(value.id);
  }
  const numberValue = typeof value === "number" ? value : Number(value);
  return Number.isFinite(numberValue) ? numberValue : null;
}

function toSnapshot(
  clinicalPlan: ClinicalPlanSnapshotSource,
  recordId?: string,
): ClinicalPlanSnapshot {
  return {
    medicalRecordId: clinicalPlan.medical_record_id || recordId || "",
    version: clinicalPlan.version,
    physicalExam: clinicalPlan.physical_exam ?? "",
    plan: clinicalPlan.treatment_policy ?? "",
    assessment: clinicalPlan.diagnosis_details ?? "",
    diagnosis1CategoryId:
      toOptionalNumber(clinicalPlan.diagnosis_type_id) ??
      toOptionalNumber(clinicalPlan.diagnosis_type),
    diagnosis1NameId:
      toOptionalNumber(clinicalPlan.diagnosis_name_id) ??
      toOptionalNumber(clinicalPlan.diagnosis_name),
    diagnosis2CategoryId:
      toOptionalNumber(clinicalPlan.diagnosis_2_type_id) ??
      toOptionalNumber(clinicalPlan.diagnosis_2_type),
    diagnosis2NameId:
      toOptionalNumber(clinicalPlan.diagnosis_2_name_id) ??
      toOptionalNumber(clinicalPlan.diagnosis_2_name),
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

function hasSameSnapshot(left: ClinicalPlanSnapshot, right: ClinicalPlanSnapshot): boolean {
  return (
    left.medicalRecordId === right.medicalRecordId &&
    left.version === right.version &&
    hasSameDraft(left, right)
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
    },
    [applySnapshot],
  );

  return { clinicalPlanVersion, onClinicalPlanSaved };
}
