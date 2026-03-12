export interface BackendVaccination {
  id: string;
  medical_record_id: string;
  pet_id?: string | null;
  vaccine_id: string;
  date: string;
  doctor_id?: string | null;
  next_date?: string | null;
  next_schedule_type?: "3weeks" | "4weeks" | "1year" | "other" | null;
  supplemental: string;
  lot1: string;
  lot2: string;
  lot3: string;
  lot4: string;
  remarks: string;
  created_at: string;
  updated_at: string;
  // Relations
  vaccine?: {
    id: string;
    name: string;
    code?: string;
  } | null;
  pet?: {
    id: string;
    name: string;
    species?: string;
    owner_id?: string;
  } | null;
  doctor?: {
    id: string;
    name: string;
  } | null;
}

export interface CreateVaccinationRequest {
  medical_record_id: string;
  pet_id?: string | null;
  vaccine_id: string;
  date: string;
  doctor_id?: string | null;
  next_date?: string | null;
  lot1?: string;
  remarks?: string;
}

export interface UpdateVaccinationRequest {
  date?: string;
  next_date?: string | null;
  lot1?: string;
  remarks?: string;
}
