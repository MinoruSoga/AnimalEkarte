import type { TrimmingRecord as ApiTrimmingRecord } from "@/types/generated/models";

export type BackendTrimming = ApiTrimmingRecord;

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
  status?: string;
  appointment_date?: string;
  course?: string;
  options?: string;
  style_request?: string;
  total_price?: number | null;
  notes?: string;
}
