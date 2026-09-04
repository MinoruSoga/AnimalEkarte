import type {
  AnesthesiaType,
  Consultation,
  ExaminationType,
  Procedure,
  Vaccine,
  CheckupType,
  TaxType,
} from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Consultation
// ─────────────────────────────────────────────────

export const transformConsultation = (data: Consultation) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  price: data.price ?? 0,
  isActive: data.is_active,
  description: data.description ?? "",
  sortOrder: data.sort_order ?? 0,
  timeCondition: data.time_condition ?? "",
  duration: data.duration,
  taxType: (data.tax_type ?? "excluded") as TaxType,
  taxRate: data.tax_rate ?? 0.1,
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
});

export type ConsultationItem = ReturnType<typeof transformConsultation>;

// ─────────────────────────────────────────────────
// ExaminationType
// ─────────────────────────────────────────────────

export const transformExaminationType = (data: ExaminationType) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  price: data.price ?? 0,
  isActive: data.is_active,
  description: data.description ?? "",
  sortOrder: data.sort_order ?? 0,
  isNonInsurance: data.is_non_insurance ?? false,
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
});

// ─────────────────────────────────────────────────
// Procedure
// ─────────────────────────────────────────────────

export const transformProcedure = (data: Procedure) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  price: data.price ?? 0,
  isActive: data.is_active,
  description: data.description ?? "",
  sortOrder: data.sort_order ?? 0,
  anesthesia: data.anesthesia ?? "none",
  duration: data.duration,
  taxType: (data.tax_type ?? "excluded") as TaxType,
  taxRate: data.tax_rate ?? 0.1,
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
});

export type ProcedureItem = ReturnType<typeof transformProcedure>;

// ─────────────────────────────────────────────────
// Vaccine
// ─────────────────────────────────────────────────

export const transformVaccine = (data: Vaccine) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  price: data.price ?? 0,
  isActive: data.is_active,
  description: data.description ?? "",
  sortOrder: data.sort_order ?? 0,
  species: data.species,
  interval: data.interval ?? "",
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
});

export type VaccineItem = ReturnType<typeof transformVaccine>;

// ─────────────────────────────────────────────────
// CheckupType
// ─────────────────────────────────────────────────

export const transformCheckupType = (data: CheckupType) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  price: data.price ?? 0,
  isActive: data.is_active,
  description: data.description ?? "",
  sortOrder: data.sort_order ?? 0,
  interval: data.interval ?? "",
  targetAge: data.target_age ?? "",
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
});

export type CheckupTypeItem = ReturnType<typeof transformCheckupType>;

// ─────────────────────────────────────────────────
// 共通の最小型（5エンティティに共通するフィールド）
// ─────────────────────────────────────────────────

export type TreatmentItem = {
  id: string;
  name: string;
  parentId?: string;
  price: number;
  isActive: boolean;
  description: string;
  sortOrder: number;
  taxType?: TaxType;
  taxRate?: number;
  isNonInsurance?: boolean;
  /** Procedure only — present when the item is a procedure master row */
  anesthesia?: AnesthesiaType;
};
