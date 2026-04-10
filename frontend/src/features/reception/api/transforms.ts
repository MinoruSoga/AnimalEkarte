import { format } from "date-fns";
import type { ReservationStatus } from "@/types";
import type { ReservationAppointment as BackendReceptionReservation } from "@/types/generated/models";
import type { ReceptionAppointment, ReceptionColumn } from "./types";

/** Backend status 値 → カンバンカラム ID のマッピング */
const STATUS_TO_COLUMN_ID: Record<string, ReservationStatus> = {
  pending: "pending",
  confirmed: "pending",
  checked_in: "checked_in",
  in_consultation: "in_consultation",
  accounting: "accounting",
  completed: "completed",
  cancelled: "cancelled",
};

/** カンバンカラム定義（表示順） */
export const RECEPTION_COLUMNS: Omit<ReceptionColumn, "appointments">[] = [
  { id: "pending", title: "受付予約" },
  { id: "checked_in", title: "受付済" },
  { id: "in_consultation", title: "診療中" },
  { id: "accounting", title: "会計待ち" },
  { id: "completed", title: "会計済" },
];

/** カラム ID → 日本語タイトルマッピング */
export const COLUMN_ID_TO_TITLE: Record<ReservationStatus, string> = {
  pending: "受付予約",
  confirmed: "受付予約",
  checked_in: "受付済",
  in_consultation: "診療中",
  accounting: "会計待ち",
  completed: "会計済",
  cancelled: "会計済",
};

/** 日本語タイトル → カラム ID マッピング */
export const COLUMN_TITLE_TO_STATUS: Record<string, ReservationStatus> = {
  受付予約: "pending",
  受付済: "checked_in",
  診療中: "in_consultation",
  会計待ち: "accounting",
  会計済: "completed",
};

/** visit_type の英語 → 日本語変換 */
function visitTypeToJapanese(visitType: string): "初診" | "再診" {
  return visitType === "first" ? "初診" : "再診";
}

/** animal_species_id → 動物種名マッピング */
const ANIMAL_SPECIES_MAP: Record<number, string> = {
  1: "犬",
  2: "猫",
  3: "ウサギ",
  4: "ハムスター",
  5: "モルモット",
  6: "フェレット",
  7: "鳥",
  8: "爬虫類",
  9: "その他",
};

/** BackendReceptionReservation → ReceptionAppointment 変換 */
export function transformReservationToReceptionAppointment(
  reservation: BackendReceptionReservation
): ReceptionAppointment {
  const startDate = new Date(reservation.start_time);
  const time = format(startDate, "HH:mm");

  const petName = reservation.pet?.name ?? "";
  // animal_species ネストがないため、animal_species_id からマッピング
  const petType = reservation.pet?.animal_species?.name
    ?? (reservation.pet?.animal_species_id ? ANIMAL_SPECIES_MAP[reservation.pet.animal_species_id] : "犬")
    ?? "犬";
  const ownerName = reservation.owner?.owner_name ?? "";

  const status = STATUS_TO_COLUMN_ID[reservation.status] ?? "pending";

  return {
    id: String(reservation.id ?? 0),
    time,
    ownerName,
    petType,
    petName,
    visitType: visitTypeToJapanese(reservation.visit_type),
    reservationCategory: reservation.reservation_category?.name ?? "",
    isDesignated: reservation.is_designated,
    doctor: reservation.doctor?.name ?? (reservation.doctor_id ? String(reservation.doctor_id) : undefined),
    petId: String(reservation.pet_id ?? 0),
    ownerId: String(reservation.owner_id ?? 0),
    status,
    notes: reservation.notes || undefined,
    source: (reservation.source as "manual" | "line") ?? "manual",
  };
}

/**
 * Reservation 配列 → ReceptionColumn 配列変換
 * キャンセル済みの予約はカンバンに表示しない
 */
export function transformReservationsToReceptionColumns(
  reservations: BackendReceptionReservation[]
): ReceptionColumn[] {
  const activeReservations = reservations.filter((r) => r.status !== "cancelled");
  const appointments = activeReservations.map(transformReservationToReceptionAppointment);

  return RECEPTION_COLUMNS.map((col) => ({
    ...col,
    appointments: appointments.filter((a) => {
      if (col.id === "pending") {
        return a.status === "pending" || a.status === "confirmed";
      }
      return a.status === col.id;
    }),
  }));
}
