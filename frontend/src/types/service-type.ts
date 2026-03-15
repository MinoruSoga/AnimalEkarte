import type { ServiceType as ModelServiceType } from "@/types/generated/models";

// Server-managed fields excluded from request types
type ServerFields = "id" | "clinic_id" | "created_at" | "updated_at";

// ─────────────────────────────────────────────────
// ServiceType request types (models.ts から導出)
// ─────────────────────────────────────────────────

export type CreateServiceTypeRequest =
  Required<Pick<ModelServiceType, "name">> &
  Partial<Omit<ModelServiceType, ServerFields | "name">>;

export type UpdateServiceTypeRequest =
  Partial<Omit<ModelServiceType, ServerFields>>;

// ─────────────────────────────────────────────────
// Reorder request type (手書き)
// ─────────────────────────────────────────────────

export interface ReorderServiceTypeRequest {
  ids: number[];
}
