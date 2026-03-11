// Backend Hospitalization types

interface BackendPetRelation {
  id: string;
  name: string;
  species: string;
  breed?: string;
  gender?: string;
}

interface BackendOwnerRelation {
  id: string;
  name: string;
}

interface BackendCarePlanItem {
  id: string;
  hospitalization_id: string;
  master_id?: string;
  type: string;
  name: string;
  description: string;
  timing: string;
  status: string;
  unit_price?: number;
  category?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

interface BackendVital {
  id: string;
  daily_record_id: string;
  staff_id?: string;
  recorded_time: string;
  temperature?: number;
  heart_rate?: number;
  respiration_rate?: number;
  weight?: number;
  notes?: string;
  created_at: string;
}

interface BackendCareLog {
  id: string;
  daily_record_id: string;
  staff_id?: string;
  recorded_time: string;
  type: string;
  status: string;
  value?: string;
  notes?: string;
  created_at: string;
}

interface BackendStaffNote {
  id: string;
  daily_record_id: string;
  staff_id?: string;
  recorded_time: string;
  content: string;
  created_at: string;
}

interface BackendDailyRecord {
  id: string;
  hospitalization_id: string;
  record_date: string;
  created_at: string;
  updated_at: string;
  vitals?: BackendVital[];
  care_logs?: BackendCareLog[];
  staff_notes?: BackendStaffNote[];
}

export interface BackendHospitalization {
  id: string;
  hospitalization_no: string;
  pet_id: string;
  owner_id: string;
  cage_id?: string;
  type: string; // 入院, ホテル
  start_date: string;
  end_date: string;
  status: string;
  owner_request?: string;
  staff_notes?: string;
  memo?: string;
  created_at: string;
  updated_at: string;
  // Relations (Preload済み)
  pet?: BackendPetRelation;
  owner?: BackendOwnerRelation;
  care_plan_items?: BackendCarePlanItem[];
  daily_records?: BackendDailyRecord[];
}

export interface CreateHospitalizationRequest {
  pet_id: string;
  owner_id: string;
  cage_id?: string;
  type: string;
  start_date: string;
  end_date: string;
  owner_request?: string;
  staff_notes?: string;
  memo?: string;
}

export interface UpdateHospitalizationRequest {
  type?: string;
  cage_id?: string;
  end_date?: string;
  status?: string;
  owner_request?: string;
  staff_notes?: string;
  memo?: string;
}
