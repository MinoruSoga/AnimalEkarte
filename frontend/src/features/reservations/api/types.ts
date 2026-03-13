import type {
  ReservationAppointment as ApiReservation,
  Owner as ApiOwner,
  Pet as ApiPet,
} from "@/types/generated/models";

export type BackendReservation = ApiReservation;
export type BackendOwner = ApiOwner;
export type BackendPet = ApiPet;

export interface CreateReservationRequest {
  pet_id: string;
  owner_id: string;
  doctor_id?: string;
  start_time: string;
  end_time: string;
  visit_type: string;
  service_type: string;
  is_designated?: boolean;
  notes?: string;
}

export interface UpdateReservationRequest {
  start_time?: string;
  end_time?: string;
  visit_type?: string;
  service_type?: string;
  doctor_id?: string;
  is_designated?: boolean;
  status?: string;
  notes?: string;
}
