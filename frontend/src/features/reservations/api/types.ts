// Backend Reservation types - JSON keys match Go model snake_case
export interface BackendReservation {
  id: string;
  pet_id: string;
  owner_id: string;
  doctor_id?: string;
  start_time: string;
  end_time: string;
  visit_type: string; // first, revisit
  service_type: string;
  is_designated: boolean;
  status: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

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
