/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { ReservationAppointment } from "@/types/generated/models";

// Backend 型エイリアス
export type BackendReservationAppointment = ReservationAppointment;

/**
 * 予約作成リクエスト
 * models.ts の ReservationAppointment から導出
 */
export interface CreateReservationRequest {
  pet_id: number;
  owner_id: number;
  doctor_id?: number;
  start_time: string;
  end_time: string;
  visit_type: string;
  service_type_id: number;
  is_designated?: boolean;
  notes?: string;
  source?: "manual" | "line";
}

/**
 * 予約更新リクエスト
 * models.ts の ReservationAppointment から導出
 */
export interface UpdateReservationRequest {
  start_time?: string;
  end_time?: string;
  visit_type?: string;
  service_type_id?: number;
  doctor_id?: number;
  is_designated?: boolean;
  status?: string;
  notes?: string;
}
