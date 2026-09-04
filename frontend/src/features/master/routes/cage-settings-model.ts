import type { CreateCageRequest, UpdateCageRequest } from "../api/cages";
import type { CageFormData } from "../lib/cage-side-panel-model";

export function buildCageCreateRequest(data: CageFormData): CreateCageRequest {
  return {
    name: data.name,
    cage_type: data.cageType,
    cage_size: data.cageSize,
    price: data.price,
    description: data.description || undefined,
    is_active: data.isActive,
  };
}

export function buildCageUpdateRequest(data: CageFormData): UpdateCageRequest {
  return buildCageCreateRequest(data);
}
