// API response types matching backend daily_record_response.go

export interface ApiVitalRecord {
    id: string;
    daily_record_id: string;
    time: string;
    temperature?: number;
    heart_rate?: number;
    respiration_rate?: number;
    weight?: number;
    notes: string;
    staff_id?: string;
    created_at: string;
}

export interface ApiCareLogRecord {
    id: string;
    daily_record_id: string;
    time: string;
    type: string;
    status: string;
    value: string;
    staff_id?: string;
    notes: string;
    created_at: string;
}

export interface ApiStaffNoteRecord {
    id: string;
    daily_record_id: string;
    time: string;
    content: string;
    staff_id?: string;
    created_at: string;
}

export interface ApiDailyRecord {
    id: string;
    hospitalization_id: string;
    date: string; // ISO8601 e.g. "2026-03-14T00:00:00Z"
    created_at: string;
    updated_at: string;
    vital_records: ApiVitalRecord[];
    care_logs: ApiCareLogRecord[];
    staff_notes: ApiStaffNoteRecord[];
}

// Request types

export interface CreateVitalRecordRequest {
    time: string;
    temperature?: number;
    heart_rate?: number;
    respiration_rate?: number;
    weight?: number;
    notes?: string;
    staff_id?: number;
}

export interface CreateCareLogRecordRequest {
    time: string;
    type: string;
    status?: string;
    value?: string;
    staff_id?: number;
    notes?: string;
}

export interface CreateStaffNoteRecordRequest {
    time: string;
    content: string;
    staff_id?: number;
}
