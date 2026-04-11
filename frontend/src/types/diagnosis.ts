import type {
  DiagnosisCategory as ModelDiagnosisCategory,
  DiagnosisName as ModelDiagnosisName,
} from "@/types/generated/models";

// Server-managed fields excluded from request types
type CategoryServerFields = "id" | "clinic_id" | "created_at" | "updated_at" | "names";
type NameServerFields = "id" | "clinic_id" | "created_at" | "updated_at" | "category";

// ─────────────────────────────────────────────────
// DiagnosisCategory request types (models.ts から導出)
// ─────────────────────────────────────────────────

export type CreateDiagnosisCategoryRequest =
  Required<Pick<ModelDiagnosisCategory, "name">> &
  Partial<Omit<ModelDiagnosisCategory, CategoryServerFields | "name">>;

export type UpdateDiagnosisCategoryRequest =
  Partial<Omit<ModelDiagnosisCategory, CategoryServerFields>>;

// ─────────────────────────────────────────────────
// DiagnosisName request types (models.ts から導出)
// ─────────────────────────────────────────────────

export type CreateDiagnosisNameRequest =
  Required<Pick<ModelDiagnosisName, "name" | "diagnosis_type_id">> &
  Partial<Omit<ModelDiagnosisName, NameServerFields | "name" | "diagnosis_type_id">>;

export type UpdateDiagnosisNameRequest =
  Partial<Omit<ModelDiagnosisName, NameServerFields>>;

// ─────────────────────────────────────────────────
// Reorder request types (models.ts に対応フィールドなし → 手書き)
// ─────────────────────────────────────────────────

export interface ReorderDiagnosisCategoryRequest {
  ids: number[];
}

export interface ReorderDiagnosisNameRequest {
  ids: number[];
}
