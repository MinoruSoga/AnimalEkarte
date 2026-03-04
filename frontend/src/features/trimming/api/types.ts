export interface BackendTrimming {
  id: string;
  pet_id: string;
  owner_id: string;
  staff_id?: string | null;
  appointment_date: string;
  course: string;
  options?: string;
  style_request?: string;
  status: "予約" | "進行中" | "完了";
  total_price?: number | null;
  notes?: string;
  created_at: string;
  updated_at: string;
  // Preloaded relations
  pet?: {
    id: string;
    name: string;
    species?: string;
    pet_number?: string;
    weight?: number;
  } | null;
  owner?: {
    id: string;
    name: string;
  } | null;
}

export interface CreateTrimmingRequest {
  pet_id: string;
  owner_id: string;
  staff_id?: string | null;
  appointment_date: string;
  course: string;
  options?: string;
  style_request?: string;
  notes?: string;
}

export interface UpdateTrimmingRequest {
  status?: "予約" | "進行中" | "完了";
  appointment_date?: string;
  course?: string;
  options?: string;
  style_request?: string;
  total_price?: number | null;
  notes?: string;
}
