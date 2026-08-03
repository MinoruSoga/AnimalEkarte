import type {
  CreateChiefComplaintTypeRequest,
  UpdateChiefComplaintTypeRequest,
} from "../api/chief-complaint-types";
import type { ChiefComplaintFormData } from "../components/chief-complaint-side-panel-model";

export function buildChiefComplaintCreateRequest(
  data: ChiefComplaintFormData,
): CreateChiefComplaintTypeRequest {
  return {
    name: data.name,
    description: data.description || undefined,
    is_active: data.isActive,
  };
}

export function buildChiefComplaintUpdateRequest(
  data: ChiefComplaintFormData,
): UpdateChiefComplaintTypeRequest {
  return {
    name: data.name,
    description: data.description || undefined,
    is_active: data.isActive,
  };
}
