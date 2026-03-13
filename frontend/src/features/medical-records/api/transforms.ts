import type { MedicalRecord } from "@/types";
import type { BackendMedicalRecord } from "./types";

export const transformMedicalRecord = (
  record: BackendMedicalRecord
): MedicalRecord => {
  const statusMap: Record<string, MedicalRecord["status"]> = {
    draft: "作成中",
    finalized: "確定済",
  };

  return {
    id: String(record.id ?? 0),
    recordNo: record.record_no,
    date: record.date,
    ownerId: record.owner_id ? String(record.owner_id) : undefined,
    ownerName: record.owner?.owner_name ?? "",
    petId: record.pet_id ? String(record.pet_id) : undefined,
    petName: record.pet?.name ?? "",
    species: record.pet?.animal_species?.name ?? "",
    chiefComplaint: record.inquiry?.chief_complaint ?? "",
    doctor: record.doctor?.name ?? String(record.doctor_id ?? ""),
    status: statusMap[record.status] ?? "作成中",
    visitType: undefined,
    subjective: undefined,
    objective: record.clinical_plan?.physical_exam,
    assessment: record.clinical_plan?.diagnosis_details,
    plan: record.clinical_plan?.treatment_policy,
    surgeryNotes: undefined,
    diagnosis: record.clinical_plan?.diagnosis_details,
    treatment: record.clinical_plan?.treatment_policy,
    prescription: undefined,
    notes: record.inquiry?.notes,
    accountingId: undefined,
  };
};
