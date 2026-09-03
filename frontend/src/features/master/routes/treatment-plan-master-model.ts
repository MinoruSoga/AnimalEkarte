import type { TreatmentFormData } from "../components/TreatmentItemSidePanel";
import type { TreatmentTabConfig } from "../lib/treatment-plan-tab-content-model";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import type {
  CreateCheckupTypeRequest,
  CreateConsultationRequest,
  CreateExaminationTypeRequest,
  CreateProcedureRequest,
  CreateVaccineRequest,
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
      data: procedureData,
      entityLabel: "処置",
      emptyMessage: "処置が登録されていません",
      searchPlaceholder: "処置名で検索...",
      onReorder: onReorderProcedures,
    },
    vaccine: {
      data: vaccineData,
      entityLabel: "予防接種",
      emptyMessage: "予防接種が登録されていません",
      searchPlaceholder: "予防接種名で検索...",
      onReorder: onReorderVaccines,
    },
    checkup: {
      data: checkupData,
      entityLabel: "定期健診",
      emptyMessage: "定期健診が登録されていません",
      searchPlaceholder: "定期健診名で検索...",
      onReorder: onReorderCheckups,
    },
  };
}

export function buildConsultationCreateRequest(data: TreatmentFormData): CreateConsultationRequest {
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
  const base = buildConsultationCreateRequest(data);
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
    is_non_insurance: data.isNonInsurance,
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

export function buildProcedureCreateRequest(data: TreatmentFormData): CreateProcedureRequest {
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
  const base = buildProcedureCreateRequest(data);
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

export function buildVaccineUpdateRequest(data: TreatmentFormData): UpdateVaccineRequest {
  const base = buildVaccineCreateRequest(data);
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

export function buildCheckupUpdateRequest(data: TreatmentFormData): UpdateCheckupTypeRequest {
  const base = buildCheckupCreateRequest(data);
  if (data.parentId === "") return { ...base, clear_parent_id: true };
  return base;
}
