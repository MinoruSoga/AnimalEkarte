import { ReservationAppointment } from "@/types";
import type { BackendReservation, CreateReservationRequest } from "./types";

export const transformReservation = (
  reservation: BackendReservation
): ReservationAppointment => {
  return {
    id: reservation.id,
    start: new Date(reservation.start_time),
    end: new Date(reservation.end_time),
    ownerName: "",
    petName: "",
    visitType: (reservation.visit_type as "first" | "revisit") || "first",
    type: reservation.service_type || "",
    doctor: reservation.doctor_id || "",
    isDesignated: reservation.is_designated || false,
    status: (reservation.status as ReservationAppointment["status"]) || "pending",
    notes: reservation.notes,
    petId: reservation.pet_id,
  };
};

export const transformToCreateRequest = (
  data: Partial<ReservationAppointment>,
  petId: string,
  ownerId: string
): CreateReservationRequest => {
  return {
    pet_id: petId,
    owner_id: ownerId,
    start_time: data.start ? data.start.toISOString() : "",
    end_time: data.end ? data.end.toISOString() : "",
    visit_type: data.visitType || "first",
    service_type: data.type || "",
    doctor_id: data.doctor || undefined,
    is_designated: data.isDesignated || false,
    notes: data.notes,
  };
};
