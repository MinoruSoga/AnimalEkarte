import type { ReservationSetting, ReservationCustomer } from "@/types/generated/models";

// ── Re-export backend types ──
export type { ReservationSetting, ReservationCustomer };

// ── API Request types ──
export type UpdateReservationSettingRequest = Omit<ReservationSetting, "id" | "clinic_id" | "created_at" | "updated_at">;

export interface LinkOwnerRequest {
  owner_id: number;
}
