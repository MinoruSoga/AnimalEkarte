/**
 * Reception API types
 * ReceptionAppointment は @/types に移動済み（shared layer）。
 * このファイルは後方互換のため re-export を維持する。
 */
import type { ReservationStatus, ReceptionAppointment } from "@/types";

export type { ReceptionAppointment };

/** 当日の受付カンバンカラム */
export interface ReceptionColumn {
  id: ReservationStatus;
  title: string;
  appointments: ReceptionAppointment[];
}

