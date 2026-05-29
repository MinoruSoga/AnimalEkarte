import type { TreatmentFormData } from "../components/TreatmentItemSidePanel";
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

export function buildConsultationCreateRequest(
  data: TreatmentFormData,
): CreateConsultationRequest {
  return {
    name: data.name,
    price: data.price,
    description: data.description || undefined,
    is_active: data.isActive,
    tax_type: data.taxType,
    tax_rate: data.taxRate,
  };
}

export function buildConsultationUpdateRequest(
  data: TreatmentFormData,
): UpdateConsultationRequest {
  return buildConsultationCreateRequest(data);
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
  };
}

export function buildExaminationUpdateRequest(
  data: TreatmentFormData,
): UpdateExaminationTypeRequest {
  return buildExaminationCreateRequest(data);
}

export function buildProcedureCreateRequest(
  data: TreatmentFormData,
): CreateProcedureRequest {
  return {
    name: data.name,
    price: data.price,
    description: data.description || undefined,
    is_active: data.isActive,
    tax_type: data.taxType,
    tax_rate: data.taxRate,
  };
}

export function buildProcedureUpdateRequest(
  data: TreatmentFormData,
): UpdateProcedureRequest {
  return buildProcedureCreateRequest(data);
}

export function buildVaccineCreateRequest(
  data: TreatmentFormData,
): CreateVaccineRequest {
  return {
    name: data.name,
    price: data.price,
    description: data.description || undefined,
    is_active: data.isActive,
  };
}

export function buildVaccineUpdateRequest(
  data: TreatmentFormData,
): UpdateVaccineRequest {
  return buildVaccineCreateRequest(data);
}

export function buildCheckupCreateRequest(
  data: TreatmentFormData,
): CreateCheckupTypeRequest {
  return {
    name: data.name,
    price: data.price,
    description: data.description || undefined,
    is_active: data.isActive,
  };
}

export function buildCheckupUpdateRequest(
  data: TreatmentFormData,
): UpdateCheckupTypeRequest {
  return buildCheckupCreateRequest(data);
}
