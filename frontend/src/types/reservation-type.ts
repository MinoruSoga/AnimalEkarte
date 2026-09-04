import type { ReservationType as ModelReservationType } from "@/types/generated/models";
import type { ReorderRequest } from "@/types/form";

// Server-managed fields excluded from request types
type ServerFields = "id" | "clinic_id" | "created_at" | "updated_at";

// ─────────────────────────────────────────────────
// ReservationType request types (models.ts から導出)
// ─────────────────────────────────────────────────

export type CreateReservationTypeRequest = Required<Pick<ModelReservationType, "name">> &
  Partial<Omit<ModelReservationType, ServerFields | "name">>;

export type UpdateReservationTypeRequest = Partial<Omit<ModelReservationType, ServerFields>>;

// ─────────────────────────────────────────────────
// Reorder request type (手書き)
// ─────────────────────────────────────────────────

export type ReorderReservationTypeRequest = ReorderRequest;
