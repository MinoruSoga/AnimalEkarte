import { jstWallDateToISOString } from "@/lib/jst-date";
import type { UpdateReservationRequest } from "@/hooks/use-update-reservation";

import type { Reservation, ReservationFormData, ReservationStatus } from "../types";

/** 詳細モーダルからの破壊的 status 変更（確認ダイアログ対象）。BUG-020 */
const DESTRUCTIVE_RESERVATION_STATUSES: readonly ReservationStatus[] = [
  "cancelled",
  "no_show",
] as const;

export function isDestructiveReservationStatus(status: ReservationStatus): boolean {
  return (DESTRUCTIVE_RESERVATION_STATUSES as readonly string[]).includes(status);
}

/** Notes-only edits must omit schedule/doctor so BE skips on-duty conflict checks (BUG-012). */
export function buildReservationUpdateRequest(
  current: ReservationFormData,
  data: ReservationFormData,
  targetDoctor: string,
): { id: string; req: UpdateReservationRequest } | null {
  if (!current.id) return null;

  const req: UpdateReservationRequest = {};
  if (data.start) {
    const nextStart = jstWallDateToISOString(data.start);
    const prevStart = current.start ? jstWallDateToISOString(current.start) : "";
    if (nextStart !== prevStart) req.start_time = nextStart;
  }
  if (data.end) {
    const nextEnd = jstWallDateToISOString(data.end);
    const prevEnd = current.end ? jstWallDateToISOString(current.end) : "";
    if (nextEnd !== prevEnd) req.end_time = nextEnd;
  }
  const nextVisit = data.visitType || "first";
  if (nextVisit !== (current.visitType || "first")) req.visit_type = nextVisit;
  const nextType = data.type ? Number(data.type) : undefined;
  const prevType = current.type ? Number(current.type) : undefined;
  if (nextType !== undefined && nextType !== prevType) req.reservation_type_id = nextType;
  const nextDoctor = targetDoctor ? Number(targetDoctor) : undefined;
  const prevDoctor = current.doctor ? Number(current.doctor) : undefined;
  if (nextDoctor !== prevDoctor && nextDoctor !== undefined) req.doctor_id = nextDoctor;
  if ((data.isDesignated ?? false) !== (current.isDesignated ?? false)) {
    req.is_designated = data.isDesignated ?? false;
  }
  const nextStatus = data.status || "confirmed";
  if (nextStatus !== (current.status || "confirmed")) req.status = nextStatus;
  if ((data.notes ?? "") !== (current.notes ?? "")) req.notes = data.notes;

  if (Object.keys(req).length === 0) {
    req.notes = data.notes ?? "";
  }
  return { id: current.id, req };
}

export interface StatusConfirmTarget {
  reservation: Reservation;
  status: ReservationStatus;
}

export function buildUpdatePayload(
  reservation: Reservation,
  start: Date,
  end: Date,
  status: ReservationStatus,
) {
  return {
    id: reservation.id,
    req: {
      start_time: jstWallDateToISOString(start),
      end_time: jstWallDateToISOString(end),
      visit_type: reservation.visitType,
      doctor_id: reservation.doctor ? Number(reservation.doctor) : undefined,
      is_designated: reservation.isDesignated,
      status,
      notes: reservation.notes,
    },
  };
}
