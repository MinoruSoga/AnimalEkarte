/**
 * 予防接種 API の移行 HTTP DTO（非永続）。
 *
 * 保存 model（生成 models.ts）ではなく、現行 handler が返す JSON の列をそのまま写す。
 * Go の json tag に合わせ、omitempty は省略可（?）で表す。
 * pet / doctor の入れ子は同じ Go struct の tygo 生成型（medicalrecord-responses）を使う。
 * vaccinations feature と shared hooks が同じ wire を読むため共有層に置く
 * （feature 間 import は禁止）。予防接種の wire 契約を生成型へ移すときに、
 * この adapter ごと置き換える（EMR-527）。
 */
import type {
  PetSummaryResponse,
  StaffSummaryResponse,
} from "@/types/generated/medicalrecord-responses";
// enum union は allowlist 済みの ./index から type-only で受け取る。
import type { NextScheduleType } from "./index";

/** @see backend/internal/medicalrecord/vaccination_response.go vaccineSummaryResponse */
interface LegacyVaccineSummary {
  id: number;
  name: string;
  /**
   * Go の vaccineSummaryResponse には無い列（常に undefined）。
   * 既存の履歴 transform（price ?? null）の型を保つためだけに残す。要確認（EMR-527）。
   */
  price?: number;
}

/** @see backend/internal/medicalrecord/vaccination_response.go vaccinationResponse */
export interface LegacyVaccinationResponse {
  id: number;
  clinic_id: number;
  medical_record_id?: number;
  pet_id?: number;
  vaccine_id: number;
  date: string;
  /** 解決済みの doctor と一致するときだけ出力される。 */
  doctor_id?: number;
  next_date?: string;
  next_schedule_type?: NextScheduleType;
  supplemental: string;
  lot1: string;
  lot2: string;
  lot3: string;
  lot4: string;
  remarks: string;
  version: number;
  created_at: string;
  updated_at: string;
  /** Preload 時のみ埋まる。 */
  pet?: PetSummaryResponse;
  doctor?: StaffSummaryResponse;
  vaccine?: LegacyVaccineSummary;
}
