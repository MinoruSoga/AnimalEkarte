import { useState } from "react";

import type { MedicalRecord } from "../api/transforms";

interface UseApplyMedicalRecordArgs {
  existingRecord?: MedicalRecord;
  setChiefComplaint: (value: string) => void;
  setChiefComplaintTypeId?: (id: number | null) => void;
  setTreatmentPolicy: (value: string) => void;
  setPlan: (value: string) => void;
  setAssessment: (value: string) => void;
  setVisitType?: (value: string) => void;
  setNextVisitDate?: (value: string) => void;
}

export function useApplyMedicalRecord({
  existingRecord,
  setChiefComplaint,
  setChiefComplaintTypeId,
  setTreatmentPolicy,
  setPlan,
  setAssessment,
  setVisitType,
  setNextVisitDate,
}: UseApplyMedicalRecordArgs) {
  const [prevExistingRecord, setPrevExistingRecord] = useState(existingRecord);

  if (prevExistingRecord !== existingRecord && existingRecord) {
    setPrevExistingRecord(existingRecord);
    if (existingRecord.chiefComplaint) setChiefComplaint(existingRecord.chiefComplaint);
    if (existingRecord.chiefComplaintTypeId != null && setChiefComplaintTypeId) {
      setChiefComplaintTypeId(existingRecord.chiefComplaintTypeId);
    }
    if (existingRecord.plan) setPlan(existingRecord.plan);
    if (existingRecord.assessment) setAssessment(existingRecord.assessment);
    if (existingRecord.notes) setTreatmentPolicy(existingRecord.notes);
    if (existingRecord.visitType && setVisitType) setVisitType(existingRecord.visitType);
    if (setNextVisitDate) setNextVisitDate(existingRecord.nextVisitRecommendedDate);
  }
}
