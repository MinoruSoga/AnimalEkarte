import { useState } from "react";

import type { MedicalRecord } from "../api/transforms";

interface UseApplyMedicalRecordArgs {
  existingRecord?: MedicalRecord;
  setChiefComplaint: (value: string) => void;
  setTreatmentPolicy: (value: string) => void;
  setPlan: (value: string) => void;
  setAssessment: (value: string) => void;
}

export function useApplyMedicalRecord({
  existingRecord,
  setChiefComplaint,
  setTreatmentPolicy,
  setPlan,
  setAssessment,
}: UseApplyMedicalRecordArgs) {
  const [prevExistingRecord, setPrevExistingRecord] = useState(existingRecord);

  if (prevExistingRecord !== existingRecord && existingRecord) {
    setPrevExistingRecord(existingRecord);
    if (existingRecord.chiefComplaint) setChiefComplaint(existingRecord.chiefComplaint);
    if (existingRecord.plan) setPlan(existingRecord.plan);
    if (existingRecord.assessment) setAssessment(existingRecord.assessment);
    if (existingRecord.notes) setTreatmentPolicy(existingRecord.notes);
  }
}
