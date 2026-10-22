import type {
  AnesthesiaType,
  ExaminationType,
  Vaccine,
  CheckupType,
  TaxType,
} from "@/types/generated/models";
import type {
  ItemInsuranceSetting,
  LegacyConsultationResponse,
  LegacyExactNumericJson,
  LegacyProcedureResponse,
} from "@/types/treatment";

// ─────────────────────────────────────────────────
// 品目別保険設定（EMR-454）
// ─────────────────────────────────────────────────

/**
 * 明示値だけを保持する。null / 未指定は null のまま返し、false へ補完しない
 * （AC-accounting-item-insurance-legacy-missing / AC-accounting-item-insurance-read-failure）。
 */
export const toItemInsuranceSetting = (value: boolean | null | undefined): ItemInsuranceSetting =>
  typeof value === "boolean" ? value : null;

/**
 * 診察・処置の応答は保存 model ではなく HTTP DTO（@/types/treatment・EMR-527）で受ける。
 * READ の sort_order / anesthesia / is_non_insurance は nullable。
 * 下の変換後 sortOrder / anesthesia は表示用の既定値であり、wire の NULL を変更しない。
 */
export type ConsultationWire = LegacyConsultationResponse;
export type ProcedureWire = LegacyProcedureResponse;
export type VaccineWire = Omit<Vaccine, "is_non_insurance"> & {
  is_non_insurance?: ItemInsuranceSetting;
  version?: number | null;
};

/**
 * tax_rate は model.ExactNumeric の JSON 表現（null / "NaN" / number）で届く。
 * null・欠落は従来どおり 0.1、"NaN" は文字列のまま通さず数値 NaN にする（0.1 へ黙って補完しない）。
 */
const toTaxRate = (value: LegacyExactNumericJson | undefined): number => Number(value ?? 0.1);

// ─────────────────────────────────────────────────
// Consultation
// ─────────────────────────────────────────────────

export const transformConsultation = (data: ConsultationWire) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  price: data.price ?? 0,
  isActive: data.is_active,
  description: data.description ?? "",
  sortOrder: data.sort_order ?? 0,
  timeCondition: data.time_condition ?? "",
  duration: data.duration,
  taxType: (data.tax_type ?? "excluded") as TaxType,
  taxRate: toTaxRate(data.tax_rate),
  isNonInsurance: toItemInsuranceSetting(data.is_non_insurance),
  version: typeof data.version === "number" ? data.version : null,
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
});

export type ConsultationItem = ReturnType<typeof transformConsultation>;

// ─────────────────────────────────────────────────
// ExaminationType
// ─────────────────────────────────────────────────

export const transformExaminationType = (data: ExaminationType) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  price: data.price ?? 0,
  isActive: data.is_active,
  description: data.description ?? "",
  sortOrder: data.sort_order ?? 0,
  isNonInsurance: data.is_non_insurance ?? false,
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
});

// ─────────────────────────────────────────────────
// Procedure
// ─────────────────────────────────────────────────

export const transformProcedure = (data: ProcedureWire) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  price: data.price ?? 0,
  isActive: data.is_active,
  description: data.description ?? "",
  sortOrder: data.sort_order ?? 0,
  anesthesia: data.anesthesia ?? "none",
  duration: data.duration,
  taxType: (data.tax_type ?? "excluded") as TaxType,
  taxRate: toTaxRate(data.tax_rate),
  isNonInsurance: toItemInsuranceSetting(data.is_non_insurance),
  version: typeof data.version === "number" ? data.version : null,
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
});

export type ProcedureItem = ReturnType<typeof transformProcedure>;

// ─────────────────────────────────────────────────
// Vaccine
// ─────────────────────────────────────────────────

export const transformVaccine = (data: VaccineWire) => ({
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  price: data.price ?? 0,
  isActive: data.is_active,
  description: data.description ?? "",
  sortOrder: data.sort_order ?? 0,
  species: data.species,
  interval: data.interval ?? "",
  // Compatibility alias matches the dedicated vaccine adapter; canonical null is retained below.
  isNonInsurance: typeof data.is_non_insurance === "boolean" ? data.is_non_insurance : undefined,
  insuranceSetting: toItemInsuranceSetting(data.is_non_insurance),
  version: typeof data.version === "number" ? data.version : null,
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
});

export type VaccineItem = ReturnType<typeof transformVaccine>;

// ─────────────────────────────────────────────────
// CheckupType
// ─────────────────────────────────────────────────

export const transformCheckupType = (
  data: CheckupType & { is_non_insurance?: boolean | null; version?: number },
) => ({
  // EMR-225: サイドパネルが「この行は健診区分」と識別するための discriminant
  // （健診フィールド定義エディタの出し分けに使用。他タブの TreatmentItem には付かない）。
  masterKind: "checkup" as const,
  id: String(data.id ?? 0),
  name: data.name,
  parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
  price: data.price ?? 0,
  isActive: data.is_active,
  description: data.description ?? "",
  sortOrder: data.sort_order ?? 0,
  interval: data.interval ?? "",
  targetAge: data.target_age ?? "",
  isNonInsurance: toItemInsuranceSetting(data.is_non_insurance),
  insuranceSetting: toItemInsuranceSetting(data.is_non_insurance),
  version: typeof data.version === "number" ? data.version : null,
  createdAt: data.created_at ?? "",
  updatedAt: data.updated_at ?? "",
});

export type CheckupTypeItem = ReturnType<typeof transformCheckupType>;

// ─────────────────────────────────────────────────
// 共通の最小型（5エンティティに共通するフィールド）
// ─────────────────────────────────────────────────

export type TreatmentItem = {
  id: string;
  name: string;
  parentId?: string;
  price: number;
  isActive: boolean;
  description: string;
  sortOrder: number;
  taxType?: TaxType;
  taxRate?: number;
  /**
   * 品目別保険設定（EMR-454）。診察・処置は三状態（true=対象外 / false=対象 / null=旧欠落）。
   * 検査・薬剤など NOT NULL 列の種別は boolean のまま。
   */
  isNonInsurance?: ItemInsuranceSetting;
  /** Procedure only — present when the item is a procedure master row */
  anesthesia?: AnesthesiaType;
  /** Checkup only — present when the item is a checkup package master row (EMR-225) */
  masterKind?: "checkup";
  version?: number | null;
  insuranceSetting?: ItemInsuranceSetting;
};
