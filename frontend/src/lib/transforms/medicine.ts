import type {
  Medicine as BackendMedicine,
  MedicineCalculationType,
  TaxType,
} from "@/types/generated/models";
import { MedicineCalculationTypeNone } from "@/types/generated/models";
import { exactDecimalToNumber } from "@/lib/exact-decimal";

type MedicineDecimalKey = "default_quantity" | "tax_rate" | "strength";

/**
 * EMR-447: 薬剤 API の応答 DTO。保存 model 由来の生成型と違い、default_quantity / tax_rate / strength は
 * exact decimal 文字列で返る（default_quantity の NULL は null、strength の NULL は省略）。
 */
export type MedicineHttpResponse = Omit<BackendMedicine, MedicineDecimalKey> & {
  default_quantity: string | null;
  tax_rate: string;
  strength?: string;
};

/**
 * バックエンド Medicine レスポンスをフロントエンド Medicine 型に変換
 * ReturnType<typeof transformBackendMedicineToFrontend> が Medicine 型の正式定義
 */
export const transformBackendMedicineToFrontend = (data: MedicineHttpResponse) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  dosageForm: data.dosage_form,
  medicineUnit: data.medicine_unit,
  price: data.price ?? 0,
  // EMR-447: wire の decimal 文字列は画面・計算用の number へここで1回だけ明示変換する。
  // 欠落時の扱い（default_quantity/strength=undefined、tax_rate=0.1）は従来どおり、値があれば形式を検査する。
  defaultQuantity:
    data.default_quantity == null
      ? undefined
      : exactDecimalToNumber(data.default_quantity, "default_quantity"),
  inventoryId: data.inventory_id !== undefined ? String(data.inventory_id ?? 0) : undefined,
  description: data.description ?? "",
  isActive: data.is_active,
  sortOrder: data.sort_order ?? 0,
  taxType: (data.tax_type ?? "excluded") as TaxType,
  taxRate: data.tax_rate == null ? 0.1 : exactDecimalToNumber(data.tax_rate, "tax_rate"),
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
 * 数値3 field 以外は BackendMedicine（models.ts）と同期され、3 field は MedicineHttpResponse を経由する
 */
export type Medicine = ReturnType<typeof transformBackendMedicineToFrontend>;
