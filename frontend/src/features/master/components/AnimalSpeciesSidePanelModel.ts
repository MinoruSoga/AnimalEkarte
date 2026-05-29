import type { AnimalSpecies } from "../api/animal-species";

export interface AnimalSpeciesFormData {
  name: string;
  isActive: boolean;
}

export function animalSpeciesToFormData(item: AnimalSpecies | null): AnimalSpeciesFormData {
  return {
    name: item?.name ?? "",
    isActive: item?.isActive ?? true,
  };
}
