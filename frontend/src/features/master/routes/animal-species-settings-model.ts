import type { CreateAnimalSpeciesRequest, UpdateAnimalSpeciesRequest } from "../api/animal-species";
import type { AnimalSpeciesFormData } from "../lib/animal-species-side-panel-model";

export function buildAnimalSpeciesCreateRequest(
  data: AnimalSpeciesFormData,
): CreateAnimalSpeciesRequest {
  return {
    name: data.name,
    is_active: true,
    sort_order: 0,
  };
}

export function buildAnimalSpeciesUpdateRequest(
  data: AnimalSpeciesFormData,
): UpdateAnimalSpeciesRequest {
  return {
    name: data.name,
    is_active: data.isActive,
  };
}
