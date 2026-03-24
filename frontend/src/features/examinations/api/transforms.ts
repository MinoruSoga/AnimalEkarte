import type { ExaminationRecord, ExaminationItem } from "@/types";
import type { BackendExamination, BackendExaminationItem } from "./types";

const EXAM_STATUS_EN_TO_JA: Record<string, "依頼中" | "検査中" | "完了"> = {
  pending: "依頼中",
  in_progress: "検査中",
  completed: "完了",
};

function transformExaminationItem(
  item: BackendExaminationItem
): ExaminationItem {
  return {
    id: String(item.id ?? 0),
    name: item.name ?? "",
    result: item.result ?? "",
    unit: item.unit ?? "",
    referenceRange: item.ref ?? "",
  };
}

export function transformExamination(
  data: BackendExamination
): ExaminationRecord {
  return {
    id: String(data.id ?? 0),
    date: data.date ? data.date.split("T")[0] : "",
    ownerName: data.pet?.owner?.owner_name ?? "",
    petName: data.pet?.name ?? "",
    testType: data.exam_type?.name ?? "",
    testTypeId: String(data.exam_type_id ?? ""),
    doctor: data.doctor?.name ?? String(data.doctor_id ?? ""),
    doctorId: String(data.doctor_id ?? ""),
    status: EXAM_STATUS_EN_TO_JA[data.status ?? ""] ?? "依頼中",
    resultSummary: data.result_summary ?? undefined,
    machine: data.machine ?? undefined,
    items: data.items?.map(transformExaminationItem),
  };
}
