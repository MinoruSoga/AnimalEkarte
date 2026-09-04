import type {
  Medicine as BackendMedicine,
  MedicineCalculationType,
  TaxType,
} from "@/types/generated/models";
import { MedicineCalculationTypeNone } from "@/types/generated/models";

/**
 * バックエンド Medicine レスポンスをフロントエンド Medicine 型に変換
 * ReturnType<typeof transformBackendMedicineToFrontend> が Medicine 型の正式定義
 */
export const transformBackendMedicineToFrontend = (data: BackendMedicine) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  dosageForm: data.dosage_form,
  medicineUnit: data.medicine_unit,
  price: data.price ?? 0,
  defaultQuantity: data.default_quantity,
  inventoryId: data.inventory_id !== undefined ? String(data.inventory_id ?? 0) : undefined,
  description: data.description ?? "",
  isActive: data.is_active,
  sortOrder: data.sort_order ?? 0,
  taxType: (data.tax_type ?? "excluded") as TaxType,
  taxRate: data.tax_rate ?? 0.1,
  isNonInsurance: data.is_non_insurance ?? false,
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
  calculationType: (data.calculation_type ??
    MedicineCalculationTypeNone) as MedicineCalculationType,
  strength: data.strength,
  frequencyPerDay: data.frequency_per_day,
  defaultDurationDays: data.default_duration_days,
  doseParams: data.dose_params ?? [],
});

/**
 * Medicine フロントエンド型 — transformBackendMedicineToFrontend の戻り値から自動導出
 * 手動管理せず BackendMedicine（models.ts）と常に同期される
 */
export type Medicine = ReturnType<typeof transformBackendMedicineToFrontend>;
