import { requireMasterEditVersion } from "@/lib/master-edit-version";
import type { TreatmentFormData } from "../components/TreatmentItemSidePanel";
import type { TreatmentTabConfig } from "../lib/treatment-plan-tab-content-model";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import type {
  CreateCheckupTypeRequest,
  CreateConsultationRequestBase,
  CreateExaminationTypeRequest,
  CreateProcedureRequestBase,
  CreateVaccineRequest,
  ItemInsuranceSetting,
  UpdateCheckupTypeRequest,
  UpdateConsultationRequest,
  UpdateExaminationTypeRequest,
  UpdateProcedureRequest,
  UpdateVaccineRequest,
} from "@/types/treatment";

export const TREATMENT_PLAN_TABS = [
  { value: "consultation", label: "診察" },
  { value: "examination", label: "検査" },
  { value: "procedure", label: "処置" },
  { value: "vaccine", label: "予防接種" },
  { value: "checkup", label: "定期健診" },
] as const;

export type TreatmentPlanTabValue = (typeof TREATMENT_PLAN_TABS)[number]["value"];

const TREATMENT_PLAN_TAB_VALUES = new Set<string>(TREATMENT_PLAN_TABS.map((tab) => tab.value));

export function toTreatmentPlanTabValue(value: string | null): TreatmentPlanTabValue {
  return TREATMENT_PLAN_TAB_VALUES.has(value ?? "")
    ? (value as TreatmentPlanTabValue)
    : "consultation";
}

interface BuildTreatmentTabConfigsOptions {
  consultationData: TreatmentItem[] | undefined;
  examinationData: TreatmentItem[] | undefined;
  procedureData: TreatmentItem[] | undefined;
  vaccineData: TreatmentItem[] | undefined;
  checkupData: TreatmentItem[] | undefined;
  onReorderConsultations: (ids: number[]) => void;
  onReorderExaminations: (ids: number[]) => void;
  onReorderProcedures: (ids: number[]) => void;
  onReorderVaccines: (ids: number[]) => void;
  onReorderCheckups: (ids: number[]) => void;
}

export function buildTreatmentTabConfigs({
  consultationData,
  examinationData,
  procedureData,
  vaccineData,
  checkupData,
  onReorderConsultations,
  onReorderExaminations,
  onReorderProcedures,
  onReorderVaccines,
  onReorderCheckups,
}: BuildTreatmentTabConfigsOptions): Record<TreatmentPlanTabValue, TreatmentTabConfig> {
  return {
    consultation: {
      insuranceTriState: true,
      data: consultationData,
      entityLabel: "診察",
      emptyMessage: "診察が登録されていません",
      searchPlaceholder: "診察名で検索...",
      onReorder: onReorderConsultations,
    },
    examination: {
      data: examinationData,
      entityLabel: "検査",
      emptyMessage: "検査が登録されていません",
      searchPlaceholder: "検査名で検索...",
      onReorder: onReorderExaminations,
    },
    procedure: {
      insuranceTriState: true,
      data: procedureData,
      entityLabel: "処置",
      emptyMessage: "処置が登録されていません",
      searchPlaceholder: "処置名で検索...",
      onReorder: onReorderProcedures,
    },
    vaccine: {
      insuranceTriState: true,
      data: vaccineData,
      entityLabel: "予防接種",
      emptyMessage: "予防接種が登録されていません",
      searchPlaceholder: "予防接種名で検索...",
      onReorder: onReorderVaccines,
    },
    checkup: {
      insuranceTriState: true,
      data: checkupData,
      entityLabel: "定期健診",
      emptyMessage: "定期健診が登録されていません",
      searchPlaceholder: "定期健診名で検索...",
      onReorder: onReorderCheckups,
    },
  };
}

export function buildConsultationCreateRequest(
  data: TreatmentFormData,
): CreateConsultationRequestBase {
  return {
    name: data.name,
    price: data.price,
    description: data.description || undefined,
    is_active: data.isActive,
    tax_type: data.taxType,
    tax_rate: data.taxRate,
    ...(data.parentId ? { parent_id: Number(data.parentId) } : {}),
  };
}

