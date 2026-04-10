import type { ReservationCategory as ModelReservationCategory } from "@/types/generated/models";

// Server-managed fields excluded from request types
type ServerFields = "id" | "clinic_id" | "created_at" | "updated_at";

// ─────────────────────────────────────────────────
// ReservationCategory request types (models.ts から導出)
// ─────────────────────────────────────────────────

export type CreateReservationCategoryRequest =
  Required<Pick<ModelReservationCategory, "name">> &
  Partial<Omit<ModelReservationCategory, ServerFields | "name">>;

export type UpdateReservationCategoryRequest =
  Partial<Omit<ModelReservationCategory, ServerFields>>;

// ─────────────────────────────────────────────────
// Reorder request type (手書き)
// ─────────────────────────────────────────────────

export interface ReorderReservationCategoryRequest {
  ids: number[];
}
