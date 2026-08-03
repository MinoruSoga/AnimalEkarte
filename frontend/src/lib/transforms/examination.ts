import type {
  Examination,
  ExamResult as BackendExamResult,
} from "@/types/generated/models";

type ExamResultResponse = BackendExamResult & {
  is_assessed?: boolean;
};

export const EXAM_STATUS_EN_TO_JA: Record<
  string,
  "依頼中" | "検査中" | "結果入力済み" | "完了" | "確定"
> = {
  pending: "依頼中",
  in_progress: "検査中",
  result_entered: "結果入力済み",
  completed: "完了",
  confirmed: "確定",
};

export function transformExamResult(item: ExamResultResponse) {
  return {
    id: String(item.id ?? 0),
    examTypeFieldId: item.exam_type_field_id ?? undefined,
    name: item.name ?? "",
    result: item.result ?? "",
    inspectionValue: item.inspection_value ?? "",
    normalValue: item.normal_value ?? "",
    unit: item.unit ?? "",
    referenceValue: item.reference_value ?? "",
    refMin: item.ref_min ?? undefined,
    refMax: item.ref_max ?? undefined,
    isAssessed: item.is_assessed,
    isAbnormal: item.is_abnormal ?? false,
    // backend が ref_min/ref_max から導出した判定結果（"normal" | "high" | "low"）
    status: (item.status ?? "normal") as "normal" | "high" | "low",
    sortOrder: item.sort_order ?? 0,
  };
}

export function transformExamination(data: Examination) {
  return {
    id: String(data.id ?? 0),
    date: data.date ? data.date.split("T")[0] : "",
    ownerName: data.pet?.owner?.name ?? "",
    petName: data.pet?.name ?? "",
    petId: data.pet_id ? String(data.pet_id) : undefined,
    medicalRecordId: data.medical_record_id
      ? String(data.medical_record_id)
      : undefined,
    testType: data.exam_type?.name ?? "",
    testTypeId: String(data.exam_type_id ?? ""),
    doctor: data.doctor?.name ?? String(data.doctor_id ?? ""),
    doctorId: String(data.doctor_id ?? ""),
    status: EXAM_STATUS_EN_TO_JA[data.status ?? ""] ?? "依頼中",
    currentRevisionVersion: data.current_revision_version,
    resultSummary: data.result_summary ?? undefined,
    machine: data.machine ?? undefined,
    items: data.items?.map(transformExamResult),
  };
}

export type ExamResult = ReturnType<typeof transformExamResult>;
export type ExaminationRecord = ReturnType<typeof transformExamination>;
