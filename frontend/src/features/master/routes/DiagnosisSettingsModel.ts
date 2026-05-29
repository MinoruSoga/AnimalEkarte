import type {
  DiagnosisNameFormData,
  DiagnosisTypeFormData,
} from "../components/DiagnosisSidePanelModel";
import type {
  CreateDiagnosisNameRequest,
  CreateDiagnosisTypeRequest,
  UpdateDiagnosisNameRequest,
  UpdateDiagnosisTypeRequest,
} from "@/types/diagnosis";

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
