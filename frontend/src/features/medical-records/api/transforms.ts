import { MedicalRecord } from "@/types";
import type { BackendMedicalRecord } from "./types";

export const transformMedicalRecord = (
  record: BackendMedicalRecord
): MedicalRecord => {
  return {
    id: record.id,
    recordNo: record.record_no,
    date: new Date(record.visit_date).toISOString().split("T")[0],
    ownerId: record.owner_id,
    ownerName: "", // Will be populated by frontend if needed
    petId: record.pet_id,
    petName: "", // Will be populated by frontend if needed
    species: "", // Will be populated by frontend if needed
    chiefComplaint: record.chief_complaint || "",
    doctor: record.doctor_id || "Unknown",
    status: record.status as "作成中" | "確定済",
    visitType: record.visit_type,
    subjective: record.subjective,
    objective: record.objective,
    assessment: record.assessment,
    plan: record.plan,
    surgeryNotes: record.surgery_notes,
    diagnosis: record.diagnosis,
    treatment: record.treatment,
    prescription: record.prescription,
    notes: record.notes,
    accountingId: record.accounting_id,
  };
};
