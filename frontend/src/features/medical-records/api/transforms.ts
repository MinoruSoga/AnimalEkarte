import { formatDate } from "@/lib/format/date";
import { PetStatusDeceased } from "@/types/generated/models";
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

/**
 * MedicalRecordResponse wire → UI.
 * clinical_plan / visit_type は detail wire に無い（clinical-plan API / form 別経路）。
 * inquiry.notes は InquirySummary に載せ、問診「治療方針」を再読込 hydrate する（BUG-034）。
 * version は wire 必須（TASK-444-S2 選択肢A）。?? 1 フォールバックは置かない。
 */
export const transformMedicalRecord = (record: BackendMedicalRecord) => {
  return {
    id: String(record.id ?? 0),
    recordNo: record.record_no,
    date: formatDate(record.date),
    ownerId: record.owner_id ? String(record.owner_id) : undefined,
    ownerName: record.owner?.name ?? "",
    petId: record.pet_id ? String(record.pet_id) : undefined,
    petName: record.pet?.name ?? "",
    petIsDeceased: record.pet?.status === PetStatusDeceased,
    species: record.pet?.animal_species?.name ?? "",
    chiefComplaint: record.inquiry?.chief_complaint ?? "",
    // InquirySummary wire に chief_complaint_type_id は無い
    chiefComplaintTypeId: null as number | null,
    doctor: record.doctor?.name ?? String(record.doctor_id ?? ""),
    // visit_type は medical-record detail wire に無い（form / 別経路）
    visitType: undefined as string | undefined,
    nextVisitRecommendedDate: record.next_visit_recommended_date ?? "",
    subjective: undefined as string | undefined,
    objective: undefined as string | undefined,
    assessment: undefined as string | undefined,
    plan: undefined as string | undefined,
    surgeryNotes: undefined as string | undefined,
    diagnosis: undefined as string | undefined,
    treatment: undefined as string | undefined,
    prescription: undefined as string | undefined,
    // 構造化診断は clinical-plan 専用 GET が正本（wire に clinical_plan は載らない）
    diagnosis1CategoryId: null as number | null,
    diagnosis1NameId: null as number | null,
    diagnosis2CategoryId: null as number | null,
    diagnosis2NameId: null as number | null,
    // 問診タブ「治療方針」= inquiry.notes（clinical_plan.treatment_policy とは別 state）
    notes: record.inquiry?.notes || undefined,
    accountingId: record.accounting_id ? String(record.accounting_id) : undefined,
    visitCount: record.visit_count,
    version: record.version,
    recommendationReason: (record.recommendation_reason ?? null) as RecommendationReason | null,
    clinicId: record.clinic_id ? String(record.clinic_id) : undefined,
    status: (STATUS_MAP[record.status] ?? "作成中") as MedicalRecordStatus,
  };
};

export type MedicalRecord = ReturnType<typeof transformMedicalRecord>;

/** FEAT-003: BackendMedicalRecord → InterviewHistoryItem 変換 */
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
