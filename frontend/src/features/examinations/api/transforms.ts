import type { ExaminationRecord, ExamResult } from "@/types";
import type { BackendExamination, BackendExamResult } from "./types";

const EXAM_STATUS_EN_TO_JA: Record<string, "依頼中" | "検査中" | "結果入力済み" | "完了" | "確定"> = {
  pending: "依頼中",
  in_progress: "検査中",
  result_entered: "結果入力済み",
  completed: "完了",
  confirmed: "確定",
};

function transformExamResult(
  item: BackendExamResult
): ExamResult {
  return {
    id: String(item.id ?? 0),
    name: item.name ?? "",
    result: item.result ?? "",
    unit: item.unit ?? "",
    referenceRange: item.reference_value ?? "",
    isAbnormal: item.is_abnormal ?? false,
  };
}

export function transformExamination(
  data: BackendExamination
): ExaminationRecord {
  return {
    id: String(data.id ?? 0),
    date: data.date ? data.date.split("T")[0] : "",
    ownerName: data.pet?.owner?.name ?? "",
    petName: data.pet?.name ?? "",
    petId: data.pet_id ? String(data.pet_id) : undefined,
    medicalRecordId: data.medical_record_id ? String(data.medical_record_id) : undefined,
    testType: data.exam_type?.name ?? "",
    testTypeId: String(data.exam_type_id ?? ""),
    doctor: data.doctor?.name ?? String(data.doctor_id ?? ""),
    doctorId: String(data.doctor_id ?? ""),
    status: EXAM_STATUS_EN_TO_JA[data.status ?? ""] ?? "依頼中",
    resultSummary: data.result_summary ?? undefined,
    machine: data.machine ?? undefined,
    items: data.items?.map(transformExamResult),
  };
}
