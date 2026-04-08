/**
 * Master API types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 *
 * 個別マスタ（Consultation, Procedure, Medicine 等）の型定義は
 * src/types/{treatment.ts, medicine.ts} に一元化済み。
 *
 * @deprecated Legacy types from old master_items CRUD. Use dedicated types instead:
 * - Consultation: {@link import("@/types").CreateConsultationRequest}
 * - Procedure: {@link import("@/types").CreateProcedureRequest}
 * - Medicine: {@link import("@/types").CreateMedicineRequest}
 * - Vaccine: {@link import("@/types").CreateVaccineRequest}
 * - CheckupType: {@link import("@/types").CreateCheckupTypeRequest}
 */

/**
 * @deprecated Use dedicated types from src/types/{treatment.ts, medicine.ts}
 */
export interface CreateMasterItemRequest {
  name: string;
  category: string;
  code?: string;
  price?: number | null;
  description?: string;
  inventory_id?: string | null;
  default_quantity?: number | null;
  // Consultation-specific fields
  timeCondition?: string;
  duration?: number | null;
}

/**
 * @deprecated Use dedicated types from src/types/{treatment.ts, medicine.ts}
 */
export interface UpdateMasterItemRequest {
  name?: string;
  category?: string;
  code?: string;
  price?: number | null;
  status?: "active" | "inactive";
  description?: string;
  inventory_id?: string | null;
  default_quantity?: number | null;
  // Consultation-specific fields
  timeCondition?: string;
  duration?: number | null;
}
