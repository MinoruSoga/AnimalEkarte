import type { CreateOccupationRequest, UpdateOccupationRequest } from "../api/occupations";
import type { OccupationFormData } from "../lib/occupation-side-panel-model";

export function buildOccupationCreateRequest(data: OccupationFormData): CreateOccupationRequest {
  return {
    name: data.name,
    description: data.description || undefined,
    is_active: data.isActive,
  };
}

export function buildOccupationUpdateRequest(data: OccupationFormData): UpdateOccupationRequest {
  return buildOccupationCreateRequest(data);
}
