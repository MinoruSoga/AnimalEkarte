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
  hospitalizationId: string;
  masterId?: string;
  type: string;
  name: string;
  description: string;
  timing: string;
  status: string;
  unitPrice?: number;
  category?: string;
  notes?: string;
  createdAt: string;
  updatedAt: string;
}

interface BackendVital {
  id: string;
  dailyRecordId: string;
  staffId?: string;
  recordedTime: string;
  temperature?: number;
  heartRate?: number;
  respirationRate?: number;
  weight?: number;
  notes?: string;
  createdAt: string;
}

interface BackendCareLog {
  id: string;
  dailyRecordId: string;
  staffId?: string;
  recordedTime: string;
  type: string;
  status: string;
  value?: string;
  notes?: string;
  createdAt: string;
}

interface BackendStaffNote {
  id: string;
  dailyRecordId: string;
  staffId?: string;
  recordedTime: string;
  content: string;
  createdAt: string;
}

interface BackendDailyRecord {
  id: string;
  hospitalizationId: string;
  recordDate: string;
  createdAt: string;
  updatedAt: string;
  vitals?: BackendVital[];
  careLogs?: BackendCareLog[];
  staffNotes?: BackendStaffNote[];
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
