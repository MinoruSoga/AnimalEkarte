import { formatDate } from "@/lib/format/date";
import type { BackendMedicalRecord } from "./types";
import type { InterviewHistoryCopySource, InterviewHistoryItem } from "../types";

export {
  transformMedicalRecord,
  toBackendMedicalRecordStatus,
  type MedicalRecord,
} from "@/lib/transforms/medical-record";

/** EMR-182: 前回複写ペイロード。複写可能な値が1つも無いときは undefined。 */
function toCopySource(
  inquiry: BackendMedicalRecord["inquiry"],
): InterviewHistoryCopySource | undefined {
  if (!inquiry) return undefined;
  const source: InterviewHistoryCopySource = {};
  if (inquiry.chief_complaint) source.chiefComplaint = inquiry.chief_complaint;
  if (inquiry.notes) source.treatmentPolicy = inquiry.notes;
  if (inquiry.chief_complaint_type_id != null) {
    source.chiefComplaintTypeId = inquiry.chief_complaint_type_id;
  }
  return Object.keys(source).length > 0 ? source : undefined;
}

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
    copySource: toCopySource(record.inquiry),
  };
};
