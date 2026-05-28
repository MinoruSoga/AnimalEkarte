import type {
  CreateOccupationRequest,
  UpdateOccupationRequest,
} from "../api/occupations";
import type { OccupationFormData } from "../components/OccupationSidePanelModel";

export function buildOccupationCreateRequest(
  data: OccupationFormData,
): CreateOccupationRequest {
  return {
    name: data.name,
    description: data.description || undefined,
    is_active: data.isActive,
  };
}

export function buildOccupationUpdateRequest(
  data: OccupationFormData,
): UpdateOccupationRequest {
  return buildOccupationCreateRequest(data);
}
