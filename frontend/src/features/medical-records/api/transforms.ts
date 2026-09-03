import { formatDate } from "@/lib/format/date";
import type { BackendMedicalRecord } from "./types";
import type { InterviewHistoryItem } from "../types";

export {
  transformMedicalRecord,
  toBackendMedicalRecordStatus,
  type MedicalRecord,
} from "@/lib/transforms/medical-record";

/** FEAT-003: BackendMedicalRecord → InterviewHistoryItem 変換（feature 専用 UI 型）。 */
export const transformToHistoryItem = (record: BackendMedicalRecord): InterviewHistoryItem => {
  const chiefComplaint = record.inquiry?.chief_complaint ?? "";
  const content = chiefComplaint || "（記録なし）";

  return {
    id: String(record.id ?? 0),
    date: formatDate(record.date),
    author: record.doctor?.name ?? "-",
    type: record.status === "finalized" ? "確定済" : "作成中",
    title: chiefComplaint || record.record_no,
    content,
  };
};
