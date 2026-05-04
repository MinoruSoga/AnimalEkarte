/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { Reservation as ApiReservation } from "@/types/generated/models";

/** codegen 前の暫定拡張 — codegen 実行後に削除して ApiReservation に戻す */
export type BackendReservation = ApiReservation & {
  reservation_route?: string | null; // 暫定 — codegen 後に削除
};

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
