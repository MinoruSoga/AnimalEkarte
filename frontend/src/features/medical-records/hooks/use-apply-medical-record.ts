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
  // BUG-410: 未配線だと保存済みの構造化診断(diagnosis1/2)が編集保存で無言クリアされる。
  setDiagnosis1CategoryId?: (id: number | null) => void;
  setDiagnosis1NameId?: (id: number | null) => void;
  setDiagnosis2CategoryId?: (id: number | null) => void;
  setDiagnosis2NameId?: (id: number | null) => void;
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
  setDiagnosis1CategoryId,
  setDiagnosis1NameId,
  setDiagnosis2CategoryId,
  setDiagnosis2NameId,
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
    if (existingRecord.diagnosis1CategoryId != null && setDiagnosis1CategoryId) {
      setDiagnosis1CategoryId(existingRecord.diagnosis1CategoryId);
    }
    if (existingRecord.diagnosis1NameId != null && setDiagnosis1NameId) {
      setDiagnosis1NameId(existingRecord.diagnosis1NameId);
    }
    if (existingRecord.diagnosis2CategoryId != null && setDiagnosis2CategoryId) {
      setDiagnosis2CategoryId(existingRecord.diagnosis2CategoryId);
    }
    if (existingRecord.diagnosis2NameId != null && setDiagnosis2NameId) {
      setDiagnosis2NameId(existingRecord.diagnosis2NameId);
    }
  }
}
