import { formatDate } from "@/utils/format/date";
import type { BackendMedicalRecord } from "./types";
import type { InterviewHistoryItem } from "../types";

type MedicalRecordStatus = "作成中" | "確定済";

export const transformMedicalRecord = (
  record: BackendMedicalRecord
) => {
  const statusMap: Record<string, MedicalRecordStatus> = {
    draft: "作成中",
    finalized: "確定済",
  };

  return {
    id: String(record.id ?? 0),
    recordNo: record.record_no,
    date: formatDate(record.date),
    ownerId: record.owner_id ? String(record.owner_id) : undefined,
    ownerName: record.owner?.name ?? "",
    petId: record.pet_id ? String(record.pet_id) : undefined,
    petName: record.pet?.name ?? "",
    species: record.pet?.animal_species?.name ?? "",
    chiefComplaint: record.inquiry?.chief_complaint ?? "",
    doctor: record.doctor?.name ?? String(record.doctor_id ?? ""),
    status: (statusMap[record.status] ?? "作成中") as MedicalRecordStatus,
    visitType: undefined,
    subjective: undefined,
    objective: record.clinical_plan?.physical_exam,
    assessment: record.clinical_plan?.diagnosis_details,
    plan: record.clinical_plan?.treatment_policy,
    surgeryNotes: undefined,
    diagnosis: undefined, // clinical_plan.diagnosis_details は assessment にマップ済み
    treatment: undefined, // clinical_plan.treatment_policy は plan にマップ済み
    prescription: undefined,
    notes: record.inquiry?.notes,
    accountingId: record.accounting_id ? String(record.accounting_id) : undefined,
    visitCount: record.visit_count,
    version: record.version ?? 1,
  };
};

export type MedicalRecord = ReturnType<typeof transformMedicalRecord>;

/** FEAT-003: BackendMedicalRecord → InterviewHistoryItem 変換 */
export const transformToHistoryItem = (record: BackendMedicalRecord): InterviewHistoryItem => {
  const chiefComplaint = record.inquiry?.chief_complaint ?? "";
  const notes = record.inquiry?.notes ?? "";
  const content = [chiefComplaint, notes].filter(Boolean).join("\n") || "（記録なし）";

  return {
    id: String(record.id ?? 0),
    date: formatDate(record.date),
    author: record.doctor?.name ?? "-",
    type: record.status === "finalized" ? "確定済" : "作成中",
    title: chiefComplaint || record.record_no,
    content,
  };
};
