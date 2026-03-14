export type ShiftType =
  | "full"
  | "morning"
  | "afternoon"
  | "off"
  | "paid_leave";

export interface Shift {
  id: string;
  clinic_id: string;
  staff_id: string;
  staff_name?: string;
  date: string; // YYYY-MM-DD
  shift_type: ShiftType;
  start_time: string;
  end_time: string;
  note: string;
  created_at: string;
  updated_at: string;
}

export interface CreateShiftInput {
  staff_id: string;
  date: string; // YYYY-MM-DD
  shift_type: ShiftType;
  start_time?: string;
  end_time?: string;
  note?: string;
}

export interface UpdateShiftInput {
  shift_type?: ShiftType;
  start_time?: string;
  end_time?: string;
  note?: string;
}

export const SHIFT_TYPE_LABELS: Record<ShiftType, string> = {
  full: "全日",
  morning: "午前",
  afternoon: "午後",
  off: "休日",
  paid_leave: "有休",
};

export const SHIFT_TYPE_COLORS: Record<ShiftType, string> = {
  full: "bg-blue-100 text-blue-800 border-blue-200",
  morning: "bg-green-100 text-green-800 border-green-200",
  afternoon: "bg-teal-100 text-teal-800 border-teal-200",
  off: "bg-gray-100 text-gray-600 border-gray-200",
  paid_leave: "bg-purple-100 text-purple-800 border-purple-200",
};

// シフトカレンダーで使用するスタッフの最小型（feature間importを避けるため）
export interface ShiftStaff {
  id: string;
  name: string;
}
