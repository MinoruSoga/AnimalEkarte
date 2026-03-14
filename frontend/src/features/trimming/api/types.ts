import type { TrimmingRecord as ApiTrimmingRecord } from "@/types/generated/models";

export type BackendTrimming = ApiTrimmingRecord;

export interface CreateTrimmingRequest {
  pet_id: number;
  staff_id: number;
  course_id: number;
  date?: string;
  style_request?: string;
  remarks?: string;
  option_ids?: number[];
}

export interface UpdateTrimmingRequest {
  status?: string;
  date?: string;
  style_request?: string;
  bw?: string;
  bw_unit?: string;
  bt?: string;
  used_shampoo?: string;
  used_ribbon?: string;
  treatment?: string;
  charge?: string;
  remarks?: string;
  option_ids?: number[];
}
