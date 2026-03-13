/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { Vaccination } from "@/types/generated/models";

export type BackendVaccination = Vaccination;

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
