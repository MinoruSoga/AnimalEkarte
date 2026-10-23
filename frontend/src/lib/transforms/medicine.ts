import type { MedicineCalculationType, TaxType } from "@/types/generated/models";
import { MedicineCalculationTypeNone } from "@/types/generated/models";
import { exactDecimalToNumber } from "@/lib/exact-decimal";
import type { LegacyMedicineResponse } from "@/types/medicine";

/**
 * バックエンド Medicine レスポンスをフロントエンド Medicine 型に変換
 * ReturnType<typeof transformBackendMedicineToFrontend> が Medicine 型の正式定義
 * 入力は保存 model ではなく HTTP DTO（LegacyMedicineResponse・EMR-527）。
 */
export const transformBackendMedicineToFrontend = (data: LegacyMedicineResponse) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  dosageForm: data.dosage_form,
  medicineUnit: data.medicine_unit,
  price: data.price ?? 0,
  // EMR-447: wire の decimal 文字列は画面・計算用の number へここで1回だけ明示変換する。
  // default_quantity の null と strength の欠落は undefined。tax_rate は BE が常に返す（EMR-527: null 不可）ため、
  // 欠落を 0.1 で補わず形式違反と同じく例外にする。
  defaultQuantity:
    data.default_quantity == null
      ? undefined
      : exactDecimalToNumber(data.default_quantity, "default_quantity"),
  inventoryId: data.inventory_id !== undefined ? String(data.inventory_id ?? 0) : undefined,
  description: data.description ?? "",
  isActive: data.is_active,
  sortOrder: data.sort_order ?? 0,
  taxType: (data.tax_type ?? "excluded") as TaxType,
  taxRate: exactDecimalToNumber(data.tax_rate, "tax_rate"),
  // EMR-536: number へ丸める前の wire 文字列。PATCH で「明示再選択」と「未編集/取消」を区別する比較元。
  taxRateExact: data.tax_rate,
  isNonInsurance: data.is_non_insurance ?? false,
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
  calculationType: (data.calculation_type ??
    MedicineCalculationTypeNone) as MedicineCalculationType,
  strength: data.strength == null ? undefined : exactDecimalToNumber(data.strength, "strength"),
  frequencyPerDay: data.frequency_per_day,
  defaultDurationDays: data.default_duration_days,
  doseParams: data.dose_params ?? [],
});

/**
 * Medicine フロントエンド型 — transformBackendMedicineToFrontend の戻り値から自動導出
 * 応答の列は LegacyMedicineResponse（@/types/medicine）が handler の response struct を写す。
 * taxRateExact は変換元が無い Medicine（画面内で組み立てた値など）が既存 caller にあるため省略可にする
 * （省略時は保存 exact 値を証明できないので PATCH は税率を送る側へ倒す）。
 */
export type Medicine = Omit<
  ReturnType<typeof transformBackendMedicineToFrontend>,
  "taxRateExact"
> & {
  taxRateExact?: string;
};
