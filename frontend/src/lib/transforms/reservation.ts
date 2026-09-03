import type { ReservationStatus } from "@/types";
import { toJSTWallDate } from "@/lib/jst-date";
import type { Reservation as BackendReservation } from "@/types/generated/models";

// BE 契約: reservation_route は "line" | "phone" | "reception" | "exam_room" | "record_shortcut" のみ
// （features/reservations/constants/reservation-route.ts の RESERVATION_ROUTE_VALUES と同一値域）。
// lib/transforms は feature に依存しないためここで narrow union を再定義する。
export type ReservationRoute =
  | "line"
  | "phone"
  | "reception"
  | "exam_room"
  | "record_shortcut";

/** customer_fields JSON（LINE予約のオーナー未紐付け時のフォールバック用） */
interface CustomerFieldsJSON {
  customer_name?: string;
  owner_name?: string;
  pets?: Array<{ name?: string; type?: string }>;
}

/**
 * customer_fields は json.RawMessage（Go）なので JSON オブジェクトとしてそのまま届く。
 * 入力は unknown。オブジェクトであることを確認してから CustomerFieldsJSON として扱う。
 */
function extractCustomerFields(raw: unknown): CustomerFieldsJSON {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) return {};
  return raw as CustomerFieldsJSON;
}

/**
 * R-F2-S12: reservations feature から昇格。reception (use-reception-modal-handlers)
 * が reservations feature を直接 import しないための cross-feature 共有 transform。
 */
export const transformReservation = (reservation: BackendReservation) => {
  // LINE予約でオーナー未紐付けの場合、customer_fields をフォールバックとして使用
  const cf = extractCustomerFields(reservation.customer_fields);
  const ownerName =
    reservation.owner?.name ??
    reservation.pet?.owner?.name ??
    cf.owner_name ??
    cf.customer_name ??
    "";
  const petName = reservation.pet?.name ?? cf.pets?.[0]?.name ?? "";
  // ペットの種類（犬種等）: カルテ紐付け前は customer_fields から取得
  const petType = reservation.pet?.animal_species?.name ?? cf.pets?.[0]?.type;

  return {
    id: String(reservation.id ?? 0),
    start: toJSTWallDate(reservation.start_time),
    end: toJSTWallDate(reservation.end_time),
    ownerName,
    petName,
    petType,
    visitType: (reservation.visit_type as "first" | "revisit") ?? "first",
    type: reservation.reservation_type?.name ?? "",
    category: reservation.reservation_type?.category ?? "general",
    reservationTypeId: reservation.reservation_type_id ? String(reservation.reservation_type_id) : undefined,
    doctor: reservation.doctor?.name ?? "",
    doctorId: reservation.doctor_id ? String(reservation.doctor_id) : undefined,
    isDesignated: reservation.is_designated ?? false,
    status: (reservation.status as ReservationStatus) ?? "pending",
    notes: reservation.notes || undefined,
    petId: reservation.pet_id ? String(reservation.pet_id) : undefined,
    ownerId: reservation.owner_id ? String(reservation.owner_id) : reservation.owner?.id ? String(reservation.owner.id) : undefined,
    source: (reservation.source as "manual" | "line") ?? "manual",
    reservationRoute: (reservation.reservation_route ?? null) as ReservationRoute | null,
    clinicId: reservation.clinic_id ? String(reservation.clinic_id) : undefined,
  };
};

export type Reservation = ReturnType<typeof transformReservation>;
