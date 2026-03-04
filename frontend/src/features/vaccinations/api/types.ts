export interface BackendVaccination {
  id: string;
  pet_id: string;
  owner_id: string;
  doctor_id?: string | null;
  vaccine_master_id?: string | null;
  vaccine_name: string;
  vaccination_date: string;
  next_date?: string | null;
  lot_number?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
  // Preloaded relations
  pet?: {
    id: string;
    name: string;
    species?: string;
  } | null;
  owner?: {
    id: string;
    name: string;
  } | null;
}

export interface CreateVaccinationRequest {
  pet_id: string;
  owner_id: string;
  doctor_id?: string | null;
  vaccine_master_id?: string | null;
  vaccine_name: string;
  vaccination_date: string;
  next_date?: string | null;
  lot_number?: string;
  notes?: string;
}

export interface UpdateVaccinationRequest {
  vaccine_name?: string;
  vaccination_date?: string;
  next_date?: string | null;
  lot_number?: string;
  notes?: string;
}
