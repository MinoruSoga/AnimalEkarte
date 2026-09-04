import type {
  DiagnosisNameFormData,
  DiagnosisTypeFormData,
} from "../lib/diagnosis-side-panel-model";
import type {
  CreateDiagnosisNameRequest,
  CreateDiagnosisTypeRequest,
  UpdateDiagnosisNameRequest,
  UpdateDiagnosisTypeRequest,
} from "@/types/diagnosis";

export const DIAGNOSIS_TABS = [
  { value: "diagnosis_type", label: "診断病名カテゴリ" },
  { value: "diagnosis_name", label: "診断病名" },
] as const;

export type DiagnosisTabValue = (typeof DIAGNOSIS_TABS)[number]["value"];

const DIAGNOSIS_TAB_VALUES = new Set<string>(DIAGNOSIS_TABS.map((tab) => tab.value));

export function toDiagnosisTabValue(value: string | null): DiagnosisTabValue {
  return DIAGNOSIS_TAB_VALUES.has(value ?? "") ? (value as DiagnosisTabValue) : "diagnosis_type";
}

export function buildDiagnosisTypeCreateRequest(
  data: DiagnosisTypeFormData,
): CreateDiagnosisTypeRequest {
  return {
    name: data.name,
    description: data.description || undefined,
    is_active: true,
  };
}

export function buildDiagnosisTypeUpdateRequest(
  data: DiagnosisTypeFormData,
): UpdateDiagnosisTypeRequest {
  return {
    ...buildDiagnosisTypeCreateRequest(data),
    is_active: data.isActive,
  };
}

export function buildDiagnosisNameCreateRequest(
  data: DiagnosisNameFormData,
): CreateDiagnosisNameRequest {
  return {
    name: data.name,
    diagnosis_type_id: Number(data.diagnosisTypeId),
    description: data.description || undefined,
    is_active: true,
  };
}

export function buildDiagnosisNameUpdateRequest(
  data: DiagnosisNameFormData,
): UpdateDiagnosisNameRequest {
  return {
    ...buildDiagnosisNameCreateRequest(data),
    is_active: data.isActive,
  };
}
