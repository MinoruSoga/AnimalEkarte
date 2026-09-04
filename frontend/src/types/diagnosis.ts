import type {
  DiagnosisType as ModelDiagnosisType,
  DiagnosisName as ModelDiagnosisName,
} from "@/types/generated/models";
import type { ReorderRequest } from "@/types/form";

// Server-managed fields excluded from request types
type CategoryServerFields = "id" | "clinic_id" | "created_at" | "updated_at" | "names";
type NameServerFields = "id" | "clinic_id" | "created_at" | "updated_at" | "category";

// ─────────────────────────────────────────────────
// DiagnosisType request types (models.ts から導出)
// ─────────────────────────────────────────────────

export type CreateDiagnosisTypeRequest = Required<Pick<ModelDiagnosisType, "name">> &
  Partial<Omit<ModelDiagnosisType, CategoryServerFields | "name">>;

export type UpdateDiagnosisTypeRequest = Partial<Omit<ModelDiagnosisType, CategoryServerFields>>;

// ─────────────────────────────────────────────────
// DiagnosisName request types (models.ts から導出)
// ─────────────────────────────────────────────────

export type CreateDiagnosisNameRequest = Required<
  Pick<ModelDiagnosisName, "name" | "diagnosis_type_id">
> &
  Partial<Omit<ModelDiagnosisName, NameServerFields | "name" | "diagnosis_type_id">>;

export type UpdateDiagnosisNameRequest = Partial<Omit<ModelDiagnosisName, NameServerFields>>;

// ─────────────────────────────────────────────────
// Reorder request types (models.ts に対応フィールドなし → 手書き)
// ─────────────────────────────────────────────────

export type ReorderDiagnosisTypeRequest = ReorderRequest;

export type ReorderDiagnosisNameRequest = ReorderRequest;
