/**
 * Backend API response types
 * Generated from backend/docs/api.yaml via openapi-typescript
 * DO NOT EDIT BackendClinic/BackendStaff manually — run `make codegen` to regenerate
 */
import type { components } from "@/types/generated/api";

export type BackendClinic = components["schemas"]["Clinic"];
export type BackendStaff = components["schemas"]["Staff"];

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
}

export type CreateStaffRequest = components["schemas"]["StaffRegistrationRequest"];

export interface UpdateStaffRequest {
  code?: string;
  name?: string;
  staff_role?: "veterinarian" | "nurse" | "trimmer" | "reception" | "manager";
  license_number?: string;
  job_title_id?: string | null;
  sort_order?: number;
  is_active?: boolean;
}