export function buildConsultationUpdateRequest(data: TreatmentFormData): UpdateConsultationRequest {
  const base = {
    ...buildConsultationCreateRequest(data),
    version: requireMasterEditVersion(data.version),
  };
  if (data.parentId === "") return { ...base, clear_parent_id: true };
  return base;
}

export function buildExaminationCreateRequest(
  data: TreatmentFormData,
): CreateExaminationTypeRequest {
  return {
    name: data.name,
    price: data.price,
    description: data.description || undefined,
    is_active: data.isActive,
    // 検査は NOT NULL 列（旧欠落の概念がない）ため、既存どおり false 補完を維持する。
    is_non_insurance: data.isNonInsurance ?? false,
    ...(data.parentId ? { parent_id: Number(data.parentId) } : {}),
  };
}

export function buildExaminationUpdateRequest(
  data: TreatmentFormData,
): UpdateExaminationTypeRequest {
  const base = buildExaminationCreateRequest(data);
  if (data.parentId === "") return { ...base, clear_parent_id: true };
  return base;
}

export function buildProcedureCreateRequest(data: TreatmentFormData): CreateProcedureRequestBase {
  return {
    name: data.name,
    price: data.price,
    description: data.description || undefined,
    is_active: data.isActive,
    tax_type: data.taxType,
    tax_rate: data.taxRate,
    anesthesia: data.anesthesia,
    ...(data.parentId ? { parent_id: Number(data.parentId) } : {}),
  };
}

export function buildProcedureUpdateRequest(data: TreatmentFormData): UpdateProcedureRequest {
  const base = {
    ...buildProcedureCreateRequest(data),
    version: requireMasterEditVersion(data.version),
  };
  if (data.parentId === "") return { ...base, clear_parent_id: true };
  return base;
}

export function buildVaccineCreateRequest(data: TreatmentFormData): CreateVaccineRequest {
  return {
    name: data.name,
    price: data.price,
    description: data.description || undefined,
    is_active: data.isActive,
    ...(data.parentId ? { parent_id: Number(data.parentId) } : {}),
  };
}

export function buildVaccineUpdateRequest(
  data: TreatmentFormData,
): UpdateVaccineRequest & { version: number } {
  const base = {
    ...buildVaccineCreateRequest(data),
    version: requireMasterEditVersion(data.version),
  };
  if (data.parentId === "") return { ...base, clear_parent_id: true };
  return base;
}

export function buildCheckupCreateRequest(data: TreatmentFormData): CreateCheckupTypeRequest {
  return {
    name: data.name,
    price: data.price,
    description: data.description || undefined,
    is_active: data.isActive,
    ...(data.parentId ? { parent_id: Number(data.parentId) } : {}),
  };
}

export function buildCheckupUpdateRequest(
  data: TreatmentFormData,
): UpdateCheckupTypeRequest & { version: number } {
  const base = {
    ...buildCheckupCreateRequest(data),
    version: requireMasterEditVersion(data.version),
  };
  if (data.parentId === "") return { ...base, clear_parent_id: true };
  return base;
}

// ─────────────────────────────────────────────────
// 品目別保険設定（EMR-454）
// ─────────────────────────────────────────────────

/**
 * create: 新規登録の初期値は「保険対象外（true）」を API へ明示する
 * （AC-accounting-item-insurance-default-unresolved）。
 * null は旧欠落 = 保存値ではないため、そのまま送らず初期値を明示する。
 */
export function applyItemInsuranceForCreate<T extends object>(
  request: T,
  setting: ItemInsuranceSetting | undefined,
): T & { is_non_insurance: boolean } {
  return { ...request, is_non_insurance: setting ?? true };
}

/**
 * update: 未指定（undefined）と旧欠落（null）は key を送らない＝更新しない。
 * false / true だけを明示送信し、未選択の既存値を反転させない
 * （AC-accounting-item-insurance-legacy-missing）。
 * 契約（backend/docs/api.yaml）では update の null は 400、省略は「変更なし」なので、
 * 旧欠落は必ず省略して送る。
 */
export function applyItemInsuranceForUpdate<T extends object>(
  request: T,
  setting: ItemInsuranceSetting | undefined,
): T & { is_non_insurance?: boolean } {
  if (typeof setting !== "boolean") return request;
  return { ...request, is_non_insurance: setting };
}
