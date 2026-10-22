/**
 * Backend API response types
 * Source: @/types/vaccination-legacy-wire（現行 handler の JSON を写す移行 HTTP DTO。EMR-527）
 */
import type { LegacyVaccinationResponse } from "@/types/vaccination-legacy-wire";

export type BackendVaccination = LegacyVaccinationResponse;

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
  /**
   * UAT-R2-EXCLUSIVE-LOCK: 楽観的ロック expectedVersion。
   * 呼出側は読取済み VaccinationRecord.version を必ず同送する（省略時は BE が照合スキップ＝後方互換）。
   */
  version?: number;
}
