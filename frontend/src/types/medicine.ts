import type {
  DosageForm,
  MedicineCalculationType,
  MedicineDoseBasis,
  MedicineDoseSpecies,
  MedicineRoundingMode,
  MedicineUnit,
  TaxType,
} from "@/types/generated/models";
import type { ReorderRequest } from "@/types/form";

// 薬剤マスタ API の移行 HTTP DTO（非永続・EMR-527）。
// 保存 model（生成 Medicine / MedicineDoseParam）から Omit / Pick せず、現行 handler の
// response struct と binding struct の列をそのまま写す。omitempty は省略可（?）で表す。
// Go 側で string(model.X) として直列化される列は、値域どおり enum union で受ける。
// EMR-447: default_quantity / tax_rate / strength は exact decimal 文字列 wire（JSON number は 400）。
// 画面・計算用の number への変換は transformBackendMedicineToFrontend で1回だけ行う。
// 新しい薬剤・投与量の契約へ切り替えるときに、この DTO ごと置き換える。

/**
 * 薬剤マスタ応答。inventory は返らない。
 * default_quantity の NULL は null、strength の NULL は省略で返る。
 * @see backend/internal/medicalrecord/medicine_response.go medicineResponse
 */
export interface LegacyMedicineResponse {
  id: number;
  clinic_id: number;
  name: string;
  parent_id?: number;
  price?: number;
  is_active: boolean;
  description: string;
  dosage_form?: DosageForm;
  medicine_unit?: MedicineUnit;
  inventory_id?: number;
  default_quantity: string | null;
  tax_type: TaxType;
  tax_rate: string;
  sort_order: number;
  is_non_insurance: boolean;
  calculation_type: MedicineCalculationType;
  strength?: string;
  frequency_per_day?: number;
  default_duration_days?: number;
  created_at: string;
  updated_at: string;
  /**
   * medicineResponse には無い列（現行 handler では常に欠落）。変換の互換キー doseParams の
   * 受け口としてだけ残す。投与量パラメータは /dose-params を別途取得する。
   */
  dose_params?: LegacyMedicineDoseParamResponse[];
}

/**
 * 薬剤 × 種の投与量パラメータ応答（#201）。
 * @see backend/internal/medicalrecord/medicine_dose_param_response.go medicineDoseParamResponse
 */
export interface LegacyMedicineDoseParamResponse {
  id: number;
  clinic_id: number;
  medicine_id: number;
  species: MedicineDoseSpecies;
  dose_basis: MedicineDoseBasis;
  dose_per_kg: number;
  min_mg_per_kg?: number;
  max_mg_per_kg?: number;
  absolute_max_dose?: number;
  rounding_step?: number;
  rounding_mode?: MedicineRoundingMode;
  notes: string;
  created_at: string;
  updated_at: string;
}

/**
 * 薬剤作成リクエスト（バックエンドAPI）
 * name のみ必須。送信側（buildMedicineCreateRequest）が送るキーだけを列挙する。
 * @see backend/internal/medicalrecord/medicine_request.go createMedicineRequest
 */
export interface CreateMedicineRequest {
  name: string;
  parent_id?: number;
  price?: number;
  is_active?: boolean;
  description?: string;
  dosage_form?: DosageForm;
  medicine_unit?: MedicineUnit;
  tax_type?: TaxType;
  /** EMR-447: exact decimal 文字列。省略は既定 0.1 */
  tax_rate?: string;
  is_non_insurance?: boolean;
  calculation_type?: MedicineCalculationType;
  /** EMR-447: exact decimal 文字列（入力字句のまま） */
  strength?: string;
  frequency_per_day?: number;
  default_duration_days?: number;
}

/**
 * 薬剤更新リクエスト（バックエンドAPI）
 * PATCH: 全フィールドoptional
 * clear_parent_id: true を送ると parent_id を NULL にクリアする
 * @see backend/internal/medicalrecord/medicine_request.go updateMedicineRequest
 */
export interface UpdateMedicineRequest {
  name?: string;
  parent_id?: number;
  clear_parent_id?: boolean;
  price?: number;
  is_active?: boolean;
  description?: string;
  dosage_form?: DosageForm;
  medicine_unit?: MedicineUnit;
  tax_type?: TaxType;
  /** EMR-447: exact decimal 文字列。省略は保持 */
  tax_rate?: string;
  is_non_insurance?: boolean;
  calculation_type?: MedicineCalculationType;
  /** EMR-447: exact decimal 文字列。省略は保持 */
  strength?: string;
  /** EMR-447 Q4: 既存含量を空へ編集し計算方式 none で保存したときだけ true（strength と同時送信しない） */
  clear_strength?: boolean;
  frequency_per_day?: number;
  default_duration_days?: number;
}

/**
 * 薬剤並び替えリクエスト
 * models.ts に対応フィールドなし → 手書き
 */
export type ReorderMedicinesRequest = ReorderRequest;
