import type { ReservationSetting, ServiceType, Staff, ReservationAppointment, ReservationCustomer } from "@/types/generated/models";

// ── Re-export backend types ──
export type { ReservationSetting, ServiceType, Staff, ReservationAppointment, ReservationCustomer };

// ── API Request types ──
export type UpdateReservationSettingRequest = Omit<ReservationSetting, "id" | "clinic_id" | "created_at" | "updated_at">;

export type CreateServiceTypeRequest = {
  name: string;
  short_name: string;
  duration_minutes: number;
  reservation_day_option: string;
  is_internal: boolean;
  reservation_visible: boolean;
  reservation_comment: string;
  description: string;
  color: string;
};

export type UpdateServiceTypeRequest = CreateServiceTypeRequest & {
  is_active: boolean;
  sort_order: number;
};

export type CreateStaffRequest = {
  name: string;
  staff_type: string;
  reservation_visible: boolean;
  reservation_comment: string;
};

export type UpdateStaffRequest = CreateStaffRequest & {
  is_active: boolean;
  sort_order: number;
};

export type StaffScheduleShiftType = "full" | "morning" | "afternoon" | "off" | "paid_leave";

export interface StaffSchedule {
  id: number;
  staff_id: number;
  date: string;
  shift_type: StaffScheduleShiftType;
  work_start?: string;
  work_end?: string;
  break_start?: string;
  break_end?: string;
  created_at: string;
  updated_at: string;
}

export type UpsertScheduleRequest = {
  shift_type: StaffScheduleShiftType;
  work_start?: string;
  work_end?: string;
  break_start?: string;
  break_end?: string;
};

export type CreateReservationRequest = {
  start_time: string;
  end_time: string;
  service_type_id: number;
  doctor_id?: number;
  notes: string;
};

export interface StatusPatchRequest {
  is_active: boolean;
}

export interface SortOrderPatchRequest {
  sort_order: number;
}

export interface LinkOwnerRequest {
  owner_id: number;
}
