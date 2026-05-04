import type { ReservationStatus } from "@/types";
import type { BackendReservation, CreateReservationRequest } from "./types";
import type { ReservationRoute } from "../constants/reservation-route";

/** customer_fields JSON（LINE予約のオーナー未紐付け時のフォールバック用） */
interface CustomerFieldsJSON {
  customer_name?: string;
  owner_name?: string;
  pets?: Array<{ name?: string; type?: string }>;
}

/**
 * customer_fields は json.RawMessage（Go）なので JSON オブジェクトとしてそのまま届く。
 * 型は any だが、JSONB 由来のオブジェクトとして扱う。
 */
function extractCustomerFields(raw: unknown): CustomerFieldsJSON {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) return {};
  return raw as CustomerFieldsJSON;
}

export const transformReservation = (
  reservation: BackendReservation
) => {
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
    start: new Date(reservation.start_time),
    end: new Date(reservation.end_time),
    ownerName,
    petName,
    petType,
    visitType: (reservation.visit_type as "first" | "revisit") ?? "first",
    type: reservation.reservation_type?.name ?? "",
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
  };
};

export type Reservation = ReturnType<typeof transformReservation>;

export const transformToCreateRequest = (
  data: Partial<Reservation>,
  petId: string,
  ownerId: string
): CreateReservationRequest => {
  return {
    pet_id: Number(petId),
    owner_id: Number(ownerId),
    start_time: data.start ? data.start.toISOString() : "",
    end_time: data.end ? data.end.toISOString() : "",
    visit_type: data.visitType ?? "first",
    reservation_type_id: Number(data.type ?? 0),
    doctor_id: data.doctor ? Number(data.doctor) : undefined,
    is_designated: data.isDesignated ?? false,
    notes: data.notes,
    source: data.source,
  };
};
