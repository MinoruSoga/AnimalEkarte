import type {
  CreateChiefComplaintTypeRequest,
  UpdateChiefComplaintTypeRequest,
} from "../api/chief-complaint-types";
import type { ChiefComplaintFormData } from "../components/ChiefComplaintSidePanelModel";

export function buildChiefComplaintCreateRequest(
  data: ChiefComplaintFormData,
): CreateChiefComplaintTypeRequest {
  return {
    name: data.name,
    description: data.description || undefined,
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
