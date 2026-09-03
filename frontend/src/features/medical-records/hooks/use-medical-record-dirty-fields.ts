import { useCallback } from "react";

interface UseMedicalRecordDirtyFieldsParams {
  markDirty: () => void;
  setChiefComplaint: (value: string) => void;
  setChiefComplaintTypeId: (id: number | null) => void;
  setTreatmentPolicy: (value: string) => void;
  setPhysicalExam: (value: string) => void;
  setPlan: (value: string) => void;
  setAssessment: (value: string) => void;
  setDiagnosis1CategoryId: (id: number | null) => void;
  setDiagnosis1NameId: (id: number | null) => void;
  setDiagnosis2CategoryId: (id: number | null) => void;
  setDiagnosis2NameId: (id: number | null) => void;
}

export function useMedicalRecordDirtyFields({
  markDirty,
  setChiefComplaint,
  setChiefComplaintTypeId,
  setTreatmentPolicy,
  setPhysicalExam,
  setPlan,
  setAssessment,
  setDiagnosis1CategoryId,
  setDiagnosis1NameId,
  setDiagnosis2CategoryId,
  setDiagnosis2NameId,
}: UseMedicalRecordDirtyFieldsParams) {
  const handleSetChiefComplaint = useCallback(
    (value: string) => {
      markDirty();
      setChiefComplaint(value);
    },
    [markDirty, setChiefComplaint],
  );

  const handleSetChiefComplaintTypeId = useCallback(
    (id: number | null) => {
      markDirty();
      setChiefComplaintTypeId(id);
    },
    [markDirty, setChiefComplaintTypeId],
  );

  const handleSetTreatmentPolicy = useCallback(
    (value: string) => {
      markDirty();
      setTreatmentPolicy(value);
    },
    [markDirty, setTreatmentPolicy],
  );

  const handleSetPhysicalExam = useCallback(
    (value: string) => {
      markDirty();
      setPhysicalExam(value);
    },
    [markDirty, setPhysicalExam],
  );

  const handleSetPlan = useCallback(
    (value: string) => {
      markDirty();
      setPlan(value);
    },
    [markDirty, setPlan],
  );

  const handleSetAssessment = useCallback(
    (value: string) => {
      markDirty();
      setAssessment(value);
    },
    [markDirty, setAssessment],
  );

  const handleSetDiagnosis1CategoryId = useCallback(
    (id: number | null) => {
      markDirty();
      setDiagnosis1CategoryId(id);
    },
    [markDirty, setDiagnosis1CategoryId],
  );

  const handleSetDiagnosis1NameId = useCallback(
    (id: number | null) => {
      markDirty();
      setDiagnosis1NameId(id);
    },
    [markDirty, setDiagnosis1NameId],
  );

  const handleSetDiagnosis2CategoryId = useCallback(
    (id: number | null) => {
      markDirty();
      setDiagnosis2CategoryId(id);
    },
    [markDirty, setDiagnosis2CategoryId],
  );

  const handleSetDiagnosis2NameId = useCallback(
    (id: number | null) => {
      markDirty();
      setDiagnosis2NameId(id);
    },
    [markDirty, setDiagnosis2NameId],
  );

  return {
    handleSetAssessment,
    handleSetChiefComplaint,
    handleSetChiefComplaintTypeId,
    handleSetDiagnosis1CategoryId,
    handleSetDiagnosis1NameId,
    handleSetDiagnosis2CategoryId,
    handleSetDiagnosis2NameId,
    handleSetPhysicalExam,
    handleSetPlan,
    handleSetTreatmentPolicy,
  };
}
