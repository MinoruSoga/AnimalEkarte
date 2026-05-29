import { format } from "date-fns";
import type { ReservationStatus } from "@/types";
import type { Reservation as BackendReceptionReservation } from "@/types/generated/models";

interface CustomerFieldsJSON {
  customer_name?: string;
  owner_name?: string;
  pets?: Array<{ name?: string; type?: string }>;
}

/**
 * customer_fields は json.RawMessage（Go）なので JSON オブジェクトとしてそのまま届く。
 * 型は any だが、JSONB 由来のオブジェクトとして扱う。
 */
function parseCustomerFields(raw: unknown): CustomerFieldsJSON {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) return {};
  return raw as CustomerFieldsJSON;
}

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

function optionalID(value: number | null | undefined): string {
  return value ? String(value) : "";
}

function padDatePart(value: number): string {
  return String(value).padStart(2, "0");
}

function formatJSTDate(date: Date): string {
  const jstDate = new Date(date.getTime() + 9 * 60 * 60 * 1000);
  return `${jstDate.getUTCFullYear()}-${padDatePart(jstDate.getUTCMonth() + 1)}-${padDatePart(jstDate.getUTCDate())}`;
}

/** BackendReceptionReservation → ReceptionAppointment 変換 */
export function transformReservationToReceptionAppointment(
  reservation: BackendReceptionReservation
) {
  const startDate = new Date(reservation.start_time);
  const time = format(startDate, "HH:mm");
  const cf = parseCustomerFields(reservation.customer_fields);

  const petName = reservation.pet?.name ?? cf.pets?.[0]?.name ?? "";
  // animal_species ネストがないため、animal_species_id からマッピング
  const petType = reservation.pet?.animal_species?.name
    ?? (reservation.pet?.animal_species_id ? ANIMAL_SPECIES_MAP[reservation.pet.animal_species_id] : undefined)
    ?? cf.pets?.[0]?.type
    ?? "犬";
  const ownerName = reservation.owner?.name ?? cf.owner_name ?? cf.customer_name ?? "";

  const status = STATUS_TO_COLUMN_ID[reservation.status] ?? "pending";

  return {
    id: String(reservation.id ?? 0),
    time,
    visitDate: formatJSTDate(startDate),
    ownerName,
    petType,
    petName,
    visitType: visitTypeToJapanese(reservation.visit_type),
    reservationType: reservation.reservation_type?.name ?? "",
    reservationCategory: reservation.reservation_type?.category ?? "general",
    isDesignated: reservation.is_designated,
    doctor: reservation.doctor?.name ?? (reservation.doctor_id ? String(reservation.doctor_id) : undefined),
    petId: optionalID(reservation.pet_id),
    ownerId: optionalID(reservation.owner_id),
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
  const activeReservations = reservations.filter((r) => r.status !== "cancelled" && r.status !== "no_show");
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

export type ReceptionAppointment = ReturnType<typeof transformReservationToReceptionAppointment>;
export type ReceptionColumn = {
  id: string;
  title: string;
  appointments: ReceptionAppointment[];
};
