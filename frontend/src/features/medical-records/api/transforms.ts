import { formatDate } from "@/utils/format/date";
import { fromVisitTypeValue } from "../routes/medical-record-form-model";
import type { BackendMedicalRecord } from "./types";
import type { InterviewHistoryItem } from "../types";
import type { RecommendationReason } from "../constants/recommendation-reason";

type MedicalRecordStatus = "作成中" | "確定済";

const STATUS_MAP: Record<string, MedicalRecordStatus> = {
  draft: "作成中",
  finalized: "確定済",
};

/** 日本語UIラベル（作成中/確定済）→ BE enum（draft/finalized）の逆引き。BUG-B1: 一覧の server-side status フィルタ送信用 */
const STATUS_MAP_TO_BACKEND: Record<MedicalRecordStatus, string> = {
  作成中: "draft",
  確定済: "finalized",
};

export const toBackendMedicalRecordStatus = (label: string): string | undefined =>
  STATUS_MAP_TO_BACKEND[label as MedicalRecordStatus];

export const transformMedicalRecord = (
  record: BackendMedicalRecord
) => {
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
    chiefComplaintTypeId: record.inquiry?.chief_complaint_type_id ?? null,
    doctor: record.doctor?.name ?? String(record.doctor_id ?? ""),
    visitType: record.visit_type != null ? fromVisitTypeValue(record.visit_type) : undefined,
    nextVisitRecommendedDate: record.next_visit_recommended_date ?? "",
    subjective: undefined,
    objective: record.clinical_plan?.physical_exam,
    assessment: record.clinical_plan?.diagnosis_details,
    plan: record.clinical_plan?.treatment_policy,
    surgeryNotes: undefined,
    diagnosis: undefined, // clinical_plan.diagnosis_details は assessment にマップ済み
    treatment: undefined, // clinical_plan.treatment_policy は plan にマップ済み
    prescription: undefined,
    // BUG-410: 構造化診断(diagnosis1/2)のカテゴリ/名称IDを再読込経路で復元するためのマップ。
    // 欠落していると保存済み診断が編集保存で無言クリアされる(特に diagnosis2 は backend の
    // null=クリア契約のため直接データ喪失に繋がる)。
    diagnosis1CategoryId: record.clinical_plan?.diagnosis_type_id ?? null,
    diagnosis1NameId: record.clinical_plan?.diagnosis_name_id ?? null,
    diagnosis2CategoryId: record.clinical_plan?.diagnosis_2_type_id ?? null,
    diagnosis2NameId: record.clinical_plan?.diagnosis_2_name_id ?? null,
    notes: record.inquiry?.notes,
    accountingId: record.accounting_id ? String(record.accounting_id) : undefined,
    visitCount: record.visit_count,
    version: record.version ?? 1,
    recommendationReason: (record.recommendation_reason ?? null) as RecommendationReason | null,
    clinicId: record.clinic_id ? String(record.clinic_id) : undefined,
    status: (STATUS_MAP[record.status] ?? "作成中") as MedicalRecordStatus,
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
