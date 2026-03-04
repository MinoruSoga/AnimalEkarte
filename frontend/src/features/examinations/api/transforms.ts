import type { ExaminationRecord } from "@/types";
import type { BackendExamination } from "./types";

export function transformExamination(data: BackendExamination): ExaminationRecord {
  let items = undefined;
  if (data.items) {
    try {
      items = JSON.parse(data.items);
    } catch {
      items = undefined;
    }
  }

  return {
    id: data.id,
    date: data.examination_date,
    ownerName: data.owner?.name ?? "",
    petName: data.pet?.name ?? "",
    testType: data.test_type,
    doctor: data.doctor_id || "",
    status: data.status,
    resultSummary: data.result_summary,
    machine: data.machine,
    items: items,
  };
}
