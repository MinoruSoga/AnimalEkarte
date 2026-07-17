import type { DiagnosisName, DiagnosisType } from "../api/diagnosis";

export interface DiagnosisTypeFormData {
  name: string;
  description: string;
  isActive: boolean;
}

export interface DiagnosisNameFormData {
  name: string;
  diagnosisTypeId: string;
  description: string;
  isActive: boolean;
}

export function diagnosisTypeToFormData(item: DiagnosisType | null): DiagnosisTypeFormData {
  return {
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  };
}

export function diagnosisNameToFormData(
  item: DiagnosisName | null,
  categories: DiagnosisType[],
): DiagnosisNameFormData {
  return {
    name: item?.name ?? "",
    diagnosisTypeId: item
      ? String(item.diagnosisTypeId)
      : categories[0]
        ? String(categories[0].id)
        : "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  };
}
