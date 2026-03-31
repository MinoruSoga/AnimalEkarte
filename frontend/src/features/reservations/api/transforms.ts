import type { ReservationAppointment } from "@/types";
import type { ReservationAppointment as BackendReservation } from "@/types/generated/models";
import type { CreateReservationRequest } from "./types";

export const transformReservation = (
  reservation: BackendReservation
): ReservationAppointment => {
  return {
    id: String(reservation.id ?? 0),
    start: new Date(reservation.start_time),
    end: new Date(reservation.end_time),
    ownerName: reservation.owner?.owner_name ?? reservation.pet?.owner?.owner_name ?? "",
    petName: reservation.pet?.name ?? "",
    visitType: (reservation.visit_type as "first" | "revisit") ?? "first",
    type: reservation.service_type?.name ?? "",
    serviceTypeId: reservation.service_type_id ? String(reservation.service_type_id) : undefined,
    doctor: reservation.doctor?.name ?? "",
    doctorId: reservation.doctor_id ? String(reservation.doctor_id) : undefined,
    isDesignated: reservation.is_designated ?? false,
    status: (reservation.status as ReservationAppointment["status"]) ?? "pending",
    notes: reservation.notes || undefined,
    petId: reservation.pet_id ? String(reservation.pet_id) : undefined,
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
    service_type_id: Number(data.type ?? 0),
    doctor_id: data.doctor ? Number(data.doctor) : undefined,
    is_designated: data.isDesignated ?? false,
    notes: data.notes,
  };
};
