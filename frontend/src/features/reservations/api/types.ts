/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { Reservation as ApiReservation } from "@/types/generated/models";
import type { ReservationRoute } from "../constants/reservation-route";

export type BackendReservation = ApiReservation;

/**
 * 予約作成リクエスト
 * models.ts の Appointment から導出
 */
export interface CreateReservationRequest {
  pet_id: number;
  owner_id: number;
  doctor_id?: number;
  start_time: string;
  end_time: string;
  visit_type: string;
  reservation_type_id: number;
  is_designated?: boolean;
  status?: string;
  notes?: string;
  source?: "manual" | "line";
  reservation_route?: ReservationRoute;
}

// R-F2-S12: UpdateReservationRequest は src/hooks/use-update-reservation.ts へ単一ソース化。
// ここは feature 内既存 import (`./types`) を壊さないための re-export（S8 パターン踏襲）。
export type { UpdateReservationRequest } from "@/hooks/use-update-reservation";
