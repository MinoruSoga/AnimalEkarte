import type { LineReservationSetting, LineCustomer } from "@/types/generated/models";

// ── Re-export backend types ──
export type { LineReservationSetting, LineCustomer };
// Backward-compat aliases
export type ReservationSetting = LineReservationSetting;
export type ReservationCustomer = LineCustomer;

// ── API Request types ──
export type UpdateReservationSettingRequest = Omit<LineReservationSetting, "id" | "clinic_id" | "created_at" | "updated_at">;

export interface LinkOwnerRequest {
  owner_id: number;
}
