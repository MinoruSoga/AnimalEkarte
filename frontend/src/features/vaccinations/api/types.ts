/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { Vaccination } from "@/types/generated/models";

export type BackendVaccination = Vaccination;

export interface CreateVaccinationRequest {
  medical_record_id?: number | null;
  pet_id?: number | null;
  vaccine_id: number;
  date: string;
  doctor_id?: number | null;
  next_date?: string | null;
  lot1?: string;
  lot2?: string;
  lot3?: string;
  lot4?: string;
  remarks?: string;
  supplemental?: string;
  next_schedule_type?: string;
}

export interface UpdateVaccinationRequest {
  date?: string;
  next_date?: string | null;
  lot1?: string;
  lot2?: string;
  lot3?: string;
  lot4?: string;
  remarks?: string;
  supplemental?: string;
  next_schedule_type?: string;
}
