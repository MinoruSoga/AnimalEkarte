/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { Clinic, Staff } from "@/types/generated/models";

export type BackendClinic = Clinic;
export type BackendStaff = Staff;

export interface UpdateClinicRequest {
  name?: string;
  postal_code?: string;
  address?: string;
  phone_number?: string;
  fax_number?: string;
  registration_number?: string;
  director_name?: string;
  email?: string;
  website?: string;
  logo_url?: string;
  standard_tax_rate?: number;
  reduced_tax_rate?: number;
}

export interface CreateStaffRequest {
  name: string;
  email: string;
  password: string;
  staff_code?: string;
  license_number?: string;
  occupation_id?: string | null;
  sort_order: number;
}

export interface UpdateStaffRequest {
  name?: string;
  license_number?: string;
  occupation_id?: string | null;
  sort_order?: number;
  is_active?: boolean;
}
