import { formatDate } from "@/lib/format/date";
import type { MedicalRecordResponse } from "@/types/generated/medicalrecord-responses";

/**
 * FE-RC-015 followup3: medical-record wire→UI 変換の単一正本。
 * hooks / features は本モジュールを参照する（複製禁止）。
 * TASK-444-S1: @/types/generated/models は使わない。
 */

type MedicalRecordStatus = "作成中" | "確定済";

type MedicalRecordRecommendationReason = "revisit" | "checkup" | "prevention" | "exam";

const STATUS_MAP: Record<string, MedicalRecordStatus> = {
  draft: "作成中",
  finalized: "確定済",
};

/** 日本語UIラベル（作成中/確定済）→ BE enum（draft/finalized）の逆引き。BUG-B1 */
const STATUS_MAP_TO_BACKEND: Record<MedicalRecordStatus, string> = {
  作成中: "draft",
  確定済: "finalized",
};

export const toBackendMedicalRecordStatus = (label: string): string | undefined =>
  STATUS_MAP_TO_BACKEND[label as MedicalRecordStatus];

function toVisitTypeLabel(visitType?: string | null): string | undefined {
  if (!visitType) return undefined;
  if (visitType === "first" || visitType === "初診") return "初診";
  if (visitType === "revisit" || visitType === "再診") return "再診";
  return visitType;
}

/**
 * MedicalRecordResponse wire → UI.
 * clinical_plan / visit_type は detail wire に無い（clinical-plan API / form 別経路）。
 * inquiry.notes は InquirySummary に載せ、問診「治療方針」を再読込 hydrate する（BUG-034）。
 * version は wire 必須（TASK-444-S2 選択肢A）。?? 1 フォールバックは置かない。
 */
export const transformMedicalRecord = (record: MedicalRecordResponse) => {
  return {
    id: String(record.id ?? 0),
    recordNo: record.record_no,
    date: formatDate(record.date),
    ownerId: record.owner_id ? String(record.owner_id) : undefined,
    ownerName: record.owner?.name ?? "",
    petId: record.pet_id ? String(record.pet_id) : undefined,
    petName: record.pet?.name ?? "",
    petIsDeceased: record.pet?.status === "deceased",
    species: record.pet?.animal_species?.name ?? "",
    chiefComplaint: record.inquiry?.chief_complaint ?? "",
    chiefComplaintTypeId: record.inquiry?.chief_complaint_type_id ?? null,
    doctor: record.doctor?.name ?? String(record.doctor_id ?? ""),
    visitType: toVisitTypeLabel(record.visit_type),
    nextVisitRecommendedDate: record.next_visit_recommended_date ?? "",
    subjective: undefined as string | undefined,
    objective: undefined as string | undefined,
    assessment: undefined as string | undefined,
    plan: undefined as string | undefined,
    surgeryNotes: undefined as string | undefined,
    diagnosis: undefined as string | undefined,
    treatment: undefined as string | undefined,
    prescription: undefined as string | undefined,
    diagnosis1CategoryId: null as number | null,
    diagnosis1NameId: null as number | null,
    diagnosis2CategoryId: null as number | null,
    diagnosis2NameId: null as number | null,
    notes: record.inquiry?.notes || undefined,
    accountingId: record.accounting_id ? String(record.accounting_id) : undefined,
    visitCount: record.visit_count,
    version: record.version,
    recommendationReason: (record.recommendation_reason ??
      null) as MedicalRecordRecommendationReason | null,
    clinicId: record.clinic_id ? String(record.clinic_id) : undefined,
    status: (STATUS_MAP[record.status] ?? "作成中") as MedicalRecordStatus,
  };
};

export type MedicalRecord = ReturnType<typeof transformMedicalRecord>;
