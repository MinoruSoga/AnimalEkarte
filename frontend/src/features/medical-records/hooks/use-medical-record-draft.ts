import { useEffect } from "react";
import { toast } from "sonner";

interface MedicalRecordDraft {
  chiefComplaint: string;
  treatmentPolicy: string;
  plan: string;
  assessment: string;
  visitType: string;
  diagnosis1CategoryId: number | null;
  diagnosis1NameId: number | null;
  diagnosis2CategoryId: number | null;
  diagnosis2NameId: number | null;
}

interface UseMedicalRecordDraftArgs extends MedicalRecordDraft {
  recordId?: string;
  setChiefComplaint: (value: string) => void;
  setTreatmentPolicy: (value: string) => void;
  setPlan: (value: string) => void;
  setAssessment: (value: string) => void;
  setVisitType: (value: string) => void;
  setDiagnosis1CategoryId: (value: number | null) => void;
  setDiagnosis1NameId: (value: number | null) => void;
  setDiagnosis2CategoryId: (value: number | null) => void;
  setDiagnosis2NameId: (value: number | null) => void;
}

export function useMedicalRecordDraft({
  recordId,
  chiefComplaint,
  treatmentPolicy,
  plan,
  assessment,
  visitType,
  diagnosis1CategoryId,
  diagnosis1NameId,
  diagnosis2CategoryId,
  diagnosis2NameId,
  setChiefComplaint,
  setTreatmentPolicy,
  setPlan,
  setAssessment,
  setVisitType,
  setDiagnosis1CategoryId,
  setDiagnosis1NameId,
  setDiagnosis2CategoryId,
  setDiagnosis2NameId,
}: UseMedicalRecordDraftArgs) {
  const draftKey = `medical-record-draft-${recordId}`;

  useEffect(() => {
    if (!recordId) return;
    const saved = localStorage.getItem(draftKey);
    if (!saved) return;

    try {
      const draft = JSON.parse(saved) as Partial<MedicalRecordDraft>;
      if (draft.chiefComplaint) setChiefComplaint(draft.chiefComplaint);
      if (draft.treatmentPolicy) setTreatmentPolicy(draft.treatmentPolicy);
      if (draft.plan) setPlan(draft.plan);
      if (draft.assessment) setAssessment(draft.assessment);
      if (draft.visitType) setVisitType(draft.visitType);
      if (draft.diagnosis1CategoryId) setDiagnosis1CategoryId(draft.diagnosis1CategoryId);
      if (draft.diagnosis1NameId) setDiagnosis1NameId(draft.diagnosis1NameId);
      if (draft.diagnosis2CategoryId) setDiagnosis2CategoryId(draft.diagnosis2CategoryId);
      if (draft.diagnosis2NameId) setDiagnosis2NameId(draft.diagnosis2NameId);
      toast.info("未保存の下書きを復元しました", { duration: 2000 });
    } catch {
      localStorage.removeItem(draftKey);
    }
  }, [
    recordId,
    draftKey,
    setChiefComplaint,
    setTreatmentPolicy,
    setPlan,
    setAssessment,
    setVisitType,
    setDiagnosis1CategoryId,
    setDiagnosis1NameId,
    setDiagnosis2CategoryId,
    setDiagnosis2NameId,
  ]);

  useEffect(() => {
    if (!recordId) return;
    const draft: MedicalRecordDraft = {
      chiefComplaint,
      treatmentPolicy,
      plan,
      assessment,
      visitType,
      diagnosis1CategoryId,
      diagnosis1NameId,
      diagnosis2CategoryId,
      diagnosis2NameId,
    };
    localStorage.setItem(draftKey, JSON.stringify(draft));
  }, [
    recordId,
    draftKey,
    chiefComplaint,
    treatmentPolicy,
    plan,
    assessment,
    visitType,
    diagnosis1CategoryId,
    diagnosis1NameId,
    diagnosis2CategoryId,
    diagnosis2NameId,
  ]);

  return { draftKey };
}
