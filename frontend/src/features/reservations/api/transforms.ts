import type { ReservationAppointment } from "@/types";
import type { ReservationAppointment as BackendReservation } from "@/types/generated/models";
import type { CreateReservationRequest } from "./types";

/** customer_fields JSON（LINE予約のオーナー未紐付け時のフォールバック用） */
interface CustomerFieldsJSON {
  customer_name?: string;
  owner_name?: string;
  pets?: Array<{ name?: string; type?: string }>;
}

function parseCustomerFields(raw: string | undefined): CustomerFieldsJSON {
  if (!raw) return {};
  try {
    return JSON.parse(raw) as CustomerFieldsJSON;
  } catch {
    return {};
  }
}

export const transformReservation = (
  reservation: BackendReservation
): ReservationAppointment => {
  // LINE予約でオーナー未紐付けの場合、customer_fields をフォールバックとして使用
  const cf = parseCustomerFields(reservation.customer_fields);
  const ownerName =
    reservation.owner?.name ??
    reservation.pet?.owner?.name ??
    cf.owner_name ??
    cf.customer_name ??
    "";
  const petName = reservation.pet?.name ?? cf.pets?.[0]?.name ?? "";

  return {
    id: String(reservation.id ?? 0),
    start: new Date(reservation.start_time),
    end: new Date(reservation.end_time),
    ownerName,
    petName,
    visitType: (reservation.visit_type as "first" | "revisit") ?? "first",
    type: reservation.reservation_category?.name ?? "",
    reservationCategoryId: reservation.reservation_category_id ? String(reservation.reservation_category_id) : undefined,
    doctor: reservation.doctor?.name ?? "",
    doctorId: reservation.doctor_id ? String(reservation.doctor_id) : undefined,
    isDesignated: reservation.is_designated ?? false,
    status: (reservation.status as ReservationAppointment["status"]) ?? "pending",
    notes: reservation.notes || undefined,
    petId: reservation.pet_id ? String(reservation.pet_id) : undefined,
    source: (reservation.source as "manual" | "line") ?? "manual",
  };
};

export const transformToCreateRequest = (
  data: Partial<ReservationAppointment>,
  petId: string,
  ownerId: string
): CreateReservationRequest => {
  return {
    pet_id: Number(petId),
    owner_id: Number(ownerId),
    start_time: data.start ? data.start.toISOString() : "",
    end_time: data.end ? data.end.toISOString() : "",
    visit_type: data.visitType ?? "first",
    reservation_category_id: Number(data.type ?? 0),
    doctor_id: data.doctor ? Number(data.doctor) : undefined,
    is_designated: data.isDesignated ?? false,
    notes: data.notes,
    source: data.source,
  };
};
