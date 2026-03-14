import type { ShiftType } from "../types";

// バックエンドレスポンス型（snake_case）
export interface BackendShift {
  id: string;
  clinic_id: string;
  staff_id: string;
  staff_name?: string;
  date: string;
  shift_type: ShiftType;
  start_time: string;
  end_time: string;
  note: string;
  created_at: string;
  updated_at: string;
  staff?: {
    id: string;
    name: string;
  };
}
