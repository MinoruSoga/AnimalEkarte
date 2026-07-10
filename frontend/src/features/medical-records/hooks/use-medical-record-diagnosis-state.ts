import { useState } from "react";
import type { RecommendationReason } from "../constants/recommendation-reason";
import {
  DEFAULT_ASSESSMENT,
  DEFAULT_CHIEF_COMPLAINT,
  DEFAULT_PLAN,
  DEFAULT_TREATMENT_POLICY,
} from "./use-medical-record-form-model";

export function useMedicalRecordDiagnosisState() {
  // 問診タブの状態
  const [chiefComplaint, setChiefComplaint] = useState(DEFAULT_CHIEF_COMPLAINT);
  const [chiefComplaintTypeId, setChiefComplaintTypeId] = useState<number | null>(null);
  const [treatmentPolicy, setTreatmentPolicy] = useState(DEFAULT_TREATMENT_POLICY);

  // 診察/治療プランタブの状態（SOAPS）
  const [plan, setPlan] = useState(DEFAULT_PLAN);
  const [assessment, setAssessment] = useState(DEFAULT_ASSESSMENT);

  // 診断マスタの状態
  const [diagnosis1CategoryId, setDiagnosis1CategoryId] = useState<number | null>(null);
  const [diagnosis1NameId, setDiagnosis1NameId] = useState<number | null>(null);
  const [diagnosis2CategoryId, setDiagnosis2CategoryId] = useState<number | null>(null);
  const [diagnosis2NameId, setDiagnosis2NameId] = useState<number | null>(null);

  // 推奨理由 (create mode 専用 state; edit mode では existingRecord から取得)
  const [createRecommendationReason, setCreateRecommendationReason] =
    useState<RecommendationReason | null>(null);

  return {
    chiefComplaint,
    setChiefComplaint,
    chiefComplaintTypeId,
    setChiefComplaintTypeId,
    treatmentPolicy,
    setTreatmentPolicy,
    plan,
    setPlan,
    assessment,
    setAssessment,
    diagnosis1CategoryId,
    setDiagnosis1CategoryId,
    diagnosis1NameId,
    setDiagnosis1NameId,
    diagnosis2CategoryId,
    setDiagnosis2CategoryId,
    diagnosis2NameId,
    setDiagnosis2NameId,
    createRecommendationReason,
    setCreateRecommendationReason,
  };
}
