import { useEffect } from "react";

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
  // ⚠️ レンダー中比較 (inline-comparison: useState(existingRecord) + if (prev !== existingRecord))
  // に書き換えてはならない。この state は useMedicalRecordSaveAction の useActionState
  // action closure から直接参照される (diagnosis1CategoryId 等)。render-phase setState は
  // 「マウント時点で既に existingRecord が値を持つ」場合 (TanStack Query のウォームキャッシュ。
  // useGetMedicalRecord は staleTime=QUERY_STALE_TIMES.MEDIUM=5分で、同一カルテの短時間内の
  // 再訪問では初回レンダーから data を返す) に prevExistingRecord の初期値が existingRecord と
  // 同一参照になり、hydrate が一度も発火しない (BUG-410 react-reviewer 指摘・RED で実証済み)。
  // useEffect ならマウント含め毎回のコミット後に必ず一度実行されるため安全。
  // 先例: 10f69364 (useAccountingDetailState で同型の render-phase setState バグを effect に差し戻し)。
  useEffect(() => {
    if (!existingRecord) return;
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
    // eslint-disable-next-line react-hooks/exhaustive-deps -- existingRecord のみで判定 (setter 群は安定参照)
  }, [existingRecord]);
}
