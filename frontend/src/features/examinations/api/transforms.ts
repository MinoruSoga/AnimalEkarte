import type { ExaminationRecord, ExaminationItem } from "@/types";
import type { BackendExamination, BackendExaminationItem } from "./types";

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
    date: data.date ?? "",
    ownerName: "",
    petName: data.pet?.name ?? "",
    testType: data.exam_type?.name ?? "",
    doctor: data.doctor?.name ?? String(data.doctor_id ?? ""),
    status: (data.status ?? "依頼中") as "依頼中" | "検査中" | "完了",
    resultSummary: data.result_summary ?? undefined,
    machine: data.machine ?? undefined,
    items: data.items?.map(transformExaminationItem),
  };
}
