import { useState } from "react";
import type { RecommendationReason } from "../constants/recommendation-reason";
import { DEFAULT_CHIEF_COMPLAINT, DEFAULT_TREATMENT_POLICY } from "./use-medical-record-form-model";

export function useMedicalRecordDiagnosisState() {
  // 問診タブの状態
  const [chiefComplaint, setChiefComplaint] = useState(DEFAULT_CHIEF_COMPLAINT);
  const [chiefComplaintTypeId, setChiefComplaintTypeId] = useState<number | null>(null);
  // 問診 notes 用（clinical_plan.treatment_policy ではない）
  const [treatmentPolicy, setTreatmentPolicy] = useState(DEFAULT_TREATMENT_POLICY);

  // 診察/治療プランタブの状態 — clinical_plan 3欄の単一 owner（BUG-010）
  // 初期値は空。テンプレート固定文字列を初期表示に載せて保存すると入力が置換されるため使わない。
  const [physicalExam, setPhysicalExam] = useState("");
  const [plan, setPlan] = useState(""); // treatment_policy
  const [assessment, setAssessment] = useState(""); // diagnosis_details

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
    physicalExam,
    setPhysicalExam,
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
