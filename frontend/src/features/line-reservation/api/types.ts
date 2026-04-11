import type { LineReservationSetting, LineCustomer as LineCustomerModel } from "@/types/generated/models";

// ── Re-export backend types ──
export type { LineReservationSetting };
export type { LineCustomerModel as LineCustomer };
// Backward-compat aliases
export type ReservationSetting = LineReservationSetting;

// ── API Request types ──
export type UpdateLineReservationSettingRequest = Omit<LineReservationSetting, "id" | "clinic_id" | "created_at" | "updated_at">;

export interface LinkOwnerRequest {
  owner_id: number;
}
