import type { ReservationType as ModelReservationType } from "@/types/generated/models";
import type { ReorderRequest } from "@/types/form";

// EMR-527: 予約の移行 HTTP DTO（./reservation-legacy-wire.ts）が使う enum union の
// type-only 再 export。新規ファイルは generated/models を直接 import できない
// （TASK-444-S1）ため、既存の import 箇所を経由する。
export type {
  DangerLevel,
  PetStatus,
  ReservationDayOption,
  ReservationSource,
  ReservationTypeCategory,
} from "@/types/generated/models";

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
