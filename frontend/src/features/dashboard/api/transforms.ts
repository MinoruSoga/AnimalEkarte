import { format } from "date-fns";
import type {
  BackendDashboardReservation,
  DashboardAppointment,
  DashboardColumn,
  DashboardStatus,
} from "./types";

/** Backend status 値 → カンバンカラム ID のマッピング */
const STATUS_TO_COLUMN_ID: Record<string, DashboardStatus> = {
  pending: "pending",
  confirmed: "pending",    // confirmed は受付予約として扱う
  checked_in: "checked_in",
  in_consultation: "in_consultation",
  accounting: "accounting",
  completed: "completed",
  canceled: "canceled",
};

/** カンバンカラム定義（表示順） */
export const DASHBOARD_COLUMNS: Omit<DashboardColumn, "appointments">[] = [
  { id: "pending", title: "受付予約" },
  { id: "checked_in", title: "受付済" },
  { id: "in_consultation", title: "診療中" },
  { id: "accounting", title: "会計待ち" },
  { id: "completed", title: "会計済" },
];

/** カラム ID → 日本語タイトルマッピング */
export const COLUMN_ID_TO_TITLE: Record<DashboardStatus, string> = {
  pending: "受付予約",
  confirmed: "受付予約",
  checked_in: "受付済",
  in_consultation: "診療中",
  accounting: "会計待ち",
  completed: "会計済",
  canceled: "会計済", // キャンセルは会計済カラムには表示しない（後述のfilterで除外）
};

/** 日本語タイトル → カラム ID マッピング */
export const COLUMN_TITLE_TO_STATUS: Record<string, DashboardStatus> = {
  受付予約: "pending",
  受付済: "checked_in",
  診療中: "in_consultation",
  会計待ち: "accounting",
  会計済: "completed",
};

/** visit_type の英語 → 日本語変換 */
function mapVisitType(visitType: string): "初診" | "再診" {
  return visitType === "first" ? "初診" : "再診";
}

/** BackendDashboardReservation → DashboardAppointment 変換 */
export function transformReservationToDashboardAppointment(
  reservation: BackendDashboardReservation
): DashboardAppointment {
  const startDate = new Date(reservation.start_time);
  const time = format(startDate, "HH:mm");

  const petName = reservation.pet?.name ?? "";
  const petType = reservation.pet?.species ?? "犬";
  const ownerName = reservation.owner?.name ?? "";

  const status = STATUS_TO_COLUMN_ID[reservation.status] ?? "pending";

  return {
    id: reservation.id,
    time,
    ownerName,
    petType,
    petName,
    visitType: mapVisitType(reservation.visit_type),
    serviceType: reservation.service_type,
    isDesignated: reservation.is_designated,
    doctor: reservation.doctor_id,
    petId: reservation.pet_id,
    ownerId: reservation.owner_id,
    status,
  };
}

/**
 * Reservation 配列 → DashboardColumn 配列変換
 * キャンセル済みの予約はカンバンに表示しない
 */
export function transformReservationsToDashboardColumns(
  reservations: BackendDashboardReservation[]
): DashboardColumn[] {
  // キャンセル除外
  const active = reservations.filter((r) => r.status !== "canceled");
  const appointments = active.map(transformReservationToDashboardAppointment);

  return DASHBOARD_COLUMNS.map((col) => ({
    ...col,
    appointments: appointments.filter((a) => {
      // "confirmed" は "pending" カラムに配置
      if (col.id === "pending") {
        return a.status === "pending" || a.status === "confirmed";
      }
      return a.status === col.id;
    }),
  }));
}
