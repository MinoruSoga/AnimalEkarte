import type { ExaminationRecord, ExaminationItem } from "@/types";
import type { BackendExamination, BackendExaminationItem } from "./types";

function transformExaminationItem(
  item: BackendExaminationItem
): ExaminationItem {
  return {
    id: item.id,
    name: item.name,
    result: item.result,
    unit: item.unit,
    referenceRange: item.ref,
  };
}

export function transformExamination(
  data: BackendExamination
): ExaminationRecord {
  return {
    id: data.id,
    date: data.date,
    ownerName: "",
    petName: data.pet?.name ?? "",
    testType: data.exam_type?.name ?? "",
    doctor: data.doctor?.name ?? data.doctor_id ?? "",
    status: data.status,
    resultSummary: data.result_summary,
    machine: data.machine,
    items: data.items?.map(transformExaminationItem),
  };
}
