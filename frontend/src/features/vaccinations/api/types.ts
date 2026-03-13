/**
 * Backend API response types
 * Generated from backend/docs/api.yaml via openapi-typescript
 * DO NOT EDIT manually — run `make codegen` to regenerate
 */
import type { components } from "@/types/generated/api";

export type BackendVaccination = components["schemas"]["Vaccination"];

export interface CreateVaccinationRequest {
  medical_record_id: string;
  pet_id?: string | null;
  vaccine_id: string;
  date: string;
  doctor_id?: string | null;
  next_date?: string | null;
  lot1?: string;
  remarks?: string;
}

export interface UpdateVaccinationRequest {
  date?: string;
  next_date?: string | null;
  lot1?: string;
  remarks?: string;
}
