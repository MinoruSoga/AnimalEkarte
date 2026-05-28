/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { Reservation as ApiReservation } from "@/types/generated/models";

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
}

/**
 * 予約更新リクエスト
 * models.ts の Appointment から導出
 */
export interface UpdateReservationRequest {
  start_time?: string;
  end_time?: string;
  visit_type?: string;
  reservation_type_id?: number;
  doctor_id?: number;
  is_designated?: boolean;
  status?: string;
  notes?: string;
}
